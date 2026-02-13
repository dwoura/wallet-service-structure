package grpc

import (
	"context"
	"log"

	pb "wallet-core/api/pb"
	"wallet-core/internal/service"
)

// AddressHandler implements pb.AddressServiceServer
type AddressHandler struct {
	pb.UnimplementedAddressServiceServer
	service service.AddressService
}

// NewAddressHandler creates a new gRPC handler
func NewAddressHandler(svc service.AddressService) *AddressHandler {
	return &AddressHandler{
		service: svc,
	}
}

// GetAddress handles the GetAddress gRPC request
func (h *AddressHandler) GetAddress(ctx context.Context, req *pb.GetAddressRequest) (*pb.GetAddressResponse, error) {
	log.Printf("[gRPC] GetAddress request for UserID: %d, Currency: %s", req.UserId, req.Currency)

	address, index, err := h.service.GetDepositAddress(uint64(req.UserId), req.Currency)
	if err != nil {
		log.Printf("[gRPC] Failed to get address: %v", err)
		return nil, err
	}

	return &pb.GetAddressResponse{
		Address:   address,
		PathIndex: uint32(index),
		Currency:  req.Currency,
	}, nil
}
