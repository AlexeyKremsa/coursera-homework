package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type response struct {
	Error    string      `json:"error"`
	Response interface{} `json:"response,omitempty"`
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

func (api *MyApi) wrapperProfile(w http.ResponseWriter, r *http.Request) {
	var login string

	if r.Method == http.MethodPost {
		login = r.FormValue("login")
	}

	if r.Method == http.MethodGet {
		login = r.URL.Query().Get("login")
	}

	if login == "" {
		writeResponseJSON(w, http.StatusBadRequest, nil, "login must me not empty")
		return
	}

	profileParams := ProfileParams{Login: login}
	user, err := api.Profile(r.Context(), profileParams)
	if err != nil {
		apiErr, ok := err.(ApiError)
		if ok {
			writeResponseJSON(w, apiErr.HTTPStatus, nil, apiErr.Err.Error())
			return
		}

		writeResponseJSON(w, http.StatusInternalServerError, nil, err.Error())
		return
	}

	writeResponseJSON(w, http.StatusOK, user, "")
}

func (api *MyApi) wrapperCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeResponseJSON(w, http.StatusNotAcceptable, nil, "bad method")
		return
	}

	if r.Header.Get("X-Auth") != "100500" {
		writeResponseJSON(w, http.StatusForbidden, nil, "unauthorized")
		return
	}

	login := r.FormValue("login")
	if login == "" {
		writeResponseJSON(w, http.StatusBadRequest, nil, "login must me not empty")
		return
	}
	if len(login) < 10 {
		writeResponseJSON(w, http.StatusBadRequest, nil, "login len must be >= 10")
		return
	}

	full_name := r.FormValue("full_name")

	status := r.FormValue("status")
	if status == "" {
		status = "user"
	} else {
		statusList := []string{"user", "moderator", "admin"}
		isStatusValid := false

		for _, item := range statusList {
			if item == status {
				isStatusValid = true
				break
			}
		}

		if !isStatusValid {
			writeResponseJSON(w, http.StatusBadRequest, nil, "status must be one of [user, moderator, admin]")
			return
		}
	}

	age := r.FormValue("age")
	var ageInt int
	if age != "" {
		ageInt, err := strconv.Atoi(age)
		if err != nil {
			writeResponseJSON(w, http.StatusBadRequest, nil, "age must be int")
			return
		}

		if ageInt < 0 {
			writeResponseJSON(w, http.StatusBadRequest, nil, "age must be >= 0")
			return
		}

		if ageInt > 128 {
			writeResponseJSON(w, http.StatusBadRequest, nil, "age must be <= 128")
			return
		}
	}

	createParams := CreateParams{
		Login:  login,
		Name:   full_name,
		Status: status,
		Age:    ageInt,
	}

	newUser, err := api.Create(r.Context(), createParams)
	if err != nil {
		apiErr, ok := err.(ApiError)
		if ok {
			writeResponseJSON(w, apiErr.HTTPStatus, nil, apiErr.Err.Error())
			return
		}

		writeResponseJSON(w, http.StatusInternalServerError, nil, err.Error())
		return
	}

	writeResponseJSON(w, http.StatusOK, newUser, "")
}
