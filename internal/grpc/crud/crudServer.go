package crud

import (
	"context"
	"errors"
	"github.com/shipho-pluto/MESSAGE/internal/domain/models"
	jwtVal "github.com/shipho-pluto/MESSAGE/internal/lib/jwt"
	"github.com/shipho-pluto/MESSAGE/internal/storage"
	crudv1 "github.com/shipho-pluto/PROTO/gen/go/crud"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CRUD interface {
	GetMessage(ctx context.Context, mid int64) (models.Message, error)
	UpdateMessage(ctx context.Context, mid int64, newContent string) (bool, error)
	SentMessage(ctx context.Context, uid int64, content string, typeOf int32, datetime string) (int64, error)
	DeleteMessage(ctx context.Context, uid int64) (bool, error)
	ShowAllMessages(ctx context.Context, uid int64) ([]models.Message, error)
}

type serverCRUD struct {
	crudv1.UnimplementedMessageServer
	crud   CRUD
	Secret string
}

func RegisterServer(gRPCServer *grpc.Server, crud CRUD, secret string) {
	crudv1.RegisterMessageServer(gRPCServer, &serverCRUD{crud: crud, Secret: secret})
}

func (s *serverCRUD) SentMessage(ctx context.Context, req *crudv1.SentMessageRequest) (*crudv1.SentMessageResponse, error) {
	TokenResponse := jwtVal.ValidateToken(req.GetToken(), s.Secret)
	if TokenResponse.Error != nil {
		return nil, status.Error(codes.Unauthenticated, "failed in decoding token "+TokenResponse.Error.Error())
	}

	id, err := s.crud.SentMessage(ctx, TokenResponse.UserID, req.GetContent(), req.GetType(), req.GetDatetime())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to create message")
	}
	return &crudv1.SentMessageResponse{Mid: id}, nil
}

func (s *serverCRUD) DeleteMessage(ctx context.Context, req *crudv1.DeleteMessageRequest) (*crudv1.DeleteMessageResponse, error) {
	TokenResponse := jwtVal.ValidateToken(req.GetToken(), s.Secret)
	if TokenResponse.Error != nil {
		return nil, status.Error(codes.Unauthenticated, "failed in decoding token")
	}

	answer, err := s.crud.DeleteMessage(ctx, req.GetMid())
	if err != nil {
		if errors.Is(err, storage.ErrMessageNotExist) {
			return nil, status.Error(codes.PermissionDenied, "message not found")
		}
		return nil, status.Error(codes.Unauthenticated, "failed to delete message")
	}
	return &crudv1.DeleteMessageResponse{Status: answer}, nil
}

func (s *serverCRUD) GetMessage(ctx context.Context, req *crudv1.GetMessageRequest) (*crudv1.GetMessageResponse, error) {
	TokenResponse := jwtVal.ValidateToken(req.GetToken(), s.Secret)
	if TokenResponse.Error != nil {
		return nil, status.Error(codes.Unauthenticated, "failed in decoding token")
	}

	message, err := s.crud.GetMessage(ctx, req.GetMid())
	if err != nil {
		if errors.Is(err, storage.ErrMessageNotExist) {
			return nil, status.Error(codes.PermissionDenied, "message not found")
		}
		return nil, status.Error(codes.Unauthenticated, "failed to get message")
	}
	return &crudv1.GetMessageResponse{Id: message.ID, Content: message.Content, Type: message.Type, Uid: message.UserID}, nil
}

func (s *serverCRUD) UpdateMessage(ctx context.Context, req *crudv1.UpdateMessageRequest) (*crudv1.UpdateMessageResponse, error) {
	TokenResponse := jwtVal.ValidateToken(req.GetToken(), s.Secret)
	if TokenResponse.Error != nil {
		return nil, status.Error(codes.Unauthenticated, "failed in decoding token")
	}

	answer, err := s.crud.UpdateMessage(ctx, req.GetMid(), req.GetNewContent())
	if err != nil {
		if errors.Is(err, storage.ErrMessageNotExist) {
			return nil, status.Error(codes.PermissionDenied, "message not found")
		}
		return nil, status.Error(codes.Unauthenticated, "failed to update message")
	}
	return &crudv1.UpdateMessageResponse{Status: answer}, nil
}

func (s *serverCRUD) ShowMessages(ctx context.Context, req *crudv1.ShowMessagesRequest) (*crudv1.ShowMessagesResponse, error) {
	tokenResponse := jwtVal.ValidateToken(req.GetToken(), s.Secret)
	if tokenResponse.Error != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid authentication token")
	}

	messages, err := s.crud.ShowAllMessages(ctx, tokenResponse.UserID)
	if err != nil {
		if errors.Is(err, storage.ErrNoMessagesFound) {
			return &crudv1.ShowMessagesResponse{}, nil
		}
		return nil, status.Error(codes.Internal, "failed to retrieve messages")
	}

	var pbMessages []*crudv1.GetMessageResponse
	for _, msg := range messages {
		pbMessages = append(pbMessages, &crudv1.GetMessageResponse{
			Id:       msg.ID,
			Content:  msg.Content,
			Uid:      msg.UserID,
			Type:     msg.Type,
			Datetime: msg.DateTime,
		})
	}

	return &crudv1.ShowMessagesResponse{
		Message: pbMessages,
	}, nil
}
