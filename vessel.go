package vessel

import (
	"fmt"

	"github.com/docker/docker/client"
)

type Container struct {
	image   string
	name    string
	ports   []string // format: "host:container"
	envs    []string // format: "KEY=VALUE"
	volumes []string // format: "/host/path:/container/path"

	id  string         // id created container
	cli *client.Client // instance client docker

	err error
}

func NewContainer(image string) *Container {
	if image == "" {
		return &Container{err: fmt.Errorf("vessel: image name cannot be empty")}
	}
	return &Container{
		image: image,
	}
}

// WithName sets the name of the container
func (c *Container) WithName(name string) *Container {
	if c.err != nil {
		return c
	}

	if name == "" {
		c.err = fmt.Errorf("vessel: container name cannot be empty")
		return c
	}

	c.name = name
	return c
}

// WithPort throws the port.
// hostPort is the port on your machine, and containerPort is the port inside the container
func (c *Container) WithPort(hostPort, containerPort string) *Container {
	if c.err != nil {
		return c
	}

	if hostPort == "" || containerPort == "" {
		c.err = fmt.Errorf("vessel: ports cannot be empty")
		return c
	}
	c.ports = append(c.ports, fmt.Sprintf("%s:%s", hostPort, containerPort))

	return c
}

// WithEnv add  an environment variable.
func (c *Container) WithEnv(key, value string) *Container {
	if c.err != nil {
		return c
	}

	if key == "" {
		c.err = fmt.Errorf("vessel: env key cannot be empty")
		return c
	}

	c.envs = append(c.envs, fmt.Sprintf("%s=%s", key, value))
	return c
}

// WithVolume mounts the directory
func (c *Container) WithVolume(hostPath, containerPath string) *Container {
	if c.err != nil {
		return c
	}

	if hostPath == "" || containerPath == "" {
		c.err = fmt.Errorf("vessel: volume paths cannot be empty")
		return c
	}

	// in the future, we can add a check for absolute paths on the host
	// but for now, we'll just use a lightweight wrapper.

	c.volumes = append(c.volumes, fmt.Sprintf("%s:%s", hostPath, containerPath))
	return c

}
