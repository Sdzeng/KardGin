package main

import (
	"context"
	"flag"
	"fmt"
	_ "kard/src/client"
	"kard/src/client/cmd/razor"
	"kard/src/global/variable"
	"kard/src/model/dto"
	"kard/src/repository"
	"strconv"
	"time"

	"github.com/olivere/elastic/v7"
)

// var (
// 	ai *helper.AutoInc
// )

// func init() {
// 	ai = helper.NewAi(0, 1)
// }

func main() {

	seedUrl := flag.String("seed-url", "", "useage to search")
	comp := flag.Bool("comp", false, "useage to completion data")
	flag.Parse()

	variable.ZapLog.Sugar().Infof("seedUrl=%s comp=%v\n", *seedUrl, *comp)
	seedUrlStr := *seedUrl

	// a4kRazor := razor.NewA4KRazor(seedUrlStr)
	zmkRazor := razor.NewZmkRazor(seedUrlStr)
	razorWork(zmkRazor)
	// if *comp {
	// 	zmkRazor.CompletionData(store, 430)
	// }
	// zmkRazor.CompletionData(store, 763)

	var quit string
	fmt.Scan(&quit)
	variable.ZapLog.Info("退出")
}

func razorWork(razors ...razor.IRazor) {
	for _, raz := range razors {
		go raz.Work(store)
	}
}

func store(taskDto *dto.TaskDto) {
	downloadFileRepository := repository.DownloadPathsFactory()
	//清洗数据4
	err := downloadFileRepository.KSave(taskDto)
	if err != nil {
		variable.ZapLog.Sugar().Errorf("保存数据报错：%v", err)
		return
	}

	variable.ZapLog.Sugar().Infof("新加数据：%v", taskDto.Name)

	if taskDto.Error != nil {
		return
	}

	if variable.ES == nil {
		toConsole(taskDto)
	} else {
		toEs(taskDto)
	}
}

func toEs(taskDto *dto.TaskDto) {
	indexName := variable.IndexName //+ time.Now().Format("20060102")
	indexType := "_doc"             // time.Now().Format("20060102")
	nowStr := time.Now().Format(variable.TimeFormat)

	deleteEs(indexName, taskDto.DelDownloadPathIds)

	for _, subtitlesFile := range taskDto.SubtitlesFiles {
		if subtitlesFile.DbNew {
			toEsByBulk(indexName, indexType, nowStr, taskDto, subtitlesFile)
		}
	}
}

func deleteEs(indexName string, pathIds []int32) {

	if len(pathIds) <= 0 {
		return
	}
	ids := make([]interface{}, 0)
	for _, pathId := range pathIds {
		ids = append(ids, pathId)
	}

	q := elastic.NewTermsQuery("path_id", ids...)

	variable.ES.DeleteByQuery().
		Index(indexName).
		Query(q).
		Pretty(true).
		Do(context.TODO())
}

func toEsByBulk(indexName, indexType string, nowStr string, taskDto *dto.TaskDto, subtitlesFile *dto.SubtitlesFileDto) {

	bulkRequest := variable.ES.Bulk()

	indexId := strconv.FormatInt(int64(subtitlesFile.DownloadPathId), 10)
	partId := 1

	for _, itemDto := range subtitlesFile.SubtitleItems {

		indexDto := &dto.SubtitlesIndexDto{
			DownloadPathId: subtitlesFile.DownloadPathId,
			PartId:         int32(partId),
			Title:          taskDto.Name,
			SubTitle:       subtitlesFile.Name,
			Texts:          itemDto.Texts,
			StartAt:        int32(itemDto.StartAt.Seconds()),
			Lan:            taskDto.Lan,
			CreateTime:     nowStr,
			// PicPath:        "",
		}

		indexReq := elastic.NewBulkIndexRequest().Index(indexName).Type(indexType).Id(indexId + "_" + strconv.Itoa(partId)).Doc(indexDto)
		bulkRequest = bulkRequest.Add(indexReq)

		partId++
	}

	if bulkRequest.NumberOfActions() <= 0 {
		return
	}

	bulkResponse, err := bulkRequest.Do(context.TODO())
	if err != nil {
		variable.ZapLog.Sugar().Infof("批量插入es失败：%v", err)
	}
	if bulkResponse == nil {
		variable.ZapLog.Sugar().Infof("批量插入es：expected bulkResponse to be != nil; got nil")
	}
	if bulkRequest.NumberOfActions() != 0 {
		variable.ZapLog.Sugar().Infof("expected bulkRequest.NumberOfActions %d; got %d", 0, bulkRequest.NumberOfActions())
	}

}

func toConsole(taskDto *dto.TaskDto) {

	// for _, dto := range dtoSlice {
	// 	variable.ZapLog.Sugar().Infof("%v:%v", dto.TimeDuration, dto.Text)
	// }
	for _, subtitlesFile := range taskDto.SubtitlesFiles {
		variable.ZapLog.Sugar().Infof("解析成功：%v-%v", taskDto.Name, subtitlesFile.FileName)
	}
}
