package internal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/hashicorp/go-version"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	dockerRequiredVersion        = "20.4.0"
	dockerComposeRequiredVersion = "1.25.0"
)

// NewDockerClient creates a docker client and return with it.
func NewDockerClient() (*client.Client, error) {
	log.Debugf("Creating a new Docker Client. Host: %v", viper.GetString("docker_host"))

	return client.NewClientWithOpts(client.FromEnv, client.WithHost(viper.GetString("docker_host")))
}

func getDockerVersion() (*version.Version, error) {
	cli, err := NewDockerClient()
	if err != nil {
		return nil, err
	}

	data, err := cli.ServerVersion(context.Background())
	if err != nil {
		return nil, err
	}

	v, err := version.NewVersion(data.Version)
	if err != nil {
		return nil, err
	}

	return v, err
}

func dockerIsRunning() bool {
	_, err := getDockerVersion()

	return err == nil
}

func checkDockerVersion() bool {
	v, err := getDockerVersion()
	if err != nil {
		return false
	}

	requiredVersion, err := version.NewVersion(dockerRequiredVersion)
	if err != nil {
		return false
	}

	if v.LessThan(requiredVersion) {
		return false
	}

	return true
}

func getDockerComposeVersion() (*version.Version, error) {
	data, err := RunDockerComposeCommand([]string{"version", "--short"}, true)
	if err != nil {
		return nil, err
	}

	v, err := version.NewVersion(strings.TrimSpace(data))
	if err != nil {
		return nil, err
	}

	return v, err
}

func checkDockerComposeVersion() bool {
	v, err := getDockerComposeVersion()
	if err != nil {
		return false
	}

	requiredVersion, err := version.NewVersion(dockerComposeRequiredVersion)
	if err != nil {
		return false
	}

	if v.LessThan(requiredVersion) {
		return false
	}

	return true
}

// CheckDocker checks if docker-engine is running or not.
func CheckDocker() error {
	if !dockerIsRunning() {
		return ErrDockerIsNotRunning
	}

	if !checkDockerVersion() {
		ver, err := getDockerVersion()
		if err != nil {
			return err
		}

		return DockerVersionMismatchError(fmt.Sprintf(
			"your docker version is %v, required version: %v",
			ver.String(),
			dockerRequiredVersion,
		))
	}

	if !checkDockerComposeVersion() {
		ver, err := getDockerComposeVersion()
		if err != nil {
			return err
		}

		return DockerComposeVersionMismatchError(fmt.Sprintf(
			"your docker-compose version is %v, required version: %v",
			ver.String(),
			dockerComposeRequiredVersion,
		))
	}

	return nil
}

// LookupContainerAddressInNetwork returns the container IP address in the specific network.
func LookupContainerAddressInNetwork(containerName, networkName string) (string, error) {
	ctx := context.Background()

	c, err := NewDockerClient()
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	f := filters.NewArgs()

	f.Add("name", containerName)

	filterName := types.ContainerListOptions{
		Filters: f,
	}

	containers, err := c.ContainerList(ctx, filterName)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	if len(containers) != 1 {
		return "", CannotFindContainerError(containerName)
	}

	var ipAddress string

	for _, container := range containers {
		inspect, err := c.ContainerInspect(ctx, container.ID)
		if err != nil {
			log.Debugf("%v", err)
		}

		if val, ok := inspect.NetworkSettings.Networks[networkName]; ok {
			ipAddress = val.IPAddress
		}
	}

	return ipAddress, nil
}

// LookupContainerGatewayInNetwork returns the container IP address in the specific network.
func LookupContainerGatewayInNetwork(containerName, networkName string) (string, error) {
	ctx := context.Background()

	c, err := NewDockerClient()
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	f := filters.NewArgs()

	f.Add("name", containerName)

	filterName := types.ContainerListOptions{
		Filters: f,
	}

	containers, err := c.ContainerList(ctx, filterName)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	if len(containers) != 1 {
		return "", CannotFindContainerError(containerName)
	}

	var gatewayAddress string

	for _, container := range containers {
		inspect, err := c.ContainerInspect(ctx, container.ID)
		if err != nil {
			log.Debugf("%v", err)
		}

		if val, ok := inspect.NetworkSettings.Networks[networkName]; ok {
			gatewayAddress = val.Gateway
		}
	}

	return gatewayAddress, nil
}

// GetContainerIDByName returns a container ID of the containerName running in
//   the current environment.
func GetContainerIDByName(containerName string) (string, error) {
	ctx := context.Background()

	c, err := NewDockerClient()
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	f := filters.NewArgs()

	f.Add("name", containerName)
	f.Add("label", "com.docker.compose.project="+GetEnvName())

	filterName := types.ContainerListOptions{
		Filters: f,
	}

	containers, err := c.ContainerList(ctx, filterName)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	if len(containers) != 1 {
		return "", CannotFindContainerError(containerName)
	}

	id := containers[0].ID

	return id, nil
}

// GetContainerStateByName returns the container state of the containerName running in
//   the current environment.
func GetContainerStateByName(containerName string) (string, error) {
	ctx := context.Background()

	c, err := NewDockerClient()
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	f := filters.NewArgs()

	f.Add("name", containerName)
	f.Add("label", "com.docker.compose.project="+GetEnvName())

	filterName := types.ContainerListOptions{
		Filters: f,
	}

	containers, err := c.ContainerList(ctx, filterName)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	if len(containers) != 1 {
		return "", CannotFindContainerError(containerName)
	}

	state := containers[0].State

	return state, nil
}

// RunDockerComposeCommand runs the passed parameters with docker-compose and returns the output.
func RunDockerComposeCommand(args []string, suppressOsStdOut ...bool) (string, error) {
	log.Tracef("args: %#v", args)
	log.Debugf("Running command: docker-compose %v", strings.Join(args, " "))

	cmd := exec.Command("docker-compose", args...)

	var combinedOutBuf bytes.Buffer

	cmd.Stdin = os.Stdin
	if len(suppressOsStdOut) > 0 && suppressOsStdOut[0] {
		cmd.Stdout = io.Writer(&combinedOutBuf)
		cmd.Stderr = io.Writer(&combinedOutBuf)
	} else {
		cmd.Stdout = io.Writer(os.Stdout)
		cmd.Stderr = io.Writer(os.Stderr)
	}

	err := cmd.Run()
	outStr := combinedOutBuf.String()

	// if err != nil {
	// 	return outStr, err
	// }

	return outStr, err //nolint:wrapcheck
}

func getDockerNetworksWithLabel(label string) ([]string, error) {
	ctx := context.Background()

	c, err := NewDockerClient()
	if err != nil {
		return []string{}, fmt.Errorf("%w", err)
	}

	f := filters.NewArgs()

	f.Add("label", label)

	filterName := types.NetworkListOptions{
		Filters: f,
	}

	networks, err := c.NetworkList(ctx, filterName)
	if err != nil {
		return []string{}, fmt.Errorf("%w", err)
	}

	log.Traceln("networks: ", networks)

	if len(networks) == 0 {
		return []string{}, nil
	}

	var result []string
	for _, network := range networks {
		result = append(result, network.Name)
	}

	return result, nil
}
