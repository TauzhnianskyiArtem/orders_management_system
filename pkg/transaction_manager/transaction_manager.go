package transaction_manager

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/moguchev/microservices_courcse/orders_management_system/pkg/postgres"
)

type TransactionManager struct {
	connection *postgres.Connection
}

func New(connection *postgres.Connection) *TransactionManager {
	return &TransactionManager{connection: connection}
}

type key string

const (
	txKey key = "tx"
)

func (m *TransactionManager) runTransaction(ctx context.Context, txOpts pgx.TxOptions, fn func(ctx context.Context) error) (err error) {
	tx, ok := ctx.Value(txKey).(*postgres.Transaction)
	if ok {
		return fn(ctx)
	}

	pgxTx, err := m.connection.BeginTx(ctx, txOpts)
	if err != nil {
		return fmt.Errorf("can't begin transaction: %v", err)
	}

	tx = &postgres.Transaction{Tx: pgxTx}
	ctx = context.WithValue(ctx, txKey, tx)

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic recovered: %v", r)
		}

		if err == nil {
			err = tx.Commit(ctx)
			if err != nil {
				err = fmt.Errorf("commit failed: %v", err)
			}
		}

		if err != nil {
			if errRollback := tx.Rollback(ctx); errRollback != nil {
				err = fmt.Errorf("rollback failed: %v", errRollback)
			}
		}
	}()
	err = fn(ctx)

	return err
}

func (m *TransactionManager) GetQueryEngine(ctx context.Context) QueryEngine {
	if tx, ok := ctx.Value(txKey).(QueryEngine); ok {
		return tx
	}

	return m.connection
}

func (m *TransactionManager) RunReadCommitted(ctx context.Context, accessMode pgx.TxAccessMode, f func(ctx context.Context) error) error {
	return m.runTransaction(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: accessMode,
	}, f)
}

func (m *TransactionManager) RunRepeatableRead(ctx context.Context, accessMode pgx.TxAccessMode, f func(ctx context.Context) error) error {
	return m.runTransaction(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: accessMode,
	}, f)
}

func (m *TransactionManager) RunSerializable(ctx context.Context, accessMode pgx.TxAccessMode, f func(ctx context.Context) error) error {
	return m.runTransaction(ctx, pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: accessMode,
	}, f)
}
