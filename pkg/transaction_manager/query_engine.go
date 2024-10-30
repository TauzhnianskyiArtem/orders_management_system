package transaction_manager

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/moguchev/microservices_courcse/orders_management_system/pkg/postgres"
)

type QueryEngineProvider interface {
	GetQueryEngine(ctx context.Context) QueryEngine
}

type PgxCommonAPI interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

type PgxCommonScanAPI interface {
	// Getx - aka QueryRow
	Getx(ctx context.Context, dest interface{}, sqlizer postgres.Sqlizer) error
	// Selectx - aka Query
	Selectx(ctx context.Context, dest interface{}, sqlizer postgres.Sqlizer) error
	// Execx - aka Exec
	Execx(ctx context.Context, sqlizer postgres.Sqlizer) (pgconn.CommandTag, error)
}

type PgxExtendedAPI interface {
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
}

type QueryEngine interface {
	PgxCommonAPI
	PgxCommonScanAPI
	PgxExtendedAPI
}

type TxAccessMode = pgx.TxAccessMode

// Transaction access modes
const (
	ReadWrite = pgx.ReadWrite
	ReadOnly  = pgx.ReadOnly
)
