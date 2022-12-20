package cmd

import (
	"github.com/spf13/cobra"

	"reward/internal/app"
)

type Command struct {
	*cobra.Command
	App *app.App
}

func (c *Command) AddCommands(commands ...*Command) {
	for _, command := range commands {
		c.AddCommand(command.Command)
	}
}

func (c *Command) AddGroups(title string, cmds ...*Command) {
	g := &cobra.Group{
		Title: title,
		ID:    title,
	}
	c.AddGroup(g)
	for _, cmd := range cmds {
		cmd.GroupID = g.ID
		c.AddCommands(cmd)
	}
}
