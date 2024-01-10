package search

import (
	"context"
	"encoding/json"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type SearchStorage struct {
	db     neo4j.DriverWithContext
	dbName string
}

func NewSearchStorage(db neo4j.DriverWithContext, dbName string) *SearchStorage {
	return &SearchStorage{
		db:     db,
		dbName: dbName,
	}
}

type searchTextResult struct {
	UserName string `json:"userName"`
	Tag      string `json:"tag"`
	Hashtag  string `json:"hashtag"`
	UserPic  string `json:"userPic"`
}

func (s *SearchStorage) searchText(text string, ctx context.Context) ([]searchTextResult, error) {
	session := s.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: s.dbName, AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	searchResult, err := session.ExecuteRead(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			result, err := tx.Run(ctx,
				`CALL db.index.fulltext.queryNodes("SearchWithTitleAndName", $text) YIELD node, score
        RETURN node.userName AS userName, node.profilePic AS userPic, node.name AS tag, node.title AS hashtag,score ORDER BY score`,
				map[string]any{
					"text": text + "*",
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

	var arr []searchTextResult

	for _, searchText := range searchResult.([]*neo4j.Record) {
		jsonData, _ := json.Marshal(searchText.AsMap())

		var structData searchTextResult
		json.Unmarshal(jsonData, &structData)

		arr = append(arr, searchTextResult{
			UserName: structData.UserName,
			Tag:      structData.Tag,
			Hashtag:  structData.Hashtag,
			UserPic:  structData.UserPic,
		})
	}

	return arr, nil
}

type stylesByTextResult struct {
	Id         string `json:"id"`
	Image      string `json:"image"`
	Links      []link `json:"links"`
	User       user   `json:"user"`
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
}

func (s *SearchStorage) stylesByText(text string, ctx context.Context) ([]stylesByTextResult, error) {
	session := s.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: s.dbName, AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	searchResult, err := session.ExecuteRead(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			result, err := tx.Run(ctx,
				`CALL db.index.fulltext.queryNodes("stylesByTagsAndHastags", $text) YIELD node, score
        MATCH (node)<-[r]-(s:Style)
        MATCH (s)-[:CREATED_BY]->(p:User)
        OPTIONAL MATCH (s)-[:LINKED_TO]->(l:Link)
        OPTIONAL MATCH (:User)-[m:MARKED_TREND]->(s)
        WITH s,l,p, COUNT(m) AS trendCount
        RETURN s.uuid as id, s.image as image, s.created_at as created_at, collect(l{id:l.uuid,url:l.url,image:l.image}) AS links, {userName:p.userName, profilePic:p.profilePic} as user, trendCount
        `,
				map[string]any{
					"text": text + "*",
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

	var arr []stylesByTextResult

	for _, style := range searchResult.([]*neo4j.Record) {
		jsonData, _ := json.Marshal(style.AsMap())

		var structData stylesByTextResult
		json.Unmarshal(jsonData, &structData)

		arr = append(arr, stylesByTextResult{
			Id:         structData.Id,
			Image:      structData.Image,
			Links:      structData.Links,
			User:       structData.User,
			TrendCount: structData.TrendCount,
			Created_at: structData.Created_at,
		})
	}

	return arr, nil
}
