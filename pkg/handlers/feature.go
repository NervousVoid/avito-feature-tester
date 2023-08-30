package handlers

import (
	"database/sql"
	"encoding/json"
	"featuretester/pkg/errors"
	"featuretester/pkg/feature"
	"log"
	"math"
	"net/http"
	"os"
)

type FeaturesHandler struct {
	FeaturesRepo feature.Repository
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

func (fh *FeaturesHandler) AutoAssignFeature(w http.ResponseWriter, r *http.Request) {
	f := &feature.Template{}

	err := errors.ValidateAndParseJSON(r, f)
	if err != nil {
		fh.ErrLog.Printf("%s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if f.Fraction < 1 || f.Fraction > 100 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	activeUsers, err := fh.FeaturesRepo.GetActiveUsersAmount(r.Context())
	if err != nil {
		fh.ErrLog.Printf("%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sampleSize := int(math.Ceil(float64(activeUsers) * (float64(f.Fraction) / 100))) //nolint:gomnd // creating percents

	users, err := fh.FeaturesRepo.GetNRandomUsersWithoutFeature(sampleSize, f.FeatureSlug)
	if err != nil {
		fh.ErrLog.Printf("%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = fh.FeaturesRepo.AssignFeatures(r.Context(), users, []string{f.FeatureSlug})
	if err != nil {
		fh.ErrLog.Printf("%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (fh *FeaturesHandler) AddFeature(w http.ResponseWriter, r *http.Request) {
	f := &feature.Template{}

	err := errors.ValidateAndParseJSON(r, f)
	if err != nil {
		fh.ErrLog.Printf("%s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = fh.FeaturesRepo.InsertFeature(r.Context(), f.FeatureSlug)
	if err != nil {
		fh.ErrLog.Printf("%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (fh *FeaturesHandler) DeleteFeature(w http.ResponseWriter, r *http.Request) {
	f := &feature.Template{}

	err := errors.ValidateAndParseJSON(r, f)
	if err != nil {
		fh.ErrLog.Printf("%s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = fh.FeaturesRepo.DeleteFeature(r.Context(), f.FeatureSlug)
	if err != nil {
		fh.ErrLog.Printf("%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (fh *FeaturesHandler) UpdateUserFeatures(w http.ResponseWriter, r *http.Request) {
	f := &feature.Template{}

	err := errors.ValidateAndParseJSON(r, f)
	if err != nil {
		fh.ErrLog.Printf("%s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = fh.FeaturesRepo.AssignFeatures(r.Context(), []int{f.UserID}, f.AssignFeatures)
	if err != nil {
		fh.ErrLog.Printf("%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = fh.FeaturesRepo.UnassignFeatures(r.Context(), []int{f.UserID}, f.UnassignFeatures)
	if err != nil {
		fh.ErrLog.Printf("%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (fh *FeaturesHandler) GetUserFeatures(w http.ResponseWriter, r *http.Request) {
	receivedUserID := &feature.Template{}

	err := errors.ValidateAndParseJSON(r, receivedUserID)
	if err != nil {
		fh.ErrLog.Printf("%s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userFeatures, err := fh.FeaturesRepo.GetUserFeatures(r.Context(), receivedUserID.UserID)
	if err != nil {
		fh.ErrLog.Printf("%s", err)
		w.WriteHeader(http.StatusInternalServerError)
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
