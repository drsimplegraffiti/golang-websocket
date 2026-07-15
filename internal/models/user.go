package models

import (
	"database/sql"
	"errors"
	"time"

	"golangchatapp/internal/db"
	"golangchatapp/internal/middlewares"
)

// *string/*time.Time because those DB columns are nullable — a pointer can be nil
// to represent SQL NULL, which a plain string/time.Time can't. Plain value types
// are used for Email, Name, etc. because those columns are NOT NULL, so no need to
// represent "missing." Bonus: Scan actually requires pointer types to scan NULL
// columns without erroring.

type User struct {
	ID                   int64      `json:"id"`
	Email                string     `json:"email"`
	Name                 string     `json:"name"`
	Password             string     `json:"-"`
	RefreshTokenWeb      *string    `json:"-"` // we need * here because
	RefreshTokenWebAt    *time.Time `json:"-"`
	RefreshTokenMobile   *string    `json:"-"`
	RefreshTokenMobileAt *time.Time `json:"-"`
	CreatedAt            time.Time  `json:"created_at"`
}

func CreateUserByEmail(name, email, hashedPassword string) (*User, error) {
	res, err := db.DB.Exec(`INSERT INTO users (name, email, password) VALUES (?,?,?)`,
		name, email, hashedPassword)
	if err != nil {
		return nil, err
	}

	id, _ := res.LastInsertId()

	createdAt := time.Now()

	return &User{ID: id, Name: name, Email: email, CreatedAt: createdAt}, nil
}

func GetUserByEmail(email string) (*User, error) {
	u := &User{}
	row := db.DB.QueryRow(`SELECT id, name,email, password, refresh_token_web,
    refresh_token_web_at, refresh_token_mobile, refresh_token_mobile_at,
    created_at FROM users where email = ?`, email)

	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.RefreshTokenWeb,
		&u.RefreshTokenWebAt, &u.RefreshTokenMobile, &u.RefreshTokenMobileAt,
		&u.CreatedAt)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func UpdateUserUserRefreshToken(userId int64, platform, refreshToken string) error {
	switch platform {
	case middlewares.PlatformWeb:
		_, err := db.DB.Exec(`UPDATE users SET refresh_token_web = ?, refresh_token_web_at =
        CURRENT_TIMESTAMP WHERE id = ?`, refreshToken, userId)
		return err
	case middlewares.PlatformMobile:
		_, err := db.DB.Exec(`UPDATE users SET refresh_token_mobile = ?, refresh_token_mobile_at =
        CURRENT_TIMESTAMP WHERE id = ?`, refreshToken, userId)
		return err

	default:
		return errors.New("invalid platform")

	}
}

func DeleteUserRefreshToken(userId int64, platform string) error {
	switch platform {
	case middlewares.PlatformWeb:
		_, err := db.DB.Exec(`UPDATE users SET refresh_token_web = NULL, refresh_token_web_at = NULL WHERE id = ?`, userId)
		return err
	case middlewares.PlatformMobile:
		_, err := db.DB.Exec(`UPDATE users SET refresh_token_mobile = NULL, refresh_token_mobile_at =
        NULL WHERE id = ?`, userId)
		return err

	default:
		return errors.New("invalid platform")

	}
}

func GetUserByRefreshToken(refreshToken, platform string) (*User, error) {
	u := &User{}
	var row *sql.Row

	switch platform {
	case middlewares.PlatformWeb:
		row = db.DB.QueryRow(`SELECT id, name,email, password, refresh_token_web,
            refresh_token_web_at, refresh_token_mobile, refresh_token_mobile_at,
            created_at FROM users where refresh_token_web = ?`, refreshToken)

	case middlewares.PlatformMobile:
		row = db.DB.QueryRow(`SELECT id, name,email, password, refresh_token_web,
            refresh_token_web_at, refresh_token_mobile, refresh_token_mobile_at,
            created_at FROM users where refresh_token_mobile = ?`, refreshToken)
	default:
		return nil, errors.New("invalid platform")
	}

	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.RefreshTokenWeb,
		&u.RefreshTokenWebAt, &u.RefreshTokenMobile, &u.RefreshTokenMobileAt,
		&u.CreatedAt)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func GetUserById(userId int64) (*User, error) {
	u := &User{}
	row := db.DB.QueryRow(`SELECT id, name,email, password, refresh_token_web,
    refresh_token_web_at, refresh_token_mobile, refresh_token_mobile_at,
    created_at FROM users where id= ?`, userId)

	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.RefreshTokenWeb,
		&u.RefreshTokenWebAt, &u.RefreshTokenMobile, &u.RefreshTokenMobileAt,
		&u.CreatedAt)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (u *User) ToMap() map[string]any {
	return map[string]any{
		"id":    u.ID,
		"name":  u.Name,
		"email": u.Email,
	}
}
