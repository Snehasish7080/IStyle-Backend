package feed

import (
	"context"
	"encoding/json"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
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

type feedStyle struct {
	Id         string `json:"id"`
	Image      string `json:"image"`
	Links      []link `json:"links"`
	User       user   `json:"user"`
	IsMarked   bool   `json:"isMarked"`
	TrendCount int    `json:"trendCount"`
	Created_at string `json:"created_at"`
}

type link struct {
	Id    string `json:"id"`
	Image string `json:"image"`
	Url   string `json:"url"`
}

type user struct {
	UserName   string `json:"userName"`
	ProfilePic string `json:"profilePic"`
	IsFollwing bool   `json:"isFollowing"`
}

func (f *FeedStorage) feed(userName string, cursor string, ctx context.Context) ([]feedStyle, error) {
	session := f.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: f.dbName, AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	styles, err := session.ExecuteRead(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			if cursor != "" {
				result, err := tx.Run(ctx,
					`
      MATCH(u:User{userName:$userName})
      MATCH(p:User)
      MATCH(s:Style) 
      WHERE ((s)-[:TAG_TO]->(:Tag)<-[:MARK_FAV]-(u) AND NOT (s)-[:CREATED_BY]->(u) AND (s)-[:CREATED_BY]->(p)) OR ((s)-[:CREATED_BY]->(p)<-[:FOLLOWING]-(u))
      OPTIONAL MATCH (:User)-[r:MARKED_TREND]->(s)
      OPTIONAL MATCH (s)-[:LINKED_TO]->(l:Link)
      WITH s,l,p,u, COUNT(r) AS trendCount
      WHERE s.created_at<datetime($cursor)
      RETURN s.uuid AS id, s.image AS image, collect(l{id:l.uuid,url:l.url,image:l.image}) AS links, {userName:p.userName, profilePic:p.profilePic, isFollowing:EXISTS((u)-[:FOLLOWING]->(p))} AS user, EXISTS((u)-[:MARKED_TREND]->(s)) AS isMarked, trendCount,s.created_at AS created_at ORDER BY s.created_at DESC
      LIMIT 4
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

			} else {

				result, err := tx.Run(ctx,
					`
      MATCH(u:User{userName:$userName})
      MATCH(p:User)
      MATCH(s:Style) 
      WHERE ((s)-[:TAG_TO]->(:Tag)<-[:MARK_FAV]-(u) AND NOT (s)-[:CREATED_BY]->(u) AND (s)-[:CREATED_BY]->(p)) OR ((s)-[:CREATED_BY]->(p)<-[:FOLLOWING]-(u))
      OPTIONAL MATCH (:User)-[r:MARKED_TREND]->(s)
      OPTIONAL MATCH (s)-[:LINKED_TO]->(l:Link)
      WITH s,l,p,u, COUNT(r) AS trendCount
      RETURN s.uuid AS id, s.image AS image, collect(l{id:l.uuid,url:l.url,image:l.image}) AS links, {userName:p.userName, profilePic:p.profilePic, isFollowing:EXISTS((u)-[:FOLLOWING]->(p))} AS user, EXISTS((u)-[:MARKED_TREND]->(s)) AS isMarked, trendCount,s.created_at AS created_at ORDER BY s.created_at DESC
      LIMIT 4
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

			}
		})
	if err != nil {
		return nil, err
	}

	var arr []feedStyle
	for _, style := range styles.([]*neo4j.Record) {

		jsonData, _ := json.Marshal(style.AsMap())

		var structData feedStyle
		json.Unmarshal(jsonData, &structData)

		arr = append(arr, feedStyle{
			Id:         structData.Id,
			Image:      structData.Image,
			Links:      structData.Links,
			User:       structData.User,
			IsMarked:   structData.IsMarked,
			TrendCount: structData.TrendCount,
			Created_at: structData.Created_at,
		})
	}

	return arr, nil
}
