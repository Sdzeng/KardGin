package main

import (
	"context"
	"fmt"
	"kard/src/dto"

	"github.com/asticode/go-astisub"
	"github.com/olivere/elastic/v7"
)

var es *elastic.Client
var ctx = context.Background()
var esUrl string = "http://localhost:9200"

func init() {
	var err error
	es, err = elastic.NewClient(elastic.SetURL(esUrl), elastic.SetSniff(false))

	if err != nil {
		fmt.Println("初始化es客户端连接失败", err)
	}

}

func parseFile(urlDto *dto.UrlDto, workerQueue chan *dto.UrlDto) {
	// filePathLower := ""
	// var subtitles *astisub.Subtitles
	//decoder := mahonia.NewDecoder("utf8")
	for _, filePath := range urlDto.FilePaths {
		subtitles, err := astisub.Open(astisub.Options{Filename: filePath})
		if err != nil {
			continue
		}
		// filePathLower = strings.ToLower(filePath)
		// switch {
		// case strings.HasSuffix(filePathLower, ".ass"):
		// 	file, err := os.Open(filePath)
		// 	if err != nil {
		// 		continue
		// 	}
		// 	defer file.Close()

		// 	subtitles, err = astisub.Open(file)
		// 	if err != nil {
		// 		continue
		// 	}

		// case strings.HasSuffix(filePathLower, ".srt"):
		// 	file, err := os.Open(filePath)
		// 	if err != nil {
		// 		continue
		// 	}
		// 	defer file.Close()

		// 	subtitles, err = astisub.ReadFromSRT(file)
		// 	if err != nil {
		// 		continue
		// 	}

		// default:
		// 	fmt.Println("未识别的文件" + urlDto.FileName)
		// 	continue
		// }

		for _, item := range subtitles.Items {
			for _, line := range item.Lines {
				lineText := line.VoiceName + "："
				for _, lineItem := range line.Items {
					lineText += lineItem.Text + "\n"

					e1 := Employee{"Jane", "Smith", 32, "I like to collect rock albums", []string{"music"}}
					put1, err := es.Index().
						Index("subtitles").
						Type("employee").
						Id(strings.i).
						BodyJson(e1).
						Do(context.Background())
					if err != nil {
						panic(err)
					}

				}
				fmt.Println(lineText)
			}

		}

	}

}
