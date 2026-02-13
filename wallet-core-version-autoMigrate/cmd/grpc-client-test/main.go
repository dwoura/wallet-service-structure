package main

import (
	"context"
	"log"
	"time"

	pb "wallet-core/api/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	addr := "127.0.0.1:50051"
	// Set up a connection to the server.
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewAddressServiceClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test 1: Get BTC Address
	r, err := c.GetAddress(ctx, &pb.GetAddressRequest{UserId: 1, Currency: "BTC"})
	if err != nil {
		log.Fatalf("could not get BTC address: %v", err)
	}
	log.Printf("BTC Address: %s (Path: m/44'/0'/0'/0/%d)", r.Address, r.PathIndex)

	// Test 2: Get ETH Address
	r2, err := c.GetAddress(ctx, &pb.GetAddressRequest{UserId: 1, Currency: "ETH"})
	if err != nil {
		log.Fatalf("could not get ETH address: %v", err)
	}
	log.Printf("ETH Address: %s (Path: m/44'/60'/0'/0/%d)", r2.Address, r2.PathIndex)
}
