package orchestrator

import (
	"context"
	"time"

	"golang.org/x/exp/slog"
)

const CleanupInterval = -1 * time.Hour

type Cleaner struct {
	driver Driver
	ticker *time.Ticker
	opts   *CleanerOpts
}

type CleanerOpts struct {
	Namespace string
	Interval  *time.Duration
}

func NewCleaner(driver Driver, opts *CleanerOpts) *Cleaner {
	if opts.Interval == nil {
		ivl := CleanupInterval
		opts.Interval = &ivl
	}
	return &Cleaner{
		driver: driver,
		ticker: time.NewTicker(5 * time.Minute),
		opts:   opts,
	}
}

func (c *Cleaner) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			slog.Info("shutting down job cleaner")
			return
		case <-c.ticker.C:
			err := c.driver.CleanExpiredJobs(ctx, c.opts.Namespace, c.opts.Interval)
			if err != nil {
				slog.Error("failed to cleanup jobs", slog.Any("err", err))
			}
		}
	}
}
