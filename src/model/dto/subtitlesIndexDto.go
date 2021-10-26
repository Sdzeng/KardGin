package dto

type SubtitlesIndexDto struct {
	DownloadPathId int32    `json:"path_id"`
	Title          string   `json:"title"`
	SubTitle       string   `json:"subtitle"`
	Texts          []string `json:"texts"`
	StartAt        int32    `json:"start_at"`
	Lan            string   `json:"lan"`
	PicPath        string   `json:"pic_path"`
	// Making         string   `json:"making"` //制作
	// Edit           string   `json:"edit"`   //校订
	// Source         string   `json:"source"` //来源
}
