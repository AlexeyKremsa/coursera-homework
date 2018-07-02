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
	Error    string                 `json:"error,omitempty"`
	Response map[string]interface{} `json:"response,omitempty"`
}

// type records struct {
// 	Records interface{} `json:"records,omitempty"`
// }

// func writeResponseJSON(w http.ResponseWriter, status int, responseName string, data interface{}, errorText string) {
// 	w.Header().Set("Content-Type", "application/json")
// 	resMap := make(map[string]interface{})
// 	resMap[responseName] = data
// 	resp := response{
// 		Error:    errorText,
// 		Response: resMap,
// 	}

// 	jsonResp, err := json.Marshal(resp)
// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		fmt.Fprint(w, err.Error())
// 	} else {
// 		w.WriteHeader(status)
// 		w.Write(jsonResp)
// 	}
// }

func writeResponseJSON(w http.ResponseWriter, status int, responseName string, data interface{}, errorText string) {
	w.Header().Set("Content-Type", "application/json")
	var resp response
	if errorText != "" {
		resp = response{
			Error: errorText,
		}
	} else {
		resMap := make(map[string]interface{})
		resMap[responseName] = data
		resp = response{
			Response: resMap,
		}
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
