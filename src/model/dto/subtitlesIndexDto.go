package dto

type SubtitlesIndexDto struct {
	IndexId  string   `json:"index_id"`
	Title    string   `json:"title"`
	SubTitle string   `json:"subtitle"`
	Text     []string `json:"text"`
	StartAt  int32    `json:"time_duration"`
	Lan      string   `json:"lan"`
}
