package server

import (
	"context"

	userv1 "wallet-core/api/gen/user/v1"
	"wallet-core/internal/service/user"
)

// UserGRPCServer 实现 user.v1.UserServiceServer 接口
type UserGRPCServer struct {
	userv1.UnimplementedUserServiceServer
	svc *user.Service
}

func NewUserGRPCServer(svc *user.Service) *UserGRPCServer {
	return &UserGRPCServer{svc: svc}
}

func (s *UserGRPCServer) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
	userId, err := s.svc.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	return &userv1.RegisterResponse{
		UserId: userId,
	}, nil
}

func (s *UserGRPCServer) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	token, userId, username, err := s.svc.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	return &userv1.LoginResponse{
		Token:    token,
		UserId:   userId,
		Username: username,
	}, nil
}

func (s *UserGRPCServer) GetUserInfo(ctx context.Context, req *userv1.GetUserInfoRequest) (*userv1.GetUserInfoResponse, error) {
	u, err := s.svc.GetUserInfo(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	return &userv1.GetUserInfoResponse{
		UserId:   int64(u.ID),
		Username: u.Username,
		Email:    u.Email,
	}, nil
}
