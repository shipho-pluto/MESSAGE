package crud

import (
	"context"
	"errors"
	"fmt"
	"github.com/shipho-pluto/MESSAGE/internal/domain/models"
	"github.com/shipho-pluto/MESSAGE/internal/storage"
	"log/slog"
)

type CRUD struct {
	Log           *slog.Logger
	MessageCRUDer MessageCRUDer
}

type MessageCRUDer interface {
	CreateMessage(ctx context.Context, uid int64, content string, typeOf int32, datetime string) (int64, error)
	GetMessage(ctx context.Context, mid int64) (models.Message, error)
	DeleteMessage(ctx context.Context, mid int64) (bool, error)
	UpdateMessage(ctx context.Context, mid int64, newContent string) (bool, error)
	ShowAllMessages(ctx context.Context, uid int64) ([]models.Message, error)
}

func (m *CRUD) SentMessage(ctx context.Context, uid int64, content string, typeOf int32, datetime string) (int64, error) {
	const op = "services.crud.SentMessage"
	log := m.Log.With(slog.String("op", op))

	id, err := m.MessageCRUDer.CreateMessage(ctx, uid, content, typeOf, datetime)
	if err != nil {
		log.Error("Failed to create message", slog.String("err", err.Error()))
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

func (m *CRUD) DeleteMessage(ctx context.Context, uid int64) (bool, error) {
	const op = "services.crud.DeleteMessage"
	log := m.Log.With(slog.String("op", op))
	answer, err := m.MessageCRUDer.DeleteMessage(ctx, uid)
	if err != nil {
		if errors.Is(err, storage.ErrMessageNotExist) {
			log.Error("Message does not exist")
			return false, fmt.Errorf("%s: %w", op, err)
		}
		log.Error("Failed to delete message", slog.String("err", err.Error()))
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return answer, err
}

func (m *CRUD) GetMessage(ctx context.Context, mid int64) (models.Message, error) {
	const op = "services.crud.GetMessage"
	log := m.Log.With(slog.String("op", op))
	content, err := m.MessageCRUDer.GetMessage(ctx, mid)
	if err != nil {
		if errors.Is(err, storage.ErrMessageNotExist) {
			log.Error("Message does not exist")
			return models.Message{}, fmt.Errorf("%s: %w", op, err)
		}
		log.Error("Failed to get message", slog.String("err", err.Error()))
		return models.Message{}, fmt.Errorf("%s: %w", op, err)
	}
	return content, err
}

func (m *CRUD) UpdateMessage(ctx context.Context, mid int64, newContent string) (bool, error) {
	const op = "services.crud.UpdateMessage"
	log := m.Log.With(slog.String("op", op))
	answer, err := m.MessageCRUDer.UpdateMessage(ctx, mid, newContent)
	if err != nil {
		if errors.Is(err, storage.ErrMessageNotExist) {
			log.Error("Message does not exist")
			return false, fmt.Errorf("%s: %w", op, err)
		}
		log.Error("Failed to update message", slog.String("err", err.Error()))
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return answer, nil
}

func (m *CRUD) ShowAllMessages(ctx context.Context, uid int64) ([]models.Message, error) {
	const op = "services.crud.ShowAllMessages"
	log := m.Log.With(slog.String("op", op))
	answer, err := m.MessageCRUDer.ShowAllMessages(ctx, uid)
	if err != nil {
		if errors.Is(err, storage.ErrNoMessagesFound) {
			return answer, storage.ErrNoMessagesFound
		}
		log.Error("Failed to show all messages", slog.String("err", err.Error()))
	}
	return answer, nil
}
