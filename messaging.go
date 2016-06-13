package main

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/irrenhaus/pushmearound_server/httpresponse"
	"github.com/irrenhaus/pushmearound_server/models"
	"github.com/satori/go.uuid"
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
		httpresponse.InternalServerError("Could not convert user").WriteJSON(resp)
		return
	}

	// Store to 100M in memory
	if err := req.ParseMultipartForm(100 * 1024 * 1024); err != nil {
		log.WithFields(log.Fields{"error": err}).Warn("Failed to parse multipart form for message sending")
		httpresponse.InternalServerError("Failed to parse multipart form").WriteJSON(resp)
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
		httpresponse.BadRequest("No such device").WriteJSON(resp)
		return
	}
	msg.DeviceID = device.ID

	msg.Title = req.PostFormValue("title")
	msg.Msg = req.PostFormValue("text")
	msg.URL = req.PostFormValue("url")

	if contentType, err = strconv.ParseUint(req.PostFormValue("content_type"), 10, 32); err != nil {
		httpresponse.BadRequest("content_type has to be an uint").WriteJSON(resp)
		return
	}

	if contentType >= models.ContentTypeLast {
		httpresponse.BadRequest("Unknown content type").WriteJSON(resp)
		return
	}

	msg.ContentType = uint(contentType)

	// File upload
	if msg.ContentType == models.ContentTypeFile {
		file, _, err := req.FormFile("sendfile")
		if err != nil {
			fmt.Println("File upload failed:")
			fmt.Println(err)
			httpresponse.BadRequest("Message type file selected without uploading file").WriteJSON(resp)
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
			httpresponse.InternalServerError("File upload failed").WriteJSON(resp)
			return
		}
	}

	if err := msg.Create(DB); err != nil {
		httpresponse.InternalServerError("Sending the message failed").WriteJSON(resp)
		os.Remove(msg.File)
		return
	}

	destinationDeviceID := req.PostFormValue("dest_id")
	if err == nil {
		// We have a destination device ID
		if sendMessageToDevice(msg, destinationDeviceID) == nil {
			httpresponse.InternalServerError("Sending the message failed").WriteJSON(resp)
			msg.Delete(DB)
			os.Remove(msg.File)
		}

		return
	}

	sentMessages := []*models.ReceivedMessage{}
	for _, device := range user.Devices {
		receivedMessage := sendMessageToDevice(msg, device.ID)
		if receivedMessage == nil {
			httpresponse.InternalServerError(fmt.Sprintf("Sending the message to the device %s failed", device.Name)).WriteJSON(resp)

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
		httpresponse.InternalServerError("Could not convert user").WriteJSON(resp)
		return
	}

	receivedMessages := []models.ReceivedMessage{}

	deviceID := req.FormValue("device")
	if deviceID != "" {
		// We have got a device ID

		err := user.LoadDevices(DB)
		if err != nil {
			if err != sql.ErrNoRows {
				log.WithFields(log.Fields{"user": user.ID, "device": deviceID}).Error("SQL error upon loading a users devices")
				httpresponse.InternalServerError("Error finding users devices").WriteJSON(resp)
				return
			}

			httpresponse.NotFound("No device found for your user").WriteJSON(resp)
			return
		}

		isUserDevice := false
		for _, d := range user.Devices {
			if d.ID == deviceID {
				isUserDevice = true
				break
			}
		}

		if !isUserDevice {
			httpresponse.NotFound("No such device").WriteJSON(resp)
			return
		}

		receivedMessages, err = models.FindUnreadReceivedMessagesByDevice(DB, deviceID)
		if err != nil {
			if err != sql.ErrNoRows {
				log.WithFields(log.Fields{"device": deviceID, "error": err}).Error("SQL error while searching unread messages")
			}

			httpresponse.NotFound("Could not find any unread messages for this device").WriteJSON(resp)
			return
		}
	} else {
		// No device ID, just collect all unread messages for this user
		var err error
		receivedMessages, err = models.FindUnreadReceivedMessagesByUser(DB, user.ID)
		if err != nil {
			fmt.Println("Could not find all unread messages for user:", err)
			httpresponse.InternalServerError("Could not find any unread messages for this user").WriteJSON(resp)
			return
		}
	}

	messageIDs := []uint{}
	for _, receivedMessage := range receivedMessages {
		messageIDs = append(messageIDs, receivedMessage.MessageID)
	}

	msgs, err := models.FindMessageList(DB, messageIDs)
	if err != nil {
		if err == sql.ErrNoRows {
			log.WithFields(log.Fields{"messages": msgs, "received": receivedMessages}).Error("Could not find messages for ReceivedMessages")
		} else {
			log.WithFields(log.Fields{"messages": msgs, "error": err}).Error("SQL error while loading messages for received messages")
		}

		httpresponse.NotFound("Could not find any unread messages for this device").WriteJSON(resp)
		return
	}

	response := httpresponse.Success("")
	response.Data = msgs
	response.WriteJSON(resp)
}

func UpdateMessageHandler(resp http.ResponseWriter, req *http.Request) {
	user, ok := context.Get(req, ContextKeyUser).(models.User)
	if !ok {
		httpresponse.InternalServerError("Could not convert user").WriteJSON(resp)
		return
	}

	vars := mux.Vars(req)
	msgId, err := strconv.ParseUint(vars["msg"], 10, 32)
	if err != nil {
		log.WithFields(log.Fields{"msgId": vars["msg"], "err": ok}).Warn("Could not parse message id")
		httpresponse.BadRequest("Invalid message ID").WriteJSON(resp)
		return
	}

	msg, err := models.FindReceivedMessageByMessage(DB, uint(msgId))
	if err != nil {
		if err != sql.ErrNoRows {
			log.WithFields(log.Fields{"user": user.ID, "msg": msgId}).Error("SQL error while searching for message to update")
		}

		httpresponse.NotFound("No such message found").WriteJSON(resp)
		return
	}

	device, err := models.FindDevice(DB, msg.DeviceID)
	if err != nil {
		if err != sql.ErrNoRows {
			log.WithFields(log.Fields{"user": user.ID, "msg": msg.ID, "device": msg.DeviceID}).Error("SQL error while searching for messages device")
		} else {
			log.WithFields(log.Fields{"user": user.ID, "msg": msg.ID, "device": msg.DeviceID}).Warn("No device found for message")
		}

		httpresponse.NotFound("Error resolving receiving device").WriteJSON(resp)
		return
	}

	if device.UserID != user.ID {
		httpresponse.NotFound("No such message").WriteJSON(resp)
		return
	}

	dirty := false

	unread, err := strconv.ParseBool(req.FormValue("unread"))
	if err != nil {
		msg.Unread = unread
	}

	if dirty {
		if err := msg.Update(DB); err != nil {
			log.WithFields(log.Fields{"msg": msg.ID, "error": err}).Error("SQL error while updating received message")
		}
	}
}
