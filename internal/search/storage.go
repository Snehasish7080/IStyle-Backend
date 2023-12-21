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
}

func (s *SearchStorage) searchText(text string, ctx context.Context) ([]searchTextResult, error) {
	session := s.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: s.dbName, AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	searchResult, err := session.ExecuteRead(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			result, err := tx.Run(ctx,
				`CALL db.index.fulltext.queryNodes("SearchWithTitleAndName", "$text*") YIELD node, score
        RETURN node.userName AS userName, node.name AS tag, node.title AS hashtag,score ORDER BY score`,
				map[string]any{
					"text": text,
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
		})
	}

	return arr, nil
}
