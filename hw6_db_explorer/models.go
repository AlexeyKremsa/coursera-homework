package main

import "database/sql"

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

type Table struct {
	Name string
}

type TablesResp struct {
	Tables []string `json:"tables"`
}
