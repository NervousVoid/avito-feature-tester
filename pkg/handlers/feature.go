package handlers

import (
	"database/sql"
	"encoding/json"
	"featuretester/pkg/errors"
	"featuretester/pkg/feature"
	"fmt"
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
		FeaturesRepo: feature.NewFeaturesRepo(),
		DB:           db,
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

	err = fh.FeaturesRepo.InsertFeature(fh.DB, receivedFeature.FeatureSlug)
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

	err = fh.FeaturesRepo.DeleteFeature(fh.DB, receivedFeature.FeatureSlug)
	if err != nil {
		fh.ErrLog.Printf("%s: %s", errors.ErrorDeletingFromDB, err)
		errors.JSONError(w, r, http.StatusBadRequest, errors.ErrorDeletingFromDB)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (fh *FeaturesHandler) UpdateUserFeatures(w http.ResponseWriter, r *http.Request) {
	receivedUserFeatures := &feature.Template{}

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

	err = json.Unmarshal(body, receivedUserFeatures)
	if err != nil {
		errors.JSONError(w, r, http.StatusBadRequest, errors.ErrorCantUnpackPayload)
		return
	}

	var addFeaturesPayload, removeFeaturesPayload string

	if len(receivedUserFeatures.AddFeaturesSlugs) > 0 {
		addFeaturesPayload = "INSERT INTO user_feature_relation VALUES "
		for pos, _ := range receivedUserFeatures.AddFeaturesSlugs {
			addFeaturesPayload += fmt.Sprintf(`(%d, (SELECT id FROM features WHERE slug = ? LIMIT 1))`, receivedUserFeatures.UserID)
			if pos < len(receivedUserFeatures.AddFeaturesSlugs)-1 {
				addFeaturesPayload += ", "
			}
		}
		addFeaturesPayload += ";"

		result, err := fh.DB.Exec(
			addFeaturesPayload,
			receivedUserFeatures.AddFeaturesSlugs...,
		)

		if err != nil {
			fmt.Println(err)
			errors.JSONError(w, r, http.StatusNotFound, errors.ErrorNotFound)
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

		fmt.Printf("UpdateFeature Insert — RowsAffected: %d, LastInsertID: %d\n", affected, lastID)
	}

	if len(receivedUserFeatures.DeleteFeaturesSlugs) > 0 {
		removeFeaturesPayload = fmt.Sprintf(
			`DELETE FROM user_feature_relation WHERE userID = %d and featureID in (`,
			receivedUserFeatures.UserID,
		)
		for pos, _ := range receivedUserFeatures.DeleteFeaturesSlugs {
			removeFeaturesPayload += "(SELECT id FROM features WHERE slug = ? LIMIT 1)"
			if pos < len(receivedUserFeatures.DeleteFeaturesSlugs)-1 {
				removeFeaturesPayload += ", "
			}
		}
		removeFeaturesPayload += ");"

		result, err := fh.DB.Exec(
			removeFeaturesPayload,
			receivedUserFeatures.DeleteFeaturesSlugs...,
		)

		if err != nil {
			fmt.Println(err)
			errors.JSONError(w, r, http.StatusNotFound, errors.ErrorNotFound)
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

		fmt.Printf("UpdateFeature Remove — RowsAffected: %d, LastInsertID: %d\n", affected, lastID)
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

	rows, err := fh.DB.QueryContext(
		r.Context(),
		"SELECT slug FROM features WHERE id IN (SELECT featureID FROM user_feature_relation WHERE userID = ?)",
		receivedUserID.UserID,
	)
	if err != nil {
		errors.JSONError(w, r, http.StatusNotFound, errors.ErrorNotFound)
		return
	}

	userFeatures := &struct {
		UserID   int      `json:"userID"`
		Features []string `json:"features"`
	}{}

	for rows.Next() {
		var feature string
		err = rows.Scan(&feature)
		if err != nil {
			errors.JSONError(w, r, http.StatusInternalServerError, errors.ErrorResponseWrite)
			return
		}
		userFeatures.Features = append(userFeatures.Features, feature)
	}
	rows.Close()

	userFeatures.UserID = receivedUserID.UserID

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
