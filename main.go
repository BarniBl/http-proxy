package main

import (
	"context"
	"crypto/tls"
	"github.com/BarniBl/http-proxy/cmd/handlers"
	"github.com/rs/zerolog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	httpAddr               = ":8080"
	httpsAddr              = ":8081"
	historyAddr            = ":8090"
	certPath               = "cert.pem"
	keyPath                = "key.pem"
	serverReadWriteTimeout = 15 * time.Second
	clientReadWriteTimeout = 10 * time.Second
)

func main() {
	log := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
	/*	certs := x509.NewCertPool()

		pemData, err := ioutil.ReadFile(pemPath)
		if err != nil {
			log.Fatal().Msg(err.Error())
		}

		certs.AppendCertsFromPEM(pemData)*/

	/*	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: certs,
		},
	}*/
	var mu sync.Mutex
	reqList := handlers.MyRequestList{
		Requests: make([]handlers.MyRequest, 0),
		ReqMu:    &mu,
	}
	client := http.Client{
		Timeout: clientReadWriteTimeout,
		//Transport: tr,
	}
	proxy := handlers.Proxy{Client: &client, Log: &log, MyReqList: &reqList}

	httpAPI := &http.Server{
		Addr:        httpAddr,
		Handler:     http.HandlerFunc(proxy.Proxy),
		ReadTimeout: serverReadWriteTimeout,
	}

	httpsAPI := &http.Server{
		Addr: httpsAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				proxy.Proxy(w, r)
			} else {
				proxy.Proxy(w, r)
			}
		}),
		ReadTimeout:  serverReadWriteTimeout,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	historyAPI := &http.Server{
		Addr:        historyAddr,
		Handler:     handlers.HistoryRouter(&log, &reqList, &proxy),
		ReadTimeout: serverReadWriteTimeout,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Info().Msgf("Start listening http server %s", httpAPI.Addr)
		serverErrors <- httpAPI.ListenAndServe()
	}()
	go func() {
		log.Info().Msgf("Start listening https server %s", httpsAPI.Addr)
		serverErrors <- httpsAPI.ListenAndServeTLS(certPath, keyPath)
	}()
	go func() {
		log.Info().Msgf("Start listening history server %s", historyAPI.Addr)
		serverErrors <- historyAPI.ListenAndServe()
	}()
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)
	select {
	case <-osSignals:
		log.Info().Msg("Start shutdown...")
		if err := httpAPI.Shutdown(context.Background()); err != nil {
			log.Error().Msg("Graceful http server shutdown error: " + err.Error())
		}
		if err := httpsAPI.Shutdown(context.Background()); err != nil {
			log.Error().Msg("Graceful https server shutdown error: " + err.Error())
		}
		if err := historyAPI.Shutdown(context.Background()); err != nil {
			log.Error().Msg("Graceful history server shutdown error: " + err.Error())
		}
		os.Exit(1)
	}
}
