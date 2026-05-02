package vessel

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// Downlaod image, create and launch container
func (c *Container) Up(ctx context.Context) error {
	if c.err != nil {
		return fmt.Errorf("vessel builder error: %v", c.err)
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to init docker client: %w", err)
	}

	c.cli = cli

	fmt.Printf("Pulling image %s...\n", c.image)
	reader, err := cli.ImagePull(ctx, c.image, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	_, _ = io.Copy(os.Stdout, reader)
	reader.Close()

	exposedPorts := nat.PortSet{}
	portBindings := nat.PortMap{}

	for _, p := range c.ports {
		parts := strings.Split(p, ":")
		if len(parts) != 2 {
			continue
		}

		hostPort, containerPort := parts[0], parts[1]
		port, err := nat.NewPort("tcp", containerPort)
		if err != nil {
			return fmt.Errorf("invalid container port %s: %w", containerPort, err)
		}

		exposedPorts[port] = struct{}{}
		portBindings[port] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: hostPort,
			},
		}
	}
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        c.image,
		Env:          c.envs,
		ExposedPorts: exposedPorts,
	},
		&container.HostConfig{
			PortBindings: portBindings,
			Binds:        c.volumes,
		},
		nil, nil, c.name,
	)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}
	c.id = resp.ID

	fmt.Printf("Starting container %s (ID: %s)...\n", c.name, c.id[:12])
	if err := cli.ContainerStart(ctx, c.id, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	return nil
}

// stop gracefully stops the container, but does NOT delete it
func (c *Container) Stop(ctx context.Context) error {
	if c.cli == nil || c.id == "" {
		return nil
	}

	fmt.Printf("\nStopping container %s...\n", c.name)

	err := c.cli.ContainerStop(ctx, c.id, container.StopOptions{})
	if err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	fmt.Println("Done. See ya!")
	return nil
}

// Remove removes the container and its anonymous volumes
func (c *Container) Remove(ctx context.Context) error {
	if c.cli == nil || c.id == "" {
		return nil
	}

	fmt.Printf("Removing container %s...\n", c.name)
	err := c.cli.ContainerRemove(ctx, c.id, container.RemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	})

	if err != nil {
		return fmt.Errorf("failed to remove contaner: %w", err)
	}

	return nil
}
