// package db
//
// import (
// 	"database/sql"
// 	"fmt"
// 	"log"
// 	"os"
// 	"path/filepath"
//
// 	_ "modernc.org/sqlite"
// )
//
// var DB *sql.DB // this is a global variable that holds the database connection.
// // It is of type *sql.DB, which is a pointer to an sql.DB struct. The sql.DB struct
// // represents a database connection pool and provides methods for interacting with
// // the database.
//
// func InitDB(dbPath, dbName string) {
// 	// os.MkdirAll creates a directory named path, along with any necessary
// 	// parents, and returns nil, or else returns an error. The permission bits perm
// 	// (before umask) are used for all directories that MkdirAll creates. If path
// 	// is already a directory, MkdirAll does nothing and returns nil.
// 	err := os.MkdirAll(dbPath, os.ModePerm)
// 	if err != nil {
// 		log.Fatalf("failed to create db directory, %v", err)
// 	}
//
// 	dbFile := filepath.Join(dbPath, dbName)
// 	db, err := sql.Open("sqlite", dbFile) // dbFile is the path to the SQLite database file. If the file does not exist, it will be created.
// 	// e.g dbFile := "sqlite/dev/api.db" will create a SQLite database file named
// 	// api.db in the sqlite/dev directory.
// 	if err != nil {
// 		log.Fatalf("failed to open db, %v", err)
// 	}
//
// 	err = db.Ping() // Ping verifies a connection to the database is still
// 	// alive, establishing a connection if necessary. It returns an error if the
// 	// connection is not alive.
// 	if err != nil {
// 		log.Fatalf("failed to ping db, %v", err)
// 	}
//
// 	db.SetMaxOpenConns(30)
// 	db.SetMaxIdleConns(5)
//
// 	// sqlite specific settings
// 	// these pragmas are used to configure the behavior of the SQLite database.
// 	// They can improve performance and ensure data integrity.
// 	// the PRAGMA statements are executed using the Exec method of the sql.DB
// 	// struct.
// 	pragmas := []string{
// 		"PRAGMA journal_mode = WAL;",  // speeds up perform
// 		"PRAGMA busy_timeout = 5000;", // max time of 5 secs of one query to fininsh
// 		"PRAGMA foreign_keys= ON;",    // enable the foreign_keys
// 		"PRAGMA synchronous= NORMAL;", // FULL is safest but slow
// 	}
//
// 	for _, p := range pragmas {
// 		_, err := db.Exec(p)
// 		if err != nil {
// 			log.Fatalf("failed to execute: %s %v", p, err)
// 		}
// 	}
//
// 	tables := []string{
// 		// Users
// 		`CREATE TABLE IF NOT EXISTS users (
// 		id INTEGER PRIMARY KEY AUTOINCREMENT,
// 		email TEXT NOT NULL UNIQUE,
// 		name TEXT NOT NULL,
// 		password TEXT NOT NULL,
// 		refresh_token_web TEXT,
// 		refresh_token_web_at DATETIME,
// 		refresh_token_mobile TEXT,
// 		refresh_token_mobile_at DATETIME,
// 		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
// 	);`,
//
// 		// Privates
// 		`CREATE TABLE IF NOT EXISTS privates (
// 		id INTEGER PRIMARY KEY AUTOINCREMENT,
// 		user1_id INTEGER NOT NULL,
// 		user2_id INTEGER NOT NULL,
// 		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
// 		UNIQUE(user1_id, user2_id),
// 		CHECK(user1_id < user2_id),
// 		FOREIGN KEY(user1_id) REFERENCES users(id) ON DELETE CASCADE,
// 		FOREIGN KEY(user2_id) REFERENCES users(id) ON DELETE CASCADE
// 	);`,
//
// 		// Messages
// 		`CREATE TABLE IF NOT EXISTS messages (
// 		id INTEGER PRIMARY KEY AUTOINCREMENT,
// 		from_id INTEGER NOT NULL,
// 		private_id INTEGER,
// 		message_type TEXT NOT NULL,
// 		content TEXT NOT NULL,
// 		delivered INTEGER NOT NULL DEFAULT 0,
// 		read INTEGER NOT NULL DEFAULT 0,
// 		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
// 		FOREIGN KEY(from_id) REFERENCES users(id) ON DELETE CASCADE,
// 		FOREIGN KEY(private_id) REFERENCES privates(id) ON DELETE CASCADE
// 	);`,
// 	}
// 	for _, t := range tables {
// 		_, err := db.Exec(t)
// 		if err != nil {
// 			log.Fatalf("failed to create table: %s %v", t, err)
// 		}
// 	}
//
// 	indexes := []string{
// 		`CREATE INDEX IF NOT EXISTS idx_messages_private_id ON messages(private_id);`,
// 		`CREATE INDEX IF NOT EXISTS idx_messages_from_id ON messages(from_id);`,
// 		`CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);`,
// 		`CREATE INDEX IF NOT EXISTS idx_privates_user1_id ON privates(user1_id);`,
// 		`CREATE INDEX IF NOT EXISTS idx_privates_user2_id ON privates(user2_id);`,
// 	}
//
// 	for _, i := range indexes {
// 		_, err := db.Exec(i)
// 		if err != nil {
// 			log.Fatalf("failed to index table: %s %v", i, err)
// 		}
// 	}
//
// 	DB = db // assign the database connection to the global variable DB, so it
// 	// can be used throughout the application.
//
// 	log.Print("DB initialized")
// }
//
// func CloseDB() {
// 	if DB == nil {
// 		return
// 	}
//
// 	err := DB.Close()
// 	if err != nil {
// 		fmt.Println("Err closing db: ", err)
// 	} else {
// 		fmt.Println("Db closed!")
// 	}
// }

// with migrations:
package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	_ "modernc.org/sqlite"
)

var DB *sql.DB // package-level connection, set by InitDB

func InitDB(dbPath, dbName string) (*sql.DB, error) {
	if err := os.MkdirAll(dbPath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	dbFile := filepath.Join(dbPath, dbName)
	conn, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)

	pragmas := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA busy_timeout = 5000;",
		"PRAGMA foreign_keys = ON;",
		"PRAGMA synchronous = NORMAL;",
	}
	for _, p := range pragmas {
		if _, err := conn.Exec(p); err != nil {
			conn.Close()
			return nil, fmt.Errorf("exec pragma %q: %w", p, err)
		}
	}

	DB = conn // set the package-level var so models/* can use db.DB
	return conn, nil
}

func CloseDB(db *sql.DB) error {
	if db == nil {
		return nil
	}
	return db.Close()
}

// func RunMigrations(dbFile, migrationsPath string) error {
// 	m, err := migrate.New(
// 		fmt.Sprintf("file://%s", migrationsPath),
// 		fmt.Sprintf("sqlite://%s", dbFile),
// 	)
// 	if err != nil {
// 		return fmt.Errorf("init migrator: %w", err)
// 	}
// 	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
// 		return fmt.Errorf("run migrations: %w", err)
// 	}
// 	return nil
// }

func RunMigrations(dbFile, migrationsPath string) error {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		fmt.Sprintf("sqlite://%s", dbFile),
	)
	if err != nil {
		return fmt.Errorf("init migrator: %w", err)
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("migrations: no pending migrations, database already up to date")
			return nil
		}
		return fmt.Errorf("run migrations: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil {
		return fmt.Errorf("get migration version: %w", err)
	}

	log.Printf("migrations: applied successfully, now at version %d (dirty=%v)", version, dirty)

	return nil
}
