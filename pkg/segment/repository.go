package segment

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"usersegmentator/pkg/errors"
)

type Repository interface {
	InsertSegment(ctx context.Context, segmentSlug string) error
	DeleteSegment(ctx context.Context, segmentSlug string) error
	UnassignSegments(ctx context.Context, userID []int, segmentsToUnassign []string) error
	AssignSegments(ctx context.Context, userID []int, segmentsToAssign []string) error
	GetUserSegments(ctx context.Context, userID int) (*Template, error)
	GetNRandomUsersWithoutSegment(n int, slug string) ([]int, error)
	GetActiveUsersAmount(ctx context.Context) (int, error)
	GetSegmentsIDs(ctx context.Context, segmentSlugs []string) ([]int, error)
}

type segmentsRepository struct {
	db      *sql.DB
	InfoLog *log.Logger
	ErrLog  *log.Logger
}

func NewSegmentsRepo(db *sql.DB) Repository {
	return &segmentsRepository{
		db:      db,
		InfoLog: log.New(os.Stdout, "INFO\tSEGMENTS REPO\t", log.Ldate|log.Ltime),
		ErrLog:  log.New(os.Stdout, "ERROR\tSEGMENTS REPO\t", log.Ldate|log.Ltime),
	}
}

func (fr *segmentsRepository) GetSegmentsIDs(ctx context.Context, segmentSlugs []string) ([]int, error) {
	ids := []int{}
	for _, f := range segmentSlugs {
		var curID int
		row, err := fr.db.QueryContext(ctx, "SELECT id FROM segments WHERE slug = ? LIMIT 1", f)
		if err != nil {
			return []int{}, err
		}
		row.Next()
		err = row.Scan(&curID)
		//if err != nil {
		//	fmt.Println("12312")
		//	return []int{}, err
		//}

		err = row.Close()
		//if err != nil {
		//	return []int{}, err
		//}

		ids = append(ids, curID)
	}
	return ids, nil
}

func (fr *segmentsRepository) GetNRandomUsersWithoutSegment(n int, slug string) ([]int, error) {
	userIDs := []int{}

	rows, err := fr.db.Query(
		`SELECT DISTINCT u.id FROM users u
				WHERE (SELECT user_id 
					   FROM user_segment_relation 
					   WHERE user_id = u.id 
					   AND segment_id = 
							(SELECT id 
							FROM segments 
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

func (fr *segmentsRepository) GetActiveUsersAmount(ctx context.Context) (int, error) {
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

func (fr *segmentsRepository) InsertSegment(ctx context.Context, segmentSlug string) error {
	if segmentSlug == "" {
		return fmt.Errorf("empty segment slug")
	}

	_, err := fr.db.ExecContext(
		ctx,
		"INSERT INTO segments (`slug`) VALUES (?) ON DUPLICATE KEY UPDATE is_active = TRUE",
		segmentSlug,
	)
	if err != nil {
		return err
	}

	fr.InfoLog.Printf("InsertSegment — %s\n", segmentSlug)
	return nil
}

func (fr *segmentsRepository) DeleteSegment(ctx context.Context, segmentSlug string) error {
	segmentID, err := fr.GetSegmentsIDs(ctx, []string{segmentSlug})
	if err != nil {
		fr.ErrLog.Printf("%s: %s", errors.ErrorGettingSegmentID, err)
		return err
	}

	tx, err := fr.db.BeginTx(ctx, nil)
	if err != nil {
		fr.ErrLog.Printf("%s: %s", errors.ErrorBeginTransaction, err)
		return err
	}

	_, err = tx.ExecContext(ctx, "UPDATE segments SET is_active = FALSE WHERE id = ?", segmentID[0])
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %w, rollback error: %s", err, rbErr)
		}
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		"UPDATE user_segment_relation "+
			"SET is_active = FALSE, date_unassigned = CURRENT_TIMESTAMP "+
			"WHERE segment_id = ?",
		segmentID[0],
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

	fr.InfoLog.Printf("DeleteSegment — %s\n", segmentSlug)
	return nil
}

func (fr *segmentsRepository) UnassignSegments(ctx context.Context, userID []int, segmentsToUnassign []string) error {
	if len(segmentsToUnassign) == 0 {
		return nil
	}

	ids, err := fr.GetSegmentsIDs(ctx, segmentsToUnassign)
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
				"UPDATE user_segment_relation "+
					"SET is_active = FALSE, date_unassigned = CURRENT_TIMESTAMP "+
					"WHERE user_id = ? AND segment_id = ? AND is_active = TRUE",
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

	fr.InfoLog.Printf("UnassignSegments — %d\n", userID)
	return nil
}

func (fr *segmentsRepository) AssignSegments(ctx context.Context, userID []int, segmentsToAssign []string) error {
	if len(segmentsToAssign) == 0 {
		return nil
	}

	ids, err := fr.GetSegmentsIDs(ctx, segmentsToAssign)
	if err != nil {
		return err
	}

	tx, err := fr.db.BeginTx(ctx, nil)
	if err != nil {
		fr.ErrLog.Printf("%s: %s", errors.ErrorBeginTransaction, err)
		return err
	}

	for _, usr := range userID {
		for _, segmentID := range ids {
			var rows *sql.Rows
			rows, err = tx.QueryContext(
				ctx,
				"SELECT id FROM user_segment_relation WHERE is_active = TRUE AND user_id = ? AND segment_id = ?",
				usr,
				segmentID,
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
				"INSERT INTO user_segment_relation (`user_id`, `segment_id`) VALUES (?, ?)",
				usr,
				segmentID,
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

	fr.InfoLog.Printf("AssignSegments — %d\n", userID)
	return nil
}

func (fr *segmentsRepository) GetUserSegments(ctx context.Context, userID int) (*Template, error) {
	rows, err := fr.db.QueryContext(
		ctx,
		"SELECT slug FROM segments "+
			"WHERE id IN ("+
			"SELECT segment_id FROM user_segment_relation "+
			"WHERE user_id = ? AND is_active = TRUE"+
			") AND is_active = TRUE",
		userID,
	)
	if err != nil {
		return nil, err
	}

	userSegments := &Template{
		UserID:   userID,
		Segments: []string{},
	}

	for rows.Next() {
		var segment string
		err = rows.Scan(&segment)
		if err != nil {
			return nil, err
		}
		userSegments.Segments = append(userSegments.Segments, segment)
	}
	err = rows.Close()
	if err != nil {
		return nil, err
	}

	fr.InfoLog.Printf("GetSegments — %d\n", userID)
	return userSegments, nil
}
