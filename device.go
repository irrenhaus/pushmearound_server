package main

import (
	"encoding/json"
	"github.com/gorilla/context"
	"github.com/irrenhaus/pushmearound_server/httputils"
	"github.com/irrenhaus/pushmearound_server/models"
	"log"
	"net/http"
)

func DeviceCreateHandler(resp http.ResponseWriter, req *http.Request) {
	deviceID := req.FormValue("device_id")
	platform := req.FormValue("platform")

	if deviceID == "" || !models.DevicePlatforms[platform] {
		httputils.NewBadRequest("Please specify device ID and a valid platform").WriteJSONResponse(resp)
		return
	}

	user := context.Get(req, ContextKeyUser)
	if user == nil || user.(models.User).ID == 0 {
		httputils.NewInternalServerError("No user object found").WriteJSONResponse(resp)
		return
	}

	device := models.Device{
		UserID:   user.(models.User).ID,
		DeviceID: deviceID,
		Platform: platform,
		Options: models.DeviceOptions{
			PushNotifications: true,
		},
	}

	if err := DB.Model(&user).Association("Devices").Append(&device).Error; err != nil {
		log.Println("Could not append device", err.Error())
		httputils.NewInternalServerError("Creating the device failed").WriteJSONResponse(resp)
		return
	}

	httputils.NewSuccess("success").WriteJSONResponse(resp)
}

func DeviceOptionsHandler(resp http.ResponseWriter, req *http.Request) {
	var options interface{}
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&options); err != nil {
		httputils.NewBadRequest("Could not parse JSON request").WriteJSONResponse(resp)
		return
	}

	optionsMap := options.(map[string]interface{})
	deviceOptions := models.DeviceOptions{}
	if err := deviceOptions.ParseJSONMap(optionsMap); err != nil {
		httputils.NewBadRequest(err.Error()).WriteJSONResponse(resp)
		return
	}
}
