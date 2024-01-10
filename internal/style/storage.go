package style

import (
	"context"
	"encoding/json"
	"errors"
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

type styleById struct {
	Id         string      `json:"id"`
	Image      string      `json:"image"`
	Links      []styleLink `json:"links"`
	TrendCount int64       `json:"trendCount"`
	IsMarked   bool        `json:"isMarked"`
	User       styleUser   `json:"user"`
}
type styleLink struct {
	Id    string `json:"id"`
	Image string `json:"image"`
	Url   string `json:"url"`
}

type styleUser struct {
	UserName   string `json:"userName"`
	ProfilePic string `json:"profilePic"`
}

func (s *StyleStorage) styleById(userName string, id string, ctx context.Context) (*styleById, error) {
	session := s.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: s.dbName, AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	isStyleExist := s.checkStyleExists(id, ctx)

	if !isStyleExist {
		return nil, errors.New("invalid request")
	}

	result, _ := session.ExecuteRead(ctx,
		func(tx neo4j.ManagedTransaction) (interface{}, error) {
			result, err := tx.Run(ctx,
				`MATCH(u:User{userName:$userName})
         MATCH (s:Style{uuid: $id})
         MATCH ((s)-[:LINKED_TO]->(l:Link))
         MATCH ((s)-[:CREATED_BY]->(p:User))
         OPTIONAL MATCH ((:User)-[m:MARKED_TREND]->(s))
         WITH s,l,u,p, COUNT(m) AS trendCount
        RETURN s.uuid AS id, s.image AS image, collect({id:l.uuid, image:l.image, url:l.url}) AS links, trendCount, EXISTS((u)-[:MARKED_TREND]->(s)) AS isMarked, {userName:p.userName,profilePic:p.profilePic} AS user
        `,
				map[string]interface{}{
					"userName": userName,
					"id":       id,
				},
			)
			if err != nil {
				return nil, err
			}

			record, err := result.Single(ctx)
			if err != nil {
				return nil, err
			}
			id, _ := record.Get("id")
			image, _ := record.Get("image")
			links, _ := record.Get("links")
			trendCount, _ := record.Get("trendCount")
			isMarked, _ := record.Get("isMarked")
			user, _ := record.Get("user")

			var arr []styleLink
			jsonData, _ := json.Marshal(links)
			json.Unmarshal(jsonData, &arr)

			var postUser styleUser
			userjsonData, _ := json.Marshal(user)
			json.Unmarshal(userjsonData, &postUser)

			if isMarked == nil {
				isMarked = false
			}

			return &styleById{
				Id:         id.(string),
				Image:      image.(string),
				Links:      arr,
				TrendCount: trendCount.(int64),
				IsMarked:   isMarked.(bool),
				User:       postUser,
			}, nil
		})

	style, err := result.(*styleById)

	if !err {
		return nil, errors.New("something went wrong")
	}

	return style, nil
}

type likedUser struct {
	UserName   string `json:"userName"`
	ProfilePic string `json:"profilePic"`
}

func (s *StyleStorage) likedUsers(id string, ctx context.Context) ([]likedUser, error) {
	session := s.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: s.dbName, AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	isStyleExist := s.checkStyleExists(id, ctx)

	if !isStyleExist {
		return nil, errors.New("invalid request")
	}

	users, err := session.ExecuteRead(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			result, err := tx.Run(ctx,
				`
      MATCH(s:Style{uuid:$id})
        WHERE (s)<-[:MARKED_TREND]-(u:User)
      RETURN u.userName AS userName, s.profilePic As profilePic 
      `,
				map[string]interface{}{
					"id": id,
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
	// handle error
	if err != nil {
		return nil, err
	}

	var arr []likedUser
	for _, user := range users.([]*neo4j.Record) {
		jsonData, _ := json.Marshal(user.AsMap())

		var structData likedUser
		json.Unmarshal(jsonData, &structData)

		arr = append(arr, likedUser{
			UserName:   structData.UserName,
			ProfilePic: structData.ProfilePic,
		})
	}

	return arr, nil
}

func (s *StyleStorage) checkStyleExists(id string, ctx context.Context) bool {
	session := s.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: s.dbName, AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, _ := session.ExecuteRead(ctx,
		func(tx neo4j.ManagedTransaction) (interface{}, error) {
			result, err := tx.Run(ctx,
				"MATCH (s:Style {uuid:$id}) RETURN s.uuid AS uuid",
				map[string]interface{}{
					"id": id,
				},
			)
			if err != nil {
				return nil, err
			}
			record, err := result.Single(ctx)
			if err != nil {
				return nil, err
			}
			uuid, _ := record.Get("uuid")
			return uuid.(string), nil
		})

	return result != nil
}
