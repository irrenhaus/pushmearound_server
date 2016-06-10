package httpresponse

import (
	"encoding/json"
	"net/http"
)

type HttpResponse struct {
	Status int
	Error  string
	Msg    string
	Data   interface{}
}

func New() HttpResponse {
	return HttpResponse{
		Status: http.StatusOK,
	}
}

func Error(status int, err string) HttpResponse {
	return HttpResponse{
		Status: status,
		Error:  err,
	}
}

func BadRequest(err string) HttpResponse {
	return HttpResponse{
		Status: http.StatusBadRequest,
		Error:  err,
	}
}

func InternalServerError(err string) HttpResponse {
	return HttpResponse{
		Status: http.StatusInternalServerError,
		Error:  err,
	}
}

func Unauthorized(err string) HttpResponse {
	return HttpResponse{
		Status: http.StatusUnauthorized,
		Error:  err,
	}
}

func NotFound(err string) HttpResponse {
	return HttpResponse{
		Status: http.StatusNotFound,
		Error:  err,
	}
}

func Success(msg string) HttpResponse {
	return HttpResponse{
		Status: http.StatusOK,
		Msg:    msg,
	}
}

func (r HttpResponse) WriteJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	jsonContent, _ := json.Marshal(r)
	w.WriteHeader(r.Status)
	w.Write(jsonContent)
}
