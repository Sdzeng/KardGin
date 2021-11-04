package dto

type EsResultDto struct {
	ScrollId     string               `json:"scroll_id"`
	TookInMillis int64                `json:"took_in_millis"` //耗时ms
	Total        int64                `json:"total"`          //总数
	SearchHits   []*SubtitlesIndexDto `json:"search_hits"`
}
