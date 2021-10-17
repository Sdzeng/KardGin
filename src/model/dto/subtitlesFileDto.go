package dto

import "io"

type SubtitlesFileDto struct {
	FilePath       string //rar路径
	Reader         io.Reader
	FileName       string              //解压后文件的名字
	SubtitleItems  []*SubtitlesItemDto //解析后的文本
	DownloadPathId int32               //download_paths表的id
}
