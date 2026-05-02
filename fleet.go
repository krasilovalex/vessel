package vessel

import (
	"context"
	"fmt"

	"github.com/pterm/pterm"
	"golang.org/x/sync/errgroup"
)

type Fleet struct {
	containers []*Container
}

func NewFleet(containers ...*Container) *Fleet {
	return &Fleet{
		containers: containers,
	}
}

func (f *Fleet) Up(ctx context.Context) error {
	g, gCtx := errgroup.WithContext(ctx)

	pterm.Info.Printfln("🌊 Launching a fleet of %d containers concurrently...", len(f.containers))

	for _, c := range f.containers {
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

	pterm.Success.Println("Fleet successufully dismanted.")
	return nil
}
