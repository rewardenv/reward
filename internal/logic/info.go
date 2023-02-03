package logic

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	cmdpkg "github.com/rewardenv/reward/cmd"
)

// RunCmdInfo represents the info command.
func (c *Client) RunCmdInfo(cmd *cmdpkg.Command) error {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Info"})

	c.infoHeader(t)
	c.infoGlobalServices(t)
	c.infoEnvironment(t)
	c.infoRender(cmd, t)

	return nil
}

func (c *Client) infoHeader(t table.Writer) {
	t.AppendRows([]table.Row{
		{
			fmt.Sprintf("%s version", cases.Title(language.English).String(c.Config.AppName())),
			c.Config.AppVersion(),
		},
	})

	if len(c.Plugins()) > 0 {
		var plugins []string
		for _, plugin := range c.Plugins() {
			plugins = append(plugins, plugin.Name)
		}

		t.AppendRow([]interface{}{
			"Installed plugins",
			strings.Join(plugins, ", "),
		})
	}

	t.AppendRows([]table.Row{
		{"Docs", "https://rewardenv.readthedocs.io"},
	})
}

func (c *Client) infoRender(cmd *cmdpkg.Command, t table.Writer) {
	style, _ := cmd.Flags().GetString("style")
	switch style {
	case "csv":
		t.RenderCSV()

		return
	case "html":
		t.RenderHTML()

		return
	case "markdown":
		t.RenderMarkdown()

		return
	case "black":
		t.SetStyle(table.StyleColoredCyanWhiteOnBlack)
	case "double":
		t.SetStyle(table.StyleDouble)
	case "bright":
		t.SetStyle(table.StyleColoredBright)
	case "light":
		t.SetStyle(table.StyleLight)
	case "dark":
		t.SetStyle(table.StyleColoredDark)
	default:
		t.SetStyle(table.StyleDefault)
	}

	t.Render()
}

func (c *Client) infoEnvironment(t table.Writer) {
	if c.Config.EnvInitialized() {
		t.AppendSeparator()
		t.AppendRow([]interface{}{"Environment"})
		t.AppendSeparator()

		status := "Stopped"

		state := c.Config.Docker.ContainerRunning(c.Config.DefaultSyncedContainer(c.Config.EnvType()))
		if state {
			status = "Running"
		}

		t.AppendRow([]interface{}{"Environment status", status})
		t.AppendRow([]interface{}{
			"Environment address",
			fmt.Sprintf("https://%s", c.Config.TraefikDomain()),
		})

		if c.Config.EnvType() == "magento2" || c.Config.EnvType() == "magento1" {
			t.AppendRow([]interface{}{
				"Admin URL",
				fmt.Sprintf("https://%s/%s", c.TraefikFullDomain(), c.MagentoBackendFrontname()),
			})
		}

		if c.Config.EnvType() == "shopware" {
			t.AppendRow([]interface{}{
				"Admin URL",
				fmt.Sprintf("https://%s/%s", c.TraefikFullDomain(), c.ShopwareAdminPath()),
			})
		}

		if c.Config.EnvType() == "wordpress" {
			t.AppendRow([]interface{}{
				"Admin URL",
				fmt.Sprintf("https://%s/%s", c.TraefikFullDomain(), c.WordpressAdminPath()),
			})
		}

		svcs := []string{
			"db",
			"redis",
			"elasticsearch",
			"opensearch",
			"mercure",
		}

		for _, svc := range svcs {
			if c.IsSvcEnabled(svc) && c.Docker.ContainerRunning(svc) {
				names, _ := c.Docker.ContainerNamesByName(svc)

				t.AppendRow([]interface{}{
					fmt.Sprintf("%s Container Name", cases.Title(language.English).String(svc)),
					c.normalizeNames(names),
				})
			}
		}
	}
}

func (c *Client) infoGlobalServices(t table.Writer) {
	t.AppendSeparator()
	t.AppendRow([]interface{}{"Global Services"})
	t.AppendSeparator()

	for _, service := range c.Config.Services() {
		if c.Config.SvcEnabledPermissive(service) {
			t.AppendRow([]interface{}{
				service,
				fmt.Sprintf("https://%s.%s", service, c.Config.ServiceDomain()),
			})
		}
	}

	for _, service := range c.Config.OptionalServices() {
		if c.Config.SvcEnabledPermissive(service) {
			t.AppendRow([]interface{}{
				service,
				fmt.Sprintf("https://%s.%s", service, c.Config.ServiceDomain()),
			})
		}
	}
}

func (c *Client) normalizeNames(names []string) string {
	normalizedNames := make([]string, len(names))
	for i, name := range names {
		normalizedNames[i] = strings.TrimPrefix(name, "/")
	}

	return strings.Join(normalizedNames, ", ")
}
