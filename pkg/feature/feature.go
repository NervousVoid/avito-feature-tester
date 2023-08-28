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

type UserFeatures struct {
	UserID              int           `json:"user_id"`
	AddFeaturesSlugs    []interface{} `json:"add_features_slugs"`
	DeleteFeaturesSlugs []interface{} `json:"delete_features_slugs"`
}

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

func (h *Handler) DeleteFeature(w http.ResponseWriter, r *http.Request) {
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
		"DELETE FROM features WHERE slug = ?",
		receivedFeature.FeatureSlug,
	)
	if err != nil {
		fmt.Println(err)
		errors.JSONError(w, r, http.StatusInternalServerError, errors.ErrorDeletingFromDB)
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

	fmt.Printf("DeleteFeature — RowsAffected: %d, LastInsertID: %d\n", affected, lastID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UpdateUserFeatures(w http.ResponseWriter, r *http.Request) {
	receivedUserFeatures := &UserFeatures{}

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

		result, err := h.DB.Exec(
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
		removeFeaturesPayload = fmt.Sprintf(`DELETE FROM user_feature_relation WHERE userID = %d and featureID in (`, receivedUserFeatures.UserID)
		for pos, _ := range receivedUserFeatures.DeleteFeaturesSlugs {
			removeFeaturesPayload += "(SELECT id FROM features WHERE slug = ? LIMIT 1)"
			if pos < len(receivedUserFeatures.DeleteFeaturesSlugs)-1 {
				removeFeaturesPayload += ", "
			}
		}
		removeFeaturesPayload += ");"

		result, err := h.DB.Exec(
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

func (h *Handler) GetUserFeatures(w http.ResponseWriter, r *http.Request) {
	receivedUserID := &struct {
		UserID int `json:"user_id"`
	}{}

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

	rows, err := h.DB.QueryContext(
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
