package server

import (
	"context"

	walletv1 "wallet-core/api/gen/wallet/v1"
	"wallet-core/internal/service/wallet"
)

// WalletGRPCServer 实现 wallet.v1.WalletServiceServer 接口
type WalletGRPCServer struct {
	walletv1.UnimplementedWalletServiceServer
	svc *wallet.Service
}

func NewWalletGRPCServer(svc *wallet.Service) *WalletGRPCServer {
	return &WalletGRPCServer{svc: svc}
}

func (s *WalletGRPCServer) CreateAddress(ctx context.Context, req *walletv1.CreateAddressRequest) (*walletv1.CreateAddressResponse, error) {
	addr, err := s.svc.CreateAddress(ctx, req.UserId, req.Currency)
	if err != nil {
		return nil, err
	}

	return &walletv1.CreateAddressResponse{
		Address: addr,
	}, nil
}

func (s *WalletGRPCServer) GetBalance(ctx context.Context, req *walletv1.GetBalanceRequest) (*walletv1.GetBalanceResponse, error) {
	balances, err := s.svc.GetBalance(ctx, req.UserId, req.Currency)
	if err != nil {
		return nil, err
	}

	return &walletv1.GetBalanceResponse{
		Balances: balances,
	}, nil
}

func (s *WalletGRPCServer) CreateWithdrawal(ctx context.Context, req *walletv1.CreateWithdrawalRequest) (*walletv1.CreateWithdrawalResponse, error) {
	id, err := s.svc.CreateWithdrawal(ctx, req.UserId, req.ToAddress, req.Amount, req.Currency)
	if err != nil {
		return nil, err
	}

	return &walletv1.CreateWithdrawalResponse{
		WithdrawalId: id,
		Status:       "pending_review",
	}, nil
}
