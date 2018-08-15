package main

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
)

func StartMyMicroservice(ctx context.Context, addr, acl string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic(fmt.Sprintf("can not start the service. %s", err.Error()))
	}

	srv := grpc.NewServer()
	fmt.Println("starting server at: ", addr)

	go func() {
		select {
		case <-ctx.Done():
			srv.Stop()
			return
		}
	}()

	go func() {
		err := srv.Serve(lis)
		if err != nil {
			panic(err)
		}
		return
	}()

	return nil
}
