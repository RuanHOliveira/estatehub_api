package postgresql

import (
	"context"

	repo "github.com/RuanHOliveira/estatehub_api/internal/infra/database/postgresql/sqlc/generated"
	"github.com/jackc/pgx/v5"
)

type PgTxManager struct {
	q  *repo.Queries
	db *pgx.Conn
}

func NewPgTxManager(q *repo.Queries, db *pgx.Conn) *PgTxManager {
	return &PgTxManager{q: q, db: db}
}

func (m *PgTxManager) WithTx(ctx context.Context, fn func(q repo.Querier) error) error {
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	if err := fn(m.q.WithTx(tx)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
