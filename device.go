package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/irrenhaus/pushmearound_server/httpresponse"
	"github.com/irrenhaus/pushmearound_server/models"
)

func DeviceCreateHandler(resp http.ResponseWriter, req *http.Request) {
	platform := req.FormValue("platform")
	name := req.FormValue("name")

	if name == "" || !models.DevicePlatforms[platform] {
		httpresponse.BadRequest("Please specify device name and a valid platform").WriteJSON(resp)
		return
	}

	user, ok := context.Get(req, ContextKeyUser).(models.User)
	if !ok || user.ID == 0 {
		httpresponse.InternalServerError("No user object found").WriteJSON(resp)
		return
	}

	device := models.Device{
		UserID:   user.ID,
		Platform: platform,
		Name:     name,
		Options: models.DeviceOptions{
			PushNotifications: true,
		},
	}

	if err := device.Create(DB); err != nil {
		log.WithFields(log.Fields{"error": err, "device": device.Name}).Warn("Could not append device")
		httpresponse.InternalServerError("Creating the device failed").WriteJSON(resp)
		return
	}

	response := httpresponse.Success("")
	response.Data = map[string]string{
		"device_id": device.ID,
	}
	response.WriteJSON(resp)
}

func DeviceOptionsHandler(resp http.ResponseWriter, req *http.Request) {
	user, ok := context.Get(req, ContextKeyUser).(models.User)
	if !ok || user.ID == 0 {
		httpresponse.InternalServerError("No user object found").WriteJSON(resp)
		return
	}

	var options interface{}
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&options); err != nil {
		httpresponse.BadRequest("Could not parse JSON request").WriteJSON(resp)
		return
	}

	deviceID := req.FormValue("device")
	if deviceID == "" {
		httpresponse.BadRequest("No device specified").WriteJSON(resp)
		return
	}

	device, err := models.FindDevice(DB, deviceID)
	if err != nil {
		if err != sql.ErrNoRows {
			log.WithFields(log.Fields{"user": user.ID, "device": deviceID, "error": err}).Error("SQL error while loading device")
		}

		httpresponse.NotFound("No such device").WriteJSON(resp)
		return
	}

	if device.UserID != user.ID {
		httpresponse.NotFound("No such device").WriteJSON(resp)
		return
	}

	if err := device.Options.ParseJSONMap(options.(map[string]interface{})); err != nil {
		httpresponse.BadRequest(err.Error()).WriteJSON(resp)
		return
	}

	httpresponse.Success("").WriteJSON(resp)
}

func DeviceGetHandler(resp http.ResponseWriter, req *http.Request) {
	user, ok := context.Get(req, ContextKeyUser).(models.User)
	if !ok || user.ID == 0 {
		httpresponse.InternalServerError("No user object found").WriteJSON(resp)
		return
	}

	vars := mux.Vars(req)
	deviceID, ok := vars["id"]
	if !ok {
		httpresponse.BadRequest("Missing device ID parameter").WriteJSON(resp)
		return
	}

	device, err := models.FindDevice(DB, deviceID)
	if err != nil {
		if err != sql.ErrNoRows {
			log.WithFields(log.Fields{"user": user.ID, "device": deviceID, "error": err}).Error("SQL error while finding device")
		}

		httpresponse.NotFound("No such device").WriteJSON(resp)
		return
	}

	if device.UserID != user.ID {
		httpresponse.NotFound("No such device").WriteJSON(resp)
		return
	}

	response := httpresponse.New()
	response.Data = device
	response.WriteJSON(resp)
}
