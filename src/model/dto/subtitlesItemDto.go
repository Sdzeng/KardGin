package dto

import "time"

type SubtitlesItemDto struct {
	Text    []string      `json:"text"`
	StartAt time.Duration `json:"start_at"`
}
