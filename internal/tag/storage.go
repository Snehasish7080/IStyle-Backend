package tag

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/zone/IStyle/internal/models"
)

type TagStorage struct {
	db     neo4j.DriverWithContext
	dbName string
}

func NewTagStorage(db neo4j.DriverWithContext, dbName string) *TagStorage {
	return &TagStorage{
		db:     db,
		dbName: dbName,
	}
}

func (t *TagStorage) create(name string, ctx context.Context) (string, error) {

	isTagExist := t.isTagExists(name, ctx)

	if isTagExist {
		return "", errors.New("tag already exists")
	}

	now := time.Now()
	session := t.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: t.dbName, AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			return tx.Run(ctx,
				"CREATE (:Tag {name: $name, uuid:randomUUID(), created_at:datetime($createdAt), updated_at:datetime($updatedAt)})",
				map[string]any{"name": name, "createdAt": now.Format(time.RFC3339), "updatedAt": now.Format(time.RFC3339)})
		})

	if err != nil {
		return "", err
	}

	return "Created Successfully", nil

}

func (t *TagStorage) getAll(ctx context.Context) ([]models.Tag, error) {
	session := t.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: t.dbName, AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	tags, err := session.ExecuteRead(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			result, err := tx.Run(ctx,
				"MATCH (t:Tag) RETURN t.name AS name, t.uuid AS uuid",
				map[string]any{},
			)
			if err != nil {
				return nil, err
			}

			record, err := result.Collect(ctx)

			if err != nil {
				return nil, err
			}

			return record, nil
		})

	if err != nil {
		return nil, err
	}

	var arr []models.Tag

	for _, tag := range tags.([]*neo4j.Record) {
		jsonData, _ := json.Marshal(tag.AsMap())

		var structData models.Tag
		json.Unmarshal(jsonData, &structData)

		arr = append(arr, models.Tag{
			ID:   structData.Uuid,
			Name: structData.Name,
		})
	}

	return arr, nil

}

func (t *TagStorage) isTagExists(name string, ctx context.Context) bool {
	session := t.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: t.dbName, AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, _ := session.ExecuteRead(ctx,
		func(tx neo4j.ManagedTransaction) (interface{}, error) {
			result, err := tx.Run(ctx,
				"MATCH (t:Tag {name:$name}) RETURN t.name AS name",
				map[string]interface{}{
					"name": name,
				},
			)
			if err != nil {
				return nil, err
			}
			record, err := result.Single(ctx)
			if err != nil {
				return nil, err
			}
			name, _ := record.Get("name")
			return name.(string), nil
		})

	return result != nil
}
