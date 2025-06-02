package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"message/internal/domain/models"
	"message/internal/storage"
	"os"

	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	// TODO: with postgres
	const op = "storage.postgres.New"

	if _, err := os.Stat("sqlite3://" + storagePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("%s: database file does not exist: %w", op, err)
	}

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) CreateMessage(ctx context.Context, uid int64, content string, typeOf int32, datetime string) (int64, error) {
	const op = "storage.postgres.CreateMessage"

	stmt, err := s.db.Prepare("INSERT INTO messages (content, uid, type, datetime) VALUES (?, ?, ?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		err := stmt.Close()
		if err != nil {
		}
	}()

	res, err := stmt.ExecContext(ctx, content, uid, typeOf, datetime)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	mid, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return mid, nil
}

func (s *Storage) GetMessage(ctx context.Context, mid int64) (models.Message, error) {
	const op = "storage.postgres.GetMessage"

	stmt, err := s.db.Prepare("SELECT * FROM messages WHERE id=?")
	if err != nil {
		return models.Message{}, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		err := stmt.Close()
		if err != nil {
		}
	}()

	var message models.Message
	if err := stmt.QueryRowContext(ctx, mid).Scan(&message.ID, &message.Content, &message.UserID, &message.Type, &message.DateTime); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Message{}, fmt.Errorf("%s: %w", op, storage.ErrMessageNotExist)
		}
		return models.Message{}, fmt.Errorf("%s: %w", op, err)
	}
	switch message.Type {
	case "1":
		message.Type = "text"
	case "2":
		message.Type = "image"
	case "3":
		message.Type = "file"
	}
	return message, nil
}

func (s *Storage) DeleteMessage(ctx context.Context, mid int64) (bool, error) {
	const op = "storage.postgres.DeleteMessage"
	stmt, err := s.db.Prepare("DELETE from messages WHERE id=?")
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		err := stmt.Close()
		if err != nil {

		}
	}()
	res, err := stmt.ExecContext(ctx, mid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("%s: %w", op, storage.ErrMessageNotExist)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}
	_, err = res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return true, nil
}

func (s *Storage) UpdateMessage(ctx context.Context, mid int64, newContent string) (bool, error) {
	const op = "storage.postgres.UpdateMessage"
	stmt, err := s.db.Prepare("UPDATE messages SET content=? where id=?")
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		err := stmt.Close()
		if err != nil {
		}
	}()

	res, err := stmt.ExecContext(ctx, newContent, mid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("%s: %w", op, storage.ErrMessageNotExist)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return n > 0, nil
}

func (s *Storage) ShowAllMessages(ctx context.Context, uid int64) ([]models.Message, error) {
	const op = "storage.postgres.ShowAllMessages"
	query := `
        SELECT id, uid, content, type, datetime 
        FROM messages 
        ORDER BY id ASC
    `
	//TODO: integrate users
	/*answer, err := s.IsBanned(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if !answer {
		return nil, fmt.Errorf("%s: %w", op, storage.Banned)
	}*/

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		if err := rows.Scan(
			&msg.ID,
			&msg.UserID,
			&msg.Content,
			&msg.Type,
			&msg.DateTime,
		); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		switch msg.Type {
		case "1":
			msg.Type = "text"
		case "2":
			msg.Type = "image"
		case "3":
			msg.Type = "file"
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(messages) == 0 {
		return make([]models.Message, 0), storage.ErrNoMessagesFound
	}

	return messages, nil
}

func (s *Storage) IsBanned(ctx context.Context, uid int64) (bool, error) {
	const op = "storage.postgres.IsBanned"
	stmt, err := s.db.Prepare("SELECT role FROM users WHERE id=?")
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		err := stmt.Close()
		if err != nil {
		}
	}()

	var role int
	if err := stmt.QueryRowContext(ctx, uid).Scan(&role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("%s: %w", op, storage.ErrMessageNotExist)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return role == 0, nil
}
