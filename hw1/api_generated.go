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


func (srv *MyApi) wrapperProfile(w http.ResponseWriter, r *http.Request) {

	var Login string

func (srv *MyApi) wrapperCreate(w http.ResponseWriter, r *http.Request) {	
		if r.Method != http.MethodPost {
			writeResponseJSON(w, http.StatusNotAcceptable, nil, "bad method")
			return
		}

	var Login string
	var Name string
	var Status string
	var Age int

func (srv *OtherApi) wrapperCreate(w http.ResponseWriter, r *http.Request) {	
		if r.Method != http.MethodPost {
			writeResponseJSON(w, http.StatusNotAcceptable, nil, "bad method")
			return
		}

	var Username string
	var Name string
	var Class string
	var Level int

func (srv *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path { 
		case "/user/profile":
			srv.wrapperProfile(w, r)
		case "/user/create":
			srv.wrapperCreate(w, r)
		default:
			writeResponseJSON(w, http.StatusNotFound, nil, "unknown method")
		}
	}

func (srv *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path { 
		case "/user/create":
			srv.wrapperCreate(w, r)
		default:
			writeResponseJSON(w, http.StatusNotFound, nil, "unknown method")
		}
	}
