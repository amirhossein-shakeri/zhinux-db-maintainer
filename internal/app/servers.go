package app

import (
	"errors"
	"fmt"
	"net/http"

	grpcx "github.com/amirhossein-shakeri/zhinux-platform/grpc"
	grpcinterceptors "github.com/amirhossein-shakeri/zhinux-platform/grpc/interceptors"
	gogrpc "google.golang.org/grpc"
)

func (a *App) initServers() error {
	if err := a.initHTTPServer(); err != nil {
		return err
	}

	if err := a.initGRPCServer(); err != nil {
		return err
	}

	if a.httpServer == nil && a.grpcServer == nil {
		return errors.New("both HTTP and gRPC servers are disabled")
	}

	return nil
}

func (a *App) initHTTPServer() error {
	if !a.cfg.Runtime.HTTP.Enabled {
		return nil
	}

	mux := http.NewServeMux()
	healthHandler := a.health.Handler()
	mux.Handle("/livez", healthHandler)
	mux.Handle("/readyz", healthHandler)
	mux.Handle("/healthz", healthHandler)
	mux.Handle("/varz", healthHandler)
	mux.Handle("/flagz", healthHandler)
	mux.Handle("/statusz", healthHandler)
	mux.Handle("/configz", healthHandler)
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte("zhinux-db-maintainer"))
	})

	a.httpServer = &http.Server{
		Addr:              a.cfg.Base.HTTPListenAddr,
		Handler:           mux,
		ReadTimeout:       a.cfg.Runtime.HTTP.ReadTimeout,
		ReadHeaderTimeout: a.cfg.Runtime.HTTP.ReadHeaderTimeout,
		WriteTimeout:      a.cfg.Runtime.HTTP.WriteTimeout,
		IdleTimeout:       a.cfg.Runtime.HTTP.IdleTimeout,
	}

	return nil
}

func (a *App) initGRPCServer() error {
	if !a.cfg.Runtime.GRPC.Enabled {
		return nil
	}

	unary := []gogrpc.UnaryServerInterceptor{
		grpcinterceptors.UnaryRecovery(),
		grpcinterceptors.UnaryRequestID(),
		grpcinterceptors.UnaryTimeout(a.cfg.Base.DefaultRPCTimeout),
		grpcinterceptors.UnaryLogging(a.logger),
	}
	stream := []gogrpc.StreamServerInterceptor{
		grpcinterceptors.StreamRecovery(),
		grpcinterceptors.StreamRequestID(),
		grpcinterceptors.StreamLogging(a.logger),
	}

	serverOptions := grpcinterceptors.ChainServerOptions(unary, stream)
	server, err := grpcx.NewRuntime(grpcx.ServerConfig{
		ListenAddress:   a.cfg.Base.GRPCListenAddr,
		DrainTimeout:    a.cfg.Runtime.GRPC.DrainTimeout,
		MaxRecvMsgBytes: a.cfg.Runtime.GRPC.MaxRecvMsgBytes,
		MaxSendMsgBytes: a.cfg.Runtime.GRPC.MaxSendMsgBytes,
	}, serverOptions...)
	if err != nil {
		return fmt.Errorf("init grpc runtime: %w", err)
	}

	// TODO: register gRPC service implementations here once contracts/use-cases are ready.
	// Example: maintainerpb.RegisterDatabaseMaintainerServer(server.Server(), handler)
	a.grpcServer = server
	return nil
}
