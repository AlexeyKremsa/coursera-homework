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

	var login string


	if r.Method == http.MethodGet {
       login = r.URL.Query().Get(`login`)
    }


	if r.Method == http.MethodPost {
       login = r.FormValue(`login`)
    }


	if login == "" {
		writeResponseJSON(w, http.StatusBadRequest, nil, "login must me not empty")
		return
	}


func (srv *MyApi) wrapperCreate(w http.ResponseWriter, r *http.Request) {	
	if r.Method != http.MethodPost {
		writeResponseJSON(w, http.StatusNotAcceptable, nil, "bad method")
		return
	}

	var login string
	var name string
	var status string
	var age int


	if r.Method == http.MethodPost {
       login = r.FormValue(`login`)
       name = r.FormValue(`full_name`)
       status = r.FormValue(`status`)
       age = r.FormValue(`age`)
    }


	if login == "" {
		writeResponseJSON(w, http.StatusBadRequest, nil, "login must me not empty")
		return
	}


func (srv *OtherApi) wrapperCreate(w http.ResponseWriter, r *http.Request) {	
	if r.Method != http.MethodPost {
		writeResponseJSON(w, http.StatusNotAcceptable, nil, "bad method")
		return
	}

	var username string
	var name string
	var class string
	var level int


	if r.Method == http.MethodPost {
       username = r.FormValue(`username`)
       name = r.FormValue(`account_name`)
       class = r.FormValue(`class`)
       level = r.FormValue(`level`)
    }


	if username == "" {
		writeResponseJSON(w, http.StatusBadRequest, nil, "username must me not empty")
		return
	}


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
