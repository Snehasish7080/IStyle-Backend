package explore

import (
	"context"
	"encoding/json"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type ExploreStorage struct {
	db     neo4j.DriverWithContext
	dbName string
}

func NewExploreStorage(db neo4j.DriverWithContext, dbName string) *ExploreStorage {
	return &ExploreStorage{
		db:     db,
		dbName: dbName,
	}
}

type exploreStyle struct {
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

func (e *ExploreStorage) explore(userName string, ctx context.Context) ([]exploreStyle, error) {
	session := e.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: e.dbName, AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	styles, err := session.ExecuteRead(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
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

	var arr []exploreStyle
	for _, style := range styles.([]*neo4j.Record) {

		jsonData, _ := json.Marshal(style.AsMap())

		var structData exploreStyle
		json.Unmarshal(jsonData, &structData)

		arr = append(arr, exploreStyle{
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
