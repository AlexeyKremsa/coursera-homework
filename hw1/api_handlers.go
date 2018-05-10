package main

import (
	"net/http"
)

func (h *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/profile":
		h.wrapperProfile(w, r)
	default:
		http.Error(w, "unknown method", http.StatusNotFound)
	}
}

func (h *MyApi) wrapperProfile(w http.ResponseWriter, r *http.Request) {
	login := r.URL.Query().Get("login")
	if login == "" {
		http.Error(w, "login must me not empty", http.StatusBadRequest)
	}

	if login == "bad_user" {
		http.Error(w, "bad user", http.StatusInternalServerError)
	}


}