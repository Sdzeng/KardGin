package crawler

import "kard/src/model/dto"

type ICrawler interface {
	Work(store func([]*dto.SubtitlesIndexDto))
}