package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/superfly/fsm"

	log "github.com/sirupsen/logrus"
)

const FLYD_SKETCH_SOCKET = "flyd-sketch.sock"

func main() {
	initConfig()
	initStateDb()
	initS3Client()

	db := initStateDb()
	defer db.Close()

	cfg := fsm.Config{
		Logger: log.WithFields(log.Fields{}),
		DBPath: CFG.StateDir,
		Queues: map[string]int{"main": 5},
	}
	manager, err := fsm.New(cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	handler := NewStartHandler(manager)

	mux := http.NewServeMux()
	mux.Handle("start/", handler)
	socket := filepath.Join(CFG.StateDir, FLYD_SKETCH_SOCKET)
	os.Remove(socket)
	unixListener, _ := net.Listen("unix", socket)
	defer os.Remove(socket)
	http.Serve(unixListener, mux)
	return
}
