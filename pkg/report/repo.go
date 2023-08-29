package report

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"time"
)

type ReportRepo interface {
	GetUserHistory(ctx context.Context, userID int, dates *DatesRange) ([]HistoryRow, error)
	ParseAndValidateDates(dateStart, dateEnd string) (*DatesRange, error)
	CreateCSV(history []HistoryRow) (string, error)
}

type reportRepo struct {
	db      *sql.DB
	InfoLog *log.Logger
	ErrLog  *log.Logger
}

func NewReportRepo(db *sql.DB) ReportRepo {
	return &reportRepo{
		db:      db,
		InfoLog: log.New(os.Stdout, "INFO\tREPORT REPO\t", log.Ldate|log.Ltime),
		ErrLog:  log.New(os.Stdout, "ERROR\tREPORT REPO\t", log.Ldate|log.Ltime),
	}
}

func (rr *reportRepo) ParseAndValidateDates(dateStart, dateEnd string) (*DatesRange, error) {
	dates := &DatesRange{}

	if !regexp.MustCompile(`^\d{4}-\d{1,2}$`).MatchString(dateStart) ||
		!regexp.MustCompile(`^\d{4}-\d{1,2}$`).MatchString(dateEnd) {
		return nil, fmt.Errorf("error validating dates range. The format is yyyy-mm or yyyy-m")
	}

	var err error
	if len(dateStart) == 7 {
		dates.StartDate, err = time.Parse("2006-01", dateStart)
	} else {
		dates.StartDate, err = time.Parse("2006-1", dateStart)
	}

	if len(dateEnd) == 7 {
		dates.EndDate, err = time.Parse("2006-01", dateEnd)
	} else {
		dates.EndDate, err = time.Parse("2006-1", dateEnd)
	}

	if err != nil {
		rr.ErrLog.Printf("Error validating date: %s", err)
		return nil, err
	}
	dates.EndDate = dates.EndDate.AddDate(0, 1, 0)
	return dates, nil
}

func (rr *reportRepo) GetUserHistory(ctx context.Context, userID int, dates *DatesRange) ([]HistoryRow, error) {
	history := []HistoryRow{}
	rows, err := rr.db.QueryContext(
		ctx,
		`SELECT f.slug, ufr.date_assigned, ufr.date_unassigned 
		FROM user_feature_relation ufr 
		JOIN features f ON ufr.feature_id = f.id 
		WHERE ufr.user_id = ? AND (
		ufr.date_assigned >= ? OR 
		(ufr.date_unassigned < ? OR ufr.date_unassigned IS NULL))`,
		userID,
		dates.StartDate.String(),
		dates.EndDate.String(),
	)

	if err != nil {
		rr.ErrLog.Println(err.Error())
		return nil, err
	}

	for rows.Next() {
		var slug sql.NullString
		var dateAssigned, dateUnassigned sql.NullTime
		err = rows.Scan(&slug, &dateAssigned, &dateUnassigned)
		if err != nil {
			rr.ErrLog.Println(err.Error())
			return nil, err
		}

		var historyRowUnassign HistoryRow

		if dates.StartDate.Before(dateAssigned.Time) && dateAssigned.Time.Before(dates.EndDate) {
			historyRowAssign := HistoryRow{
				UserID:    userID,
				Feature:   slug.String,
				Operation: "assigned",
				Date:      dateAssigned.Time.String(),
			}
			history = append(history, historyRowAssign)
		}

		if dateUnassigned.Valid && dates.EndDate.After(dateUnassigned.Time) &&
			dates.StartDate.Before(dateUnassigned.Time) {
			historyRowUnassign = HistoryRow{
				UserID:    userID,
				Feature:   slug.String,
				Operation: "unassigned",
				Date:      dateUnassigned.Time.String(),
			}
			history = append(history, historyRowUnassign)
		}
	}
	err = rows.Close()
	if err != nil {
		rr.ErrLog.Println(err.Error())
		return nil, err
	}
	return history, nil
}

func (rr *reportRepo) CreateCSV(history []HistoryRow) (string, error) {
	alpa := "abcdefghijklmnopqrstuvwxyz1234567890"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	randStr := make([]byte, 10)
	for i := range randStr {
		randStr[i] = alpa[r.Intn(len(alpa))]
	}

	fileName := "report_" + string(randStr) + ".csv"
	filePath := "static/reports/" + fileName
	file, err := os.Create(filePath)
	if err != nil {
		rr.ErrLog.Println(err.Error())
		return "", err
	}
	defer file.Close()

	fileData := ""
	for _, row := range history {
		fileData += fmt.Sprintf("%d;%s;%s;%s\n", row.UserID, row.Feature, row.Operation, row.Date)
	}

	_, err = file.Write([]byte(fileData))
	if err != nil {
		rr.ErrLog.Println(err.Error())
		return "", nil
	}

	fileURL := fmt.Sprintf("localhost:8000/reports/%s", fileName) // Replace with your domain and file path
	return fileURL, nil
}
