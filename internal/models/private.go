package models

import (
	"database/sql"
	"errors"
	"time"

	"golangchatapp/internal/db"
)

type Private struct {
	ID        int64     `json:"id"`
	User1     int64     `json:"user1"`
	User2     int64     `json:"user2"`
	CreatedAt time.Time `json:"created_at"`
}

func GetPrivateById(privateId int64) (*Private, error) {
	var p Private

	err := db.DB.QueryRow(
		`
		SELECT id, user1_id, user2_id, created_at
		FROM privates
		WHERE id = ?
		`,
		privateId,
	).Scan(&p.ID, &p.User1, &p.User2, &p.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &p, nil
}

func GetPrivateByUsers(user1Id, user2Id int64) (*Private, error) {
	if user1Id > user2Id {
		user1Id, user2Id = user2Id, user1Id
	}

	var p Private

	err := db.DB.QueryRow(
		`
		SELECT id, user1_id, user2_id, created_at
		FROM privates
		WHERE user1_id = ? AND user2_id = ?
		`,
		user1Id, user2Id,
	).Scan(&p.ID, &p.User1, &p.User2, &p.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &p, nil
}

func GetPrivatesForUser(userId int64) ([]Private, error) {
	rows, err := db.DB.Query(
		`
		SELECT id, user1_id, user2_id, created_at
		FROM privates
		WHERE user1_id = ? OR user2_id = ?
		ORDER BY created_at DESC
		`,
		userId, userId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var privates []Private
	for rows.Next() {
		var p Private
		err := rows.Scan(&p.ID, &p.User1, &p.User2, &p.CreatedAt)
		if err != nil {
			return nil, err
		}

		privates = append(privates, p)
	}

	return privates, nil
}

func CreatePrivate(user1Id, user2Id int64) (*Private, error) {
	if user1Id == user2Id {
		return nil, errors.New("Cannot create private chat with the same user")
	}

	if user1Id > user2Id {
		user1Id, user2Id = user2Id, user1Id
	}

	existingPrivate, err := GetPrivateByUsers(user1Id, user2Id)
	if err != nil {
		return nil, err
	}

	if existingPrivate != nil {
		return existingPrivate, errors.New("Cannot create private, it already exists")
	}

	res, err := db.DB.Exec(`
	INSERT INTO privates (user1_id, user2_id) VALUES (?, ?)
	`, user1Id, user2Id)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	var createdAt time.Time
	err = db.DB.QueryRow(`SELECT created_at FROM privates WHERE id = ?`, id).Scan(&createdAt)
	if err != nil {
		return nil, err
	}

	return &Private{
		ID:        id,
		User1:     user1Id,
		User2:     user2Id,
		CreatedAt: createdAt,
	}, nil
}
