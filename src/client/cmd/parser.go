package main

import (
	"context"
	"fmt"
	"kard/src/dto"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/asticode/go-astisub"
	"github.com/olivere/elastic/v7"
)

var indexId = 0

var es *elastic.Client
var esUrl string = "http://localhost:9200"

func init() {
	ps := es.Ping(esUrl)
	if ps == nil {
		fmt.Println("初始化es客户端连接失败")
	}

	var err error
	es, err = elastic.NewClient(elastic.SetURL(esUrl), elastic.SetSniff(false), elastic.SetBasicAuth("elastic", "123456"))

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
		numberOfActions := 0
		subTitle := getPathFileName(filePath)
		indexName := "subtitles_" + time.Now().Format("20060102")
		indexType := "_doc" // time.Now().Format("20060102")
		bulkRequest := es.Bulk()

		for _, item := range subtitles.Items {

			lineTextSlice := []string{}
			for _, line := range item.Lines {
				//lineText := line.VoiceName + "："
				for _, lineItem := range line.Items {
					lineTextSlice = append(lineTextSlice, lineItem.Text)

				}
				//fmt.Println(lineText)
			}

			if len(lineTextSlice) <= 0 {
				continue
			}

			indexDto := dto.SubtitlesIndexDto{Title: urlDto.Name, SubTitle: subTitle, Text: lineTextSlice, TimeDuration: item.StartAt.String(), Lan: urlDto.Lan}
			indexId++
			numberOfActions++
			indexReq := elastic.NewBulkIndexRequest().Index(indexName).Type(indexType).Id(strconv.Itoa(indexId)).Doc(indexDto)

			bulkRequest = bulkRequest.Add(indexReq)

		}

		if numberOfActions <= 0 {

			return
		}

		if bulkRequest.NumberOfActions() != numberOfActions {
			fmt.Printf("expected bulkRequest.NumberOfActions %d; got %d", numberOfActions, bulkRequest.NumberOfActions())
			return
		}

		bulkResponse, err := bulkRequest.Do(context.TODO())
		if err != nil {
			fmt.Printf("批量插入es失败：%v", err)
		}
		if bulkResponse == nil {
			fmt.Printf("批量插入es：expected bulkResponse to be != nil; got nil")
		}
		if bulkRequest.NumberOfActions() != 0 {
			fmt.Printf("expected bulkRequest.NumberOfActions %d; got %d", 0, bulkRequest.NumberOfActions())
		}

		// // Document with Id="1" should not exist
		// exists, err := es.Exists().Index(indexName).Id("1").Do(context.TODO())
		// if err != nil {
		// 	fmt.Println(err)
		// }
		// if exists {
		// 	fmt.Printf("expected exists %v; got %v", false, exists)
		// }

	}

}

func getPathFileName(filePath string) string {

	filePathSlice := strings.Split(filePath, "\\")
	fileFullName := filePathSlice[len(filePathSlice)-1]
	fileSuffix := path.Ext(fileFullName)

	return strings.TrimSuffix(fileFullName, fileSuffix)
}
