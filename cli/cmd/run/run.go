package run

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	isatty "github.com/onsi/ginkgo/reporters/stenographer/support/go-isatty"
	"github.com/rancher/norman/pkg/kv"
	"github.com/rancher/rio/cli/cmd/attach"
	"github.com/rancher/rio/cli/cmd/create"
	"github.com/rancher/rio/cli/pkg/clicontext"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
)

type Run struct {
	create.Create
	Scale string `desc:"scale" default:"1"`
}

func (r *Run) Run(ctx *clicontext.CLIContext) error {
	service, err := r.RunCallback(ctx, func(service *riov1.Service) *riov1.Service {
		if strings.ContainsRune(r.Scale, '-') {
			min, max := kv.Split(r.Scale, "-")
			minScale, _ := strconv.Atoi(min)
			maxScale, _ := strconv.Atoi(max)
			service.Spec.AutoScale = &riov1.AutoscaleConfig{}
			service.Spec.AutoScale.MinScale = minScale
			service.Spec.AutoScale.MaxScale = maxScale
			service.Spec.AutoScale.Concurrency = r.Concurrency
			service.Spec.Scale = minScale
			return service
		}

		scale, _ := strconv.Atoi(r.Scale)
		if scale == 0 {
			scale = 1
		}
		service.Spec.Scale = scale
		return service
	})
	if err != nil {
		return err
	}

	istty := isatty.IsTerminal(os.Stdout.Fd()) &&
		isatty.IsTerminal(os.Stderr.Fd()) &&
		isatty.IsTerminal(os.Stdin.Fd())

	if istty && !r.Detach && service.Spec.OpenStdin && service.Spec.Tty {
		fmt.Println("Attaching...")
		return attach.RunAttach(ctx, time.Minute, true, true, service.Name)
	}
	fmt.Printf("%s/%s\n", service.Spec.StackName, service.Name)

	return nil
}
