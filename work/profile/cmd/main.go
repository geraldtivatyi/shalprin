package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/oligoden/chassis/adapter"
	"github.com/oligoden/chassis/storage/gosql"

	//xxx
	"github.com/geraldtivatyi/shalprin/work/profile"
	//end
	//+++
	//"github.com/geraldtivatyi/shalprin/src/profile"
	//end
)

const (
	dbt = "mysql"
	uri = "test:password@tcp(shalprin-db:3306)/test?charset=utf8&parseTime=True&loc=Local"
)

func main() {
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
