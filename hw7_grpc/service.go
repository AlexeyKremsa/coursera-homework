package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"google.golang.org/grpc"
)

var aclStorage map[string]json.RawMessage

type service struct {
	m                *sync.RWMutex
	incomingLogsCh   chan *logMsg
	closeListenersCh chan struct{}
	listeners        []*listener
	aclStorage       map[string][]string
	stat             stat
}

type stat struct {
	methods        map[string]uint64
	consumers      map[string]uint64
	consumersClose []chan struct{}
}

type listener struct {
	logsCh  chan *logMsg
	closeCh chan struct{}
}

type logMsg struct {
	methodName   string
	consumerName string
}

func StartMyMicroservice(ctx context.Context, addr, acl string) error {
	aclParsed, err := parseACL(acl)
	if err != nil {
		return err
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic(fmt.Sprintf("can not start the service. %s", err.Error()))
	}

	service := &service{
		m:                &sync.RWMutex{},
		incomingLogsCh:   make(chan *logMsg, 0),
		listeners:        make([]*listener, 0),
		aclStorage:       aclParsed,
		closeListenersCh: make(chan struct{}),
		stat: stat{
			methods:        make(map[string]uint64),
			consumers:      make(map[string]uint64),
			consumersClose: make([]chan struct{}, 0),
		},
	}

	go service.logsSender()

	opts := []grpc.ServerOption{grpc.UnaryInterceptor(service.unaryInterceptor),
		grpc.StreamInterceptor(service.streamInterceptor)}

	srv := grpc.NewServer(opts...)
	fmt.Println("starting server at: ", addr)

	RegisterBizServer(srv, service)
	RegisterAdminServer(srv, service)

	go func() {
		select {
		case <-ctx.Done():
			service.closeListenersCh <- struct{}{}
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

func (s *service) unaryInterceptor(ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	consumer, err := getConsumerNameFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = s.checkBizPermission(consumer, info.FullMethod)
	if err != nil {
		return nil, err
	}

	msg := logMsg{
		consumerName: consumer,
		methodName:   info.FullMethod,
	}

	s.incomingLogsCh <- &msg

	h, err := handler(ctx, req)
	return h, err
}

func (s *service) streamInterceptor(srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler) error {

	consumer, err := getConsumerNameFromContext(ss.Context())
	if err != nil {
		return err
	}

	err = s.checkBizPermission(consumer, "/main.Admin/Logging")
	if err != nil {
		return err
	}

	msg := logMsg{
		consumerName: consumer,
		methodName:   info.FullMethod,
	}

	s.m.Lock()
	for _, l := range s.listeners {
		l.logsCh <- &msg
	}
	s.m.Unlock()

	return handler(srv, ss)
}
