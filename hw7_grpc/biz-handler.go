package main

import (
	context "golang.org/x/net/context"
)

type bizHandler struct {
}

func newBizHandler() *bizHandler {
	return &bizHandler{}
}

func (b *bizHandler) Check(ctx context.Context, n *Nothing) (*Nothing, error) {
	err := checkBizPermission(ctx, bizAdmin, bizUser)
	return &Nothing{}, err
}

func (b *bizHandler) Add(ctx context.Context, n *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (b *bizHandler) Test(ctx context.Context, n *Nothing) (*Nothing, error) {
	err := checkBizPermission(ctx, bizAdmin)
	return &Nothing{}, err
}
