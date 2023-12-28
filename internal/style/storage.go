package style

import (
	"context"
	"encoding/json"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/zone/IStyle/internal/models"
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

func (s *StyleStorage) create(userName string, image string, links []map[string]interface{}, tags []string, hashtags []string, ctx context.Context) (string, error) {
	now := time.Now()
	session := s.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: s.dbName, AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			return tx.Run(ctx,
				`
	      MATCH (u:User {userName:$userName})
        CREATE (s:Style {image:$image, uuid:randomUUID(), created_at:datetime($createdAt), updated_at:datetime($updatedAt)})
        CREATE (s)-[:CREATED_BY]->(u)
        WITH s
        CALL{
          WITH s
          UNWIND $links AS link
          CREATE (l:Link {image:link.image, url:link.url, uuid:randomUUID(), created_at:datetime($createdAt), updated_at:datetime($updatedAt)})
          MERGE (s)-[:LINKED_TO]->(l)
        }
        WITH s
        CALL{
          WITH s
          UNWIND $hashtags AS hashtag
          CREATE (h:Hashtag {title:hashtag, uuid:randomUUID(), created_at:datetime($createdAt), updated_at:datetime($updatedAt)})
          MERGE (s)-[:HASHTAG_TO]->(h)
        }
        WITH s
				UNWIND $tags AS tagId
				MATCH (t:Tag {uuid:tagId})
				MERGE (s)-[:TAG_TO]->(t)
				`,
				map[string]interface{}{
					"userName":  userName,
					"image":     image,
					"links":     links,
					"tags":      tags,
					"hashtags":  hashtags,
					"createdAt": now.Format(time.RFC3339),
					"updatedAt": now.Format(time.RFC3339),
				})
		})
	if err != nil {
		return "", err
	}

	return "created successfully", nil
}

func (s *StyleStorage) getALLStyles(userName string, cursor string, ctx context.Context) ([]models.Style, error) {
	session := s.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: s.dbName, AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	styles, err := session.ExecuteRead(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			result, err := tx.Run(ctx,
				`
      MATCH(u:User{userName:$userName})
      MATCH(s:Style) 
      WHERE (s)-[:CREATED_BY]->(u) AND s.uuid>$cursor
      RETURN s.uuid AS uuid, s.image As image
      ORDER BY s.uuid
      LIMIT 30
      `,
				map[string]interface{}{
					"userName": userName,
					"cursor":   cursor,
				},
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

	var arr []models.Style
	for _, style := range styles.([]*neo4j.Record) {
		jsonData, _ := json.Marshal(style.AsMap())

		var structData models.Style
		json.Unmarshal(jsonData, &structData)

		arr = append(arr, models.Style{
			ID:    structData.Uuid,
			Image: structData.Image,
		})
	}

	return arr, nil
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
					"id":       id,
				})
		})
	if err != nil {
		return "", err
	}

	return "trend successfully", nil
}

func (s *StyleStorage) unTrend(userName string, id string, ctx context.Context) (string, error) {
	session := s.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: s.dbName, AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			return tx.Run(ctx,
				`
				MATCH (s:Style {uuid:$id})
        MATCH (u:User {userName:$userName})-[r:MARKED_TREND]->(s)
				DELETE r
				`,
				map[string]interface{}{
					"userName": userName,
					"id":       id,
				})
		})
	if err != nil {
		return "", err
	}

	return "unmarked successfully", nil
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
