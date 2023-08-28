package feature

import (
	"database/sql"
	"featuretester/pkg/errors"
	"log"
	"os"
)

type FeaturesRepo interface {
	InsertFeature(db *sql.DB, featureSlug string) error
	DeleteFeature(db *sql.DB, featureSlug string) error
}

type featuresRepo struct {
	InfoLog *log.Logger
	ErrLog  *log.Logger
}

func NewFeaturesRepo() FeaturesRepo {
	return &featuresRepo{
		InfoLog: log.New(os.Stdout, "INFO\tFEATURES REPO\t", log.Ldate|log.Ltime),
		ErrLog:  log.New(os.Stdout, "ERROR\tFEATURES REPO\t", log.Ldate|log.Ltime),
	}
}

func (fr *featuresRepo) InsertFeature(db *sql.DB, featureSlug string) error {
	result, err := db.Exec(
		"INSERT IGNORE INTO features (`slug`) VALUES (?)",
		featureSlug,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		fr.ErrLog.Printf("Error getting rows affected: %s", err)
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		fr.ErrLog.Printf("Error getting last insert ID: %s", err)
	}

	fr.InfoLog.Printf("AddFeature — RowsAffected: %d, LastInsertID: %d\n", affected, lastID)
	return nil
}

func (fr *featuresRepo) DeleteFeature(db *sql.DB, featureSlug string) error {
	result, err := db.Exec(
		"DELETE FROM features WHERE slug = ?",
		featureSlug,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		fr.ErrLog.Printf("%s: %s", errors.ErrorGettingAffectedRows, err)
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		fr.ErrLog.Printf("%s: %s", errors.ErrorGettingAffectedRows, err)
	}

	fr.InfoLog.Printf("DeleteFeature — RowsAffected: %d, LastInsertID: %d\n", affected, lastID)

	return nil
}
