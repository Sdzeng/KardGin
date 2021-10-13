package dto

type SubtitlesIndexDto struct {
	Title    string   `json:"title"`
	SubTitle string   `json:"subtitle"`
	Texts    []string `json:"texts"`
	StartAt  int32    `json:"start_at"`
	Lan      string   `json:"lan"`
}
