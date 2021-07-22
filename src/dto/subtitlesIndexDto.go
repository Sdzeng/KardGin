package dto

type SubtitlesIndexDto struct {
	Title        string   `json:"title"`
	SubTitle     string   `json:"subtitle"`
	Text         []string `json:"text"`
	TimeDuration string   `json:"time_duration"`
	Lan          string   `json:"lan"`
}
