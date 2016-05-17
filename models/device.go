package models

var deviceSchema = ``

type Device struct {
	User              *User
	DeviceId          string `db:device_id`
	Platform          string
	PushNotifications bool `db:push_notifications`
}
