package models

import (
	"time"

	"golangchatapp/internal/db"
)

type Message struct {
	ID          int64     `json:"id"`
	FromID      int64     `json:"from_id"`
	PrivateID   int64     `json:"private_id"`
	MessageType string    `json:"message_type"`
	Content     string    `json:"content"`
	Delivered   bool      `json:"delivered"`
	Read        bool      `json:"read"`
	CreatedAt   time.Time `json:"created_at"`
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func CreateMessage(m *Message) error {
	res, err := db.DB.Exec(
		`INSERT INTO messages (from_id, private_id, message_type, content, delivered, read)
		VALUES (?, ?, ?, ?, ?, ?)`,
		m.FromID, m.PrivateID, m.MessageType, m.Content, boolToInt(m.Delivered), boolToInt(m.Read),
	)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	m.ID = id

	return db.DB.QueryRow(`SELECT created_at FROM messages WHERE id = ?`, id).Scan(&m.CreatedAt)
}

func GetMessagesByPrivateID(privateID int64, page, limit int) ([]Message, error) {
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 20
	}

	offset := (page - 1) * limit

	rows, err := db.DB.Query(
		`SELECT id, from_id, private_id, message_type, content, delivered, read, created_at
		 FROM messages
		 WHERE private_id = ?
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`,
		privateID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message

	for rows.Next() {
		var m Message
		err := rows.Scan(&m.ID, &m.FromID, &m.PrivateID, &m.MessageType, &m.Content, &m.Delivered, &m.Read, &m.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}

	return messages, nil
}

func GetMessageByID(messageID int64) (*Message, error) {
	row := db.DB.QueryRow(
		`SELECT id, from_id, private_id, message_type, content, delivered, read, created_at 
		 FROM messages 
		 WHERE id = ?`, messageID,
	)

	var m Message

	err := row.Scan(&m.ID, &m.FromID, &m.PrivateID, &m.MessageType, &m.Content, &m.Delivered, &m.Read, &m.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func GetUndeliveredMessagesByPrivateID(privateID int64) ([]Message, error) {
	rows, err := db.DB.Query(
		`SELECT id, from_id, private_id, message_type, content, delivered, read, created_at
		 FROM messages
		 WHERE private_id = ? AND delivered = 0
		 ORDER BY created_at ASC`, privateID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message

	for rows.Next() {
		var m Message
		err := rows.Scan(&m.ID, &m.FromID, &m.PrivateID, &m.MessageType, &m.Content, &m.Delivered, &m.Read, &m.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}

	return messages, nil
}

func MarkMessageRead(messageID int64) error {
	_, err := db.DB.Exec(`UPDATE messages SET read = 1 WHERE id = ? AND read = 0`, messageID)
	return err
}

func MarkMessageDelivered(messageID int64) error {
	_, err := db.DB.Exec(`UPDATE messages SET delivered = 1 WHERE id = ? AND delivered = 0`, messageID)
	return err
}
