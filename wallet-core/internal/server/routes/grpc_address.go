package routes

import (
	"google.golang.org/grpc"

	"wallet-core/api/pb"
	handler_grpc "wallet-core/internal/handler/grpc"
	"wallet-core/internal/service"
)

// RegisterAddressGRPC 注册 AddressService gRPC 服务
func RegisterAddressGRPC(s *grpc.Server, addressService service.AddressService) {
	pb.RegisterAddressServiceServer(s, handler_grpc.NewAddressHandler(addressService))
}
