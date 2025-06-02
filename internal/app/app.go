package app

import (
	"log/slog"
	crudApp "message/internal/app/crud"
	grpcApp "message/internal/app/grpc"
	"message/internal/clients/service"
	"message/internal/storage/postgres"
)

type App struct {
	GRPCServer *grpcApp.App
	SSOClient  *service.ClientCRUD
}

func New(log *slog.Logger, storagePath string, secret string, port int, ssoClient *service.ClientCRUD) *App {

	storagePostgres, err := postgres.New(storagePath)
	if err != nil {
		panic(err)
	}
	log.Info("Starting storage")
	crudService := crudApp.New(log, storagePostgres)
	grpcSever := grpcApp.New(log, crudService, secret, port)
	return &App{
		GRPCServer: grpcSever,
		SSOClient:  ssoClient,
	}
}
