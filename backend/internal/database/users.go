package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserModel struct {
	Pool *pgxpool.Pool
}
type User struct {
	UserId    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (m *UserModel) Insert(userId uuid.UUID, email, password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	query := "INSERT INTO users(user_id, email, password) VALUES($1, $2, $3)"
	_, err := m.Pool.Exec(ctx, query, userId, email, password)
	if err != nil {
		log.Printf("failed insert users:%s", err)
		return fmt.Errorf("failed insert users:%w", err)
	}
	return nil
}
func (m *UserModel) GetByEmail(email string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	query := "SELECT user_id, email, password FROM users WHERE email=$1"
	var u User
	row := m.Pool.QueryRow(ctx, query, email)
	err := row.Scan(&u.UserId, &u.Email, &u.Password)
	if err != nil {
		return nil, fmt.Errorf("failed get user by email: %w", err)
	}
	return &u, nil
}
func (m *UserModel) GetById(id uuid.UUID) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	query := "SELECT user_id FROM users WHERE user_id=$1"
	var u User
	row := m.Pool.QueryRow(ctx, query, id)
	err := row.Scan(&u.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed get id: %w", err)
	}
	return &u, nil
}

func (m *UserModel) Delete(id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	query := "DELETE FROM users WHERE user_id=$1"
	_, err := m.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed delete id: %w", err)
	}
	return nil
}
func (m *UserModel) GetRecommendation(userId uuid.UUID, getURL func(uuid.UUID) string) ([]*Video, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	query := `
	SELECT v.video_id, v.name,
	FROM video v
	JOIN video_categories vc ON v.video_id = vc.video_id
	WHERE vc.category_id IN (
	SELECT DISTINCT vc2.category_id
	FROM video_categories vc2
	JOIN user_video_actions a ON vc2.video_id = a.video_id
	WHERE a.user_id = $1 AND (a.liked = TRUE OR a.viewed_at IS NOT NULL)
	)
	GROUP BY v.video_id
	ORDER BY COUNT(*) DESC, v.views DESC
	LIMIT 20;
	`
	rows, err := m.Pool.Query(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("failed get recommendation:%w", err)
	}
	defer rows.Close()
	var videos []*Video
	for rows.Next() {
		var v Video
		err = rows.Scan(&v.VideoId, &v.Name)
		if err != nil {
			return nil, fmt.Errorf("failed scan:%w", err)
		}
		v.VideoPath = getURL(v.VideoId)
		videos = append(videos, &v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return videos, nil
}

func (m *UserModel) GetMixedRecommendation(userId uuid.UUID, getURL func(uuid.UUID) string) ([]*Video, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	videos, err := m.GetRecommendation(userId, getURL)
	if err != nil {
		return nil, err
	}
	query := `
		SELECT v.video_id, v.name
		FROM videos v
		WHERE v.video_id NOT IN (
			SELECT a.video_id
			FROM user_video_actions a
			WHERE a.user_id = $1
		)
		ORDER BY random()
		LIMIT 5;
	`
	rows, err := m.Pool.Query(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("failed mix recommendation:%w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var v Video
		rows.Scan(&v.VideoId, &v.Name)
		v.VideoPath = getURL(v.VideoId)
		videos = append(videos, &v)
	}
	return videos, nil
}

func (m *UserModel) Videos(userId uuid.UUID, getURL func(uuid.UUID) string) ([]*Video, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	query := `
		SELECT v.*
		FROM videos v
		WHERE v.user_id = $1 
	`
	rows, err := m.Pool.Query(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("failed mix recommendation:%w", err)
	}
	defer rows.Close()
	var videos []*Video
	for rows.Next() {
		var v Video
		rows.Scan(&v.VideoId, &v.Name)
		v.VideoPath = getURL(v.VideoId)
		videos = append(videos, &v)
	}
	return videos, nil
}
