package dto

type SubtitlesFileDto struct {
	FilePath       string              //本地存储路径
	FileName       string              //解压后文件的名字
	SubtitleItems  []*SubtitlesItemDto //解析后的文本
	DownloadPathId int32               //download_paths表的id
}
