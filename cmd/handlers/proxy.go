package handlers

import (
	"bytes"
	"fmt"
	"github.com/rs/zerolog"
	"io"
	"io/ioutil"
	"net/http"
)

type Proxy struct {
	Client    *http.Client
	Log       *zerolog.Logger
	MyReqList *MyRequestList
}

func (h *Proxy) Proxy(w http.ResponseWriter, r *http.Request) {
	if r.URL.Scheme == "https" {
		h.Log.Info().Msg("https")
		return
	}
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
		return
	}
	defer resp.Body.Close()
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)

	info := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	h.Log.Info().Str("Request", info).Msg("Success")
}

func copyHeader(copyTo, copyFrom http.Header) {
	for key, values := range copyFrom {
		for _, value := range values {
			copyTo.Add(key, value)
		}
	}
}
