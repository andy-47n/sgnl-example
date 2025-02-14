package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/andy-47n/sgnl-example/pkg/adapter"
	api_adapter_v1 "github.com/sgnl-ai/adapter-framework/api/adapter/v1"
	"github.com/sgnl-ai/adapter-framework/server"
	"google.golang.org/grpc"
)

var (
	// Port is the port at which the gRPC server will listen.
	Port = flag.Int("port", 8080, "The server port")

	// Timeout is the timeout for the HTTP client used to make requests to the datasource (seconds).
	Timeout = flag.Int("timeout", 30, "The timeout for the HTTP client used to make requests to the datasource (seconds)")
)

func main() {
	flag.Parse()

	logger := log.New(os.Stdout, "adapter", log.Lmicroseconds|log.LUTC|log.Lshortfile)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *Port))
	if err != nil {
		logger.Fatalf("Failed to open server port: %v", err)
	}

	s := grpc.NewServer()

	stop := make(chan struct{})

	adapterServer := server.New(stop)

	err = server.RegisterAdapter(adapterServer, "Test-1.0.0", adapter.NewAdapter(adapter.NewClient(*Timeout)))
	if err != nil {
		logger.Fatalf("Failed to register adapter: %v", err)
	}

	api_adapter_v1.RegisterAdapterServer(s, adapterServer)

	logger.Printf("Started adapter gRPC server on port %d", *Port)

	if err := s.Serve(listener); err != nil {
		logger.Fatalf("Failed to listen on server port: %v", err)
	}
}
