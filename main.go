package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/openebs/scope-plugin/plugin"
	log "github.com/sirupsen/logrus"
)

func setupSocket(socketPath string) (net.Listener, error) {
	os.RemoveAll(filepath.Dir(socketPath))
	if err := os.MkdirAll(filepath.Dir(socketPath), 0700); err != nil {
		return nil, fmt.Errorf("failed to create directory %q: %v", filepath.Dir(socketPath), err)
	}
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %q: %v", socketPath, err)
	}
	log.Printf("Listening on: unix://%s", socketPath)
	return listener, nil
}

func setupSignals(socketPath string) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-interrupt
		os.RemoveAll(filepath.Dir(socketPath))
		os.Exit(0)
	}()
}

func main() {

	// Put socket in sub-directory to have more control on permissions.
	const socketPath = "/var/run/scope/plugins/openebs/openebs.sock"

	// Handle the exit signal
	setupSignals(socketPath)
	listener, err := setupSocket(socketPath)
	if err != nil {
		log.Error(err)
	}

	defer func() {
		listener.Close()
		os.RemoveAll(filepath.Dir(socketPath))
	}()

	plugin := &plugin.Plugin{
		HostID: "host",
	}

	go plugin.UpdateMetrics()

	http.HandleFunc("/report", plugin.Report)
	if err := http.Serve(listener, nil); err != nil {
		log.Errorf("error: %v", err)
	}
}
