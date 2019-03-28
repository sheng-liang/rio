package project

import (
	"io"
	"os"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/clientcfg"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/urfave/cli"
)

type Data struct {
	ID      string
	Project *clientcfg.Project
}

type Ls struct {
	A_All bool `desc:"List projects from all clusters"`
}

func (l *Ls) Customize(cmd *cli.Command) {
	cmd.Flags = append(cmd.Flags, table.WriterFlags()...)
}

func (l *Ls) Run(ctx *clicontext.CLIContext) error {
	var (
		clusters []clientcfg.Cluster
		err      error
	)

	var writer *table.Writer
	if l.A_All {
		clusters, err = ctx.Clusters()
		if err != nil {
			return err
		}

		if len(clusters) == 1 {
			// For the purpose of listing, this cluster is default
			clusters[0].Default = true
		}

		writer = table.NewWriter([][]string{
			{"NAME", "Project.Project.Name"},
			{"CLUSTER", "Project.Cluster.URL"},
			{"DEFAULT", "{{and .Project.Default .Project.Cluster.Default | boolToStar}}"},
		}, ctx, os.Stdout)
	} else {
		cluster, err := ctx.Cluster()
		if err != nil {
			return err
		}
		clusters = []clientcfg.Cluster{
			*cluster,
		}
		writer = NewWriter(ctx, os.Stdout)
	}

	writer.AddFormatFunc("boolToStar", BoolToStar)
	defer writer.Close()

	data, err := ListProject(clusters)
	if err != nil {
		return err
	}
	for _, d := range data {
		writer.Write(d)
	}
	return writer.Err()
}

func ListProject(clusters []clientcfg.Cluster) ([]Data, error) {
	var data []Data
	for _, cluster := range clusters {
		projects, err := cluster.Projects()
		if err != nil {
			return nil, err
		}
		for i, project := range projects {
			data = append(data, Data{
				ID:      project.Project.Name,
				Project: &projects[i],
			})
		}
	}
	return data, nil
}

func NewWriter(ctx *clicontext.CLIContext, w io.Writer) *table.Writer {
	writer := table.NewWriter([][]string{
		{"NAME", "Project.Project.Name"},
		{"DEFAULT", "{{.Project.Default | boolToStar}}"},
	}, ctx, w)
	writer.AddFormatFunc("boolToStar", BoolToStar)
	return writer
}

func BoolToStar(obj interface{}) (string, error) {
	if b, ok := obj.(bool); ok && b {
		return "*", nil
	}
	return "", nil
}
