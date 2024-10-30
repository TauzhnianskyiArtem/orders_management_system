package postgres

import (
	"github.com/jackc/pgx/v5"
)

type Transaction struct {
	pgx.Tx
}
