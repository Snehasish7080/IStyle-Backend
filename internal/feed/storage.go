package feed

import (
	"context"
	"encoding/json"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/zone/IStyle/internal/models"
)

type FeedStorage struct {
	db     neo4j.DriverWithContext
	dbName string
}

func NewFeedStorage(db neo4j.DriverWithContext, dbName string) *FeedStorage {
	return &FeedStorage{
		db:     db,
		dbName: dbName,
	}
}

func (f *FeedStorage) feed(userName string, ctx context.Context) ([]models.Style, error) {
	session := f.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: f.dbName, AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	styles, err := session.ExecuteRead(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			result, err := tx.Run(ctx,
				`
      MATCH(u:User{userName:$userName})
      MATCH(s:Style) 
      MATCH(l:Link)
      WHERE (s)-[:TAG_TO]->(:Tag)<-[:MARK_FAV]-(u) AND NOT (s)-[:CREATED_BY]->(u) AND (s)-[:LINKED_TO]->(l)
      RETURN s.uuid AS uuid, s.image AS image, collect(l{id:l.uuid,url:l.url,image:l.image}) AS links, s.created_at AS created_at ORDER BY s.created_at DESC
      `,
				map[string]interface{}{
					"userName": userName,
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
		var links []models.Link
		jsonData, _ := json.Marshal(style.AsMap())

		var structData models.Style
		json.Unmarshal(jsonData, &structData)

		for _, link := range structData.Links {
			links = append(links, models.Link{
				Id:    link.Uuid,
				Image: link.Image,
				Url:   link.Url,
			})
		}
		arr = append(arr, models.Style{
			ID:         structData.Uuid,
			Image:      structData.Image,
			Links:      links,
			Created_at: structData.Created_at,
		})
	}

	return arr, nil
}
