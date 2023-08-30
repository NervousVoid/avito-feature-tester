package errors

import (
	"encoding/json"
	"io"
	"net/http"
)

const (
	ErrorBeginTransaction      = "error beginning transaction"
	ErrorGettingAffectedRows   = "error getting rows affected"
	ErrorGettingLastID         = "error getting last affected ID"
	ErrorGettingFeatureID      = "error getting feature id"
	ErrorCommittingTransaction = "error committing transaction"
)

func ValidateAndParseJSON(r *http.Request, parseInto interface{}) error {
	var body []byte
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = r.Body.Close()
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, parseInto)
	if err != nil {
		return err
	}

	return nil
}
