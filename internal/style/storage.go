package style

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type StyleStorage struct {
	db     neo4j.DriverWithContext
	dbName string
}

func NewStyleStorage(db neo4j.DriverWithContext, dbName string) *StyleStorage {
	return &StyleStorage{
		db:     db,
		dbName: dbName,
	}
}

func (t *StyleStorage) create(name string, ctx context.Context) (string, error) {

	return "", nil
}
