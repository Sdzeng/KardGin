package dto

import "github.com/olivere/elastic/v7"

type SearchResultDto struct {
	ScrollId   string //rar路径
	SearchHits []*elastic.SearchHit
}
