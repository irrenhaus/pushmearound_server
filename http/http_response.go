package httpresponse

import "net/http"

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

func NewSuccess(msg string) HttpResponse {
	return HttpResponse{
		Status: http.StatusOK,
		Error:  "",
		Msg:    msg,
	}
}
