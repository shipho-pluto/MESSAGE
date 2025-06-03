package crud

import (
	"github.com/shipho-pluto/MESSAGE/internal/services/crud"
	"log/slog"
)

func New(log *slog.Logger, cruder crud.MessageCRUDer) *crud.CRUD {
	return &crud.CRUD{
		Log:           log,
		MessageCRUDer: cruder,
	}
}
