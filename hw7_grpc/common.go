package main

import (
	"fmt"

	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

const (
	bizAdmin = "biz_admin"
	bizUser  = "biz_user"
	logger   = "logger"
)

func checkBizPermission(ctx context.Context, rolesAllowed ...string) error {
	var found bool
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return grpc.Errorf(codes.Unauthenticated, "can not get metadata")
	}

	for _, role := range rolesAllowed {
		val, ok := md["consumer"]
		fmt.Print(role)
		if !ok || len(val) != 1 || val[0] != role {
			continue
		} else {
			found = true
			break
		}
	}

	if !found {
		return grpc.Errorf(codes.Unauthenticated, "permission denied")
	}

	return nil
}
