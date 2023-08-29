package handlers

import (
	"database/sql"
	"encoding/json"
	"featuretester/pkg/errors"
	"featuretester/pkg/report"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type ReportHandler struct {
	ReportRepo report.ReportRepo
	InfoLog    *log.Logger
	ErrLog     *log.Logger
}

func NewReportHandler(db *sql.DB) *ReportHandler {
	return &ReportHandler{
		ReportRepo: report.NewReportRepo(db),
		InfoLog:    log.New(os.Stdout, "INFO\tREPORT HANDLER\t", log.Ldate|log.Ltime),
		ErrLog:     log.New(os.Stdout, "ERROR\tREPORT HANDLER\t", log.Ldate|log.Ltime),
	}
}

func (rh *ReportHandler) GetFeatureHistory(w http.ResponseWriter, r *http.Request) {
	receivedRequest := &report.Request{}

	var body []byte
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		rh.ErrLog.Println(err.Error())
		errors.JSONError(w, r, http.StatusInternalServerError, errors.ErrorCantReadPayload)
		return
	}

	err = r.Body.Close()
	if err != nil {
		rh.ErrLog.Println(err.Error())
		errors.JSONError(w, r, http.StatusInternalServerError, errors.ErrorBodyCloseError)
		return
	}

	err = json.Unmarshal(body, receivedRequest)
	if err != nil {
		rh.ErrLog.Println(err.Error())
		errors.JSONError(w, r, http.StatusBadRequest, errors.ErrorCantUnpackPayload)
		return
	}

	dates, err := rh.ReportRepo.ParseAndValidateDates(receivedRequest.StartDate, receivedRequest.EndDate)
	if err != nil {
		rh.ErrLog.Println(err.Error())
		errors.JSONError(w, r, http.StatusBadRequest, errors.ErrorValidatingData)
		return
	}

	userHistory, err := rh.ReportRepo.GetUserHistory(r.Context(), receivedRequest.UserID, dates)
	if err != nil {
		rh.ErrLog.Println(err.Error())
		errors.JSONError(w, r, http.StatusBadRequest, errors.ErrorGettingDataFromDB)
		return
	}
	fmt.Println(userHistory)

	url, err := rh.ReportRepo.CreateCSV(userHistory)
	if err != nil {
		errors.JSONError(w, r, http.StatusBadRequest, errors.ErrorGettingDataFromDB)
		return
	}

	resp, err := json.Marshal(struct {
		CsvUrl string `json:"csv_url"`
	}{
		CsvUrl: url,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(resp)
	if err != nil {
		rh.ErrLog.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
