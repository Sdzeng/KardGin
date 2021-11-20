package razor

import "kard/src/model/dto"

type IRazor interface {
	Work(storeFunc func(taskDto *dto.TaskDto))
	CompletionData(storeFunc func(taskDto *dto.TaskDto), downloadIds ...int32)
}
