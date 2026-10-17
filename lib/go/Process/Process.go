package Process

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	PARENT_CHECK_INTERVAL = time.Second
	SHUTDOWN_DEADLINE     = 15 * time.Second
)

func Context() (context.Context, context.CancelFunc) {
	parent := os.Getppid()
	bound := bindToParent(parent)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	ctx, cancel := context.WithCancel(ctx)

	if bound {
		go watchParent(ctx, cancel, parent)
	}
	go deadline(ctx)

	return ctx, func() {
		cancel()
		stop()
	}
}

func watchParent(ctx context.Context, cancel context.CancelFunc, parent int) {
	ticker := time.NewTicker(PARENT_CHECK_INTERVAL)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if os.Getppid() != parent {
				cancel()

				return
			}
		}
	}
}

func deadline(ctx context.Context) {
	<-ctx.Done()
	time.Sleep(SHUTDOWN_DEADLINE)
	os.Exit(1)
}
