package feature

import (
	"context"
	"database/sql"
	"featuretester/pkg/errors"
	"fmt"
	"log"
	"os"
)

type Repository interface {
	InsertFeature(ctx context.Context, featureSlug string) error
	DeleteFeature(ctx context.Context, featureSlug string) error
	UnassignFeatures(ctx context.Context, userID []int, featuresToUnassign []string) error
	AssignFeatures(ctx context.Context, userID []int, featuresToAssign []string) error
	GetUserFeatures(ctx context.Context, userID int) (*Template, error)
	GetNRandomUsersWithoutFeature(n int, slug string) ([]int, error)
	GetActiveUsersAmount(ctx context.Context) (int, error)
	GetFeaturesIDs(ctx context.Context, featureSlugs []string) ([]int, error)
}

type featuresRepository struct {
	db      *sql.DB
	InfoLog *log.Logger
	ErrLog  *log.Logger
}

func NewFeaturesRepo(db *sql.DB) Repository {
	return &featuresRepository{
		db:      db,
		InfoLog: log.New(os.Stdout, "INFO\tFEATURES REPO\t", log.Ldate|log.Ltime),
		ErrLog:  log.New(os.Stdout, "ERROR\tFEATURES REPO\t", log.Ldate|log.Ltime),
	}
}

func (fr *featuresRepository) GetFeaturesIDs(ctx context.Context, featureSlugs []string) ([]int, error) {
	ids := []int{}
	for _, f := range featureSlugs {
		var curID int
		row, err := fr.db.QueryContext(ctx, "SELECT id FROM features WHERE slug = ? LIMIT 1", f)
		if err != nil {
			return []int{}, err
		}
		row.Next()
		err = row.Scan(&curID)
		if err != nil {
			return []int{}, err
		}

		err = row.Close()
		if err != nil {
			return []int{}, err
		}

		ids = append(ids, curID)
	}
	return ids, nil
}

func (fr *featuresRepository) GetNRandomUsersWithoutFeature(n int, slug string) ([]int, error) {
	userIDs := []int{}

	rows, err := fr.db.Query(
		`SELECT DISTINCT u.id FROM users u
				WHERE (SELECT user_id 
					   FROM user_feature_relation 
					   WHERE user_id = u.id 
					   AND feature_id = 
							(SELECT id 
							FROM features 
							WHERE slug = ? 
							LIMIT 1) 
					   AND is_active = TRUE 
					   ORDER BY date_assigned 
					   LIMIT 1) IS NULL 
					   AND is_active = TRUE
				ORDER BY RAND() LIMIT ?`,
		slug,
		n,
	)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		userIDs = append(userIDs, id)
	}
	err = rows.Close()
	if err != nil {
		return nil, err
	}

	return userIDs, nil
}

func (fr *featuresRepository) GetActiveUsersAmount(ctx context.Context) (int, error) {
	var amount int

	row, err := fr.db.QueryContext(ctx, "SELECT COUNT(id) FROM users WHERE is_active = TRUE")
	if err != nil {
		return -1, err
	}

	row.Next()
	err = row.Scan(&amount)
	if err != nil {
		return -1, err
	}

	err = row.Close()
	if err != nil {
		return -1, err
	}

	return amount, nil
}

func (fr *featuresRepository) InsertFeature(ctx context.Context, featureSlug string) error {
	_, err := fr.db.ExecContext(
		ctx,
		"INSERT INTO features (`slug`) VALUES (?) ON DUPLICATE KEY UPDATE is_active = TRUE",
		featureSlug,
	)
	if err != nil {
		return err
	}

	fr.InfoLog.Printf("InsertFeature — %s\n", featureSlug)
	return nil
}

func (fr *featuresRepository) DeleteFeature(ctx context.Context, featureSlug string) error {
	featureID, err := fr.GetFeaturesIDs(ctx, []string{featureSlug})
	if err != nil {
		fr.ErrLog.Printf("%s: %s", errors.ErrorGettingFeatureID, err)
		return err
	}

	tx, err := fr.db.BeginTx(ctx, nil)
	if err != nil {
		fr.ErrLog.Printf("%s: %s", errors.ErrorBeginTransaction, err)
		return err
	}

	_, err = tx.ExecContext(ctx, "UPDATE features SET is_active = FALSE WHERE id = ?", featureID[0])
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %w, rollback error: %s", err, rbErr)
		}
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		"UPDATE user_feature_relation "+
			"SET is_active = FALSE, date_unassigned = CURRENT_TIMESTAMP "+
			"WHERE feature_id = ?",
		featureID,
	)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %w, rollback error: %s", err, rbErr)
		}
		return err
	}

	err = tx.Commit()
	if err != nil {
		fr.ErrLog.Printf("%s: %s", errors.ErrorCommittingTransaction, err)
		return err
	}

	fr.InfoLog.Printf("DeleteFeature — %s\n", featureSlug)
	return nil
}

func (fr *featuresRepository) UnassignFeatures(ctx context.Context, userID []int, featuresToUnassign []string) error {
	if len(featuresToUnassign) == 0 {
		return nil
	}

	ids, err := fr.GetFeaturesIDs(ctx, featuresToUnassign)
	if err != nil {
		return err
	}

	tx, err := fr.db.BeginTx(ctx, nil)
	if err != nil {
		fr.ErrLog.Printf("%s: %s", errors.ErrorBeginTransaction, err)
		return err
	}

	for _, usr := range userID {
		for _, id := range ids {
			_, err = tx.ExecContext(
				ctx,
				"UPDATE user_feature_relation "+
					"SET is_active = FALSE, date_unassigned = CURRENT_TIMESTAMP "+
					"WHERE user_id = ? AND feature_id = ? AND is_active = TRUE",
				usr,
				id,
			)

			if err != nil {
				if rbErr := tx.Rollback(); rbErr != nil {
					return fmt.Errorf("transaction error: %w, rollback error: %s", err, rbErr)
				}
				return err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		fr.ErrLog.Printf("%s: %s", errors.ErrorCommittingTransaction, err)
		return err
	}

	fr.InfoLog.Printf("UnassignFeatures — %d\n", userID)
	return nil
}

func (fr *featuresRepository) AssignFeatures(ctx context.Context, userID []int, featuresToAssign []string) error {
	if len(featuresToAssign) == 0 {
		return nil
	}

	ids, err := fr.GetFeaturesIDs(ctx, featuresToAssign)
	if err != nil {
		return err
	}

	tx, err := fr.db.BeginTx(ctx, nil)
	if err != nil {
		fr.ErrLog.Printf("%s: %s", errors.ErrorBeginTransaction, err)
		return err
	}

	for _, usr := range userID {
		for _, featureID := range ids {
			var rows *sql.Rows
			rows, err = tx.QueryContext(
				ctx,
				"SELECT id FROM user_feature_relation WHERE is_active = TRUE AND user_id = ? AND feature_id = ?",
				usr,
				featureID,
			)

			if err != nil {
				if rbErr := tx.Rollback(); rbErr != nil {
					return fmt.Errorf("transaction error: %w, rollback error: %s", err, rbErr)
				}
				return err
			}

			ok := rows.Next()

			err = rows.Close()
			if err != nil {
				return err
			}

			if ok {
				return nil
			}

			_, err = tx.ExecContext(
				ctx,
				"INSERT INTO user_feature_relation (`user_id`, `feature_id`) VALUES (?, ?)",
				usr,
				featureID,
			)

			if err != nil {
				if rbErr := tx.Rollback(); rbErr != nil {
					return fmt.Errorf("transaction error: %w, rollback error: %s", err, rbErr)
				}
				return err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		fr.ErrLog.Printf("%s: %s", errors.ErrorCommittingTransaction, err)
		return err
	}

	fr.InfoLog.Printf("AssignFeatures — %d\n", userID)
	return nil
}

func (fr *featuresRepository) GetUserFeatures(ctx context.Context, userID int) (*Template, error) {
	rows, err := fr.db.QueryContext(
		ctx,
		"SELECT slug FROM features "+
			"WHERE id IN ("+
			"SELECT feature_id FROM user_feature_relation "+
			"WHERE user_id = ? AND is_active = TRUE"+
			") AND is_active = TRUE",
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

	fr.InfoLog.Printf("GetFeatures — %d\n", userID)
	return userFeatures, nil
}
