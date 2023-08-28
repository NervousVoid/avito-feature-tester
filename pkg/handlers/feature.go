package handlers

import (
	"database/sql"
	"encoding/json"
	"featuretester/pkg/errors"
	"featuretester/pkg/feature"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type FeaturesHandler struct {
	FeaturesRepo feature.FeaturesRepo
	DB           *sql.DB
	InfoLog      *log.Logger
	ErrLog       *log.Logger
}

func NewFeaturesHandler(db *sql.DB) *FeaturesHandler {
	return &FeaturesHandler{
		FeaturesRepo: feature.NewFeaturesRepo(db),
		InfoLog:      log.New(os.Stdout, "INFO\tFEATURES HANDLER\t", log.Ldate|log.Ltime),
		ErrLog:       log.New(os.Stdout, "ERROR\tFEATURES HANDLER\t", log.Ldate|log.Ltime),
	}
}

func (fh *FeaturesHandler) AddFeature(w http.ResponseWriter, r *http.Request) {
	receivedFeature := &feature.Template{}

	var body []byte
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errors.JSONError(w, r, http.StatusBadRequest, errors.ErrorCantReadPayload)
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

	err = fh.FeaturesRepo.InsertFeature(receivedFeature.FeatureSlug)
	if err != nil {
		fh.ErrLog.Printf("%s: %s", errors.ErrorInsertingDB, err)
		errors.JSONError(w, r, http.StatusBadRequest, errors.ErrorInsertingDB)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (fh *FeaturesHandler) DeleteFeature(w http.ResponseWriter, r *http.Request) {
	receivedFeature := &feature.Template{}

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

	err = fh.FeaturesRepo.DeleteFeature(receivedFeature.FeatureSlug)
	if err != nil {
		fh.ErrLog.Printf("%s: %s", errors.ErrorDeletingFromDB, err)
		errors.JSONError(w, r, http.StatusBadRequest, errors.ErrorDeletingFromDB)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (fh *FeaturesHandler) UpdateUserFeatures(w http.ResponseWriter, r *http.Request) {
	receivedFeatures := &feature.Template{}

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

	err = json.Unmarshal(body, receivedFeatures)
	if err != nil {
		errors.JSONError(w, r, http.StatusBadRequest, errors.ErrorCantUnpackPayload)
		return
	}

	err = fh.FeaturesRepo.AssignFeatures(receivedFeatures.UserID, receivedFeatures.AssignFeatures)
	if err != nil {
		fh.ErrLog.Printf("%s: %s", errors.ErrorInsertingDB, err)
		errors.JSONError(w, r, http.StatusBadRequest, errors.ErrorInsertingDB)
		return
	}

	err = fh.FeaturesRepo.UnassignFeatures(receivedFeatures.UserID, receivedFeatures.UnassignFeatures)
	if err != nil {
		fh.ErrLog.Printf("%s: %s", errors.ErrorDeletingFromDB, err)
		errors.JSONError(w, r, http.StatusBadRequest, errors.ErrorDeletingFromDB)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (fh *FeaturesHandler) GetUserFeatures(w http.ResponseWriter, r *http.Request) {
	receivedUserID := &feature.Template{}

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

	err = json.Unmarshal(body, receivedUserID)
	if err != nil {
		errors.JSONError(w, r, http.StatusBadRequest, errors.ErrorCantUnpackPayload)
		return
	}

	userFeatures, err := fh.FeaturesRepo.GetUserFeatures(r.Context(), receivedUserID.UserID)
	if err != nil {
		fh.ErrLog.Printf("%s: %s", errors.ErrorGettingDataFromBD, err)
		errors.JSONError(w, r, http.StatusBadRequest, errors.ErrorGettingDataFromBD)
		return
	}

	resp, err := json.Marshal(userFeatures)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
