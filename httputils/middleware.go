package httputils

import (
	"net/http"
)

type HttpHandler func(w http.ResponseWriter, r *http.Request)
