package dto

import "time"

type CoverDto struct {
	Id                  int32
	ShowDate            time.Time
	EssayId             int64
	EssayCoverMediaType string
	EssayCoverPath      string
	EssayCoverExtension string
	EssayTitle          string
	EssayContent        string
	EssayPageUrl        string
	EssayLocation       string `json:"essayLocation"`
	EssayCreationTime   time.Time
	KuserNickName       string
	KuserIntroduction   string
	KuserAvatarUrl      string
}
