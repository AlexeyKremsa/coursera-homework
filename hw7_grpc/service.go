package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"google.golang.org/grpc"
)

var aclStorage map[string]json.RawMessage

func StartMyMicroservice(ctx context.Context, addr, acl string) error {
	err := json.Unmarshal([]byte(acl), &aclStorage)
	if err != nil {
		return err
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic(fmt.Sprintf("can not start the service. %s", err.Error()))
	}

	srv := grpc.NewServer()
	fmt.Println("starting server at: ", addr)

	RegisterBizServer(srv, newBizHandler())
	RegisterAdminServer(srv, newAdminHandler())

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
