package main

import (
	"fmt"
	"kard/src/dto"
	"os"
	"strings"

	"github.com/asticode/go-astisub"
	"github.com/axgle/mahonia"
)

func parseFile(urlDto *dto.UrlDto, workerQueue chan *dto.UrlDto) {
	filePathLower := ""
	var subtitles *astisub.Subtitles
	decoder := mahonia.NewDecoder("utf8")
	for _, filePath := range urlDto.FilePaths {

		filePathLower = strings.ToLower(filePath)
		switch {
		case strings.HasSuffix(filePathLower, ".ass"):
			file, err := os.Open(filePath)
			if err != nil {
				continue
			}
			defer file.Close()

			subtitles, err = astisub.ReadFromSRT(decoder.NewReader(file))
			if err != nil {
				continue
			}

		case strings.HasSuffix(filePathLower, ".srt"):
			file, err := os.Open(filePath)
			if err != nil {
				continue
			}
			defer file.Close()

			subtitles, err = astisub.ReadFromSRT(decoder.NewReader(file))
			if err != nil {
				continue
			}

		default:
			fmt.Println("未识别的文件" + urlDto.FileName)
			continue
		}

		for _, item := range subtitles.Items {
			for _, line := range item.Lines {
				lineText := line.VoiceName + "："
				for _, lineItem := range line.Items {
					lineText += lineItem.Text + "\n"
				}
				fmt.Println(lineText)
			}

		}

	}

}
