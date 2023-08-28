package feature

import (
	"context"
	"database/sql"
	"featuretester/pkg/errors"
	"fmt"
	"log"
	"os"
)

type FeaturesRepo interface {
	InsertFeature(featureSlug string) error
	DeleteFeature(featureSlug string) error
	UnassignFeatures(userID int, featuresToUnassign []interface{}) error
	AssignFeatures(userID int, featuresToAssign []interface{}) error
	GetUserFeatures(ctx context.Context, userID int) (*Template, error)
}

type featuresRepo struct {
	db      *sql.DB
	InfoLog *log.Logger
	ErrLog  *log.Logger
}

func NewFeaturesRepo(db *sql.DB) FeaturesRepo {
	return &featuresRepo{
		db:      db,
		InfoLog: log.New(os.Stdout, "INFO\tFEATURES REPO\t", log.Ldate|log.Ltime),
		ErrLog:  log.New(os.Stdout, "ERROR\tFEATURES REPO\t", log.Ldate|log.Ltime),
	}
}

// INSERT IGNORE INTO features (`slug`) VALUES
// (SOME_SLUG_1),
// (SOME_SLUG_2),
// ...
// (SOME_SLUG_99);
func (fr *featuresRepo) InsertFeature(featureSlug string) error {
	result, err := fr.db.Exec(
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

// DELETE FROM features WHERE slug = SOME_SLUG
func (fr *featuresRepo) DeleteFeature(featureSlug string) error {
	result, err := fr.db.Exec(
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

// DELETE FROM user_feature_relation WHERE userID = 1234 AND featureID IN
// (
// (SELECT id FROM features WHERE slug = SOME_SLUG_1 LIMIT 1),
// ((SELECT id FROM features WHERE slug = SOME_SLUG_2 LIMIT 1),
// ...
// ((SELECT id FROM features WHERE slug = SOME_SLUG_99 LIMIT 1)
// )
func (fr *featuresRepo) UnassignFeatures(userID int, featuresToUnassign []interface{}) error {
	if len(featuresToUnassign) == 0 {
		return nil
	}

	payload := fmt.Sprintf(
		`DELETE FROM user_feature_relation WHERE userID = %d AND featureID IN (`,
		userID,
	)
	for pos, _ := range featuresToUnassign {
		payload += "(SELECT id FROM features WHERE slug = ?)"
		if pos < len(featuresToUnassign)-1 {
			payload += ", "
		}
	}
	payload += ")"

	result, err := fr.db.Exec(
		payload,
		featuresToUnassign...,
	)

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		fr.ErrLog.Printf("%s: %s", errors.ErrorGettingAffectedRows)
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		fr.ErrLog.Printf("%s: %s", errors.ErrorGettingLastID)
	}

	fr.InfoLog.Printf("UnassignFeatures — RowsAffected: %d, LastInsertID: %d\n", affected, lastID)
	return nil
}

// INSERT IGNORE INTO user_feature_relation VALUES
// (111, (SELECT id FROM features WHERE slug = SOME_SLUG_1)),
// (112, (SELECT id FROM features WHERE slug = SOME_SLUG_2)),
// ...
// (111, (SELECT id FROM features WHERE slug = SOME_SLUG_99))
func (fr *featuresRepo) AssignFeatures(userID int, featuresToAssign []interface{}) error {
	if len(featuresToAssign) == 0 {
		return nil
	}
	payload := "INSERT IGNORE INTO user_feature_relation VALUES "
	for pos, _ := range featuresToAssign {
		payload += fmt.Sprintf(`(%d, (SELECT id FROM features WHERE slug = ?))`, userID)
		if pos < len(featuresToAssign)-1 {
			payload += ", "
		}
	}

	result, err := fr.db.Exec(
		payload,
		featuresToAssign...,
	)

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		fr.ErrLog.Printf("%s: %s", errors.ErrorGettingAffectedRows)
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		fr.ErrLog.Printf("%s: %s", errors.ErrorGettingLastID)
	}

	fr.InfoLog.Printf("AssignFeatures — RowsAffected: %d, LastInsertID: %d\n", affected, lastID)
	return nil
}

func (fr *featuresRepo) GetUserFeatures(ctx context.Context, userID int) (*Template, error) {
	rows, err := fr.db.QueryContext(
		ctx,
		"SELECT slug FROM features WHERE id IN (SELECT featureID FROM user_feature_relation WHERE userID = ?)",
		userID,
	)
	if err != nil {
		return nil, err
	}

	userFeatures := &Template{
		UserID:   userID,
		Features: []string{},
	}

	for rows.Next() {
		var feature string
		err = rows.Scan(&feature)
		if err != nil {
			return nil, err
		}
		userFeatures.Features = append(userFeatures.Features, feature)
	}
	err = rows.Close()
	if err != nil {
		return nil, err
	}

	return userFeatures, nil
}
