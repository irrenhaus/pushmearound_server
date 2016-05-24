package main

import (
	"fmt"
	"github.com/gorilla/context"
	"github.com/irrenhaus/pushmearound_server/httputils"
	"github.com/irrenhaus/pushmearound_server/models"
	"log"
	"net/http"
	"strconv"
)

func SendMessageHandler(resp http.ResponseWriter, req *http.Request) {
	user, ok := context.Get(req, ContextKeyUser).(models.User)
	if !ok {
		httputils.NewInternalServerError("Could not convert user").WriteJSONResponse(resp)
		return
	}

	// Store to 100M in memory
	if err := req.ParseMultipartForm(100 * 1024 * 1024); err != nil {
		log.Println("Failed to parse multipart form for message sending", err.Error())
		httputils.NewInternalServerError("Failed to parse multipart form").WriteJSONResponse(resp)
		return
	}

	msg := models.Message{
		User:   user,
		UserID: user.ID,
	}

	msg.Title = req.PostFormValue("title")
	msg.Text = req.PostFormValue("text")
	msg.URL = req.PostFormValue("url")

	contentType, err := strconv.ParseUint(req.PostFormValue("content_type"), 10, 32)
	if err != nil {
		httputils.NewBadRequest("Content type is not an integer").WriteJSONResponse(resp)
		return
	}
	msg.ContentType = uint(contentType)
	if msg.ContentType >= models.ContentTypeLast {
		httputils.NewBadRequest("Unknown content type").WriteJSONResponse(resp)
		return
	}

	// Parse and find destination device
	toDeviceID, err := strconv.ParseUint(req.PostFormValue("dest_device_id"), 10, 32)
	if err == nil {
		toDevice := models.Device{}
		if err := DB.First(&toDevice, toDeviceID).Error; err != nil {
			httputils.NewBadRequest(fmt.Sprintf("No such destination device: %d", toDeviceID)).WriteJSONResponse(resp)
			return
		}
		msg.Device = toDevice
		msg.DeviceID = toDevice.ID

		msg.User = toDevice.User
		msg.UserID = toDevice.User.ID
	} else {
		toUserID, err := strconv.ParseUint(req.PostFormValue("dest_user_id"), 10, 32)
		if err != nil {
			httputils.NewBadRequest("Destination device or user needed").WriteJSONResponse(resp)
			return
		}

		toUser := models.User{}
		if err := DB.First(&toUser, toUserID).Error; err != nil {
			httputils.NewBadRequest(fmt.Sprintf("No such destination user: %d", toUserID)).WriteJSONResponse(resp)
			return
		}

		msg.User = toUser
		msg.UserID = toUser.ID
	}

	fromDeviceID, err := strconv.ParseUint(req.PostFormValue("device_id"), 10, 32)
	if err != nil {
		httputils.NewBadRequest("Device ID is not an integer").WriteJSONResponse(resp)
		return
	}
	fromDevice := models.Device{}
	if err := DB.First(&fromDevice, fromDeviceID).Error; err != nil {
		httputils.NewBadRequest(fmt.Sprintf("No such source device: %d", fromDeviceID)).WriteJSONResponse(resp)
		return
	}
	msg.FromDevice = fromDevice
	msg.FromDeviceID = fromDevice.ID

	if !user.FriendsWith(&msg.User) {
		httputils.NewBadRequest(fmt.Sprintf("You're not allowed to send messages to user %d", msg.User.ID)).WriteJSONResponse(resp)
		return
	}

	if msg.DeviceID == 0 {
		var errs []error

		for _, d := range msg.User.Devices {
			msgClone := msg
			msgClone.Device = d
			msgClone.DeviceID = d.ID

			if err := DB.Create(&msgClone).Error; err != nil {
				log.Println("Sending message to", d.ID, err.Error())
				errs = append(errs, err)
			}
		}

		if len(errs) > 0 && len(errs) < len(msg.User.Devices) {
			httputils.NewSuccess(fmt.Sprintf("Message sending failed for %d devices", len(errs)))
		} else if len(errs) > 0 {
			httputils.NewInternalServerError("Sending the message failed")
		}
		return
	}

	if err := DB.Create(&msg).Error; err != nil {
		log.Println("Sending message to", msg.DeviceID, err.Error())
		httputils.NewInternalServerError("Sending the message failed")
		return
	}

	httputils.NewSuccess("Message sent").WriteJSONResponse(resp)
}
