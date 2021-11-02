package crawler

import "kard/src/model/dto"

type ICrawler interface {
	Work(seedUrlStr, qStr string, store func(taskDto *dto.TaskDto))
}
