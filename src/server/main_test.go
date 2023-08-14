package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
)

type killServer struct {
	server http.Server
	cancel context.CancelFunc
}

func newKillServer(addr string, cancel context.CancelFunc) *killServer {
	return &killServer{
		server: http.Server{
			Addr: addr,
		},
		cancel: cancel,
	}
}

func (s *killServer) Start() {
	s.server.Handler = s

	err := s.server.ListenAndServe()
	if err != nil {
		fmt.Println("KillServer Error:", err)
	}
}

func (s *killServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	// cancel the context
	s.cancel()
}

func runService(ctx context.Context) {

}

func TestRun(t *testing.T) {
	channel := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())

	killPort := os.Getenv("SCAFFOLD_KILL_SERVER_PORT")

	killServer := newKillServer(fmt.Sprintf(":%s", killPort), cancel)
	go killServer.Start()
	go run(ctx, channel)

	<-ctx.Done()

	killServer.server.Shutdown(context.Background())
}
