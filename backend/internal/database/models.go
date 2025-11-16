package database

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type Models struct {
	Videos  VideoModel
	Users   UserModel
	Actions ActionModel
}

func NewModel(pool *pgxpool.Pool) *Models {
	return &Models{
		Videos:  VideoModel{Pool: pool},
		Users:   UserModel{Pool: pool},
		Actions: ActionModel{Pool: pool},
	}
}
