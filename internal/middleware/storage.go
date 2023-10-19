package middleware

import "github.com/neo4j/neo4j-go-driver/v5/neo4j"

type MiddlewareStorage struct {
	db     neo4j.DriverWithContext
	dbName string
}

func NewMiddlewareStorage(db neo4j.DriverWithContext, dbName string) *MiddlewareStorage {
	return &MiddlewareStorage{
		db:     db,
		dbName: dbName,
	}
}
