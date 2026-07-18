CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	email TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,
	password TEXT NOT NULL,
	refresh_token_web TEXT,
	refresh_token_web_at DATETIME,
	refresh_token_mobile TEXT,
	refresh_token_mobile_at DATETIME,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
