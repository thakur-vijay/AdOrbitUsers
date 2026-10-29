package app

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	usersv1 "github.com/adorbit/users/gen/users/v1"
	usersgrpc "github.com/adorbit/users/internal/transport/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func (a *App) Run(addr string) error {
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           a.Router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	grpcListener, err := net.Listen("tcp", a.GRPCPort)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	usersv1.RegisterUsersServiceServer(
		grpcServer,
		usersgrpc.NewUsersServer(a.UsersSvc),
	)
	reflection.Register(grpcServer)
	grpcErrors := make(chan error, 1)
	go func() {
		log.Printf("gRPC server listening on %s", a.GRPCPort)
		grpcErrors <- grpcServer.Serve(grpcListener)
	}()

	consumerCtx, cancelConsumer := context.WithCancel(context.Background())
	defer cancelConsumer()

	consumerErrors := make(chan error, 1)

	go func() {
		consumerErrors <- a.Consumer.Start(consumerCtx)
	}()

	httpErrors := make(chan error, 1)
	go func() {
		log.Printf("HTTP server listening on %s", addr)
		httpErrors <- httpServer.ListenAndServe()
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)
	select {
	case sig := <-stop:
		log.Printf("shutdown signal received: %s", sig)
	case err := <-httpErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			grpcServer.Stop()
			return err
		}
	case err := <-grpcErrors:
		if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			_ = httpServer.Close()
			return err
		}
	case err := <-consumerErrors:
		if err != nil {
			grpcServer.Stop()
			_ = httpServer.Close()
			return err
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		_ = httpServer.Close()
	}
	grpcServer.GracefulStop()
	return nil

}

func (a *App) Close() error {
	if a == nil {
		return nil
	}

	var firstErr error

	if a.RabbitMQ != nil {
		if err := a.RabbitMQ.Close(); err != nil {
			firstErr = err
		}
	}

	if a.DB != nil {
		if err := a.DB.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}
