package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type response struct {
	Error    string                 `json:"error,omitempty"`
	Response map[string]interface{} `json:"response,omitempty"`
}

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
