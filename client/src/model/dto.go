package model

type ChatDto struct {
	From      string   `json:"from"`
	To        string   `json:"to"`
	Msg       string   `json:"message"`
	Type      string   `json:"type"`
	Users     []string `json:"users"`
	Timestamp int64    `json:"timestamp"`
}