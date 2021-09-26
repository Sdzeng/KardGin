package main

import (
	"context"
	"fmt"
	"kard/src/dto"
	"kard/src/global/variable"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/asticode/go-astisub"
	"github.com/olivere/elastic/v7"
)

var (
	ai *AutoInc
)

func init() {
	ai = NewAi(0, 1)
}

func parseFile(urlDto *dto.UrlDto, workerQueue chan *dto.UrlDto) {
	batchNum := 10
	dtoSlice := []*dto.SubtitlesIndexDto{}

	for _, filePath := range urlDto.FilePaths {
		subtitles, err := astisub.Open(astisub.Options{Filename: filePath})
		if err != nil {
			continue
		}

		timeDuration := ""
		subTitle := getPathFileName(filePath)
		itemLen := len(subtitles.Items)
		lineTextSlice := []string{}

		for itemIndex, item := range subtitles.Items {

			for _, line := range item.Lines {
				//lineText := line.VoiceName + "："
				for _, lineItem := range line.Items {
					lineTextSlice = append(lineTextSlice, lineItem.Text)

				}
				//fmt.Println(lineText)
			}

			if itemIndex%batchNum == 0 {
				timeDuration = item.StartAt.String()
			}

			if ((itemIndex+1)%batchNum == 0 || (itemIndex+1) == itemLen) && len(lineTextSlice) > 0 {

				indexDto := &dto.SubtitlesIndexDto{
					IndexId:      strconv.Itoa(ai.Id()),
					Title:        urlDto.Name,
					SubTitle:     subTitle,
					Text:         lineTextSlice,
					TimeDuration: timeDuration,
					Lan:          urlDto.Lan,
				}

				dtoSlice = append(dtoSlice, indexDto)

				lineTextSlice = (lineTextSlice)[0:0]
				timeDuration = ""
			}

		}

	}

	toEs(dtoSlice)
}

func toEs(dtoSlice []*dto.SubtitlesIndexDto) {
	indexName := "subtitles_" + time.Now().Format("20060102")
	indexType := "_doc" // time.Now().Format("20060102")

	bulkRequest := variable.ES.Bulk()
	for _, dto := range dtoSlice {
		indexReq := elastic.NewBulkIndexRequest().Index(indexName).Type(indexType).Id(dto.IndexId).Doc(dto)
		bulkRequest = bulkRequest.Add(indexReq)

		fmt.Println(dto.Text)
	}

	if bulkRequest.NumberOfActions() <= 0 {
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
}

func getPathFileName(filePath string) string {

	filePathSlice := strings.Split(filePath, "\\")
	fileFullName := filePathSlice[len(filePathSlice)-1]
	fileSuffix := path.Ext(fileFullName)

	return strings.TrimSuffix(fileFullName, fileSuffix)
}
