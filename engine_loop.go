package main

import (
	"context"
	"fmt"
	"github.com/distributed-monitoring/policy-engine-sandbox/policyexpr"
	"github.com/distributed-monitoring/policy-engine-sandbox/threshold"
	"github.com/oklog/run"
	"os"
	"os/signal"
	"syscall"
	"time"
)

/*
  requirements:
*/

func policyProcess(p PolicyYaml) error {
	for _, g := range p.Groups {
		for i, r := range g.Rules {
			expr := policyexpr.Policyexpr_main(r.Expr)
			rdlist := threshold.Read(expr)
			rllist := threshold.Evaluate(expr, rdlist)
			fmt.Printf("transmit: %s[%d], %+v\n", g.Name, i, rllist)
		}
	}
	fmt.Printf("\n")
	return nil
}

func engineLoop(ctx context.Context, p PolicyYaml) error {
	fmt.Printf("loop start!\n")
	ticker := time.NewTicker(time.Second) // need to change Nanoseconds()
	for {
		select {
		case t := <-ticker.C:
			fmt.Printf("Current time: %v\n", t)
			policyProcess(p)
		case <-ctx.Done():
			fmt.Printf("canceled!\n")
			return nil
		}
	}
	return nil
}

func engine_loop_main(p PolicyYaml) {
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
				case s := <-signal_chan:
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
				return engineLoop(ctx, p)
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
