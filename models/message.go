package models

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"time"
)

const (
	ContentTypeMessage = iota
	ContentTypeURL     = iota
	ContentTypeFile    = iota
	ContentTypeLast    = iota
)

type Message struct {
	ID             uint
	CreatedAt      time.Time
	LastModifiedAt time.Time
	UserID         uint
	DeviceID       string
	ContentType    uint
	Title          string
	Text           string
	URL            string
	File           string
}

type ReceivedMessage struct {
	ID        uint
	CreatedAt time.Time
	DeviceID  string
	MessageID uint
	Unread    bool
}

func scanMessage(msg *Message, rows *sql.Rows) error {
	return rows.Scan(&msg.ID, &msg.CreatedAt, &msg.LastModifiedAt, &msg.UserID, &msg.DeviceID, &msg.ContentType, &msg.Title, &msg.Text, &msg.URL, &msg.File)
}

func scanMultiMessages(msgs *[]Message, rows *sql.Rows) error {
	for rows.Next() {
		var msg Message
		if err := scanMessage(&msg, rows); err != nil {
			return err
		}
		*msgs = append(*msgs, msg)
	}

	return nil
}

func scanReceivedMessage(msg *ReceivedMessage, row *sql.Row) error {
	return row.Scan(&msg.ID, &msg.CreatedAt, &msg.DeviceID, &msg.MessageID, &msg.Unread)
}

func scanMultiReceivedMessages(msgs *[]ReceivedMessage, rows *sql.Rows) error {
	for rows.Next() {
		var msg ReceivedMessage
		err := rows.Scan(&msg.ID, &msg.CreatedAt, &msg.DeviceID, &msg.MessageID, &msg.Unread)
		if err != nil {
			return err
		}
		*msgs = append(*msgs, msg)
	}

	return nil
}

func FindMessagesByDevice(DB *sql.DB, deviceID string) ([]Message, error) {
	query := "SELECT * FROM messages WHERE device_id=$1"

	msgs := []Message{}
	rows, err := DB.Query(query, deviceID)
	if err != nil {
		return msgs, errors.New(fmt.Sprintf("Error finding Messages by device: %s", err.(*pq.Error).Code.Name()))
	}
	defer rows.Close()

	err = scanMultiMessages(&msgs, rows)

	return msgs, err
}

func FindReceivedMessage(DB *sql.DB, id uint) (ReceivedMessage, error) {
	query := "SELECT * FROM received_messages WHERE id=$1"

	row := DB.QueryRow(query, id)

	msg := ReceivedMessage{}
	err := scanReceivedMessage(&msg, row)

	return msg, err
}

func FindReceivedMessagesByDevice(DB *sql.DB, deviceID string) ([]ReceivedMessage, error) {
	query := "SELECT * FROM received_messages WHERE device_id=$1"

	msgs := []ReceivedMessage{}
	rows, err := DB.Query(query, deviceID)
	if err != nil {
		return msgs, errors.New(fmt.Sprintf("Error finding ReceivedMessages by device: %s", err.(*pq.Error).Code.Name()))
	}
	defer rows.Close()

	err = scanMultiReceivedMessages(&msgs, rows)

	return msgs, err
}

func (msg *Message) Create(DB *sql.DB) error {
	return DB.QueryRow("INSERT INTO messages(created_at, last_modified_at, user_id, device_id, content_type, title, text, url, file) VALUES (current_timestamp(), current_timestamp(), $1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at, last_modified_at", msg.UserID, msg.DeviceID, msg.ContentType, msg.Title, msg.Text, msg.URL, msg.File).Scan(&msg.ID, &msg.CreatedAt, &msg.LastModifiedAt)
}

func (msg *Message) Delete(DB *sql.DB) error {
	if msg.ID == 0 {
		return errors.New("Message object has no ID")
	}

	res, err := DB.Exec("DELETE FROM messages WHERE id=$1", msg.ID)

	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected <= 0 {
		return errors.New("No such database entry")
	}

	return nil
}

func FindUnreadReceivedMessagesByDevice(DB *sql.DB, deviceID string) ([]ReceivedMessage, error) {
	query := "SELECT * FROM received_messages WHERE device_id=$1 AND unread=false"

	msgs := []ReceivedMessage{}
	rows, err := DB.Query(query, deviceID)
	if err != nil {
		return msgs, errors.New(fmt.Sprintf("Error finding unread ReceivedMessages by device: %s", err.(*pq.Error).Code.Name()))
	}
	defer rows.Close()

	err = scanMultiReceivedMessages(&msgs, rows)

	return msgs, err
}

func FindUnreadReceivedMessagesByUser(DB *sql.DB, userID uint) ([]ReceivedMessage, error) {
	query := "SELECT * FROM received_messages JOIN devices on received_messages.device_id = devices.id WHERE devices.user_id=$1 AND unread=false"

	msgs := []ReceivedMessage{}
	rows, err := DB.Query(query, userID)
	if err != nil {
		return msgs, errors.New(fmt.Sprintf("Error finding unread ReceivedMessages by user: %s", err.(*pq.Error).Code.Name()))
	}
	defer rows.Close()

	err = scanMultiReceivedMessages(&msgs, rows)

	return msgs, err
}

func (rm *ReceivedMessage) Create(DB *sql.DB) error {
	return DB.QueryRow("INSERT INTO received_messages(created_at, device_id, message_id, unread) VALUES (current_timestamp(), $1, $2, true) RETURNING id, created_at, unread", rm.DeviceID, rm.MessageID).Scan(&rm.ID, &rm.CreatedAt, &rm.Unread)
}

func (rm *ReceivedMessage) Delete(DB *sql.DB) error {
	if rm.ID == 0 {
		return errors.New("ReceivedMessage object has no ID")
	}

	res, err := DB.Exec("DELETE FROM messages WHERE id=$1", rm.ID)

	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected <= 0 {
		return errors.New("No such database entry")
	}

	return nil
}
