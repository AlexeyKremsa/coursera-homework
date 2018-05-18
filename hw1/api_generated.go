package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type response struct {
	Error    string      "json:'error'"
	Response interface{} "json:'response,omitempty'"
}

func writeResponseJSON(w http.ResponseWriter, status int, data interface{}, errorText string) {
	w.Header().Set("Content-Type", "application/json")
	resp := response{
		Error:    errorText,
		Response: data,
	}

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
	} else {
		w.WriteHeader(status)
		w.Write(jsonResp)
	}
}

func (api *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path { 
	case "/user/profile":
		api.wrapperProfile(w, r)
	case "/user/create":
		api.wrapperCreate(w, r)
	default:
		writeResponseJSON(w, http.StatusNotFound, nil, "unknown method")
	}
}
func (api *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path { 
	case "/user/create":
		api.wrapperCreate(w, r)
	default:
		writeResponseJSON(w, http.StatusNotFound, nil, "unknown method")
	}
}