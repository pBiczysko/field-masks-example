package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	// Static files

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	pbExample "github.com/pBiczysko/field-masks-example/proto"
	"github.com/pBiczysko/field-masks-example/server"
)

func main() {
	// Adds gRPC internal logs. This is quite verbose, so adjust as desired!
	log := grpclog.NewLoggerV2(os.Stdout, ioutil.Discard, ioutil.Discard)
	grpclog.SetLoggerV2(log)

	addr := "0.0.0.0:10000"
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	s := grpc.NewServer()
	pbExample.RegisterUserServiceServer(s, server.New())

	// Serve gRPC Server
	log.Info("Serving gRPC on http://", addr)
	go func() {
		log.Fatal(s.Serve(lis))
	}()

	// See https://github.com/grpc/grpc/blob/master/doc/naming.md
	// for gRPC naming standard information.
	dialAddr := fmt.Sprintf("dns:///%s", addr)
	// Create a client connection to the gRPC Server we just started.
	// This is where the gRPC-Gateway proxies the requests.
	conn, err := grpc.DialContext(
		context.Background(),
		dialAddr,
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	jsonpb := &runtime.JSONPb{
		EmitDefaults: true,
		Indent:       "  ",
		OrigName:     true,
	}
	gwmux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, jsonpb),
	)
	err = pbExample.RegisterUserServiceHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "11000"
	}
	gatewayAddr := "0.0.0.0:" + port
	mux := http.NewServeMux()
	mux.HandleFunc("/", defaultHandler)

	gwServer := &http.Server{
		Addr: gatewayAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api") {
				gwmux.ServeHTTP(w, r)
				return
			}
			mux.ServeHTTP(w, r)
		}),
	}

	log.Info("Serving gRPC-Gateway on http://", gatewayAddr)
	log.Fatalln(gwServer.ListenAndServe())
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello")
}
