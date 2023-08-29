package report

import (
	"time"
)

type Request struct {
	UserID    int    `json:"user_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type DatesRange struct {
	StartDate time.Time
	EndDate   time.Time
}

type HistoryRow struct {
	UserID    int
	Feature   string
	Operation string
	Date      string
}
