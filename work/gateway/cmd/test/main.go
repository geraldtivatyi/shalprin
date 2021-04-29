package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/oligoden/chassis/adapter"
	"github.com/oligoden/chassis/storage/gosql"

	//---
	"github.com/geraldtivatyi/shalprin/work/gateway"
	"github.com/geraldtivatyi/shalprin/work/gateway/session"
	//end
	//+++
	//"github.com/geraldtivatyi/shalprin/src/gateway"
	//"github.com/geraldtivatyi/shalprin/src/gateway/session"
	//end
)

func serveFile(f string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, f)
	})
}

func serveFiles(p, d string) http.Handler {
	return http.StripPrefix(p, http.FileServer(http.Dir(d)))
}

func main() {
	hIndex := gateway.NewIndex()
	profile, _ := url.Parse("http://profile:9000/")
	hIndex.SetProxy("profiles", httputil.NewSingleHostReverseProxy(profile))

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
		log.Fatal(fmt.Errorf("starting store, %w", store.Err()))
	}

	dSession := session.NewDevice(store)
	store.Migrate(session.NewRecord())
	store.Migrate(session.NewSessionUsersRecord())

	mux := http.NewServeMux()
	mux.Handle("/", adapter.Core(serveFile("static")).Notify().Entry())
	mux.Handle("/static/", adapter.Core(serveFiles("/static/", "static")).Entry())

	mwProfile := adapter.Core(hIndex).And(dSession.CreateUser())
	mux.Handle("/api/v1/profiles", mwProfile.And(dSession.Validate()).Notify().Entry())

	mwSession := adapter.MNA().Delete(dSession.Signout()).Post(dSession.Signin())
	mux.Handle("/api/v1/sessions", mwSession.And(dSession.Validate()).Notify().Entry())

	mux.Handle("/api/v1/signin", mwSession.And(dSession.Validate()).Notify().Entry())

	mux.Handle("/api/v1/", adapter.Core(hIndex).And(dSession.Validate()).Notify().Entry())

	httpServer := &http.Server{
		Addr:           ":9000",
		Handler:        mux,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	serverError := make(chan error)
	go func() {
		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			serverError <- err
			return
		}
		fmt.Println("http server shutdown")
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	select {
	case err := <-serverError:
		fmt.Println("http server error", err)
		time.Sleep(100 * time.Millisecond)
		os.Exit(1)
	case sig := <-quit:
		fmt.Println("\ngot signal", sig)
	}

	ctxHTTPServer, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := httpServer.Shutdown(ctxHTTPServer)
	if err != nil && err != http.ErrServerClosed {
		fmt.Println("http server shutdown error", err)
	}

	time.Sleep(100 * time.Millisecond)
}
