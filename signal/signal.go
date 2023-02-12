package sig

import (
	"os"
	"os/signal"
	"syscall"
)

const signalChannelSize = 2

func Wait() {
	sigint := make(chan os.Signal, signalChannelSize)
	defer close(sigint)

	signal.Notify(sigint,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGHUP)

	<-sigint

	signal.Stop(sigint)
}
