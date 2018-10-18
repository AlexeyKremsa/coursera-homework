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
}

type listener struct {
	logsCh  chan *logMsg
	closeCh chan struct{}
}

type logMsg struct {
	methodName   string
	consumerName string
}

func (srv *service) addListener(l *listener) {
	srv.m.Lock()
	srv.listeners = append(srv.listeners, l)
	srv.m.Unlock()
}

func (srv *service) logsSender() {
	for {
		select {
		case log := <-srv.incomingLogsCh:
			fmt.Println("Got: ", log)
			for _, l := range srv.listeners {
				l.logsCh <- log
			}

		case <-srv.closeListenersCh:
			fmt.Println("CLOSE LISTENERS")
			for _, l := range srv.listeners {
				l.closeCh <- struct{}{}
			}

			return
		}
	}
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
			fmt.Println("CLOSE SERVER")
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
	fmt.Println("OLO")
	return handler(srv, ss)
}
