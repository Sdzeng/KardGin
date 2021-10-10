package main

import (
	"context"
	"fmt"
	_ "kard/src/client"
	"kard/src/client/cmd/crawler"
	"kard/src/global/variable"
	"kard/src/model/dto"
	"time"

	"github.com/olivere/elastic"
)

func main() {

	zimuCrawler := &crawler.ZimuCrawler{Open: true}

	crawlerWork(zimuCrawler)
}

func crawlerWork(crawler crawler.ICrawler) {
	crawler.Work(store)
}

func store(dtos []*dto.SubtitlesIndexDto) {
	if variable.ES == nil {
		toConsole(dtos)
	} else {
		toEs(dtos)
	}
}

func toEs(dtos []*dto.SubtitlesIndexDto) {
	indexName := "subtitles_" + time.Now().Format("20060102")
	indexType := "_doc" // time.Now().Format("20060102")

	bulkRequest := variable.ES.Bulk()
	for _, dto := range dtos {
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

func toConsole(dtos []*dto.SubtitlesIndexDto) {

	// for _, dto := range dtoSlice {
	// 	fmt.Printf("\n%v:%v", dto.TimeDuration, dto.Text)
	// }

	fmt.Printf("\n解析成功：%v", dtos[0].Title)
}
