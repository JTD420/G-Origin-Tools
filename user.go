package main

import "time"

type User struct {
	Index      int
	Name       string
	Figure     string
	Gender     string
	Custom     string
	X, Y       int
	Z          float64
	PoolFigure string
	BadgeCode  string
	Type       int
}

type ChatMessage struct {
	Timestamp time.Time `json:"timestamp"`
	UserID    int       `json:"userId"`
	Username  string    `json:"username"`
	Message   string    `json:"message"`
	Type      string    `json:"type"`
}
