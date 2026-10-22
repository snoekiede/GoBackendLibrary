// internal/api/querierwithtx.go
package api

import (
	"context"

	db "bookbackend/internal/database"
	"github.com/jackc/pgx/v5"
)

type TxBeginner interface {
	BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
}

type QuerierWithTx interface {
	db.Querier
	WithTx(tx pgx.Tx) *db.Queries
}
