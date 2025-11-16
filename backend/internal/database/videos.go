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
	UserId     uuid.UUID  `json:"user_id"`
	Name       string     `json:"name"`
	VideoPath  string     `json:"video_path"`
	LanguageId int        `json:"language_id"`
	Language   string     `json:"language"`
	Quality    int        `json:"quality"`
	Qualities  []int      `json:"qualities"`
	Views      int        `json:"views"`
	Likes      int        `json:"likes"`
	Dislikes   int        `json:"dislikes"`
	Upscaled   bool       `json:"upscaled"`
	CreatedAt  *time.Time `json:"created_at"`
	UpdatedAt  *time.Time `json:"update_at"`
}

var standardHeights = []int{144, 240, 360, 480, 720, 1080, 1440, 2160, 4320}

func (m *VideoModel) Insert(video *Video) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	query := "INSERT INTO videos(video_id, user_id, name) VALUES($1, $2, $3)"
	_, err := m.Pool.Exec(ctx, query, video.VideoId, video.UserId, video.Name)
	if err != nil {
		return err
	}
	return nil
}

func (m *VideoModel) GetAll(limit, offset string, getURL func(uuid.UUID) string) ([]*Video, error) {
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
	query := "SELECT video_id, user_id, name,created_at FROM videos LIMIT $1 OFFSET $2"
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
			&video.UserId,
			&video.Name,
			&video.CreatedAt,
		); err != nil {
			return nil, err
		}
		video.VideoPath = getURL(video.VideoId)
		videos = append(videos, &video)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return videos, nil
}

func (m *VideoModel) GetByID(id uuid.UUID, getURL func(uuid.UUID) string) (*Video, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT 
			v.video_id,
			v.user_id,
			v.name,
			v.quality,
			v.views,
			v.likes,
			v.dislikes,
			v.upscaled,
			v.created_at
		FROM videos AS v
		WHERE v.video_id = $1;
	`

	row := m.Pool.QueryRow(ctx, query, id)

	var v Video
	var lang sql.NullString

	err := row.Scan(&v.VideoId, &v.UserId, &v.Name, &v.Quality, &v.Views, &v.Likes, &v.Dislikes, &v.Upscaled, &v.CreatedAt)
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
	if v.Upscaled {
		v.Qualities = standardHeights[0 : slices.Index(standardHeights, v.Quality)+3]
	} else {
		v.Qualities = standardHeights[0 : slices.Index(standardHeights, v.Quality)+1]
	}
	v.VideoPath = getURL(v.VideoId)
	return &v, nil
}

func (m *VideoModel) Delete(id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	query := "DELETE FROM videos WHERE video_id=$1"
	_, err := m.Pool.Exec(ctx, query, id)
	return err
}

func (m *VideoModel) GetLanguage(id uuid.UUID) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	query := `SELECT l.code 
	FROM languages AS l 
	JOIN videos AS v ON v.video_id = $1
	WHERE v.language_id = l.language_id`
	row := m.Pool.QueryRow(ctx, query, id)
	var lang string
	err := row.Scan(&lang)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return lang, nil
}
func (m *VideoModel) UpdateLanguage(videoId uuid.UUID, lang string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	query := `UPDATE videos SET language_id = (SELECT language_id FROM languages WHERE code = $1) WHERE video_id = $2`
	res, err := m.Pool.Exec(ctx, query, lang, videoId)
	if err != nil {
		return err
	}
	rows := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("no rows updated for video_id:%s", videoId)
	}
	return nil
}
func (m *VideoModel) UpdateUpscaled(videoId uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	query := `UPDATE videos SET upscaled = true WHERE video_id = $1`
	res, err := m.Pool.Exec(ctx, query, videoId)
	if err != nil {
		return err
	}
	rows := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("no rows updated for video_id:%s", videoId)
	}
	return nil
}
func (m *VideoModel) UpdateQuality(videoId uuid.UUID, height int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	query := `UPDATE videos SET quality = $1 WHERE video_id = $2`
	res, err := m.Pool.Exec(ctx, query, height, videoId)
	if err != nil {
		return err
	}
	rows := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("no rows updated for video_id:%s", videoId)
	}
	return nil
}
func (m *VideoModel) UpdateViews(videoId uuid.UUID, views int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	query := `UPDATE videos SET views = views + $1 WHERE video_id = $2`
	res, err := m.Pool.Exec(ctx, query, views, videoId)
	if err != nil {
		return err
	}
	rows := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("no rows updated for video_id:%s", videoId)
	}
	return nil
}
func (m *VideoModel) UpdateLikes(videoId uuid.UUID, likes int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	query := `UPDATE videos SET likes = likes + $1 WHERE video_id = $2`
	res, err := m.Pool.Exec(ctx, query, likes, videoId)
	if err != nil {
		return err
	}
	rows := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("no rows updated for video_id:%s", videoId)
	}
	return nil
}
func (m *VideoModel) UpdateDislikes(videoId uuid.UUID, dislikes int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	query := `UPDATE videos SET dislikes = dislikes + $1 WHERE video_id = $2`
	res, err := m.Pool.Exec(ctx, query, dislikes, videoId)
	if err != nil {
		return err
	}
	rows := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("no rows updated for video_id:%s", videoId)
	}
	return nil
}
