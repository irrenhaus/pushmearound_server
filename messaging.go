package main

import (
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

	// Up to 100M in memory
	if err := req.ParseMultipartForm(100 * 1024 * 1024); err != nil {
		log.Println("Failed to parse multipart form for message sending", err.Error())
		httputils.NewInternalServerError("Failed to parse multipart form").WriteJSONResponse(resp)
		return
	}

	msg := models.Message{
		User:   user,
		UserID: user.ID,
	}

	fromDeviceID := req.PostFormValue("device_id")
	destType := req.PostFormValue("dest_type")
	contentType := req.PostFormValue("content_type")

	msg.Title = req.PostFormValue("title")
	msg.Text = req.PostFormValue("text")
	msg.URL = req.PostFormValue("url")

	toDeviceID, err := strconv.ParseUint(req.PostFormValue("dest_id"), 10, 32)
	if err != nil {

	}

	device := models.Device{}
	if err := DB.First(&device, toDeviceID); err != nil {

	}
}
