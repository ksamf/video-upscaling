package database

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type Models struct {
	Videos VideoModel
}

func NewModels(pool *pgxpool.Pool) Models {
	return Models{
		Videos: VideoModel{Pool: pool},
	}
}
