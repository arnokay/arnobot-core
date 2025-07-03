package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func (app *application) Start() {
	startError := make(chan error)
	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		s := <-quit

		app.logger.Info("shutting down server", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		shutdownError <- app.Shutdown(ctx)
	}()

	go func() {
		err := startMBServer(app)
		if err != nil {
			startError <- err
		}
	}()
	select {
	case err := <-startError:
		app.logger.Error("application start error", "err", err)
	case err := <-shutdownError:
		if err != nil {
			app.logger.Error("application shutdown error", "err", err)
		}
	}
	close(startError)
	close(shutdownError)
}

func (app *application) Shutdown(ctx context.Context) error {
	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		app.logger.Debug("#shutdown.mb: gracefully closing mb connection")

		done := make(chan struct{})
		go func() {
			err := app.pubSub.Drain()
			if err != nil {
				errCh <- err
			}
			close(done)
		}()

		select {
		case <-done:
			app.logger.Debug("#shutdown.mb: gracefully closed mb connection")
		case <-ctx.Done():
			app.logger.Debug("#shutdown.mb: context timeout, force closing message broker connection")
			app.pubSub.Close()
			app.logger.Debug("#shutdown:mb: force closed mb connection")
		}
	}()

	wg.Wait()
	close(errCh)

	app.logger.Debug("#shutdown.db: gracefully closing db")
	app.db.Close()
	app.logger.Debug("#shutdown.db: gracefully closed db")

	for err := range errCh {
		return err
	}

	return nil
}

func startMBServer(a *application) error {
	if a.pubSub == nil {
		return errors.New("startMBServer: msgBroker is nil")
	}

	if a.pubSub.IsClosed() {
		return errors.New("startMBServer: msgBroker is closed")
	}

	a.mbControllers.Connect(a.pubSub)

	return nil
}
