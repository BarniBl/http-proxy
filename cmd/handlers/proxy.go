package handlers

import (
	"fmt"
	"github.com/rs/zerolog"
	"io"
	"net/http"
)

type Proxy struct {
	Client *http.Client
	Log    *zerolog.Logger
}

func (h *Proxy) Proxy(w http.ResponseWriter, r *http.Request) {
	if r.URL.Scheme == "https" {
		h.Log.Info().Msg("https")
		return
	}
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

func copyHeader(res, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			res.Add(key, value)
		}
	}
}
