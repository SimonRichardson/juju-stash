// Copyright 2019 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.
package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/juju/cmd"
	"github.com/juju/errors"
	"github.com/juju/gnuflag"

	jujucmd "github.com/juju/juju/cmd"
	"github.com/juju/juju/cmd/modelcmd"
	"github.com/juju/juju/jujuclient"
)

func newPushCommand(clientStore jujuclient.ClientStore, history *history) cmd.Command {
	return modelcmd.Wrap(&pushCommand{
		clientStore: clientStore,
		history:     history,
	})
}

type pushCommand struct {
	modelcmd.ModelCommandBase

	clientStore jujuclient.ClientStore
	history     *history

	target string
	status bool
}

// SetFlags implements Command.SetFlags.
func (c *pushCommand) SetFlags(f *gnuflag.FlagSet) {
	c.ModelCommandBase.SetFlags(f)

	f.BoolVar(&c.status, "status", false, "show juju status after a change is made")
}

// Init implements Command.Init.
func (c *pushCommand) Init(args []string) (err error) {
	c.SetClientStore(c.clientStore)
	c.target, err = cmd.ZeroOrOneArgs(args)
	return err
}

// Info implements Command.Info.
func (c *pushCommand) Info() *cmd.Info {
	return jujucmd.Info(&cmd.Info{
		Name:    "push",
		Purpose: "push adds a model to the stash history",
		Doc: `
Push adds a model to the stash history

Examples:
	juju stash push mymodel
	juju stash push mymodel --status

See:
	juju stash pop
	juju stash list
	juju switch
`,
	})
}

// Run implements Command.Run.
func (c *pushCommand) Run(ctx *cmd.Context) (err error) {
	store := modelcmd.QualifyingClientStore{ClientStore: c.clientStore}
	var controllerName string
	controllerName, err = modelcmd.DetermineCurrentController(store)
	if err != nil {
		return errors.Trace(err)
	}
	var modelName string
	modelName, _, err = c.ModelDetails()
	if err != nil {
		return errors.Trace(err)
	}

	var targetName string
	targetName, err = store.QualifiedModelName(controllerName, c.target)
	if err != nil {
		return errors.Trace(err)
	}

	defer func() {
		if !c.status {
			return
		}
		if err == nil {
			fmt.Fprintln(ctx.GetStdout(), fmt.Sprintf("%s\n", strings.Repeat("-", len(modelName)+len(targetName)+4)))
			cmd := exec.Command("juju", "status")
			cmd.Stdout = ctx.GetStdout()
			cmd.Stderr = ctx.GetStderr()
			if err := cmd.Run(); err != nil {
				fmt.Fprintln(ctx.GetStderr(), err)
			}
		}
	}()

	if err = store.SetCurrentModel(controllerName, targetName); err != nil {
		return errors.Trace(err)
	}
	logSwitch(ctx, modelName, targetName)
	if modelName == targetName {
		return nil
	}
	return c.history.Push(historySnapshot{
		controllerName: controllerName,
		modelName:      modelName,
	})
}

func logSwitch(ctx *cmd.Context, oldName string, newName string) {
	if newName == oldName {
		ctx.Infof("%s (no change)", oldName)
	} else {
		ctx.Infof("%s -> %s", oldName, newName)
	}
}
