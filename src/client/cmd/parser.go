package main

import (
	"fmt"
	"kard/src/dto"
	"strings"
)

func parseFile(urlDto *dto.UrlDto, workerQueue chan *dto.UrlDto) {
	fileNameLower := strings.ToLower(urlDto.FileName)
	switch {
	case strings.HasSuffix(fileNameLower, ".rar"):
		uncompress(urlDto, workerQueue)
		//解压rar
	case strings.HasSuffix(fileNameLower, ".zip"):
		uncompress(urlDto, workerQueue)
	case strings.HasSuffix(fileNameLower, ".7z"):
		uncompress(urlDto, workerQueue)
	case strings.HasSuffix(fileNameLower, ".ass"):
		parse(urlDto, workerQueue)
	case strings.HasSuffix(fileNameLower, ".srt"):
		parse(urlDto, workerQueue)
	default:
		fmt.Println("未识别的文件" + urlDto.FileName)
	}
}

func uncompress(urlDto *dto.UrlDto, workerQueue chan *dto.UrlDto) {

}

func parse(urlDto *dto.UrlDto, workerQueue chan *dto.UrlDto) {

}
