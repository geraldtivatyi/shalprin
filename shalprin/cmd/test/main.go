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

	//---
	"github.com/geraldtivatyi/shalprin/shalprin/cleaning"
	"github.com/geraldtivatyi/shalprin/shalprin/electrician"
	"github.com/geraldtivatyi/shalprin/shalprin/gardening"
	"github.com/geraldtivatyi/shalprin/shalprin/plumbing"
	"github.com/geraldtivatyi/shalprin/shalprin/profile"
	"github.com/geraldtivatyi/shalprin/shalprin/session"
	//end
	//+++
	//"github.com/geraldtivatyi/shalprin/work/shalprin/profile"
	//"github.com/geraldtivatyi/shalprin/work/shalprin/session"
	//"github.com/geraldtivatyi/shalprin/work/shalprin/cleaning"
	//"github.com/geraldtivatyi/shalprin/work/shalprin/gardening"
	//"github.com/geraldtivatyi/shalprin/work/shalprin/electrician"
	//"github.com/geraldtivatyi/shalprin/work/shalprin/plumbing"
	//end
)

const (
	dbt = "mysql"
	uri = "test:password@tcp(shalprin-db:3306)/test?charset=utf8&parseTime=True&loc=Local"
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
	// tmpl := template.Must(template.ParseFiles("signup.html"))

	store := gosql.New(dbt, uri)
	if store.Err() != nil {
		log.Fatal(store.Err())
	}

	dProfile := profile.NewDevice(store)
	store.Migrate(profile.NewRecord())

	dSession := session.NewDevice(store)
	store.Migrate(session.NewRecord())
	store.Migrate(session.NewSessionUsersRecord())

	dCleaningServices := cleaning.NewDevice(store)
	store.Migrate(cleaning.NewRecord())

	dGardeningServices := gardening.NewDevice(store)
	store.Migrate(gardening.NewRecord())

	dElectricianServices := electrician.NewDevice(store)
	store.Migrate(electrician.NewRecord())

	dPlumbingServices := plumbing.NewDevice(store)
	store.Migrate(plumbing.NewRecord())

	mwProfileCore := adapter.MNA()
	mwProfileMethodHandlers := mwProfileCore.Put(dProfile.Update()).Get(dProfile.Read()).Post(dSession.CreateUser())

	mwCleaningServicesCore := adapter.MNA()
	mwCleaningServicesMethodHandlers := mwCleaningServicesCore.Put(dCleaningServices.Update()).Get(dCleaningServices.List()).Post(dCleaningServices.Create())

	mwGardeningServicesCore := adapter.MNA()
	mwGardeningServicesMethodHandlers := mwGardeningServicesCore.Put(dGardeningServices.Update()).Get(dGardeningServices.List()).Post(dCleaningServices.Create())

	mwElectricianServicesCore := adapter.MNA()
	mwElectricianServicesMethodHandlers := mwElectricianServicesCore.Put(dElectricianServices.Update()).Get(dElectricianServices.List()).Post(dCleaningServices.Create())

	mwPlumbingServicesCore := adapter.MNA()
	mwPlumbingServicesMethodHandlers := mwPlumbingServicesCore.Put(dPlumbingServices.Update()).Get(dPlumbingServices.List()).Post(dCleaningServices.Create())

	mux := http.NewServeMux()
	mux.Handle("/", adapter.Core(serveFile("static/")).Notify().Entry())
	mux.Handle("/static/", adapter.Core(serveFiles("/static/", "static")).Entry())

	mwSignin := adapter.MNA().Post(dProfile.Read()).And(dSession.Signin())
	mwSignup := adapter.MNA().Post(dProfile.Create()).And(dSession.CreateUser())
	mwSignout := adapter.MNA().Delete(dSession.Signout())

	mux.Handle("/signin", mwSignin.And(dSession.Authenticate()).Notify().Entry())
	mux.Handle("/signup", mwSignup.And(dSession.Authenticate()).Notify().Entry())
	mux.Handle("/signout", mwSignout.And(dSession.Authenticate()).Notify().Entry())

	mux.Handle("/api/v1/cleaning", mwCleaningServicesMethodHandlers.And(dSession.Authenticate()).Notify().Entry())
	mux.Handle("/api/v1/gardening", mwGardeningServicesMethodHandlers.And(dSession.Authenticate()).Notify().Entry())
	mux.Handle("/api/v1/electrician", mwElectricianServicesMethodHandlers.And(dSession.Authenticate()).Notify().Entry())
	mux.Handle("/api/v1/plumbing", mwPlumbingServicesMethodHandlers.And(dSession.Authenticate()).Notify().Entry())

	mux.Handle("/api/v1/profile", mwProfileMethodHandlers.And(dSession.Authenticate()).Notify().Entry())

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
