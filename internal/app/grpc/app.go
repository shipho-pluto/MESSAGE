package grpc

import (
	"fmt"
	"github.com/shipho-pluto/MESSAGE/internal/grpc/crud"
	"google.golang.org/grpc"
	"log/slog"
	"net"
)

type App struct {
	logger     *slog.Logger
	port       int
	GrpcServer *grpc.Server
}

func New(log *slog.Logger, crudService crud.CRUD, secret string, port int) *App {
	gRPCServer := grpc.NewServer()
	crud.RegisterServer(gRPCServer, crudService, secret)
	return &App{
		logger:     log,
		port:       port,
		GrpcServer: gRPCServer,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpc.App.Run"
	log := a.logger.With(slog.String("op", op))
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	log.Info("grpc server started")
	if err := a.GrpcServer.Serve(lis); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (a *App) Stop() {
	const op = "grpc.App.Stop"
	a.logger.With(slog.String("op", op)).
		Info("stopping grpc server", slog.Int("port", a.port))
	a.GrpcServer.GracefulStop()
}
