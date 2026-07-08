package api

import (
	db "bookbackend/internal/database"

	"github.com/jackc/pgx/v5"
)

type QuerierWithTx interface {
	db.Querier
	WithTx(tx pgx.Tx) *db.Queries
}
