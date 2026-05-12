package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const authHeader = "Authorization"

var rateLimiter = newIPRateLimiter(100, time.Minute)

type API struct {
	store  *Store
	config *Config
}

func NewRouter(store *Store, cfg *Config) http.Handler {
	a := &API{store: store, config: cfg}
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/health", a.healthHandler)
	mux.HandleFunc("/api/v1/auth/register", a.registerHandler)
	mux.HandleFunc("/api/v1/auth/login", a.loginHandler)
	mux.HandleFunc("/api/v1/dashboard", a.dashboardHandler)
	mux.HandleFunc("/api/v1/blocks", a.blocksHandler)
	mux.HandleFunc("/api/v1/blocks/stats", a.blockchainStatsHandler)
	mux.HandleFunc("/api/v1/validators", a.validatorsHandler)
	mux.HandleFunc("/api/v1/sentience/learn", a.learnHandler)
	mux.HandleFunc("/api/v1/sentience/docs", a.docsHandler)
	mux.HandleFunc("/api/v1/invest", a.investHandler)
	mux.HandleFunc("/api/v1/minecraft", a.minecraftHandler)
	mux.HandleFunc("/api/v1/minecraft/hosting", a.minecraftHostingHandler)
	mux.HandleFunc("/api/v1/minecraft/launcher", a.minecraftLauncherHandler)

	apiMux := a.chain(mux)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Route API requests to the API mux
		if strings.HasPrefix(r.URL.Path, "/api") {
			apiMux.ServeHTTP(w, r)
			return
		}

		// Serve static files for everything else
		a.serveStatic(w, r)
	})
}

func (a *API) chain(next http.Handler) http.Handler {
	return a.withCORS(a.withSecureHeaders(a.withRateLimit(a.withRequestLogger(next))))
}

func (a *API) withRequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func (a *API) withSecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		next.ServeHTTP(w, r)
	})
}

func (a *API) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", a.config.CorsOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *API) withRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rateLimiter.Allow(r.RemoteAddr) {
			a.writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *API) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	a.writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "service": "sphere-backend", "version": "0.1.0"})
}

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (a *API) registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req registerRequest
	if err := a.readJSON(r.Body, &req); err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if !validUsername(req.Username) || !validPassword(req.Password) {
		a.writeError(w, http.StatusBadRequest, "invalid username or password")
		return
	}

	user, err := a.store.CreateUser(strings.TrimSpace(req.Username), req.Password)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			a.writeError(w, http.StatusConflict, "username already exists")
			return
		}
		a.writeError(w, http.StatusInternalServerError, "unable to create user")
		return
	}

	a.writeJSON(w, http.StatusCreated, map[string]any{"id": user.ID, "username": user.Username, "created_at": user.CreatedAt})
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (a *API) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req loginRequest
	if err := a.readJSON(r.Body, &req); err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := a.store.AuthenticateUser(strings.TrimSpace(req.Username), req.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			a.writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		a.writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := generateSecureToken(32)
	if err != nil {
		a.writeError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	if err := a.store.CreateSession(token, user.ID, expiresAt); err != nil {
		a.writeError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	a.writeJSON(w, http.StatusOK, map[string]any{"token": token, "expires_at": expiresAt.Format(time.RFC3339)})
}

func (a *API) dashboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	blocks, err := a.store.ListBlocks()
	if err != nil {
		a.writeError(w, http.StatusInternalServerError, "failed to load blocks")
		return
	}
	validators, err := a.store.ListValidators()
	if err != nil {
		a.writeError(w, http.StatusInternalServerError, "failed to load validators")
		return
	}
	userCount, _ := a.store.CountUsers()
	blockCount, _ := a.store.CountBlocks()

	summary := DashboardSummary{
		LatestBlocks: blocks,
		Validators:   validators,
		Metrics: map[string]any{
			"users":  userCount,
			"blocks": blockCount,
			"nodes":  len(validators),
		},
	}

	a.writeJSON(w, http.StatusOK, summary)
}

func (a *API) blocksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	blocks, err := a.store.ListBlocks()
	if err != nil {
		a.writeError(w, http.StatusInternalServerError, "failed to load blocks")
		return
	}
	a.writeJSON(w, http.StatusOK, map[string]any{"blocks": blocks})
}

func (a *API) blockchainStatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	blockCount, _ := a.store.CountBlocks()
	validators, _ := a.store.ListValidators()
	blocks, _ := a.store.ListBlocks()

	var latestBlockTime time.Time
	if len(blocks) > 0 {
		latestBlockTime = blocks[0].Timestamp
	}

	stats := map[string]any{
		"total_blocks":      blockCount,
		"active_validators": len(validators),
		"latest_block_time": latestBlockTime,
		"network_status":    "operational",
		"consensus":         "Proof of Stake",
		"block_time":        "15s",
		"total_stake":       266865.7,
	}
	a.writeJSON(w, http.StatusOK, stats)
}

func (a *API) validatorsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	validators, err := a.store.ListValidators()
	if err != nil {
		a.writeError(w, http.StatusInternalServerError, "failed to load validators")
		return
	}
	a.writeJSON(w, http.StatusOK, map[string]any{"validators": validators})
}

func (a *API) learnHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	resources, err := a.store.ListLearnResources()
	if err != nil {
		a.writeError(w, http.StatusInternalServerError, "failed to load learning resources")
		return
	}
	a.writeJSON(w, http.StatusOK, map[string]any{"learn": resources})
}

func (a *API) docsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	a.writeJSON(w, http.StatusOK, map[string]any{
		"docs": map[string]any{
			"description": "Documentation for the Sphere project and deployment guides.",
			"repository":  "https://github.com/thesphere/sphere-backend",
		},
	})
}

func (a *API) investHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	options, err := a.store.ListInvestOptions()
	if err != nil {
		a.writeError(w, http.StatusInternalServerError, "failed to load investment options")
		return
	}
	a.writeJSON(w, http.StatusOK, map[string]any{"invest": options})
}

func (a *API) minecraftHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	servers, err := a.store.ListMinecraftServers()
	if err != nil {
		a.writeError(w, http.StatusInternalServerError, "failed to load minecraft servers")
		return
	}
	a.writeJSON(w, http.StatusOK, map[string]any{"minecraft": servers})
}

func (a *API) minecraftHostingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	servers, err := a.store.ListMinecraftServers()
	if err != nil {
		a.writeError(w, http.StatusInternalServerError, "failed to load minecraft servers")
		return
	}

	hosting := map[string]any{
		"service": "Sphere Minecraft Hosting",
		"status":  "operational",
		"infrastructure": map[string]any{
			"provider": "Sphere Cloud",
			"regions":  []string{"us-east", "eu-west", "asia-pacific"},
			"uptime":   "99.9%",
		},
		"servers": servers,
		"capacity": map[string]any{
			"total_slots":    500,
			"active_players": 234,
			"average_tps":    19.8,
		},
		"features": []string{
			"24/7 monitoring",
			"Automatic backups",
			"DDoS protection",
			"Custom plugins support",
			"World transfer assistance",
		},
	}
	a.writeJSON(w, http.StatusOK, hosting)
}

func (a *API) minecraftLauncherHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	servers, err := a.store.ListMinecraftServers()
	if err != nil {
		a.writeError(w, http.StatusInternalServerError, "failed to load servers")
		return
	}

	launcher := map[string]any{
		"game":     "minecraft",
		"version":  "1.20.x",
		"servers":  servers,
		"launcher": "https://launcher.thesphere.online",
		"status":   "ready_to_play",
	}
	a.writeJSON(w, http.StatusOK, launcher)
}

func (a *API) readJSON(body io.ReadCloser, dst any) error {
	defer body.Close()
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func (a *API) writeJSON(w http.ResponseWriter, status int, payload any) {
	w.WriteHeader(status)
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(payload)
}

func (a *API) writeError(w http.ResponseWriter, status int, message string) {
	a.writeJSON(w, status, map[string]any{"error": message})
}

func (a *API) serveStatic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Resolve the file path from the frontend dist directory
	distDir := filepath.Join("frontend", "dist")
	path := filepath.Join(distDir, filepath.Clean(r.URL.Path))

	// Prevent directory traversal
	if !strings.HasPrefix(path, distDir) {
		http.NotFound(w, r)
		return
	}

	// Check if the file exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// For SPA routing: serve index.html for non-existent files
			indexPath := filepath.Join(distDir, "index.html")
			http.ServeFile(w, r, indexPath)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// If it's a directory, try to serve index.html from that directory
	if info.IsDir() {
		indexPath := filepath.Join(path, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			http.ServeFile(w, r, indexPath)
			return
		}
		http.NotFound(w, r)
		return
	}

	// Serve the file
	http.ServeFile(w, r, path)
}

func validUsername(name string) bool {
	trimmed := strings.TrimSpace(name)
	return len(trimmed) >= 3 && len(trimmed) <= 64 && !strings.Contains(trimmed, " ")
}

func validPassword(password string) bool {
	return len(password) >= 8
}

func generateSecureToken(length int) (string, error) {
	buffer := make([]byte, length)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(buffer), nil
}

func authFromHeader(r *http.Request) (string, error) {
	header := r.Header.Get(authHeader)
	if header == "" {
		return "", errors.New("missing authorization header")
	}
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("invalid authorization header")
	}
	return parts[1], nil
}

type ipRateLimiter struct {
	requests map[string]rateWindow
	limit    int
	window   time.Duration
	mu       sync.Mutex
}

type rateWindow struct {
	count int
	start time.Time
}

func newIPRateLimiter(limit int, window time.Duration) *ipRateLimiter {
	return &ipRateLimiter{
		requests: make(map[string]rateWindow),
		limit:    limit,
		window:   window,
	}
}

func (rl *ipRateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry := rl.requests[key]
	if time.Since(entry.start) > rl.window {
		entry = rateWindow{count: 1, start: time.Now()}
	} else {
		entry.count++
	}
	rl.requests[key] = entry
	return entry.count <= rl.limit
}
