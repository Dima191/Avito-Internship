package main

import (
	"avito_intership/internal/app"
	"context"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	defer stop()

	a, err := app.New(ctx)
	if err != nil {
		panic(err)
	}

	go func() {
		if err = a.Run(ctx); err != nil {
			stop()
		}
	}()

	<-ctx.Done()
	a.Stop()
}
