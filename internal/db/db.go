package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(dbPath, dbName string) {
	err := os.MkdirAll(dbPath, os.ModePerm)
	if err != nil {
		log.Fatalf("failed to create db directory, %v", err)
	}

	dbFile := filepath.Join(dbPath, dbName)
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		log.Fatalf("failed to open db, %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("failed to ping db, %v", err)
	}

	db.SetMaxOpenConns(30)
	db.SetMaxIdleConns(5)

	// sqlite specific settings
	pragmas := []string{
		"PRAGMA journal_mode = WAL;",  // speeds up perform
		"PRAGMA busy_timeout = 5000;", // max time of 5 secs of one query to fininsh
		"PRAGMA foreign_keys= ON;",    // enable the foreign_keys
		"PRAGMA synchronous= NORMAL;", // FULL is safest but slow
	}

	for _, p := range pragmas {
		_, err := db.Exec(p)
		if err != nil {
			log.Fatalf("failed to execute: %s %v", p, err)
		}
	}

	tables := []string{
		// Users
		`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		password TEXT NOT NULL,
		refresh_token_web TEXT,
		refresh_token_web_at DATETIME,
		refresh_token_mobile TEXT,
		refresh_token_mobile_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`,

		// Privates
		`CREATE TABLE IF NOT EXISTS privates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user1_id INTEGER NOT NULL,
		user2_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user1_id, user2_id),
		CHECK(user1_id < user2_id),
		FOREIGN KEY(user1_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY(user2_id) REFERENCES users(id) ON DELETE CASCADE
	);`,

		// Messages
		`CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		from_id INTEGER NOT NULL,
		private_id INTEGER,
		message_type TEXT NOT NULL,
		content TEXT NOT NULL,
		delivered INTEGER NOT NULL DEFAULT 0,
		read INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(from_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY(private_id) REFERENCES privates(id) ON DELETE CASCADE
	);`,
	}
	for _, t := range tables {
		_, err := db.Exec(t)
		if err != nil {
			log.Fatalf("failed to create table: %s %v", t, err)
		}
	}

	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_messages_private_id ON messages(private_id);`,
		`CREATE INDEX IF NOT EXISTS idx_messages_from_id ON messages(from_id);`,
		`CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);`,
		`CREATE INDEX IF NOT EXISTS idx_privates_user1_id ON privates(user1_id);`,
		`CREATE INDEX IF NOT EXISTS idx_privates_user2_id ON privates(user2_id);`,
	}

	for _, i := range indexes {
		_, err := db.Exec(i)
		if err != nil {
			log.Fatalf("failed to index table: %s %v", i, err)
		}
	}

	DB = db

	log.Print("DB initialized")
}

func CloseDB() {
	if DB == nil {
		return
	}

	err := DB.Close()
	if err != nil {
		fmt.Println("Err closing db: ", err)
	} else {
		fmt.Println("Db closed!")
	}
}
