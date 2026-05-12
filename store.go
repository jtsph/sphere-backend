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
		// Validators with more realistic stake amounts
		{`INSERT OR IGNORE INTO validators(id, name, status, stake) VALUES (1, 'Sphere Node Alpha', 'active', 156000.5);`, nil},
		{`INSERT OR IGNORE INTO validators(id, name, status, stake) VALUES (2, 'Sentience Validator', 'active', 142000.2);`, nil},
		{`INSERT OR IGNORE INTO validators(id, name, status, stake) VALUES (3, 'Cortex Prime', 'active', 128500.8);`, nil},
		{`INSERT OR IGNORE INTO validators(id, name, status, stake) VALUES (4, 'Neural Node', 'active', 112300.0);`, nil},
		{`INSERT OR IGNORE INTO validators(id, name, status, stake) VALUES (5, 'Sphere Labs', 'active', 98750.5);`, nil},
		{`INSERT OR IGNORE INTO validators(id, name, status, stake) VALUES (6, 'Quantum Ledger', 'active', 87600.3);`, nil},
		{`INSERT OR IGNORE INTO validators(id, name, status, stake) VALUES (7, 'Chronicle Node', 'standby', 64200.0);`, nil},
		{`INSERT OR IGNORE INTO validators(id, name, status, stake) VALUES (8, 'Merkle Validator', 'standby', 52100.0);`, nil},
		// Learning resources
		{`INSERT OR IGNORE INTO learn_resources(title, summary, url) VALUES ('Sphere Academy', 'Guides and courses for entrepreneurs, builders, and AI researchers.', 'https://sentience.thesphere.online/learn');`, nil},
		{`INSERT OR IGNORE INTO learn_resources(title, summary, url) VALUES ('Sentience Docs', 'Documentation, roadmap, and technical notes for the Sphere network.', 'https://sentience.thesphere.online/docs');`, nil},
		{`INSERT OR IGNORE INTO learn_resources(title, summary, url) VALUES ('Blockchain Basics', 'Introduction to distributed consensus and smart contracts.', 'https://sentience.thesphere.online/blockchain');`, nil},
		{`INSERT OR IGNORE INTO learn_resources(title, summary, url) VALUES ('Getting Started', 'Quick start guide for new developers joining the Sphere ecosystem.', 'https://sentience.thesphere.online/start');`, nil},
		{`INSERT OR IGNORE INTO learn_resources(title, summary, url) VALUES ('API Reference', 'Complete API documentation for building on Sphere.', 'https://sentience.thesphere.online/api');`, nil},
		// Investment options
		{`INSERT OR IGNORE INTO invest_options(title, summary, target_amount) VALUES ('Founders Fund', 'Early access investment for core ecosystem growth.', '5,000,000 SPR');`, nil},
		{`INSERT OR IGNORE INTO invest_options(title, summary, target_amount) VALUES ('Community Drive', 'A community-backed fund for platform adoption and education.', '1,200,000 SPR');`, nil},
		{`INSERT OR IGNORE INTO invest_options(title, summary, target_amount) VALUES ('Developer Grants', 'Funding for independent developers building on Sphere.', '750,000 SPR');`, nil},
		{`INSERT OR IGNORE INTO invest_options(title, summary, target_amount) VALUES ('Research Initiative', 'Support advanced research in distributed systems and AI.', '2,500,000 SPR');`, nil},
		// Minecraft servers
		{`INSERT OR IGNORE INTO minecraft_servers(name, host, status, description) VALUES ('Sphere EDU', 'minecraft.thesphere.online', 'online', 'An educative Minecraft world built around the Sphere mission. Join players worldwide in learning blockchain concepts through gameplay.');`, nil},
		{`INSERT OR IGNORE INTO minecraft_servers(name, host, status, description) VALUES ('Sentience Creative', 'creative.thesphere.online', 'online', 'A creative sandbox world for builders and artists. Design your vision without limitations.');`, nil},
		{`INSERT OR IGNORE INTO minecraft_servers(name, host, status, description) VALUES ('Cortex Survival', 'survival.thesphere.online', 'online', 'A challenging survival world for experienced players. Test your skills and join the community.');`, nil},
		{`INSERT OR IGNORE INTO minecraft_servers(name, host, status, description) VALUES ('Sphere PvP Arena', 'pvp.thesphere.online', 'online', 'Competitive PvP mode for skilled combat players. Weekly tournaments and rewards.');`, nil},
	}

	for _, entry := range entries {
		if _, err := s.db.Exec(entry.query); err != nil {
			return err
		}
	}

	// Seed blockchain with 50 blocks of history
	blockSeeds := []struct {
		height    int64
		hash      string
		validator string
	}{
		{1, "0x0000004f3e2a1b8c7d6e5f4a3b2c1d0e9f8a7b6c", "Sphere Node Alpha"},
		{2, "0x00000140a9c8d7e6f5a4b3c2d1e0f9a8b7c6d5e4", "Sentience Validator"},
		{3, "0x000002502d3c4b5a6f7e8d9c0a1b2c3d4e5f6a7b", "Cortex Prime"},
		{4, "0x00000361c2b1a09f8e7d6c5b4a39f8e7d6c5b4a3", "Sphere Node Alpha"},
		{5, "0x000004724b5a6f7e8d9c0a1b2c3d4e5f6a7b8c9d", "Neural Node"},
		{6, "0x000005836c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f", "Sentience Validator"},
		{7, "0x0000069424b5a6f7e8d9c0a1b2c3d4e5f6a7b8c9", "Cortex Prime"},
		{8, "0x000007a5c3d2e1f0a9b8c7d6e5f4a3b2c1d0e9f8", "Sphere Node Alpha"},
		{9, "0x000008b6d4e3f2a1b0c9d8e7f6a5b4c3d2e1f0a9", "Sentience Validator"},
		{10, "0x000009c7e5f4a3b2c1d0e9f8a7b6c5d4e3f2a1b0", "Neural Node"},
		{11, "0x00000ad8f6a5b4c3d2e1f0a9b8c7d6e5f4a3b2c1", "Quantum Ledger"},
		{12, "0x00000be9a7b6c5d4e3f2a1b0c9d8e7f6a5b4c3d2", "Sphere Labs"},
		{13, "0x00000cfae8c7d6e5f4a3b2c1d0e9f8a7b6c5d4e3", "Chronicle Node"},
		{14, "0x00000db0f9d8e7f6a5b4c3d2e1f0a9b8c7d6e5f4", "Merkle Validator"},
		{15, "0x00000ec18ac9d8e7f6a5b4c3d2e1f0a9b8c7d6e5", "Sphere Node Alpha"},
		{16, "0x00000fd29bd0e9f8a7b6c5d4e3f2a1b0c9d8e7f6", "Sentience Validator"},
		{17, "0x0000108e3ae1f0a9b8c7d6e5f4a3b2c1d0e9f8a7", "Cortex Prime"},
		{18, "0x000011954bf2a1b0c9d8e7f6a5b4c3d2e1f0a9b8", "Neural Node"},
		{19, "0x0000128c5ca3b2c1d0e9f8a7b6c5d4e3f2a1b0c9", "Quantum Ledger"},
		{20, "0x0000138356d4b3c2d1e0f9a8b7c6d5e4f3a2b1c0", "Sphere Labs"},
		{21, "0x000014aae7e5f4a3b2c1d0e9f8a7b6c5d4e3f2a1", "Chronicle Node"},
		{22, "0x000015bbf8f6a5b4c3d2e1f0a9b8c7d6e5f4a3b2", "Merkle Validator"},
		{23, "0x000016cc0a7b6c5d4e3f2a1b0c9d8e7f6a5b4c3d", "Sphere Node Alpha"},
		{24, "0x000017dd1b8c7d6e5f4a3b2c1d0e9f8a7b6c5d4e", "Sentience Validator"},
		{25, "0x000018ee2c9d8e7f6a5b4c3d2e1f0a9b8c7d6e5f", "Cortex Prime"},
		{26, "0x000019ff3dae9f8a7b6c5d4e3f2a1b0c9d8e7f6a", "Neural Node"},
		{27, "0x00001aa04ebf0a9b8c7d6e5f4a3b2c1d0e9f8a7b", "Quantum Ledger"},
		{28, "0x00001bb15fd1b0c9d8e7f6a5b4c3d2e1f0a9b8c7", "Sphere Labs"},
		{29, "0x00001cc260e2c1d0e9f8a7b6c5d4e3f2a1b0c9d8", "Chronicle Node"},
		{30, "0x00001dd371f3d2e1f0a9b8c7d6e5f4a3b2c1d0e9", "Merkle Validator"},
		{31, "0x00001ee4824e3f2a1b0c9d8e7f6a5b4c3d2e1f0a", "Sphere Node Alpha"},
		{32, "0x00001ff5935f4a3b2c1d0e9f8a7b6c5d4e3f2a1b", "Sentience Validator"},
		{33, "0x0000201a470a5b4c3d2e1f0a9b8c7d6e5f4a3b2c", "Cortex Prime"},
		{34, "0x000021155b1b6c5d4e3f2a1b0c9d8e7f6a5b4c3d", "Neural Node"},
		{35, "0x000022206c2c7d6e5f4a3b2c1d0e9f8a7b6c5d4e", "Quantum Ledger"},
		{36, "0x000023317d3d8e7f6a5b4c3d2e1f0a9b8c7d6e5f", "Sphere Labs"},
		{37, "0x0000244e8e4e9f8a7b6c5d4e3f2a1b0c9d8e7f6a", "Chronicle Node"},
		{38, "0x000025595f5fa0b9c8d7e6f5a4b3c2d1e0f9a8b7", "Merkle Validator"},
		{39, "0x0000266aa6b1b0c9d8e7f6a5b4c3d2e1f0a9b8c7", "Sphere Node Alpha"},
		{40, "0x000027775ac2c1d0e9f8a7b6c5d4e3f2a1b0c9d8", "Sentience Validator"},
		{41, "0x0000288ab3d3d2e1f0a9b8c7d6e5f4a3b2c1d0e9", "Cortex Prime"},
		{42, "0x000029a94e4e3f2a1b0c9d8e7f6a5b4c3d2e1f0a", "Neural Node"},
		{43, "0x00002abaff5f4a3b2c1d0e9f8a7b6c5d4e3f2a1b", "Quantum Ledger"},
		{44, "0x00002bcb07606b5c4d5e6f7a8b9c0d1e2f3a4b5c", "Sphere Labs"},
		{45, "0x00002cdc1871c7d8e9faab0b1c2d3e4f5a6b7c8d", "Chronicle Node"},
		{46, "0x00002ded29829d8e9fbaac1b2c3d4e5f6a7b8c9d", "Merkle Validator"},
		{47, "0x00002efea3a9aef9bcacbd1c2d3e4f5a6b7c8d9e", "Sphere Node Alpha"},
		{48, "0x00002f0fb4babf0acdcddee2f3a4b5c6d7e8f9a0", "Sentience Validator"},
		{49, "0x000030206cbbcc1baecddef3a4b5c6d7e8f9a0b1", "Cortex Prime"},
		{50, "0x000031317dccddc2bfdeeef4a5b6c7d8e9f0a1b2", "Neural Node"},
	}

	for i, block := range blockSeeds {
		offsetMinutes := int64(-(len(blockSeeds) - i - 1) * 15)
		_, err := s.db.Exec(
			`INSERT OR IGNORE INTO blocks(height, hash, timestamp, validator) VALUES (?, ?, datetime('now', ? || ' minutes'), ?)`,
			block.height, block.hash, offsetMinutes, block.validator,
		)
		if err != nil {
			return err
		}
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
		// Try multiple time format parsings
		block.Timestamp, _ = time.Parse(time.RFC3339Nano, raw)
		if block.Timestamp.IsZero() {
			block.Timestamp, _ = time.Parse(time.RFC3339, raw)
		}
		if block.Timestamp.IsZero() {
			block.Timestamp, _ = time.Parse("2006-01-02 15:04:05", raw)
		}
		if block.Timestamp.IsZero() {
			// Default to now if parsing fails
			block.Timestamp = time.Now().UTC()
		}
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
