package dto

type SubtitlesFileDto struct {
	FilePath       string //rar路径
	FileSum        string //md5值
	Content        *string
	FileName       string              //解压后文件的名字
	SubtitleItems  []*SubtitlesItemDto //解析后的文本
	DownloadPathId int32               //download_paths表的id

}
