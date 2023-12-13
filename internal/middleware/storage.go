package middleware

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

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

func (m *MiddlewareStorage) userNameExists(userName string, ctx context.Context) bool {
	session := m.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: m.dbName, AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, _ := session.ExecuteRead(ctx,
		func(tx neo4j.ManagedTransaction) (interface{}, error) {
			result, err := tx.Run(ctx,
				"MATCH (u:User {userName:$userName}) RETURN u.userName AS userName",
				map[string]interface{}{
					"userName": userName,
				},
			)
			if err != nil {
				return nil, err
			}
			record, err := result.Single(ctx)
			if err != nil {
				return nil, err
			}
			userName, _ := record.Get("userName")
			return userName.(string), nil
		})

	return result != nil
}
