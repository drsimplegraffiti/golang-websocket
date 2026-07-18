go get github.com/ilyakaznacheev/cleanenv
go get modernc.org/sqlite
go get golang.org/x/crypto/bcrypt
go get github.com/golang-jwt/jwt/v5  


sqlite:/home/xybug/Downloads/golangchatapp/sqlite/dev/api.db

Add headers like:

![image](images/paste_1784133976.png)

Add message:

![image](images/paste_1784134062.png)


#### Serve the html
cd /path/to/your/html
python3 -m http.server 8000

Then open:
http://localhost:8000


Create a new folder outside the backend and make a `main.go`


Run: User A

go run main.go \
-url=ws://localhost:8082/api/ws \
-token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJuYW1lIjoiYWJheW9taSIsIlgtcGxhdGZvcm0iOiJ3ZWIiLCJzdWIiOiIxIiwiZXhwIjoxNzg0MjcwMjIwfQ.skTlZOJhNo02I8YxMtfJnP0lj9cuS28RlgZwKvnsIbU \
-name=Alice

Run: User B
go run main.go \
-token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyLCJuYW1lIjoibWltaSIsIlgtcGxhdGZvcm0iOiJ3ZWIiLCJzdWIiOiIyIiwiZXhwIjoxNzg0MjcwMzQ1fQ.pDOZ9DSp_oz9N9EeLP7CH3ynSGcC_2gsbyaRQmrrvO0 \
-name=Bob


Send a message

Copy and paste this into Alice's terminal:
{
  "event_type": "message",
  "payload": {
    "private_id": 1,
    "receiver_id": 2,
    "message_type": "text",
    "content": "Hello Bob!"
  }
}

Bob should receive something like:
{
  "event_type": "message",
  "payload": {
    "message": {
      ...
    }
  }
}



### migration
# Migration Setup

## 1. Scaffold files

mkdir -p migrations
migrate create -ext sql -dir migrations -seq create_users_table
migrate create -ext sql -dir migrations -seq create_privates_table
migrate create -ext sql -dir migrations -seq create_messages_table
migrate create -ext sql -dir migrations -seq add_indexes

## 2. migrations/000001_create_users_table.up.sql

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

## 3. migrations/000001_create_users_table.down.sql

DROP TABLE IF EXISTS users;

## 4. migrations/000002_create_privates_table.up.sql

CREATE TABLE IF NOT EXISTS privates (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user1_id INTEGER NOT NULL,
	user2_id INTEGER NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	UNIQUE(user1_id, user2_id),
	CHECK(user1_id < user2_id),
	FOREIGN KEY(user1_id) REFERENCES users(id) ON DELETE CASCADE,
	FOREIGN KEY(user2_id) REFERENCES users(id) ON DELETE CASCADE
);

## 5. migrations/000002_create_privates_table.down.sql

DROP TABLE IF EXISTS privates;

## 6. migrations/000003_create_messages_table.up.sql

CREATE TABLE IF NOT EXISTS messages (
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
);

## 7. migrations/000003_create_messages_table.down.sql

DROP TABLE IF EXISTS messages;

## 8. migrations/000004_add_indexes.up.sql

CREATE INDEX IF NOT EXISTS idx_messages_private_id ON messages(private_id);
CREATE INDEX IF NOT EXISTS idx_messages_from_id ON messages(from_id);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
CREATE INDEX IF NOT EXISTS idx_privates_user1_id ON privates(user1_id);
CREATE INDEX IF NOT EXISTS idx_privates_user2_id ON privates(user2_id);

## 9. migrations/000004_add_indexes.down.sql

DROP INDEX IF EXISTS idx_messages_private_id;
DROP INDEX IF EXISTS idx_messages_from_id;
DROP INDEX IF EXISTS idx_messages_created_at;
DROP INDEX IF EXISTS idx_privates_user1_id;
DROP INDEX IF EXISTS idx_privates_user2_id;

## 10. Run automatically (already wired into main.go)

go run ./cmd/api -config ./config/dev.env

## 11. Run manually via CLI (alternative to app-managed migrations)

migrate -path migrations -database "sqlite://./sqlite/prod/api.db" up

## 12. Roll back one step

migrate -path migrations -database "sqlite://./sqlite/prod/api.db" down 1

## 13. Add a future migration

migrate create -ext sql -dir migrations -seq add_message_edited_flag
