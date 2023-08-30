package history

import (
	"time"
)

const (
	fileIDLength         = 10
	dateFormatShortMonth = "2006-1"
	dateFormatFullMonth  = "2006-01"
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

type ReportRow struct {
	UserID    int
	Feature   string
	Operation string
	Date      string
}
