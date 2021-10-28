package main

import (
	"context"
	"fmt"
	_ "kard/src/client"
	"kard/src/client/cmd/crawler"
	"kard/src/global/variable"
	"kard/src/model/dto"
	"kard/src/repository"
	"strconv"

	"github.com/olivere/elastic"
)

// var (
// 	ai *helper.AutoInc
// )

// func init() {
// 	ai = helper.NewAi(0, 1)
// }

func main() {

	zmkCrawler := &crawler.ZmkCrawler{Open: true}

	crawlerWork(zmkCrawler)
}

func crawlerWork(crawler crawler.ICrawler) {
	crawler.Work(store)
}

func store(taskDto *dto.TaskDto) {
	downloadFileRepository := repository.DownloadPathsFactory()
	err := downloadFileRepository.Save(taskDto)
	if err != nil {
		return
	}

	fmt.Printf("\n新加数据：%v", taskDto.Name)

	if variable.ES == nil {
		toConsole(taskDto)
	} else {
		toEs(taskDto)
	}

}

func toEs(taskDto *dto.TaskDto) {
	if !taskDto.DbNew {
		fmt.Printf("\n跳过已存在数据(漏网之鱼)：%v", taskDto.Name)
		return
	}

	indexName := variable.IndexName //+ time.Now().Format("20060102")
	indexType := "_doc"             // time.Now().Format("20060102")

	for _, subtitlesFile := range taskDto.SubtitlesFiles {
		if subtitlesFile.DbNew {
			toEsByBulk(indexName, indexType, taskDto, subtitlesFile)
		}
	}
}

func toEsByBulk(indexName, indexType string, taskDto *dto.TaskDto, subtitlesFile *dto.SubtitlesFileDto) {

	bulkRequest := variable.ES.Bulk()

	indexId := strconv.FormatInt(int64(subtitlesFile.DownloadPathId), 10)
	partId := 0

	for _, itemDto := range subtitlesFile.SubtitleItems {

		indexDto := &dto.SubtitlesIndexDto{
			DownloadPathId: subtitlesFile.DownloadPathId,
			Title:          taskDto.Name,
			SubTitle:       subtitlesFile.FileName,
			Texts:          itemDto.Texts,
			StartAt:        int32(itemDto.StartAt.Seconds()),
			Lan:            taskDto.Lan,
			// PicPath:        "",
		}
		partId++
		indexReq := elastic.NewBulkIndexRequest().Index(indexName).Type(indexType).Id(indexId + "_part" + strconv.Itoa(partId)).Doc(indexDto)
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
