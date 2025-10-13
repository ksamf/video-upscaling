package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VideoModel struct {
	Pool *pgxpool.Pool
}

type Video struct {
	VideoId    uuid.UUID  `json:"video_id"`
	Name       string     `json:"name"`
	VideoPath  string     `json:"video_path"`
	LanguageId int        `json:"language_id"`
	Quality    int        `json:"quality"`
	CreatedAt  *time.Time `json:"created_at"`
	UpdatedAt  *time.Time `json:"update_at"`
}
type FullVideo struct {
	VideoId   uuid.UUID  `json:"video_id"`
	Name      string     `json:"name"`
	VideoPath string     `json:"video_path"`
	Language  string     `json:"language"`
	Qualities []int      `json:"qualities"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"update_at"`
}

var standardHeights = []int{144, 240, 360, 480, 720, 1080, 1440, 2160, 4320}

func (m *VideoModel) Insert(video *Video) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	query := "INSERT INTO videos(video_id, name, video_path, language_id, quality) VALUES($1, $2, $3, $4, $5)"
	_, err := m.Pool.Exec(ctx, query, video.VideoId, video.Name, video.VideoPath, video.LanguageId, video.Quality)
	if err != nil {
		return err
	}
	return nil
}

func (m *VideoModel) GetAll(limit, offset string) ([]*Video, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	intLimit, err := strconv.Atoi(limit)
	if err != nil {
		intLimit = 10
	}
	intOffset, err := strconv.Atoi(offset)
	if err != nil {
		intOffset = 0
	}
	query := "SELECT video_id, name, video_path, language_id, quality, created_at, updated_at FROM videos LIMIT $1 OFFSET $2"
	rows, err := m.Pool.Query(ctx, query, intLimit, intOffset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var videos []*Video
	for rows.Next() {
		var video Video
		if err := rows.Scan(
			&video.VideoId,
			&video.Name,
			&video.VideoPath,
			&video.LanguageId,
			&video.Quality,
			&video.CreatedAt,
			&video.UpdatedAt,
		); err != nil {
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
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT 
			v.video_id,
			v.name,
			v.video_path,
			l.code,
			v.quality,
			v.created_at,
			v.updated_at
		FROM videos AS v
		LEFT JOIN languages AS l ON l.language_id = v.language_id
		WHERE v.video_id = $1;
	`

	row := m.Pool.QueryRow(ctx, query, id)

	var v FullVideo
	var q int
	var lang sql.NullString

	err := row.Scan(&v.VideoId, &v.Name, &v.VideoPath, &lang, &q, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if lang.Valid {
		v.Language = lang.String
	} else {
		v.Language = ""
	}

	v.Qualities = standardHeights[0 : slices.Index(standardHeights, q)+3]
	return &v, nil
}

func (m *VideoModel) UpdatePartial(id uuid.UUID, field string, value any) error {
	validFields := map[string]bool{
		"language_id": true,
		"quality":     true,
		"name":        true,
		"video_path":  true,
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

func (m *VideoModel) GetLanguageId(lang string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	query := "SELECT language_id FROM languages WHERE code=$1"
	row := m.Pool.QueryRow(ctx, query, lang)
	var langId int
	err := row.Scan(&langId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return langId, nil
}

func GetAllRes(base int) []int {
	var result []int
	for i, h := range standardHeights {
		if standardHeights[i] == base {
			result = append(result, standardHeights[i+1])
			result = append(result, standardHeights[i+2])
		}
		if h <= base {
			result = append(result, h)
		}
	}
	return result
}
