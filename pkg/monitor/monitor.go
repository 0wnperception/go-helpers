package monitor

type MonitorObserver interface {
	OnMonitorError(err error)
}

type MonitorInterface interface {
	StartHandle(o MonitorObserver) error
	StopHandle() error
}
