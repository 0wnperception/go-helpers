package pprof_wrapper

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/pprof"
	"robot_agent/pkg/monitor"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap/zapcore"
)

type Profile struct {
	cfg      *ProfileConfig
	server   *http.Server
	listener net.Listener
	observer monitor.MonitorObserver
	done     chan struct{}
}

type ProfileConfig struct {
	ProfilePort             string
	ProfileHost             string
	GracefulShutdownTimeout time.Duration
	LogLevel                zapcore.Level
}

func NewProfile(cfg *ProfileConfig) (*Profile, error) {
	router := mux.NewRouter()
	router.HandleFunc("/pprof/", pprof.Index)
	router.HandleFunc("/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/pprof/profile", pprof.Profile)
	router.HandleFunc("/pprof/symbol", pprof.Symbol)
	router.HandleFunc("/pprof/trace", pprof.Trace)
	router.Handle("/pprof/mutex", pprof.Handler("mutex"))
	router.Handle("/pprof/block", pprof.Handler("block"))
	router.Handle("/pprof/heap", pprof.Handler("heap"))
	router.Handle("/pprof/goroutine", pprof.Handler("goroutine"))
	router.Handle("/pprof/threadcreate", pprof.Handler("threadcreate"))
	s := &http.Server{Handler: router}
	listener, err := net.Listen("tcp", cfg.ProfileHost+":"+cfg.ProfilePort)
	if err != nil {
		return nil, err
	}
	prof := &Profile{
		server:   s,
		listener: listener,
		cfg:      cfg,
		done:     make(chan struct{}, 1),
	}
	return prof, nil
}

func (p *Profile) StartHandle(o monitor.MonitorObserver) error {
	p.observer = o
	go p.serve()
	return nil
}

func (p *Profile) StopHandle() error {
	ctx, cancel := context.WithTimeout(context.Background(), p.cfg.GracefulShutdownTimeout)
	defer cancel()
	err := p.server.Shutdown(ctx)
	if err != nil {
		return err
	}
	<-p.done
	return nil
}

func (p *Profile) serve() {
	err := p.server.Serve(p.listener)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		p.observer.OnMonitorError(err)
	}
	close(p.done)
}
