package history

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"time"
	"usersegmentator/config"
)

type Repository interface {
	GetUserHistory(ctx context.Context, userID int, dates *DatesRange) ([]ReportRow, error)
	ParseAndValidateDates(dateStart, dateEnd string) (*DatesRange, error)
	CreateCSV(history []ReportRow) (string, error)
}

type historyRepository struct {
	db      *sql.DB
	cfg     *config.Config
	InfoLog *log.Logger
	ErrLog  *log.Logger
}

func NewHistoryRepo(db *sql.DB, cfg *config.Config) Repository {
	return &historyRepository{
		db:      db,
		cfg:     cfg,
		InfoLog: log.New(os.Stdout, "INFO\tREPORT REPO\t", log.Ldate|log.Ltime),
		ErrLog:  log.New(os.Stdout, "ERROR\tREPORT REPO\t", log.Ldate|log.Ltime),
	}
}

func (rr *historyRepository) ParseAndValidateDates(dateStart, dateEnd string) (*DatesRange, error) {
	dates := &DatesRange{}

	if !regexp.MustCompile(`^\d{4}-\d{1,2}$`).MatchString(dateStart) ||
		!regexp.MustCompile(`^\d{4}-\d{1,2}$`).MatchString(dateEnd) {
		return nil, fmt.Errorf("error validating dates range. The format is yyyy-mm or yyyy-m")
	}

	var err error
	if len(dateStart) == len(dateFormatFullMonth) {
		dates.StartDate, err = time.Parse("2006-01", dateStart)

		if err != nil {
			rr.ErrLog.Printf("Error validating date: %s", err)
			return nil, err
		}
	} else {
		dates.StartDate, err = time.Parse("2006-1", dateStart)

		if err != nil {
			rr.ErrLog.Printf("Error validating date: %s", err)
			return nil, err
		}
	}

	if len(dateEnd) == len(dateFormatFullMonth) {
		dates.EndDate, err = time.Parse("2006-01", dateEnd)

		if err != nil {
			rr.ErrLog.Printf("Error validating date: %s", err)
			return nil, err
		}
	} else {
		dates.EndDate, err = time.Parse("2006-1", dateEnd)

		if err != nil {
			rr.ErrLog.Printf("Error validating date: %s", err)
			return nil, err
		}
	}

	if err != nil {
		rr.ErrLog.Printf("Error validating date: %s", err)
		return nil, err
	}

	// 25.08 -> 25.09 -> 01.09
	dates.EndDate = dates.EndDate.AddDate(0, 1, 0)
	dates.EndDate = time.Date(
		dates.EndDate.Year(),
		dates.EndDate.Month(),
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)
	return dates, nil
}

func (rr *historyRepository) GetUserHistory(ctx context.Context, userID int, dates *DatesRange) ([]ReportRow, error) {
	history := []ReportRow{}
	rows, err := rr.db.QueryContext(
		ctx,
		`SELECT f.slug, ufr.date_assigned, ufr.date_unassigned 
		FROM user_segment_relation ufr 
		JOIN segments f ON ufr.segment_id = f.id 
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

		var historyRowUnassign ReportRow

		if dates.StartDate.Before(dateAssigned.Time) && dateAssigned.Time.Before(dates.EndDate) {
			historyRowAssign := ReportRow{
				UserID:    userID,
				Segment:   slug.String,
				Operation: "assigned",
				Date:      dateAssigned.Time.String(),
			}
			history = append(history, historyRowAssign)
		}

		if dateUnassigned.Valid && dates.EndDate.After(dateUnassigned.Time) &&
			dates.StartDate.Before(dateUnassigned.Time) {
			historyRowUnassign = ReportRow{
				UserID:    userID,
				Segment:   slug.String,
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

func (rr *historyRepository) CreateCSV(history []ReportRow) (string, error) {
	alpa := "abcdefghijklmnopqrstuvwxyz1234567890"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	randStr := make([]byte, fileIDLength)
	for i := range randStr {
		randStr[i] = alpa[r.Intn(len(alpa))]
	}

	fileName := rr.cfg.Report.FilePrefix + string(randStr) + rr.cfg.Report.FileExt
	filePath := rr.cfg.StorageDir + fileName

	file, err := os.Create(filePath)
	if err != nil {
		rr.ErrLog.Println(err.Error())
		return "", err
	}
	defer file.Close()

	fileData := ""
	for _, row := range history {
		fileData += fmt.Sprintf("%d;%s;%s;%s\n", row.UserID, row.Segment, row.Operation, row.Date)
	}

	_, err = file.Write([]byte(fileData))
	if err != nil {
		rr.ErrLog.Println(err.Error())
		return "", nil
	}

	fileURL := fmt.Sprintf("%s:%s/reports/%s", rr.cfg.HTTP.Host, rr.cfg.HTTP.Port, fileName)
	return fileURL, nil
}
