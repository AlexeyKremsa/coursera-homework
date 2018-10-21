package main

import (
	"encoding/json"
	"strings"

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

func getConsumerNameFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", grpc.Errorf(codes.Unauthenticated, "can not get metadata")
	}
	consumer, ok := md["consumer"]
	if !ok || len(consumer) != 1 {
		return "", grpc.Errorf(codes.Unauthenticated, "can not get metadata")
	}

	return consumer[0], nil
}

func (srv *service) checkBizPermission(consumer, method string) error {
	allowedMethods, ok := srv.aclStorage[consumer]
	if !ok {
		return grpc.Errorf(codes.Unauthenticated, "permission denied")
	}

	for _, m := range allowedMethods {
		//check if everything allowed
		splitted := strings.Split(m, "/")
		if len(splitted) == 3 && splitted[2] == "*" {
			return nil
		}

		if m == method {
			return nil
		}
	}

	return grpc.Errorf(codes.Unauthenticated, "permission denied")
}

func parseACL(acl string) (map[string][]string, error) {
	var aclParsed map[string]*json.RawMessage
	result := make(map[string][]string)

	err := json.Unmarshal([]byte(acl), &aclParsed)
	if err != nil {
		return nil, err
	}

	for k, v := range aclParsed {
		var val []string
		err := json.Unmarshal(*v, &val)
		if err != nil {
			return nil, err
		}

		result[k] = val
	}

	return result, nil
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
			for _, l := range srv.listeners {
				l.logsCh <- log
			}

		case <-srv.closeListenersCh:
			for _, l := range srv.listeners {
				l.closeCh <- struct{}{}
			}

			return
		}
	}
}

func (srv *service) addMethodStat(name string) {
	srv.m.Lock()
	res, ok := srv.stat.method[name]
	if !ok {
		srv.stat.method[name] = 0
	} else {
		res++
		srv.stat.method[name] = res
	}
}

func (srv *service) addConsumerStat(name string) {
	srv.m.Lock()
	res, ok := srv.stat.consumer[name]
	if !ok {
		srv.stat.consumer[name] = 0
	} else {
		res++
		srv.stat.consumer[name] = res
	}
}

func (srv *service) addStatCloseCh(closeCh chan struct{}) {
	srv.m.Lock()
	srv.stat.consumersClose = append(srv.stat.consumersClose, closeCh)
	srv.m.Unlock()
}
