package feature

import (
	"database/sql"
	"encoding/json"
	"featuretester/pkg/errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Feature struct {
	FeatureSlug string `json:"feature_slug"`
}

//type AddFeatureStruct struct {
//	UserID      string `json:"user_id"`
//	FeatureSlug string `json:"feature_slug"`
//}

type Handler struct {
	DB *sql.DB
}

func (h *Handler) AddFeature(w http.ResponseWriter, r *http.Request) {
	receivedFeature := &Feature{}

	var body []byte
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errors.JSONError(w, r, http.StatusInternalServerError, errors.ErrorCantReadPayload)
		return
	}

	err = r.Body.Close()
	if err != nil {
		errors.JSONError(w, r, http.StatusInternalServerError, errors.ErrorBodyCloseError)
		return
	}

	err = json.Unmarshal(body, receivedFeature)
	if err != nil {
		errors.JSONError(w, r, http.StatusBadRequest, errors.ErrorCantUnpackPayload)
		return
	}

	result, err := h.DB.Exec(
		"INSERT INTO features (`slug`) VALUES (?)",
		receivedFeature.FeatureSlug,
	)
	if err != nil {
		errors.JSONError(w, r, http.StatusInternalServerError, errors.ErrorInsertingDB)
		return
	}

	affected, err := result.RowsAffected()
	if err != nil {
		fmt.Println("Error getting rows affected")
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		fmt.Println("Error getting last insert ID")
	}

	fmt.Printf("AddFeature — RowsAffected: %d, LastInsertID: %d\n", affected, lastID)
	w.WriteHeader(http.StatusCreated)
}

func DeleteFeature(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Delete feature"))
}
