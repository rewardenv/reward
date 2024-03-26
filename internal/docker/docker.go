package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	dockerpkg "github.com/docker/docker/client"
	"github.com/hashicorp/go-version"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var requiredVersion = "20.4.0"

var (
	// ErrDockerAPIIsUnreachable occurs when Docker is not running
	// or the user who runs the application cannot call Docker API.
	ErrDockerAPIIsUnreachable = func(err error) error {
		return fmt.Errorf("docker api is unreachable: %w", err)
	}

	// ErrDockerVersionMismatch occurs when the Docker version is too old.
	ErrDockerVersionMismatch = func(s string) error {
		return fmt.Errorf("docker version is too old: %s", s)
	}

	// ErrCannotFindContainer occurs when the application cannot find the requested container.
	ErrCannotFindContainer = func(s string, err error) error {
		return fmt.Errorf("cannot find container: %s, error: %w", s, err)
	}

	// ErrNoContainersFound occurs when the application found zero containers.
	ErrNoContainersFound = func() error {
		return fmt.Errorf("no containers found")
	}

	// ErrTooManyContainersFound occurs when the application found more than 1 container.
	ErrTooManyContainersFound = func(s string) error {
		return fmt.Errorf("too many containers found: %s", s)
	}

	// ErrCannotFindNetwork occurs when the application cannot find the requested network during container inspection.
	ErrCannotFindNetwork = func(s string) error {
		return fmt.Errorf("cannot find network: %s", s)
	}
)

type Client struct {
	*dockerpkg.Client
}

func (c *Client) AppName() string {
	return viper.GetString("app_name")
}

func (c *Client) EnvName() string {
	return strings.ToLower(viper.GetString(fmt.Sprintf("%s_env_name", c.AppName())))
}

// NewClient creates a docker client and return with it.
func NewClient(dockerHost string) (*Client, error) {
	var (
		docker *dockerpkg.Client
		err    error
	)

	if dockerHost != "" {
		log.Debugf("Creating a new Docker client using host: %s...", dockerHost)

		docker, err = dockerpkg.NewClientWithOpts(dockerpkg.FromEnv, dockerpkg.WithHost(dockerHost), dockerpkg.WithAPIVersionNegotiation())
		if err != nil {
			return nil, ErrDockerAPIIsUnreachable(err)
		}
	} else {
		log.Debugln("Creating a new Docker client from the default settings...")

		docker, err = dockerpkg.NewClientWithOpts(dockerpkg.FromEnv, dockerpkg.WithAPIVersionNegotiation())
		if err != nil {
			return nil, ErrDockerAPIIsUnreachable(err)
		}
	}

	log.Debugf("...docker client created.")

	return &Client{
		Client: docker,
	}, nil
}

func Must(c *Client, err error) *Client {
	if err != nil {
		log.Fatalln(err)
	}

	return c
}

func (c *Client) dockerVersion() (*version.Version, error) {
	log.Debugln("Fetching docker version...")

	data, err := c.ServerVersion(context.Background())
	if err != nil {
		return nil, fmt.Errorf("cannot fetch docker version: %w", err)
	}

	v, err := version.NewVersion(data.Version)
	if err != nil {
		return nil, fmt.Errorf("cannot parse docker version: %w", err)
	}

	log.Debugf("...docker version is: %s.", v.String())

	return v, nil
}

func (c *Client) isMinimumVersionInstalled() bool {
	log.Debugln("Checking if minimum docker version is installed...")

	v, err := c.dockerVersion()
	if err != nil {
		log.Traceln("...cannot fetch docker version.")

		return false
	}

	if v.LessThan(version.Must(version.NewVersion(requiredVersion))) {
		log.Tracef("...docker version is too old. Your version is: %s, required version: %s.",
			v.String(),
			requiredVersion,
		)

		return false
	}

	log.Debugln("...minimum docker version is installed.")

	return true
}

// Check checks if docker-engine is running or not.
func (c *Client) Check() error {
	log.Debugln("Checking Docker...")

	if !c.isMinimumVersionInstalled() {
		ver, err := c.dockerVersion()
		if err != nil {
			log.Traceln("...cannot fetch docker version.")

			return fmt.Errorf("cannot fetch docker version: %w", err)
		}

		return ErrDockerVersionMismatch(
			fmt.Sprintf(
				"your docker version is %s, required version: %s",
				ver.String(),
				requiredVersion,
			),
		)
	}

	log.Debugln("...docker version is appropriate.")

	return nil
}

func (c *Client) verifyContainerResults(containers []types.Container) error {
	log.Debugln("Verifying container results...")

	for _, v := range containers {
		log.Tracef("Found containers: %s", v.Names)
	}

	if len(containers) == 0 {
		return ErrNoContainersFound()
	} else if len(containers) > 1 {
		var containerNames []string
		for _, c := range containers {
			containerNames = append(containerNames, c.Names...)
		}

		return ErrTooManyContainersFound("containers: " + strings.Join(containerNames, " "))
	}

	log.Debugln("...container results are verified.")

	return nil
}

// ContainerAddressInNetwork returns the container IP address in the specific network.
func (c *Client) ContainerAddressInNetwork(containerName, environmentName, networkName string) (string, error) {
	log.Debugln("Looking up container address in network...")

	ctx := context.Background()

	containers, err := c.ContainerList(ctx, container.ListOptions{
		Filters: filters.NewArgs(
			filters.KeyValuePair{
				Key:   "label",
				Value: fmt.Sprintf("dev.%s.container.name=%s", c.AppName(), containerName),
			},
			filters.KeyValuePair{
				Key:   "label",
				Value: fmt.Sprintf("dev.%s.environment.name=%s", c.AppName(), environmentName),
			},
		),
	})
	if err != nil {
		return "", fmt.Errorf("cannot list containers: %w", err)
	}

	err = c.verifyContainerResults(containers)
	if err != nil {
		return "", ErrCannotFindContainer(containerName, err)
	}

	inspect, err := c.ContainerInspect(ctx, containers[0].ID)
	if err != nil {
		log.Errorf("cannot inspect container: %s", err)
	}

	val, ok := inspect.NetworkSettings.Networks[networkName]
	if !ok {
		return "", ErrCannotFindNetwork(networkName)
	}

	log.Debugln("...container address in network found.")

	return val.IPAddress, nil
}

// ContainerGatewayInNetwork returns the container IP address in the specific network.
func (c *Client) ContainerGatewayInNetwork(containerName, networkName string) (string, error) {
	log.Debugln("Looking up container gateway in network...")

	ctx := context.Background()

	containers, err := c.ContainerList(ctx, container.ListOptions{
		Filters: filters.NewArgs(
			filters.KeyValuePair{
				Key:   "label",
				Value: fmt.Sprintf("dev.%s.container.name=%s", c.AppName(), containerName),
			},
			filters.KeyValuePair{
				Key:   "label",
				Value: fmt.Sprintf("dev.%s.environment.name=%s", c.AppName(), c.EnvName()),
			},
		),
	})
	if err != nil {
		return "", fmt.Errorf("cannot list containers: %w", err)
	}

	err = c.verifyContainerResults(containers)
	if err != nil {
		return "", ErrCannotFindContainer(containerName, err)
	}

	inspect, err := c.ContainerInspect(ctx, containers[0].ID)
	if err != nil {
		return "", fmt.Errorf("cannot inspect container: %w", err)
	}

	val, ok := inspect.NetworkSettings.Networks[networkName]
	if !ok {
		return "", ErrCannotFindNetwork(networkName)
	}

	log.Debugln("...container gateway in network found.")

	return val.Gateway, nil
}

// ContainerIDByName returns a container ID of the containerName running in the current environment.
func (c *Client) ContainerIDByName(containerName string) (string, error) {
	log.Debugln("Looking up container ID by name...")

	containers, err := c.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(
			filters.KeyValuePair{
				Key:   "label",
				Value: fmt.Sprintf("dev.%s.container.name=%s", c.AppName(), containerName),
			},
			filters.KeyValuePair{
				Key:   "label",
				Value: fmt.Sprintf("dev.%s.environment.name=%s", c.AppName(), c.EnvName()),
			},
		),
	})
	if err != nil {
		return "", fmt.Errorf("cannot list containers: %w", err)
	}

	err = c.verifyContainerResults(containers)
	if err != nil {
		return "", ErrCannotFindContainer(containerName, err)
	}

	log.Debugln("...container ID by name found.")

	return containers[0].ID, nil
}

// ContainerNamesByName returns a container names of the containerName running in the current environment.
func (c *Client) ContainerNamesByName(containerName string) ([]string, error) {
	log.Debugln("Looking up container Names by name...")

	containers, err := c.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(
			filters.KeyValuePair{
				Key:   "label",
				Value: fmt.Sprintf("dev.%s.container.name=%s", c.AppName(), containerName),
			},
			filters.KeyValuePair{
				Key:   "label",
				Value: fmt.Sprintf("dev.%s.environment.name=%s", c.AppName(), c.EnvName()),
			},
		),
	})
	if err != nil {
		return nil, fmt.Errorf("cannot list containers: %w", err)
	}

	err = c.verifyContainerResults(containers)
	if err != nil {
		return nil, ErrCannotFindContainer(containerName, err)
	}

	log.Debugln("...container ID by name found.")

	return containers[0].Names, nil
}

// ContainerStateByName returns the container state of the containerName running in the current environment.
func (c *Client) ContainerStateByName(containerName string) (string, error) {
	log.Debugln("Looking up container state by name...")

	containers, err := c.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(
			filters.KeyValuePair{
				Key:   "label",
				Value: fmt.Sprintf("dev.%s.container.name=%s", c.AppName(), containerName),
			},
			filters.KeyValuePair{
				Key:   "label",
				Value: fmt.Sprintf("dev.%s.environment.name=%s", c.AppName(), c.EnvName()),
			},
		),
	})
	if err != nil {
		return "", fmt.Errorf("cannot list containers: %w", err)
	}

	err = c.verifyContainerResults(containers)
	if err != nil {
		return "", ErrCannotFindContainer(containerName, err)
	}

	log.Debugln("...container state by name found.")

	return containers[0].State, nil
}

// NetworkNamesByLabel returns a list of network names that have the specified label.
func (c *Client) NetworkNamesByLabel(label string) ([]string, error) {
	log.Debugln("Looking up network names by label...")

	networks, err := c.NetworkList(context.Background(), types.NetworkListOptions{
		Filters: filters.NewArgs(
			filters.KeyValuePair{
				Key:   "label",
				Value: label,
			},
		),
	})
	if err != nil {
		return []string{}, fmt.Errorf("cannot list networks: %w", err)
	}

	for _, v := range networks {
		log.Tracef("Found networks: %s.", v.Name)
	}

	results := make([]string, 0, len(networks))
	for i, network := range networks {
		results[i] = network.Name
	}

	return results, nil
}

// ContainerRunning returns true if container is running.
func (c *Client) ContainerRunning(container string) bool {
	_, err := c.ContainerIDByName(container)

	return err == nil
}

// NetworkExist returns true if the docker network exists.
func (c *Client) NetworkExist(networkName string) (bool, error) {
	networks, err := c.NetworkList(context.Background(), types.NetworkListOptions{
		Filters: filters.NewArgs(
			filters.KeyValuePair{
				Key:   "name",
				Value: networkName,
			},
		),
	})
	if err != nil {
		return false, fmt.Errorf("cannot list networks: %w", err)
	}

	if len(networks) == 0 {
		return false, nil
	}

	return true, nil
}
