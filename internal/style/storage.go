package style

import (
	"context"
	"time"

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

func (s *StyleStorage) create(userName string, image string, links []map[string]interface{}, tags []string, ctx context.Context) (string, error) {
	now := time.Now()
	session := s.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: s.dbName, AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			return tx.Run(ctx,
				`
				UNWIND $tags AS tagId
				MATCH (t:Tag {uuid:tagId})
        UNWIND $links AS link
        CREATE (l:Link {image:$link.image, url:$link.url, uuid:randomUUID(), created_at:datetime($createdAt), updated_at:datetime($updatedAt)})
				MATCH (u:User {userName:$userName})
				CREATE (s:Style {image:$image, links:$links, uuid:randomUUID(), created_at:datetime($createdAt), updated_at:datetime($updatedAt)})
				CREATE (s)-[:TAG_TO]->(t)
				CREATE (s)-[:LINKED_TO]->(l)
        CREATE (s)-[:CREATED_BY]->(u)
				`,
				map[string]interface{}{
					"userName":  userName,
					"image":     image,
					"links":     links,
					"tags":      tags,
					"createdAt": now.Format(time.RFC3339),
					"updatedAt": now.Format(time.RFC3339),
				})
		})
	if err != nil {
		return "", err
	}

	return "created successfully", nil
}

func (s *StyleStorage) trend(userName string, id string, ctx context.Context) (string, error) {
	session := s.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: s.dbName, AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			return tx.Run(ctx,
				`
				MATCH (s:Style {uuid:$id})
				MATCH (u:User {userName:$userName})
				CREATE (u)-[:MARKED_TREND]->(s)
				`,
				map[string]interface{}{
					"userName": userName,
					"uuid":     id,
				})
		})
	if err != nil {
		return "", err
	}

	return "trend successfully", nil
}

func (s *StyleStorage) clicked(userName string, id string, ctx context.Context) (string, error) {
	session := s.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: s.dbName, AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			return tx.Run(ctx,
				`
				MATCH (s:Style {uuid:$id})
				MATCH (u:User {userName:$userName})
				CREATE (u)-[:CLICKED]->(s)
				`,
				map[string]interface{}{
					"userName": userName,
					"uuid":     id,
				})
		})
	if err != nil {
		return "", err
	}

	return "trend successfully", nil
}
