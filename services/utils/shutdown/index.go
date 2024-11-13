package shutdown

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type Manager struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func New() *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	m := &Manager{
		ctx:    ctx,
		cancel: cancel,
	}

	go m.handleSignals()

	return m
}

func (m *Manager) handleSignals() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	<-signalChan
	m.cancel()
}

func (m *Manager) Wait() {
	<-m.ctx.Done()
}

func (m *Manager) Shutdown(format string, v ...any) {
	if format != "" {
		log.Println(format, v)
	}

	m.cancel()
}
