package model

type Role struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsSystem    bool   `json:"is_system"`
}

type Permission struct {
	ID          int64  `json:"id"`
	Key         string `json:"key"`
	Description string `json:"description"`
	Category    string `json:"category"`
}
