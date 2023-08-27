package errors

import (
	"encoding/json"
	"net/http"
)

const (
	ErrorCantUnpackPayload  = "can't unpack payload"
	ErrorCantReadPayload    = "can't read payload"
	ErrorMarshalError       = "marshal error"
	ErrorResponseWriteError = "error writing response"
	ErrorBodyCloseError     = "error closing response body"
	ErrorInsertingDB        = "error writing to database"
)

func JSONError(w http.ResponseWriter, r *http.Request, status int, msg string) {
	w.WriteHeader(status)
	resp, err := json.Marshal(map[string]interface{}{
		"error": msg,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
