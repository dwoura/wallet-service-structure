package server

import (
	"google.golang.org/grpc"

	"wallet-core/internal/server/routes"
	"wallet-core/internal/service"
)

// NewGRPCServer 初始化并注册 gRPC 服务
func NewGRPCServer(addressService service.AddressService) *grpc.Server {
	s := grpc.NewServer()

	// 注册 AddressService
	// 注册 gRPC 服务
	routes.RegisterAddressGRPC(s, addressService)

	return s
}
