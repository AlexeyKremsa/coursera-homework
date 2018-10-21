package main

import (
	"time"
)

func (s *service) Logging(nothing *Nothing, srv Admin_LoggingServer) error {

	listener := listener{
		logsCh:  make(chan *logMsg),
		closeCh: make(chan struct{}),
	}
	s.addListener(&listener)

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

func (s *service) Statistics(interval *StatInterval, srv Admin_StatisticsServer) error {

	closeCh := make(chan struct{})
	s.addStatCloseCh(closeCh)

	ticker := time.NewTicker(time.Second * time.Duration(interval.IntervalSeconds))

	for {
		select {
		case <-ticker.C:
			c := make(map[string]uint64)
			m := make(map[string]uint64)

			for k := range s.stat.consumers {
				c[k] = s.stat.consumers[k]
			}

			for k := range s.stat.methods {
				m[k] = s.stat.methods[k]
			}
			statEvent := &Stat{
				Timestamp:  0,
				ByMethod:   m,
				ByConsumer: c,
			}

			srv.Send(statEvent)

		case <-closeCh:
			return nil
		}
	}

	return nil
}
