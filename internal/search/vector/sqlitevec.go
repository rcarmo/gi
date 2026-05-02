package vector

import (
	"context"
	"errors"
)

var ErrSQLiteVecUnavailable = errors.New("vector/sqlitevec: implementation scaffold only")

// SQLiteVecIndex is the planned sqlite-vec backed implementation.
type SQLiteVecIndex struct{}

func NewSQLiteVecIndex() *SQLiteVecIndex { return &SQLiteVecIndex{} }

func (i *SQLiteVecIndex) Upsert(ctx context.Context, rows []VectorRecord) error {
	return ErrSQLiteVecUnavailable
}
func (i *SQLiteVecIndex) DeleteByChunkIDs(ctx context.Context, ids []int64) error {
	return ErrSQLiteVecUnavailable
}
func (i *SQLiteVecIndex) Search(ctx context.Context, embedding []float32, k int) ([]Hit, error) {
	return nil, ErrSQLiteVecUnavailable
}
