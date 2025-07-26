package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VideoModel struct {
	Pool *pgxpool.Pool
}

type Video struct {
	Id        int     `json:"id"`
	Name      string  `json:"name"`
	Url       string  `json:"url"`
	Language  string  `json:"language"`
	Nsfw      bool    `json:"nsfw"`
	Qualities []int32 `json:"qualities"`
}

func (m *VideoModel) Get(id int) (*Video, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "SELECT * FROM videos WHERE id = $1"

	var video Video

	err := m.Pool.QueryRow(ctx, query, id).Scan(&video.Id, &video.Language, &video.Name, &video.Nsfw, &video.Qualities, &video.Url)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &video, nil

}

func (m *VideoModel) GetAll() ([]*Video, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	query := "SELECT * FROM videos"

	var videos []*Video

	rows, err := m.Pool.Query(ctx, query)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var video Video
		err := rows.Scan(&video.Id, &video.Language, &video.Name, &video.Nsfw, &video.Nsfw, &video.Qualities, &video.Url)
		if err != nil {
			return nil, err
		}
		videos = append(videos, &video)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return videos, nil

}
func (m *VideoModel) Delete(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	query := "DELETE FROM videos WHERE id = $1"

	_, err := m.Pool.Exec(ctx, query, id)

	if err != nil {
		return err
	}
	return nil

}
func (m *VideoModel) GetInfo(id int) (*Video, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "SELECT * FROM videos WHERE id = $1"

	var video Video

	err := m.Pool.QueryRow(ctx, query, id).Scan(&video.Url, &video.Language, &video.Nsfw, &video.Qualities)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &video, nil
}
func (m *VideoModel) GetQualities(id int) (*Video, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "SELECT qualities FROM videos WHERE id = $1"

	var video Video

	err := m.Pool.QueryRow(ctx, query, id).Scan(&video.Qualities)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &video, nil
}
func (m *VideoModel) GetSubtitles(id int) (*Video, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "SELECT * FROM videos WHERE id = $1"

	var video Video

	err := m.Pool.QueryRow(ctx, query, id).Scan(&video.Id, &video.Language, &video.Name, &video.Nsfw, &video.Qualities, &video.Url)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &video, nil
}
func (m *VideoModel) GetDubbing(id int) (*Video, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "SELECT * FROM videos WHERE id = $1"

	var video Video

	err := m.Pool.QueryRow(ctx, query, id).Scan(&video.Id, &video.Language, &video.Name, &video.Nsfw, &video.Qualities, &video.Url)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &video, nil
}
