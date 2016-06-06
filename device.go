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
	platform := req.FormValue("platform")
	name := req.FormValue("name")

	if name == "" || !models.DevicePlatforms[platform] {
		httputils.NewBadRequest("Please specify device name and a valid platform").WriteJSONResponse(resp)
		return
	}

	user := context.Get(req, ContextKeyUser)
	if user == nil || user.(models.User).ID == 0 {
		httputils.NewInternalServerError("No user object found").WriteJSONResponse(resp)
		return
	}

	device := models.Device{
		UserID:   user.(models.User).ID,
		Platform: platform,
		Name:     name,
		Options: models.DeviceOptions{
			PushNotifications: true,
		},
	}

	if err := device.Create(DB); err != nil {
		log.Println("Could not append device", err.Error())
		httputils.NewInternalServerError("Creating the device failed").WriteJSONResponse(resp)
		return
	}

	response := httputils.NewSuccess("")
	response.Data = map[string]string{
		"device_id": device.ID,
	}
	response.WriteJSONResponse(resp)
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
