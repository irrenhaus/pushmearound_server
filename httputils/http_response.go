package httputils

import (
	"encoding/json"
	"net/http"
)

type HttpResponse struct {
	Status int
	Error  string
	Msg    string
}

func NewError(status int, err string) HttpResponse {
	return HttpResponse{
		Status: status,
		Error:  err,
		Msg:    "",
	}
}

func NewBadRequest(err string) HttpResponse {
	return HttpResponse{
		Status: http.StatusBadRequest,
		Error:  err,
		Msg:    "",
	}
}

func NewInternalServerError(err string) HttpResponse {
	return HttpResponse{
		Status: http.StatusInternalServerError,
		Error:  err,
		Msg:    "",
	}
}

func NewUnauthorized(err string) HttpResponse {
	return HttpResponse{
		Status: http.StatusUnauthorized,
		Error:  err,
		Msg:    "",
	}
}

func NewSuccess(msg string) HttpResponse {
	return HttpResponse{
		Status: http.StatusOK,
		Error:  "",
		Msg:    msg,
	}
}

func (r HttpResponse) WriteJSONResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	jsonContent, _ := json.Marshal(r)
	http.Error(w, string(jsonContent), r.Status)
}
