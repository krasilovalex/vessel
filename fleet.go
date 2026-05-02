package vessel

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
	"github.com/pterm/pterm"
	"golang.org/x/sync/errgroup"
)

type Fleet struct {
	containers []*Container
	network    string
}

func NewFleet(containers ...*Container) *Fleet {
	return &Fleet{
		containers: containers,
		network:    "vessel-default",
	}
}

func (f *Fleet) Up(ctx context.Context) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to init docker client for fleet: %w", err)
	}

	pterm.Info.Printfln("🌊 Launching a fleet of %d containers concurrently...", len(f.containers))

	if err := createNetwork(ctx, cli, f.network); err != nil {
		return err
	}

	g, gCtx := errgroup.WithContext(ctx)

	for _, c := range f.containers {
		c.SetNetwork(f.network)
		container := c
		g.Go(func() error {
			return container.Up(gCtx)
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("fleet startyp failed: %w", err)
	}

	pterm.Success.Println("All containers in the fleet are up and running")
	return nil
}

func (f *Fleet) Stop(ctx context.Context) error {
	g, gCtx := errgroup.WithContext(ctx)

	pterm.Info.Println("⏸️  Stopping fleet without removing...")

	for _, c := range f.containers {
		container := c
		g.Go(func() error {
			return container.Stop(gCtx)
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("fleet stop failed: %w", err)
	}

	return nil
}

func (f *Fleet) Down(ctx context.Context) error {
	g, gCtx := errgroup.WithContext(ctx)

	pterm.Info.Println("🛑 Stopping and removing fleet...")

	for _, c := range f.containers {
		container := c
		g.Go(func() error {
			if err := container.Stop(gCtx); err != nil {
				return err
			}
			return container.Remove(gCtx)
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("Fleet teardown failed: %w", err)
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		_ = removeNetwork(ctx, cli, f.network)
	}

	pterm.Success.Println("Fleet successufully dismanted.")
	return nil
}
