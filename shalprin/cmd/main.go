package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/oligoden/chassis/adapter"
	"github.com/oligoden/chassis/storage/gosql"
	"golang.org/x/crypto/acme/autocert"

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

func main() {
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
	mux.Handle("/api/v1/profile/", mwProfileMethodHandlers.And(dSession.Authenticate()).Notify().Entry())

	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache("certs"),
		HostPolicy: autocert.HostWhitelist(strings.Split(os.Getenv("ALLOW"), ",")...),
		Email:      "geraldtivatyi@gmail.com",
	}

	httpServer := &http.Server{
		Addr:           ":80",
		Handler:        certManager.HTTPHandler(nil),
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	httpsServer := &http.Server{
		Addr:    ":443",
		Handler: mux,
		TLSConfig: &tls.Config{
			GetCertificate:           certManager.GetCertificate,
			PreferServerCipherSuites: true,
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
				tls.X25519,
			},
			MinVersion: tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				// Best disabled, as they don't provide Forward Secrecy,
				// but might be necessary for some clients
				// tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				// tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			},
		},
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	serverHTTPError := make(chan error)
	serverHTTPSError := make(chan error)
	go func() {
		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			serverHTTPError <- err
			return
		}
		fmt.Println("http server shutdown")
	}()
	go func() {
		err := httpsServer.ListenAndServeTLS("", "")
		if err != nil && err != http.ErrServerClosed {
			serverHTTPSError <- err
			return
		}
		fmt.Println("https server shutdown")
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	select {
	case err := <-serverHTTPError:
		fmt.Println("server error", err)
		shutdown(httpsServer)
		time.Sleep(100 * time.Millisecond)
		os.Exit(1)
	case err := <-serverHTTPSError:
		fmt.Println("server error", err)
		shutdown(httpServer)
		time.Sleep(100 * time.Millisecond)
		os.Exit(1)
	case sig := <-quit:
		fmt.Println("\ngot signal", sig)
	}

	shutdown(httpServer)
	shutdown(httpsServer)
	time.Sleep(100 * time.Millisecond)
}

func serveFile(f string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, f)
	})
}

func serveFiles(p, d string) http.Handler {
	return http.StripPrefix(p, http.FileServer(http.Dir(d)))
}

func shutdown(s *http.Server) {
	ctxServer, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := s.Shutdown(ctxServer)
	if err != nil && err != http.ErrServerClosed {
		fmt.Println("https server shutdown error", err)
	}
}
