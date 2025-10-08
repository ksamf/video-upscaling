package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
)

type VideoModel struct {
	Pool *pgxpool.Pool
}

type Video struct {
	VideoId    uuid.UUID `json:"video_id"`
	Name       string    `json:"name"`
	VideoPath  string    `json:"video_path"`
	LanguageId int       `json:"language_id"`
	QualityId  int       `json:"quality_id"`
}
type FullVideo struct {
	VideoId   uuid.UUID `json:"video_id"`
	Name      string    `json:"name"`
	VideoPath string    `json:"video_path"`
	Language  string    `json:"language"`
	Qualities []int     `json:"qualities"`
}

func (m *VideoModel) Insert(video *Video) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	query := "INSERT INTO videos(video_id, name, video_path) VALUES($1, $2, $3)"
	_, err := m.Pool.Exec(ctx, query, video.VideoId, video.Name, video.VideoPath)

	if err != nil {
		return err
	}
	return nil
}

func (m *VideoModel) GetAll(limit, offset string) ([]*Video, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	query := "SELECT * FROM videos LIMIT $1 OFFSET $2"
	rows, err := m.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}

	videos := []*Video{}
	defer rows.Close()
	for rows.Next() {
		var video Video
		err := rows.Scan(&video.VideoId, &video.Name, &video.VideoPath, &video.LanguageId, &video.QualityId)
		if err != nil {
			return nil, err
		}
		videos = append(videos, &video)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return videos, nil
}

func (m *VideoModel) GetByID(id uuid.UUID) (*FullVideo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	query := `SELECT 
    			v.video_id,
    			v.name,
    			v.video_path,
    			COALESCE(l.language, '') AS language,
    			COALESCE(q.qualities, '{}') AS qualities
			FROM videos AS v
			LEFT JOIN languages AS l 
			    ON l.language_id = v.language_id OR (v.language_id IS NULL AND l.language_id IS NULL)
			LEFT JOIN qualities AS q 
			    ON q.quality_id = v.quality_id OR (v.quality_id IS NULL AND q.quality_id IS NULL)
			WHERE v.video_id = $1;
			`
	row := m.Pool.QueryRow(ctx, query, id)
	var v FullVideo
	err := row.Scan(&v.VideoId, &v.Name, &v.VideoPath, &v.Language, &v.Qualities)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (m *VideoModel) UpdatePartial(id uuid.UUID, field string, value any) error {
	validFields := map[string]bool{
		"quality_id": true,
		"name":       true,
		"video_path": true,
	}
	if !validFields[field] {
		return fmt.Errorf("invalid field: %s", field)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := fmt.Sprintf(`UPDATE videos SET "%s" = $1 WHERE video_id = $2`, field)
	res, err := m.Pool.Exec(ctx, query, value, id)
	if err != nil {
		return err
	}

	rows := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("no rows updated for video_id %s", id)
	}

	return nil
}
func (m *VideoModel) Delete(id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	query := "DELETE FROM videos WHERE video_id=$1"
	_, err := m.Pool.Exec(ctx, query, id)
	return err
}

func (m *VideoModel) UpdateQualities(id uuid.UUID, value []int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "SELECT quality_id FROM qualities WHERE qualities @> $1::int[]"
	row := m.Pool.QueryRow(ctx, query, pq.Array(value))

	var r int
	err := row.Scan(&r)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no qualities found for %v", value)
		}
		return fmt.Errorf("query failed: %w", err)
	}

	if err := m.UpdatePartial(id, "quality_id", r); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	return nil
}

func (m *VideoModel) GetLanguage(id uuid.UUID) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	query := `SELECT l.language 
				FROM languages AS l
				JOIN videos AS v ON v.language_id = l.language_id 
				WHERE v.video_id = $1;`
	row := m.Pool.QueryRow(ctx, query, id)
	var language string

	err := row.Scan(&language)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
	}
	return language, err
}
