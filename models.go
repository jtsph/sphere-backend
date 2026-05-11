package main

import "time"

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type Session struct {
	Token     string    `json:"token"`
	UserID    int64     `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Block struct {
	Height    int64     `json:"height"`
	Hash      string    `json:"hash"`
	Timestamp time.Time `json:"timestamp"`
	Validator string    `json:"validator"`
}

type Validator struct {
	ID     int64   `json:"id"`
	Name   string  `json:"name"`
	Status string  `json:"status"`
	Stake  float64 `json:"stake"`
}

type LearnResource struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
	URL     string `json:"url"`
}

type InvestOption struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	Summary      string `json:"summary"`
	TargetAmount string `json:"target_amount"`
}

type MinecraftServer struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Host        string `json:"host"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

type DashboardSummary struct {
	LatestBlocks []Block     `json:"latest_blocks"`
	Validators   []Validator `json:"validators"`
	Metrics      interface{} `json:"metrics"`
}
