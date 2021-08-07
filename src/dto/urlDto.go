package dto

import "net/http"

type UrlDto struct {
	WorkType    string
	Refers      []string
	Cookies     []*http.Cookie
	DownloadUrl string
	Name        string
	FileName    string
	FilePaths   []string
	Lan         string
	Subtitles   string
}
