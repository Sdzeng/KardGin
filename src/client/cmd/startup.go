package main

import (
	"context"
	"fmt"
	_ "kard/src/client"
	"kard/src/client/cmd/crawler"
	"kard/src/global/helper"
	"kard/src/global/variable"
	"kard/src/model/dto"
	"kard/src/repository"
	"strconv"
	"time"

	"github.com/olivere/elastic"
)

var (
	ai *helper.AutoInc
)

func init() {
	ai = helper.NewAi(0, 1)
}

func main() {

	zimuCrawler := &crawler.ZimuCrawler{Open: true}

	crawlerWork(zimuCrawler)
}

func crawlerWork(crawler crawler.ICrawler) {
	crawler.Work(store)
}

func store(taskDto *dto.TaskDto) {
	downloadFileRepository := repository.DownloadFileFactory()
	err := downloadFileRepository.Save(taskDto)
	if err != nil {
		return
	}

	if variable.ES == nil {
		toConsole(taskDto)
	} else {
		toEs(taskDto)
	}

}

func toEs(taskDto *dto.TaskDto) {
	indexName := "subtitles_" + time.Now().Format("20060102")
	indexType := "_doc" // time.Now().Format("20060102")

	for _, subtitlesFile := range taskDto.SubtitlesFiles {
		toEsByBulk(indexName, indexType, taskDto, subtitlesFile)
	}
}

func toEsByBulk(indexName, indexType string, taskDto *dto.TaskDto, subtitlesFile *dto.SubtitlesFileDto) {

	bulkRequest := variable.ES.Bulk()
	batchNum := 10
	startAt := 0 * time.Second
	texts := []string{}

	for _, itemDto := range subtitlesFile.SubtitleItems {
		for _, text := range itemDto.Text {

			texts = append(texts, text)
			if (len(texts)-1)%batchNum == 0 {
				startAt = itemDto.StartAt
			}
			if len(texts)%batchNum == 0 {
				indexDto := &dto.SubtitlesIndexDto{
					Title:    taskDto.Name,
					SubTitle: subtitlesFile.FileName,
					Texts:    texts,
					StartAt:  int32(startAt.Seconds()),
					Lan:      taskDto.Lan,
				}
				indexReq := elastic.NewBulkIndexRequest().Index(indexName).Type(indexType).Id(strconv.Itoa(ai.Id())).Doc(indexDto)
				bulkRequest = bulkRequest.Add(indexReq)

				texts = []string{} //(lineTextSlice)[0:0]
				startAt = 0 * time.Second
			}
		}
	}

	if len(texts) > 0 {
		indexDto := &dto.SubtitlesIndexDto{
			Title:    taskDto.Name,
			SubTitle: subtitlesFile.FileName,
			Texts:    texts,
			StartAt:  int32(startAt.Seconds()),
			Lan:      taskDto.Lan,
		}
		indexReq := elastic.NewBulkIndexRequest().Index(indexName).Type(indexType).Id(strconv.Itoa(ai.Id())).Doc(indexDto)
		bulkRequest = bulkRequest.Add(indexReq)
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

func toConsole(taskDto *dto.TaskDto) {

	// for _, dto := range dtoSlice {
	// 	fmt.Printf("\n%v:%v", dto.TimeDuration, dto.Text)
	// }
	for _, subtitlesFile := range taskDto.SubtitlesFiles {
		fmt.Printf("\n解析成功：%v-%v", taskDto.Name, subtitlesFile.FileName)
	}
}
