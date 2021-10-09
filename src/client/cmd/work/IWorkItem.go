package work

import "kard/src/model/dto"

type IWorkItem interface {
	Do(dto *dto.UrlDto)
}
