package main

import (
	"fmt"
	"github.com/gorilla/context"
	"github.com/irrenhaus/pushmearound_server/httputils"
	"github.com/irrenhaus/pushmearound_server/models"
	"github.com/satori/go.uuid"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

func sendMessageToDevice(msg models.Message, deviceID string) *models.ReceivedMessage {
	destinationDevice, err := models.FindDevice(DB, deviceID)
	if err != nil {
		return nil
	}

	receivedMessage := models.ReceivedMessage{
		DeviceID:  destinationDevice.ID,
		MessageID: msg.ID,
		Unread:    true,
	}

	if err := receivedMessage.Create(DB); err != nil {
		return nil
	}

	return &receivedMessage
}

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
		UserID: user.ID,
	}

	var contentType uint64
	var err error

	deviceID := req.PostFormValue("device_id")
	device, err := models.FindDevice(DB, deviceID)
	if err != nil {
		httputils.NewBadRequest("No such device").WriteJSONResponse(resp)
		return
	}
	msg.DeviceID = device.ID

	msg.Title = req.PostFormValue("title")
	msg.Text = req.PostFormValue("text")
	msg.URL = req.PostFormValue("url")

	if contentType, err = strconv.ParseUint(req.PostFormValue("content_type"), 10, 32); err != nil {
		httputils.NewBadRequest("content_type has to be an uint").WriteJSONResponse(resp)
		return
	}

	if contentType >= models.ContentTypeLast {
		httputils.NewBadRequest("Unknown content type").WriteJSONResponse(resp)
		return
	}

	msg.ContentType = uint(contentType)

	// File upload
	if msg.ContentType == models.ContentTypeFile {
		file, _, err := req.FormFile("sendfile")
		if err != nil {
			fmt.Println("File upload failed:")
			fmt.Println(err)
			httputils.NewBadRequest("Message type file selected without uploading file").WriteJSONResponse(resp)
			return
		}
		defer file.Close()

		msg.File = uuid.NewV4().String()

		f, err := os.OpenFile("./upload/"+msg.File, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()

		if _, err := io.Copy(f, file); err != nil {
			fmt.Println("Copying uploaded file failed:")
			fmt.Println(err)
			httputils.NewInternalServerError("File upload failed").WriteJSONResponse(resp)
			return
		}
	}

	if err := msg.Create(DB); err != nil {
		httputils.NewInternalServerError("Sending the message failed").WriteJSONResponse(resp)
		os.Remove(msg.File)
		return
	}

	destinationDeviceID := req.PostFormValue("dest_id")
	if err == nil {
		// We have a destination device ID
		if sendMessageToDevice(msg, destinationDeviceID) == nil {
			httputils.NewInternalServerError("Sending the message failed").WriteJSONResponse(resp)
			msg.Delete(DB)
			os.Remove(msg.File)
		}

		return
	}

	sentMessages := []*models.ReceivedMessage{}
	for _, device := range user.Devices {
		receivedMessage := sendMessageToDevice(msg, device.ID)
		if receivedMessage == nil {
			httputils.NewInternalServerError(fmt.Sprintf("Sending the message to the device %s failed", device.Name)).WriteJSONResponse(resp)

			for _, sentMessage := range sentMessages {
				sentMessage.Delete(DB)
			}

			msg.Delete(DB)
			os.Remove(msg.File)

			return
		}

		sentMessages = append(sentMessages, receivedMessage)
	}

	return
}

func UnreadMessageHandler(resp http.ResponseWriter, req *http.Request) {
	user, ok := context.Get(req, ContextKeyUser).(models.User)
	if !ok {
		httputils.NewInternalServerError("Could not convert user").WriteJSONResponse(resp)
		return
	}

	deviceID := req.FormValue("device")
	if deviceID != "" {
		// We have got a device ID
		msgs, err := models.FindUnreadReceivedMessagesByDevice(DB, deviceID)
		if err != nil {
			fmt.Println("Could not find unread messages:", err)
			httputils.NewInternalServerError("Could not find any unread messages for this device").WriteJSONResponse(resp)
			return
		}

		response := httputils.NewSuccess("")
		response.Data = msgs
		response.WriteJSONResponse(resp)
		return
	}

	// No device ID, just collect all unread messages for this user
	msgs, err := models.FindUnreadReceivedMessagesByUser(DB, user.ID)
	if err != nil {
		fmt.Println("Could not find all unread messages for user:", err)
		httputils.NewInternalServerError("Could not find any unread messages for this user").WriteJSONResponse(resp)
		return
	}

	response := httputils.NewSuccess("")
	response.Data = msgs
	response.WriteJSONResponse(resp)
}
