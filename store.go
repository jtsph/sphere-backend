package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func NewStore(dsn string) (*Store, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	store := &Store{db: db}
	if err := store.initTables(); err != nil {
		return nil, err
	}
	if err := store.seedDemoData(); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *Store) initTables() error {
	statements := []string{
		`PRAGMA foreign_keys = ON;`,
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS sessions (
			token TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			expires_at TEXT NOT NULL,
			FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS blocks (
			height INTEGER PRIMARY KEY,
			hash TEXT NOT NULL,
			timestamp TEXT NOT NULL,
			validator TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS validators (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			status TEXT NOT NULL,
			stake REAL NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS learn_resources (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			summary TEXT NOT NULL,
			url TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS invest_options (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			summary TEXT NOT NULL,
			target_amount TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS minecraft_servers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			host TEXT NOT NULL,
			status TEXT NOT NULL,
			description TEXT NOT NULL
		);`,
	}

	for _, stmt := range statements {
		if _, err := s.db.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) seedDemoData() error {
	entries := []struct {
		query string
		args  []any
	}{
		{`INSERT OR IGNORE INTO validators(id, name, status, stake) VALUES (1, 'Sphere Node Alpha', 'active', 86000.5);`, nil},
		{`INSERT OR IGNORE INTO validators(id, name, status, stake) VALUES (2, 'Sentience Validator', 'active', 72000.2);`, nil},
		{`INSERT OR IGNORE INTO learn_resources(title, summary, url) VALUES ('Sphere Academy', 'Guides and courses for entrepreneurs, builders, and AI researchers.', 'https://sentience.thesphere.online/learn');`, nil},
		{`INSERT OR IGNORE INTO learn_resources(title, summary, url) VALUES ('Sentience Docs', 'Documentation, roadmap, and technical notes for the Sphere network.', 'https://sentience.thesphere.online/docs');`, nil},
		{`INSERT OR IGNORE INTO invest_options(title, summary, target_amount) VALUES ('Founders Fund', 'Early access investment for core ecosystem growth.', '5,000,000 SPR');`, nil},
		{`INSERT OR IGNORE INTO invest_options(title, summary, target_amount) VALUES ('Community Drive', 'A community-backed fund for platform adoption and education.', '1,200,000 SPR');`, nil},
		{`INSERT OR IGNORE INTO minecraft_servers(name, host, status, description) VALUES ('Sphere EDU', 'minecraft.thesphere.online', 'online', 'An educative Minecraft world built around the Sphere mission.');`, nil},
	}

	for _, entry := range entries {
		if _, err := s.db.Exec(entry.query); err != nil {
			return err
		}
	}

	_, err := s.db.Exec(`INSERT OR IGNORE INTO blocks(height, hash, timestamp, validator) VALUES (1, '0000001a2b3c4d5e', strftime('%Y-%m-%dT%H:%M:%SZ','now'), 'Sphere Node Alpha');`)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) CreateUser(username, password string) (*User, error) {
	hash, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	result, err := s.db.Exec(`INSERT INTO users(username, password_hash, created_at) VALUES (?, ?, strftime('%Y-%m-%dT%H:%M:%SZ','now'))`, username, hash)
	if err != nil {
		return nil, err
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &User{ID: userID, Username: username, CreatedAt: time.Now().UTC()}, nil
}

func (s *Store) AuthenticateUser(username, password string) (*User, error) {
	user := &User{}
	row := s.db.QueryRow(`SELECT id, username, password_hash, created_at FROM users WHERE username = ?`, username)
	var createdAt string
	if err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &createdAt); err != nil {
		return nil, err
	}
	user.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	if user.CreatedAt.IsZero() {
		user.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	}

	if !verifyPassword(user.PasswordHash, password) {
		return nil, errors.New("invalid credentials")
	}
	return user, nil
}

func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	key := pbkdf2SHA512([]byte(password), salt, 200_000, 64)
	saltEncoded := base64.RawStdEncoding.EncodeToString(salt)
	keyEncoded := base64.RawStdEncoding.EncodeToString(key)
	return fmt.Sprintf("%s$%s", saltEncoded, keyEncoded), nil
}

func verifyPassword(hash, password string) bool {
	parts := strings.Split(hash, "$")
	if len(parts) != 2 {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}

	actual := pbkdf2SHA512([]byte(password), salt, 200_000, len(expected))
	return hmac.Equal(expected, actual)
}

func pbkdf2SHA512(password, salt []byte, iter, keyLen int) []byte {
	h := hmac.New(sha512.New, password)
	h.Write(append(salt, 0, 0, 0, 1))
	dk := hmac.New(sha512.New, password).Sum(nil)
	tmp := make([]byte, len(dk))
	copy(tmp, dk)

	for i := 1; i < iter; i++ {
		h.Reset()
		h.Write(tmp)
		tmp = hmac.New(sha512.New, password).Sum(tmp)
		for j := range dk {
			dk[j] ^= tmp[j]
		}
	}

	if keyLen <= len(dk) {
		return dk[:keyLen]
	}
	result := make([]byte, keyLen)
	copy(result, dk)
	for len(result) < keyLen {
		chunk := hmac.New(sha512.New, password).Sum(nil)
		copy(result[len(result):], chunk)
	}
	return result
}

func (s *Store) CreateSession(token string, userID int64, expiresAt time.Time) error {
	_, err := s.db.Exec(`INSERT INTO sessions(token, user_id, expires_at) VALUES (?, ?, ?)`, token, userID, expiresAt.UTC().Format(time.RFC3339Nano))
	return err
}

func (s *Store) GetSession(token string) (*Session, error) {
	session := &Session{}
	row := s.db.QueryRow(`SELECT token, user_id, expires_at FROM sessions WHERE token = ?`, token)
	var raw string
	if err := row.Scan(&session.Token, &session.UserID, &raw); err != nil {
		return nil, err
	}
	expiresAt, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return nil, err
	}
	session.ExpiresAt = expiresAt
	if time.Now().UTC().After(session.ExpiresAt) {
		return nil, errors.New("session expired")
	}
	return session, nil
}

func (s *Store) ListBlocks() ([]Block, error) {
	rows, err := s.db.Query(`SELECT height, hash, timestamp, validator FROM blocks ORDER BY height DESC LIMIT 10`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []Block
	for rows.Next() {
		var block Block
		var raw string
		if err := rows.Scan(&block.Height, &block.Hash, &raw, &block.Validator); err != nil {
			return nil, err
		}
		block.Timestamp, _ = time.Parse(time.RFC3339Nano, raw)
		blocks = append(blocks, block)
	}
	return blocks, nil
}

func (s *Store) ListValidators() ([]Validator, error) {
	rows, err := s.db.Query(`SELECT id, name, status, stake FROM validators ORDER BY stake DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var validators []Validator
	for rows.Next() {
		var validator Validator
		if err := rows.Scan(&validator.ID, &validator.Name, &validator.Status, &validator.Stake); err != nil {
			return nil, err
		}
		validators = append(validators, validator)
	}
	return validators, nil
}

func (s *Store) ListLearnResources() ([]LearnResource, error) {
	rows, err := s.db.Query(`SELECT id, title, summary, url FROM learn_resources ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []LearnResource
	for rows.Next() {
		var item LearnResource
		if err := rows.Scan(&item.ID, &item.Title, &item.Summary, &item.URL); err != nil {
			return nil, err
		}
		resources = append(resources, item)
	}
	return resources, nil
}

func (s *Store) ListInvestOptions() ([]InvestOption, error) {
	rows, err := s.db.Query(`SELECT id, title, summary, target_amount FROM invest_options ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var options []InvestOption
	for rows.Next() {
		var option InvestOption
		if err := rows.Scan(&option.ID, &option.Title, &option.Summary, &option.TargetAmount); err != nil {
			return nil, err
		}
		options = append(options, option)
	}
	return options, nil
}

func (s *Store) ListMinecraftServers() ([]MinecraftServer, error) {
	rows, err := s.db.Query(`SELECT id, name, host, status, description FROM minecraft_servers ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []MinecraftServer
	for rows.Next() {
		var server MinecraftServer
		if err := rows.Scan(&server.ID, &server.Name, &server.Host, &server.Status, &server.Description); err != nil {
			return nil, err
		}
		servers = append(servers, server)
	}
	return servers, nil
}

func (s *Store) CountUsers() (int64, error) {
	var count int64
	row := s.db.QueryRow(`SELECT COUNT(*) FROM users`)
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Store) CountBlocks() (int64, error) {
	var count int64
	row := s.db.QueryRow(`SELECT COUNT(*) FROM blocks`)
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
