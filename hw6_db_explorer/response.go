package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

type Item struct {
	ID          int            `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Updated     sql.NullString `json:"updated"`
}

type User struct {
	UserID   int            `json:"user_id"`
	Login    string         `json:"login"`
	Password string         `json:"password"`
	Email    string         `json:"email"`
	Info     string         `json:"info"`
	Updated  sql.NullString `json:"updated"`
}

type TablesResp struct {
	Tables []string `json:"tables"`
}

type response struct {
	Error    string      `json:"error,omitempty"`
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
