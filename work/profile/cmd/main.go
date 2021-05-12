package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/oligoden/chassis/adapter"
	"github.com/oligoden/chassis/storage/gosql"
	"google.golang.org/grpc"

	//xxx
	"github.com/geraldtivatyi/shalprin/work/profile"
	//end
	//+++
	//"github.com/geraldtivatyi/shalprin/src/profile"
	//end
)

func main() {
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbAddr := os.Getenv("DB_ADDRESS")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	params := "charset=utf8&parseTime=True&loc=Local"
	format := "%s:%s@tcp(%s:%s)/%s?%s"

	if dbPort == "" {
		dbPort = "3306"
	}
	uri := fmt.Sprintf(format, dbUser, dbPass, dbAddr, dbPort, dbName, params)
	dbt := "mysql"

	store := gosql.New(dbt, uri)
	if store.Err() != nil {
		log.Fatal(store.Err())
	}

	dProfile := profile.NewDevice(store)
	store.Migrate(profile.NewRecord())

	mux := http.NewServeMux()
	mux.Handle("/api/v1/profile", adapter.MNA().Put(dProfile.Update()).Get(dProfile.Read()).Post(dProfile.Create()).Notify().Entry())
	mux.Handle("/api/v1/profile/", adapter.MNA().Put(dProfile.Update()).Get(dProfile.Read()).Post(dProfile.Create()).Notify().Entry())

	// mux.Handle("/api/v1/profiles/address/predict", adapter.MNA().Get(dProfile.PredictAddress()).Notify().Entry())
	mux.Handle("/api/v1/", NF().Notify().Entry())

	httpServer := &http.Server{
		Addr:           ":9000",
		Handler:        mux,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	//protoc -I . server.proto --go_out=plugins=grpc:.
	//go install github.com/golang/protobuf/protoc-gen-go

	grpcServerError := make(chan error)
	grpcServer := grpc.NewServer()
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 9001))
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		profile.RegisterReadProfileServer(grpcServer, dProfile)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %s", err)
		}
	}()

	httpServerError := make(chan error)
	go func() {
		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			httpServerError <- err
			return
		}
		fmt.Println("http server shutdown")
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	select {
	case err := <-grpcServerError:
		fmt.Println("grpc server error", err)
		shutdown(httpServer)
		time.Sleep(100 * time.Millisecond)
		os.Exit(1)
	case err := <-httpServerError:
		fmt.Println("http server error", err)
		grpcServer.GracefulStop()
		time.Sleep(100 * time.Millisecond)
		os.Exit(1)
	case sig := <-quit:
		fmt.Println("\ngot signal", sig)
	}

	grpcServer.GracefulStop()
	shutdown(httpServer)
	time.Sleep(100 * time.Millisecond)
}

func shutdown(s *http.Server) {
	ctxServer, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := s.Shutdown(ctxServer)
	if err != nil && err != http.ErrServerClosed {
		fmt.Println("https server shutdown error", err)
	}
}

func NF() adapter.Adapter {
	return adapter.Adapter{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}),
	}
}
