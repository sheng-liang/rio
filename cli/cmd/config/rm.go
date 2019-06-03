package config

import (
	"errors"

	"github.com/rancher/rio/cli/cmd/rm"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/types"
)

type Rm struct {
}

func (r *Rm) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) == 0 {
		return errors.New("at least one argument is needed")
	}
	return rm.Remove(ctx, types.ConfigType)
}
