package dto

import (
	"net/http"
	"sync"
)

type TaskDto struct {
	// SearchKeyword string //搜索的关键词
	WorkType string
	PageNum  int
	Refers   []string
	Cookies  []*http.Cookie
	Wg       *sync.WaitGroup

	Name string //页面显示的字幕名称
	// Razor         string //剃刀
	Lan           string //字幕采用的语言
	SubtitlesType string //字幕文件格式 如SSA
	InfoUrl       string //详情页url
	DownloadId    int32  //downloads表id

	DownloadUrl string //下载链接
	// DownloadUrlFileName string              //下载链接带的文件名称（比如1.rar）
	SubtitlesFiles []*SubtitlesFileDto //解压后存储的文件列表
	StoreFunc      func(taskDto *TaskDto)

	Error error
	// DbNew   bool //是否为新数据
	// EsIndex string
}

// func (dto *TaskDto) ContainsKeyword(work string) bool {
// 	if len(dto.SearchKeyword) <= 0 {
// 		return false
// 	}

// 	kws := strings.FieldsFunc(dto.SearchKeyword, unicode.IsSpace) //strings.Split(dto.SearchKeyword, " ")
// 	for _, kw := range kws {
// 		if !strings.Contains(work, kw) {
// 			return false
// 		}
// 	}
// 	return true
// }
