package main

type adminHandler struct {
}

func newAdminHandler() *adminHandler {
	return &adminHandler{}
}

func (a *adminHandler) Logging(nothing *Nothing, srv Admin_LoggingServer) error {
	err := checkBizPermission(srv.Context(), bizAdmin, logger)
	if err != nil {
		return err
	}

	methodName := "/main.Admin/Logging"
	event := Event{
		Host:     "",
		Method:   methodName,
		Consumer: "test consumer",
	}
	event.ProtoMessage()
	srv.Send(&event)

	return nil
}

func (a adminHandler) Statistics(*StatInterval, Admin_StatisticsServer) error {
	return nil
}
