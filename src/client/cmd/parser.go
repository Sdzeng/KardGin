package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"kard/src/dto"
	"strings"

	"github.com/asticode/go-astisub"
)

func parseFile(urlDto *dto.UrlDto, workerQueue chan *dto.UrlDto) {
	filePathLower := ""
	var subtitles *astisub.Subtitles
	for _, filePath := range urlDto.FilePaths {

		filePathLower = strings.ToLower(filePath)
		switch {
		case strings.HasSuffix(filePathLower, ".ass"):
			assFileBytes, err := ioutil.ReadFile(filePath)
			if err != nil {
				continue
			}

			subtitles, err = astisub.ReadFromSRT(bytes.NewReader(assFileBytes))

		case strings.HasSuffix(filePathLower, ".srt"):
			astisub.ReadFromSRT(bytes.NewReader([]byte("00:01:00.000 --> 00:02:00.000\nCredits")))

		default:
			fmt.Println("未识别的文件" + urlDto.FileName)
			return
		}
	}

}
