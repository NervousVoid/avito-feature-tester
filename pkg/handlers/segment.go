package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"usersegmentator/pkg/errors"
	"usersegmentator/pkg/segment"
)

type SegmentsHandler struct {
	SegmentsRepo segment.Repository
	InfoLog      *log.Logger
	ErrLog       *log.Logger
}

func NewSegmentsHandler(db *sql.DB) *SegmentsHandler {
	return &SegmentsHandler{
		SegmentsRepo: segment.NewSegmentsRepo(db),
		InfoLog:      log.New(os.Stdout, "INFO\tSEGMENTS HANDLER\t", log.Ldate|log.Ltime),
		ErrLog:       log.New(os.Stdout, "ERROR\tSEGMENTS HANDLER\t", log.Ldate|log.Ltime),
	}
}

func (fh *SegmentsHandler) AutoAssignSegment(w http.ResponseWriter, r *http.Request) {
	f := &segment.Template{}

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

	activeUsers, err := fh.SegmentsRepo.GetActiveUsersAmount(r.Context())
	if err != nil {
		fh.ErrLog.Printf("%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sampleSize := int(math.Ceil(float64(activeUsers) * (float64(f.Fraction) / 100))) //nolint:gomnd // creating percents

	users, err := fh.SegmentsRepo.GetNRandomUsersWithoutSegment(sampleSize, f.SegmentSlug)
	if err != nil {
		fh.ErrLog.Printf("%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = fh.SegmentsRepo.AssignSegments(r.Context(), users, []string{f.SegmentSlug})
	if err != nil {
		fh.ErrLog.Printf("%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (fh *SegmentsHandler) AddSegment(w http.ResponseWriter, r *http.Request) {
	f := &segment.Template{}

	err := errors.ValidateAndParseJSON(r, f)
	if err != nil {
		fh.ErrLog.Printf("%s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = fh.SegmentsRepo.InsertSegment(r.Context(), f.SegmentSlug)
	if err != nil {
		fh.ErrLog.Printf("%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (fh *SegmentsHandler) DeleteSegment(w http.ResponseWriter, r *http.Request) {
	f := &segment.Template{}

	err := errors.ValidateAndParseJSON(r, f)
	if err != nil {
		fh.ErrLog.Printf("%s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = fh.SegmentsRepo.DeleteSegment(r.Context(), f.SegmentSlug)
	fmt.Println(f.SegmentSlug)
	if err != nil {
		fh.ErrLog.Printf("%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (fh *SegmentsHandler) UpdateUserSegments(w http.ResponseWriter, r *http.Request) {
	f := &segment.Template{}

	err := errors.ValidateAndParseJSON(r, f)
	if err != nil {
		fh.ErrLog.Printf("%s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = fh.SegmentsRepo.AssignSegments(r.Context(), []int{f.UserID}, f.AssignSegments)
	if err != nil {
		fh.ErrLog.Printf("%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = fh.SegmentsRepo.UnassignSegments(r.Context(), []int{f.UserID}, f.UnassignSegments)
	if err != nil {
		fh.ErrLog.Printf("%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (fh *SegmentsHandler) GetUserSegments(w http.ResponseWriter, r *http.Request) {
	receivedUserID := &segment.Template{}

	err := errors.ValidateAndParseJSON(r, receivedUserID)
	if err != nil {
		fh.ErrLog.Printf("%s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userSegments, err := fh.SegmentsRepo.GetUserSegments(r.Context(), receivedUserID.UserID)
	if err != nil {
		fh.ErrLog.Printf("%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(userSegments)
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
