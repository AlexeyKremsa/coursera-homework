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

	if r.Method == http.MethodGet {
       Login = r.URL.Query().Get(`login`)
    }

	if r.Method == http.MethodPost {
       Login = r.FormValue(`login`)
    }

	if Login == "" {
		writeResponseJSON(w, http.StatusBadRequest, nil, "Login must me not empty")
		return
	}
	
	paramsToPass := ProfileParams {
		Login: Login,
	}
	
	resp, err := srv.Profile(r.Context(), paramsToPass)
	if err != nil {
		apiErr, ok := err.(ApiError)
		if ok {
			writeResponseJSON(w, apiErr.HTTPStatus, nil, apiErr.Err.Error())
			return
		}

		writeResponseJSON(w, http.StatusInternalServerError, nil, err.Error())
		return
	}

	writeResponseJSON(w, http.StatusOK, resp, "")
}

func (srv *MyApi) wrapperCreate(w http.ResponseWriter, r *http.Request) {	
	if r.Method != http.MethodPost {
		writeResponseJSON(w, http.StatusNotAcceptable, nil, "bad method")
		return
	}
	
	if r.Header.Get("X-Auth") != "100500" {
		writeResponseJSON(w, http.StatusForbidden, nil, "unauthorized")
		return
	}

	var Login string
	var Name string
	var Status string
	var Age string

	if r.Method == http.MethodPost {
       Login = r.FormValue(`login`)
       Name = r.FormValue(`full_name`)
       Status = r.FormValue(`status`)
       Age = r.FormValue(`age`)
    }

	if Login == "" {
		writeResponseJSON(w, http.StatusBadRequest, nil, "Login must me not empty")
		return
	}
	
	if len(Login) < 10 {
		writeResponseJSON(w, http.StatusBadRequest, nil, "Login must be more than 10 characters")
		return
	}
	
	if Status == "" {
		Status = "user"
	}
	statusList := []string{"user", "moderator", "admin"}
	isStatusValid := false
	for _, item := range statusList {
		if item == Status {
			isStatusValid = true
				break
		}
	}

	if !isStatusValid {
		writeResponseJSON(w, http.StatusBadRequest, nil, "unknown status: Status")
		return
	}

	AgeInt, err := strconv.Atoi(Age)
	if err != nil {
		writeResponseJSON(w, http.StatusBadRequest, nil, "Age must be int")
		return
	}
	
	if AgeInt == 0 {
		AgeInt = 3
	}
	
	if AgeInt < 0 {
		writeResponseJSON(w, http.StatusBadRequest, nil, "Age must be >= 0")
		return
	}
	
	if AgeInt > 128 {
		writeResponseJSON(w, http.StatusBadRequest, nil, "Age must be <= 128")
		return
	}
	
	paramsToPass := CreateParams {
		Login: Login,
		Name: Name,
		Status: Status,
		Age: AgeInt,
	}
	
	resp, err := srv.Create(r.Context(), paramsToPass)
	if err != nil {
		apiErr, ok := err.(ApiError)
		if ok {
			writeResponseJSON(w, apiErr.HTTPStatus, nil, apiErr.Err.Error())
			return
		}

		writeResponseJSON(w, http.StatusInternalServerError, nil, err.Error())
		return
	}

	writeResponseJSON(w, http.StatusOK, resp, "")
}

func (srv *OtherApi) wrapperCreate(w http.ResponseWriter, r *http.Request) {	
	if r.Method != http.MethodPost {
		writeResponseJSON(w, http.StatusNotAcceptable, nil, "bad method")
		return
	}
	
	if r.Header.Get("X-Auth") != "100500" {
		writeResponseJSON(w, http.StatusForbidden, nil, "unauthorized")
		return
	}

	var Username string
	var Name string
	var Class string
	var Level string

	if r.Method == http.MethodPost {
       Username = r.FormValue(`username`)
       Name = r.FormValue(`account_name`)
       Class = r.FormValue(`class`)
       Level = r.FormValue(`level`)
    }

	if Username == "" {
		writeResponseJSON(w, http.StatusBadRequest, nil, "Username must me not empty")
		return
	}
	
	if len(Username) < 3 {
		writeResponseJSON(w, http.StatusBadRequest, nil, "Username must be more than 3 characters")
		return
	}
	
	if Class == "" {
		Class = "warrior"
	}
	statusList := []string{"warrior", "sorcerer", "rouge"}
	isStatusValid := false
	for _, item := range statusList {
		if item == Class {
			isStatusValid = true
				break
		}
	}

	if !isStatusValid {
		writeResponseJSON(w, http.StatusBadRequest, nil, "unknown status: Class")
		return
	}

	LevelInt, err := strconv.Atoi(Level)
	if err != nil {
		writeResponseJSON(w, http.StatusBadRequest, nil, "Level must be int")
		return
	}
	
	if LevelInt < 1 {
		writeResponseJSON(w, http.StatusBadRequest, nil, "Level must be >= 1")
		return
	}
	
	if LevelInt > 50 {
		writeResponseJSON(w, http.StatusBadRequest, nil, "Level must be <= 50")
		return
	}
	
	paramsToPass := OtherCreateParams {
		Username: Username,
		Name: Name,
		Class: Class,
		Level: LevelInt,
	}
	
	resp, err := srv.Create(r.Context(), paramsToPass)
	if err != nil {
		apiErr, ok := err.(ApiError)
		if ok {
			writeResponseJSON(w, apiErr.HTTPStatus, nil, apiErr.Err.Error())
			return
		}

		writeResponseJSON(w, http.StatusInternalServerError, nil, err.Error())
		return
	}

	writeResponseJSON(w, http.StatusOK, resp, "")
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
