package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/oligoden/chassis/adapter"
	"github.com/oligoden/chassis/storage/gosql"

	//xxx
	"github.com/geraldtivatyi/shalprin/work/gateway"
	//end
	//+++
	//"github.com/geraldtivatyi/shalprin/src/gateway"
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
	hIndex.SetProxy("profile", httputil.NewSingleHostReverseProxy(profile))

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

	mux := http.NewServeMux()
	mux.Handle("/", adapter.Core(serveFile("static")).Notify().Entry())
	mux.Handle("/static/", adapter.Core(serveFiles("/static/", "static")).Entry())

	mux.Handle("/api/v1/", adapter.Core(hIndex).Notify().Entry())

	caCertPool := x509.NewCertPool()
	for _, c := range strings.Split(os.Getenv("CA_CERTS"), ",") {
		caCert, err := ioutil.ReadFile("certs/" + c + ".ca.crt")
		if err != nil {
			log.Fatal(err)
		}
		if !caCertPool.AppendCertsFromPEM(caCert) {
			log.Fatal("ca cert not added")
		}
	}

	cert := os.Getenv("CERT")

	tlsConfig := &tls.Config{
		ClientCAs:  caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
		ServerName: cert,
		// PreferServerCipherSuites: true,
		// CurvePreferences: []tls.CurveID{
		// 	tls.CurveP256,
		// 	tls.X25519,
		// },
		// MinVersion: tls.VersionTLS12,
		// CipherSuites: []uint16{
		// 	tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		// 	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		// 	tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		// 	tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		// 	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		// 	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		// 	// Best disabled, as they don't provide Forward Secrecy,
		// 	// but might be necessary for some clients
		// 	// tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		// 	// tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		// },
	}

	httpsServer := &http.Server{
		Addr:           ":443",
		Handler:        mux,
		TLSConfig:      tlsConfig,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	serverHTTPSError := make(chan error)
	go func() {
		err := httpsServer.ListenAndServeTLS("certs/"+cert+".crt", "certs/"+cert+".key")
		if err != nil && err != http.ErrServerClosed {
			serverHTTPSError <- err
			return
		}
		fmt.Println("https server shutdown")
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	select {
	case err := <-serverHTTPSError:
		fmt.Println("server error", err)
		time.Sleep(100 * time.Millisecond)
		os.Exit(1)
	case sig := <-quit:
		fmt.Println("\ngot signal", sig)
	}

	shutdown(httpsServer)
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
