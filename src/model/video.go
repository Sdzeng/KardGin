package model

type Video struct {
	BaseModel          `json:"-"`
	VideoName          string
	VideoCoverFragment string
	VideoCoverImg      string
	IsHomeCover        bool
	HomeCoverDate      string
}
