package database

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ActionModel struct {
	Pool *pgxpool.Pool
}

type Action struct {
	UserId    uuid.UUID `json:"user_id"`
	VideoId   uuid.UUID `json:"video_id"`
	Liked     bool      `json:"liked"`
	Disliked  bool      `json:"disliked"`
	Viewed_at time.Time `json:"viewed_at"`
}

func (m *ActionModel) Insert(userId, videoId uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	query := `INSERT INTO user_video_actions(user_id, video_id) VALUES($1, $2) ON CONFLICT DO NOTHING`
	_, err := m.Pool.Exec(ctx, query, userId, videoId)
	if err != nil {
		return fmt.Errorf("failed insert action:%w", err)
	}
	return nil

}
func (m *ActionModel) Get(userId, videoId uuid.UUID) (*Action, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	query := `
	SELECT user_id, video_id, liked, disliked, viewed_at
	FROM user_video_actions
	WHERE user_id=$1 AND video_id=$2`
	var a Action
	row := m.Pool.QueryRow(ctx, query, userId, videoId)
	err := row.Scan(&a.UserId, &a.VideoId, &a.Liked, &a.Disliked, &a.Viewed_at)
	if err != nil {
		return nil, fmt.Errorf("failed get action:%w", err)
	}
	return &a, nil
}
func (m *ActionModel) Update(userId, videoId uuid.UUID, reaction string) error {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var query string
	if reaction == "like" {
		query = "UPDATE user_video_actions SET liked = $1 WHERE user_id=$2 AND video_id = $3"
	} else {
		query = "UPDATE user_video_actions SET disliked = $1 WHERE user_id=$2 AND video_id = $3"
	}
	res, err := m.Pool.Exec(ctx, query, true, userId, videoId)
	if err != nil {
		return err
	}

	rows := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("no rows updated for user_id:%d, video_id:%s", userId, videoId)
	}

	return nil
}
func (m *ActionModel) ClearReaction(userId, videoId uuid.UUID) error {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "UPDATE user_video_actions SET liked = $1, disliked = $2 WHERE user_id=$3 AND video_id = $4"
	res, err := m.Pool.Exec(ctx, query, false, false, userId, videoId)
	if err != nil {
		return err
	}
	rows := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("no rows updated for user_id:%d, video_id:%s", userId, videoId)
	}

	return nil
}
