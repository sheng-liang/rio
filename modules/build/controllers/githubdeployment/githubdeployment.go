package githubdeployment

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/go-github/v28/github"
	webhookv1 "github.com/rancher/gitwatcher/pkg/apis/gitwatcher.cattle.io/v1"
	webhookv1controller "github.com/rancher/gitwatcher/pkg/generated/controllers/gitwatcher.cattle.io/v1"
	"github.com/rancher/rio/modules/service/controllers/serviceset"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/kv"
	pipelinev1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"golang.org/x/oauth2"
)

const (
	defaultSecretName = "githubtoken"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := handler{
		systemNamespace: rContext.Namespace,
		serviceCache:    rContext.Rio.Rio().V1().Service().Cache(),
		secretCache:     rContext.Core.Core().V1().Secret().Cache(),
		gitwatcherCache: rContext.Webhook.Gitwatcher().V1().GitWatcher().Cache(),
		appCache:        rContext.Rio.Rio().V1().App().Cache(),
	}

	rContext.Rio.Rio().V1().Service().OnChange(ctx, "github-deployment-status-update-rio-service", h.updateGithubDeploymentStatus)

	rContext.Build.Tekton().V1alpha1().TaskRun().OnChange(ctx, "github-deployment-status-update-tekton-taskrun", h.updateGithubDeploymentStatusOnbuild)
	return nil
}

type handler struct {
	systemNamespace string
	serviceCache    riov1controller.ServiceCache
	secretCache     corev1controller.SecretCache
	appCache        riov1controller.AppCache
	gitwatcherCache webhookv1controller.GitWatcherCache
}

func (h handler) getClient(ctx context.Context, ns, secretName string) (*github.Client, error) {
	secret, err := h.secretCache.Get(ns, secretName)
	if err != nil {
		return nil, err
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: string(secret.Data["accessToken"])},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	return client, nil
}

func (h handler) updateGithubDeploymentStatus(key string, svc *riov1.Service) (*riov1.Service, error) {
	if svc == nil || svc.DeletionTimestamp != nil {
		return svc, nil
	}

	if svc.Status.GitWatcherName == "" {
		return svc, nil
	}

	gitwatcher, err := h.gitwatcherCache.Get(svc.Namespace, svc.Status.GitWatcherName)
	if err != nil {
		return svc, err
	}

	if gitwatcher.Status.GithubStatus == nil {
		return svc, nil
	}

	secretName := defaultSecretName
	if svc.Spec.Build != nil && svc.Spec.Build.GithubSecretName != "" {
		secretName = svc.Spec.Build.GithubSecretName
	}

	client, err := h.getClient(context.Background(), svc.Namespace, secretName)
	if err != nil {
		return svc, err
	}

	if svc.Status.GithubStatus == nil || svc.Status.GithubStatus.PR == "" {
		if serviceset.IsReady(svc.Status.DeploymentStatus) {
			return svc, h.createDeploymentStatus("success", "production", true, svc, gitwatcher, client)
		}
	} else if svc.Status.GithubStatus != nil || svc.Status.GithubStatus.PR != "" {
		if serviceset.IsReady(svc.Status.DeploymentStatus) {
			return svc, h.createDeploymentStatus("success", "staging", false, svc, gitwatcher, client)
		}
	}

	return svc, nil
}

func (h handler) createDeploymentStatus(state string, env string, autoInactive bool, svc *riov1.Service, gitwatcher *webhookv1.GitWatcher, client *github.Client) error {
	if gitwatcher.Status.GithubStatus == nil {
		return nil
	}

	if svc.Status.GithubStatus == nil {
		return nil
	}

	logserver, err := h.serviceCache.Get(h.systemNamespace, "logserver")
	if err != nil {
		return err
	}

	deployID := gitwatcher.Status.GithubStatus.PullRequestDeployID[svc.Status.GithubStatus.PR]
	if deployID == nil {
		return nil
	}
	owner, repo, err := getOwnerAndRepo(gitwatcher.Spec.RepositoryURL)
	if err != nil {
		return err
	}
	logURL := fmt.Sprintf("%s/logs/%s/%s", logserver.Status.Endpoints[0], svc.Namespace, svc.Name)
	endpoint := ""
	if len(svc.Status.Endpoints) > 0 {
		endpoint = svc.Status.Endpoints[0]
	}
	_, resp, err := client.Repositories.CreateDeploymentStatus(context.Background(), owner, repo, *deployID, &github.DeploymentStatusRequest{
		State:          &state,
		EnvironmentURL: &endpoint,
		Environment:    &env,
		AutoInactive:   &autoInactive,
		LogURL:         &logURL,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		msg, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("failed to create deployment status, code: %v, error: %v", resp.StatusCode, msg)
	}
	return nil
}

func (h handler) updateGithubDeploymentStatusOnbuild(key string, tr *pipelinev1alpha1.TaskRun) (*pipelinev1alpha1.TaskRun, error) {
	if tr == nil || tr.DeletionTimestamp != nil {
		return tr, nil
	}

	svcNs, svcName := tr.Labels["service-namespace"], tr.Labels["service-name"]
	svc, err := h.serviceCache.Get(svcNs, svcName)
	if err != nil {
		return tr, nil
	}

	if svc.Status.GitWatcherName == "" {
		return tr, nil
	}

	gitwatcher, err := h.gitwatcherCache.Get(svc.Namespace, svc.Status.GitWatcherName)
	if err != nil {
		return tr, err
	}

	if gitwatcher.Status.GithubStatus == nil {
		return tr, nil
	}

	secretName := defaultSecretName
	if svc.Spec.Build != nil && svc.Spec.Build.GithubSecretName != "" {
		secretName = svc.Spec.Build.GithubSecretName
	}
	client, err := h.getClient(context.Background(), svc.Namespace, secretName)
	if err != nil {
		return tr, err
	}
	state := ""
	if tr.IsDone() && !tr.IsSuccessful() {
		state = "failure"
	} else if !tr.IsDone() {
		state = "in_progress"
	} else {
		return tr, nil
	}

	if svc.Status.GithubStatus == nil || svc.Status.GithubStatus.PR == "" {
		return tr, h.createDeploymentStatus(state, "production", true, svc, gitwatcher, client)
	} else if svc.Status.GithubStatus != nil || svc.Status.GithubStatus.PR != "" {
		return tr, h.createDeploymentStatus(state, "staging", true, svc, gitwatcher, client)
	}

	return tr, nil
}

func getOwnerAndRepo(repoURL string) (string, string, error) {
	u, err := url.Parse(repoURL)
	if err != nil {
		return "", "", err
	}
	repo := strings.TrimPrefix(u.Path, "/")
	repo = strings.TrimSuffix(repo, ".git")
	owner, repo := kv.Split(repo, "/")
	return owner, repo, nil
}
