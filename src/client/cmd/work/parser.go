package main

import (
	"context"
	"fmt"
	"kard/src/global/variable"
	"kard/src/model/dto"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/asticode/go-astisub"
	"github.com/olivere/elastic"
)

var (
	ai *AutoInc
)

func init() {
	ai = NewAi(0, 1)
}

func parseFile(urlDto *dto.UrlDto) {
	defer func(d *dto.UrlDto) {
		d.Wg.Done()
	}(urlDto)

	batchNum := 10
	dtoSlice := []*dto.SubtitlesIndexDto{}

	for _, filePath := range urlDto.FilePaths {
		subtitles, err := astisub.Open(astisub.Options{Filename: filePath})
		if err != nil {
			continue
		}

		subTitle := getPathFileName(filePath)

		lineTextSlice := []string{}
		lineTextSliceLen := 0
		timeDuration := ""

		for _, item := range subtitles.Items {
			for _, line := range item.Lines {
				//lineText := line.VoiceName + "："
				for _, lineItem := range line.Items {
					if len(strings.Trim(lineItem.Text, " ")) <= 0 {
						continue
					}

					lineTextSlice = append(lineTextSlice, lineItem.Text)
					lineTextSliceLen := len(lineTextSlice)

					if (lineTextSliceLen-1)%batchNum == 0 {
						timeDuration = item.StartAt.String()
					}

					if lineTextSliceLen%batchNum == 0 {

						indexDto := &dto.SubtitlesIndexDto{
							IndexId:      strconv.Itoa(ai.Id()),
							Title:        urlDto.Name,
							SubTitle:     subTitle,
							Text:         lineTextSlice,
							TimeDuration: timeDuration,
							Lan:          urlDto.Lan,
						}

						dtoSlice = append(dtoSlice, indexDto)

						lineTextSlice = []string{} //(lineTextSlice)[0:0]
						timeDuration = ""
					}

				}
				//fmt.Println(lineText)
			}
		}

		lineTextSliceLen = len(lineTextSlice)
		if lineTextSliceLen > 0 {

			indexDto := &dto.SubtitlesIndexDto{
				IndexId:      strconv.Itoa(ai.Id()),
				Title:        urlDto.Name,
				SubTitle:     subTitle,
				Text:         lineTextSlice,
				TimeDuration: timeDuration,
				Lan:          urlDto.Lan,
			}

			dtoSlice = append(dtoSlice, indexDto)
		}

	}

	if len(dtoSlice) <= 0 {
		return
	}

	if variable.ES == nil {
		toConsole(dtoSlice)
	} else {
		toEs(dtoSlice)
	}
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

func toConsole(dtoSlice []*dto.SubtitlesIndexDto) {

	// for _, dto := range dtoSlice {
	// 	fmt.Printf("\n%v:%v", dto.TimeDuration, dto.Text)
	// }

	fmt.Printf("\n解析成功：%v", dtoSlice[0].Title)
}

func getPathFileName(filePath string) string {
	filePathSlice := strings.Split(filePath, "\\")
	fileFullName := filePathSlice[len(filePathSlice)-1]
	fileSuffix := path.Ext(fileFullName)

	return strings.TrimSuffix(fileFullName, fileSuffix)
}
