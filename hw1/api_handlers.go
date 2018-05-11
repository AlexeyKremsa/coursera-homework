package main

import (
	"net/http"
	"github.com/gin-gonic/gin/json"
	"fmt"
)

type response struct {
	error string
	response interface{}
}

func writeResponseJSON(w http.ResponseWriter, status int, data interface{}, errorText string) {
	w.Header().Set("Content-Type", "application/json")
	resp := response {
		error : errorText,
		response: data,
	}

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_,_ = fmt.Fprint(w, err.Error())
	} else {
		w.WriteHeader(status)
		_,_ = fmt.Fprint(w, string(jsonResp))
	}
}

func (api *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/profile":
		api.wrapperProfile(w, r)
	default:
		http.Error(w, "unknown method", http.StatusNotFound)
	}
}

func (api *MyApi) wrapperProfile(w http.ResponseWriter, r *http.Request) {
	login := r.URL.Query().Get("login")
	if login == "" {
		//http.Error(w, "login must me not empty", http.StatusBadRequest)
		writeResponseJSON(w, http.StatusBadRequest, nil, "login must me not empty")
		return
	}

	params := ProfileParams{ Login: login}
	user, err := api.Profile(r.Context(), params)
	if err != nil {
		apiErr, ok := err.(ApiError)
		if ok {
			//http.Error(w,  apiErr.Err.Error(), apiErr.HTTPStatus)
			writeResponseJSON(w, apiErr.HTTPStatus, nil, apiErr.Err.Error())
			return
		}

		//http.Error(w,  err.Error(), http.StatusBadRequest)
		writeResponseJSON(w, http.StatusInternalServerError, nil, err.Error())
		return
	}

	writeResponseJSON(w, http.StatusOK, user, "")
}