package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"net/http"
	"strconv"
	"sync"
)

type History struct {
	Log       *zerolog.Logger
	MyReqList *MyRequestList
	Proxy     *Proxy
}

type MyRequestList struct {
	Requests []MyRequest
	ReqMu    *sync.Mutex
}

type MyRequest struct {
	ID             int         `json:"id"`
	URL            string      `json:"url"`
	Method         string      `json:"method"`
	RequestHeaders http.Header `json:"-"`
	RequestBody    []byte      `json:"-"`
}

func (mrl *MyRequestList) AddRequest(newRequest MyRequest) {
	mrl.ReqMu.Lock()
	newRequest.ID = len(mrl.Requests)
	mrl.Requests = append(mrl.Requests, newRequest)
	mrl.ReqMu.Unlock()
}

func (h *History) GetLastRequestsList(w http.ResponseWriter, r *http.Request) {
	limit, err := strconv.Atoi(mux.Vars(r)["limit"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	offset, err := strconv.Atoi(mux.Vars(r)["offset"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if offset < 0 || limit < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if limit == 0 {
		limit = len(h.MyReqList.Requests)
	}
	requests := make([]MyRequest, limit)
	for i := offset; i < limit+offset; i++ {
		if i >= len(h.MyReqList.Requests) {
			break
		}
		requests[i-offset] = h.MyReqList.Requests[i]
	}
	requestsList := struct {
		TotalCount int         `json:"totalCount"`
		Requests   []MyRequest `json:"requests"`
	}{TotalCount: len(h.MyReqList.Requests), Requests: requests}
	reqListByte, err := json.Marshal(requestsList)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Write(reqListByte)
	return
}

func (h *History) GetResponse(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if id >= len(h.MyReqList.Requests) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	request := h.MyReqList.Requests[id]
	newReq, err := http.NewRequest(request.Method, request.URL, bytes.NewReader(request.RequestBody))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	newReq.Header = request.RequestHeaders
	h.Proxy.Proxy(w, newReq)
}
