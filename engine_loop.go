package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
        "github.com/oklog/run"
)

/*
  requirements:
*/

func engineLoop(ctx context.Context) error {
	fmt.Printf("loop start!\n")
	ticker := time.NewTicker(time.Second) // need to change Nanoseconds()
	for {
		select {
		case t := <-ticker.C:
			fmt.Printf("Current time: %v\n", t)
		case <-ctx.Done():
			fmt.Printf("canceled!\n")
			return nil
		}
	}
	return nil
}

func main() {
	var g run.Group
	ctx := context.Background()
	{
		signal_chan := make(chan os.Signal, 1)
		signal.Notify(signal_chan,
			syscall.SIGHUP,
			syscall.SIGKILL,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT)

		ctx, cancel := context.WithCancel(ctx)
		g.Add(
			func() error {
				select {
				case s:= <-signal_chan:
					switch s {
					case syscall.SIGHUP:
						fmt.Printf("sighup!\n")
					default:
						fmt.Printf("sigkill/int/term/quit!\n")
						return nil
					}
				case <-ctx.Done():
					fmt.Printf("canceled!\n")
					return nil
				}
				return nil
			},
			func(err error) {
				cancel()
			},
		)
	}
			
	{
		ctx, cancel := context.WithCancel(ctx)
		g.Add(
			func() error {
				return engineLoop(ctx)
			},
			func(err error) {
				cancel()
			},
		)
	}

	if err := g.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "err: %v", err)
	}
	fmt.Printf("exit!\n")
}
