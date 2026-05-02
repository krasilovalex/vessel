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
	"github.com/pterm/pterm"
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
	_, _, inspectErr := cli.ImageInspectWithRaw(ctx, c.image)
	if inspectErr != nil {
		fmt.Printf("☁️  Image %s not found locally. Pulling from registry...\n", c.image)
		reader, pullErr := cli.ImagePull(ctx, c.image, image.PullOptions{})
		if pullErr != nil {
			return fmt.Errorf("failed to pull image %s: %w", c.image, pullErr)
		}
		_, _ = io.Copy(os.Stdout, reader)
		reader.Close()
	} else {
		fmt.Printf("📦 Image %s found locally. Skipping pull.\n", c.image)
	}

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
	spinner, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Starting container %s...", c.name))

	resp, err := cli.ContainerCreate(ctx,
		&container.Config{
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
		spinner.Fail(fmt.Sprintf("Failed to create container %s", c.name))
		return fmt.Errorf("failed to create container: %w", err)
	}
	c.id = resp.ID

	if err := cli.ContainerStart(ctx, c.id, container.StartOptions{}); err != nil {
		spinner.Fail(fmt.Sprintf("Failed to start container %s", c.name))
		return fmt.Errorf("failed to start container: %w", err)
	}

	spinner.Success(fmt.Sprintf("Container %s is up and running! (ID: %s)", c.name, c.id[:12]))
	return nil
}

// stop gracefully stops the container, but does NOT delete it
func (c *Container) Stop(ctx context.Context) error {
	if c.cli == nil || c.id == "" {
		return nil
	}

	spinner, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Stopping container %s...", c.name))

	err := c.cli.ContainerStop(ctx, c.id, container.StopOptions{})
	if err != nil {
		spinner.Fail(fmt.Sprintf("Failed to stop container %s", c.name))
		return fmt.Errorf("failed to stop container: %w", err)
	}

	spinner.Success(fmt.Sprintf("Container %s stopped.", c.name))
	return nil
}

// Remove removes the container and its anonymous volumes
func (c *Container) Remove(ctx context.Context) error {
	if c.cli == nil || c.id == "" {
		return nil
	}

	spinner, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Removing container %s...", c.name))

	err := c.cli.ContainerRemove(ctx, c.id, container.RemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	})
	if err != nil {
		spinner.Fail(fmt.Sprintf("Failed to remove container %s", c.name))
		return fmt.Errorf("failed to remove container: %w", err)
	}

	spinner.Success(fmt.Sprintf("Container %s removed.", c.name))
	return nil
}
