package models

import (
	"database/sql"
	"errors"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/satori/go.uuid"
)

// Use a map so that we can check for existence without iterating over an array
var DevicePlatforms map[string]bool = map[string]bool{
	"chrome":  true,
	"firefox": true,
	"android": true,
}

type DeviceOptions struct {
	ID                uint
	DeviceID          string
	PushNotifications bool
}

type Device struct {
	ID               string
	CreatedAt        time.Time
	LastModifiedAt   time.Time
	UserID           uint
	Platform         string
	Name             string
	Options          DeviceOptions
	SentMessages     []Message
	ReceivedMessages []ReceivedMessage
}

func scanDeviceOptions(do *DeviceOptions, row *sql.Row) error {
	return row.Scan(&do.ID, &do.DeviceID, &do.PushNotifications)
}

func scanDevice(d *Device, row *sql.Row) error {
	return row.Scan(&d.ID, &d.CreatedAt, &d.LastModifiedAt, &d.UserID, &d.Platform, &d.Name)
}

func scanDevices(rows *sql.Rows) ([]Device, error) {
	devices := []Device{}
	for i := 0; rows.Next(); i++ {
		d := Device{}
		err := rows.Scan(&devices[i].ID, &devices[i].CreatedAt, &devices[i].LastModifiedAt, &devices[i].UserID, &devices[i].Platform, &devices[i].Name)
		if err != nil {
			log.Warn(err)
			continue
		}

		devices = append(devices, d)
	}

	if err := rows.Err(); err != nil {
		log.Error(err)
		return devices, err
	}

	return devices, nil
}

func FindDevice(DB *sql.DB, id string) (Device, error) {
	query := "SELECT * FROM devices WHERE id=$1"

	row := DB.QueryRow(query, id)
	device := Device{}
	err := scanDevice(&device, row)

	return device, err
}

func FindDevicesByUserID(DB *sql.DB, userID uint) ([]Device, error) {
	query := "SELECT * FROM devices WHERE user_id=$1"

	rows, err := DB.Query(query, userID)
	if err != nil {
		if err != sql.ErrNoRows {
			log.WithFields(log.Fields{"user_id": userID, "error": err}).Error("SQL error finding user's devices")
		}

		rows.Close()

		return []Device{}, err
	}
	devices, err := scanDevices(rows)

	return devices, err
}

func (d *Device) Create(DB *sql.DB) error {
	d.ID = uuid.NewV4().String()
	return DB.QueryRow("INSERT INTO devices(id, created_at, last_modified_at, user_id, platform, name) VALUES ($1, current_timestamp(), current_timestamp(), $2, $3, $4) RETURNING created_at, last_modified_at", d.ID, d.UserID, d.Platform, d.Name).Scan(&d.CreatedAt, &d.LastModifiedAt)
}

func (d *Device) LoadOptions(DB *sql.DB) error {
	query := "SELECT * FROM device_options WHERE device_id=$1"

	row := DB.QueryRow(query, d.ID)
	return scanDeviceOptions(&d.Options, row)
}

func (d *Device) LoadSentMessages(DB *sql.DB) error {
	sentMessages, err := FindMessagesByDevice(DB, d.ID)
	d.SentMessages = sentMessages

	return err
}

func (d *Device) LoadReceivedMessages(DB *sql.DB) error {
	receivedMessages, err := FindReceivedMessagesByDevice(DB, d.ID)
	d.ReceivedMessages = receivedMessages

	return err
}

func (d *DeviceOptions) ParseJSONMap(json map[string]interface{}) error {
	var ok bool
	d.DeviceID, ok = json["device_id"].(string)
	if !ok {
		return errors.New("No device ID specified")
	}

	pn, ok := json["push_notifications"]
	if ok {
		if d.PushNotifications, ok = pn.(bool); !ok {
			return errors.New("Push notifications value is not a boolean")
		}
	}

	return nil
}
