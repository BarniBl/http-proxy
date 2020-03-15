package handlers

import (
	"bytes"
	"fmt"
	"github.com/rs/zerolog"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

type Proxy struct {
	Client    *http.Client
	Log       *zerolog.Logger
	MyReqList *MyRequestList
}

func (h *Proxy) Proxy(w http.ResponseWriter, r *http.Request) {
/*	if r.URL.Scheme == "https" {
		h.Log.Info().Msg("https")
		w.WriteHeader(http.StatusBadRequest)
		return
	}*/
	bodyByte, _ := ioutil.ReadAll(r.Body)
	newRequest := MyRequest{
		URL:            r.URL.String(),
		RequestHeaders: r.Header.Clone(),
		RequestBody:    bodyByte,
		Method:         r.Method,
	}
	h.MyReqList.AddRequest(newRequest)
	r.Write(bytes.NewBuffer(bodyByte))
	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		h.Log.Warn().Msg(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)

	info := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	h.Log.Info().Str("Request", info).Msg("Success")
}

func (h *Proxy) ProxyConnection(w http.ResponseWriter, r *http.Request) {
/*	err := h.InsertRequest(r, r.RequestURI)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}*/
	destConnection, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConnection, _, err := hijacker.Hijack()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	go h.transfer(destConnection, clientConnection)
	go h.transfer(clientConnection, destConnection)
}

func (h *Proxy) transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}

func copyHeader(copyTo, copyFrom http.Header) {
	for key, values := range copyFrom {
		for _, value := range values {
			copyTo.Add(key, value)
		}
	}
}
