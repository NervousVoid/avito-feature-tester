package errors

import (
	"encoding/json"
	"net/http"
)

const (
	ErrorCantUnpackPayload   = "can't unpack payload"
	ErrorCantReadPayload     = "can't read payload"
	ErrorMarshalJson         = "json marshal error"
	ErrorResponseWrite       = "error writing response"
	ErrorBodyCloseError      = "error closing response body"
	ErrorInsertingDB         = "error writing to database"
	ErrorDeletingFromDB      = "error deleting row from database"
	ErrorNotFound            = "source was not found"
	ErrorGettingAffectedRows = "error getting rows affected"
	ErrorGettingLastID       = "Error getting last affected ID"
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
