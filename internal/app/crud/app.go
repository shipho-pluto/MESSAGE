package crud

import (
	"log/slog"
	"message/internal/services/crud"
)

func New(log *slog.Logger, cruder crud.MessageCRUDer) *crud.CRUD {
	return &crud.CRUD{
		Log:           log,
		MessageCRUDer: cruder,
	}
}
