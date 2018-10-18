package main

func (s *service) Logging(nothing *Nothing, srv Admin_LoggingServer) error {
	consumer, err := getConsumerNameFromContext(srv.Context())
	if err != nil {
		return err
	}

	err = s.checkBizPermission(consumer, "/main.Admin/Logging")
	if err != nil {
		return err
	}

	listener := listener{
		logsCh:  make(chan *logMsg),
		closeCh: make(chan struct{}),
	}
	s.addListener(&listener)

	current := &Event{
		Consumer: consumer,
		Method:   "/main.Admin/Logging",
		Host:     "127.0.0.1:8083",
	}

	srv.Send(current)

	for {
		select {
		case logMsg := <-listener.logsCh:
			event := &Event{
				Consumer: logMsg.consumerName,
				Method:   logMsg.methodName,
				Host:     "127.0.0.1:8083",
			}
			srv.Send(event)

		case <-listener.closeCh:
			return nil
		}
	}
}

func (s *service) Statistics(*StatInterval, Admin_StatisticsServer) error {
	return nil
}
