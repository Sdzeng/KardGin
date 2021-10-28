package dto

import "time"

type SubtitlesItemDto struct {
	Texts   []string      `json:"text"`
	StartAt time.Duration `json:"start_at"`
}
