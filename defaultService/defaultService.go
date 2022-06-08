package defaultService

import "context"

type DefaultService struct {
	Ctx    context.Context
	Cancel context.CancelFunc
	Done   chan error
}

type DefaultServiceIface interface {
	Run() (done <-chan error)
	Stop() error
}
