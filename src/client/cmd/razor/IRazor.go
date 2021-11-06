package razor

import "kard/src/model/dto"

type IRazor interface {
	Work(store func(taskDto *dto.TaskDto))
}
