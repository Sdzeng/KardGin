package controller

import (
	"context"
	"fmt"
	"io"
	"kard/src/client/web/response"
	"kard/src/global/variable"
	"kard/src/model/dto"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
)

type HomeController struct {
}

func (c *HomeController) Search(context *gin.Context) {
	json := make(map[string]interface{})
	context.BindJSON(&json)
	searchWord := json["search_word"].(string)
	pageCount := int(json["page_count"].(float64))

	indexName := "subtitles_20060102" //+ time.Now().Format("20060102")
	data := search(indexName, pageCount, searchWord)

	if data != nil {
		response.Success(context, variable.CurdStatusOkMsg, data)
	} else {
		response.Fail(context, variable.CurdSelectFailCode, variable.CurdSelectFailMsg, "")
	}
}

func (c *HomeController) SearchScroll(context *gin.Context) {
	json := make(map[string]interface{})
	context.BindJSON(&json)
	scrollId := json["scroll_id"].(string)

	// indexName := "subtitles_20060102"
	data := scrollSearch(scrollId)

	if data != nil {
		response.Success(context, variable.CurdStatusOkMsg, data)
	} else {
		response.Fail(context, variable.CurdSelectFailCode, variable.CurdSelectFailMsg, "")
	}
}

func search(es_index string, pageCount int, search_word string) *dto.SearchResultDto {
	searchResultDto := new(dto.SearchResultDto)
	// p := (page - 1) * li
	// collapsedata := elastic.NewCollapseBuilder("texts")
	esq := elastic.NewBoolQuery()
	esq = esq.Should(
		elastic.NewWildcardQuery("title", search_word),
		elastic.NewWildcardQuery("subtitle", search_word),
		elastic.NewWildcardQuery("texts", search_word),
	)

	fsc := elastic.NewFetchSourceContext(true).Include("path_id", "title", "subtitle", "texts", "lan")

	hl := elastic.NewHighlight().Fields(
		elastic.NewHighlighterField("title"),
		elastic.NewHighlighterField("subtitle"),
		elastic.NewHighlighterField("texts"),
	)

	// search := variable.ES.Search().
	// 	Index(es_index).
	// 	Highlight(hl).
	// 	Query(esq).
	// 	FetchSourceContext(fsc).
	// 	// Collapse(collapsedata).
	// 	Pretty(true)

	scroll := variable.ES.Scroll(es_index).
		Scroll("5m").
		Query(esq).
		Highlight(hl).
		FetchSourceContext(fsc).
		Size(pageCount).
		TrackTotalHits(true).
		FilterPath("hits.total", "hits.hits._source", "hits.hits.highlight").
		// // Collapse(collapsedata).
		Pretty(true)

	res, err := scroll.Do(context.TODO())
	if err == io.EOF {
		return searchResultDto
	}
	if err != nil {
		fmt.Print(err)
		return searchResultDto
	}
	if res == nil {
		fmt.Printf("expected results != nil; got nil")
		return searchResultDto
	}
	if res.Hits == nil {
		fmt.Printf("expected results.Hits != nil; got nil")
		return searchResultDto
	}
	// if want, have := int64(3), res.TotalHits(); want != have {
	// 	fmt.Printf("expected results.TotalHits() = %d; got %d", want, have)
	// 	continue
	// }
	// if want, have := 1, len(res.Hits.Hits); want != have {
	// 	fmt.Printf("expected len(results.Hits.Hits) = %d; got %d", want, have)
	// 	continue
	// }

	//清除后，该查询就清除了，游标就无用
	// err = scroll.Clear(context.TODO())
	// if err != nil {
	// 	fmt.Print(err)
	// 	return searchResultDto
	// }

	searchResultDto.ScrollId = res.ScrollId
	searchResultDto.SearchHits = res.Hits.Hits
	return searchResultDto
}

func scrollSearch(scrollId string) *dto.SearchResultDto {
	searchResultDto := new(dto.SearchResultDto)
	//游标（缓存）有效期延长*分钟
	scroll := variable.ES.Scroll().Scroll("2m").ScrollId(scrollId)

	res, err := scroll.Do(context.TODO())
	if err == io.EOF {
		return searchResultDto
	}
	if err != nil {
		fmt.Print(err)
		return searchResultDto
	}
	if res == nil {
		fmt.Printf("expected results != nil; got nil")
		return searchResultDto
	}
	if res.Hits == nil {
		fmt.Printf("expected results.Hits != nil; got nil")
		return searchResultDto
	}

	// err = scroll.Clear(context.TODO())
	// if err != nil {
	// 	fmt.Print(err)
	// 	return searchResultDto
	// }

	searchResultDto.ScrollId = res.ScrollId
	searchResultDto.SearchHits = res.Hits.Hits
	return searchResultDto
}

// func (c *HomeController) GetCover(context *gin.Context) {

// 	today := time.Now()
// 	// userName := context.GetString(consts.ValidatorPrefix + "user_name")
// 	// page := context.GetFloat64(consts.ValidatorPrefix + "page")
// 	// limits := context.GetFloat64(consts.ValidatorPrefix + "limits")
// 	// limitStart := (page - 1) * limits
// 	videoDto, _ := repository.CreateVideoFactory().GetCover(today)
// 	if videoDto != nil {
// 		response.Success(context, response.CurdStatusOkMsg, videoDto)
// 	} else {
// 		response.Fail(context, response.CurdSelectFailCode, response.CurdSelectFailMsg, "")
// 	}
// }

// func (c *HomeController) ExtractSubtitles(context *gin.Context) {

// 	var txtName = time.Now().Second()
// 	cmdStr := "-i output.mkv -an -vn -bsf:s mov2textsub -scodec copy -f rawvideo " + strconv.Itoa(txtName) + ".txt"
// 	cmdArguments := strings.Split(cmdStr, " ")
// 	// []string{"-i", "divx.avi", "-c:v", "libx264", "-crf", "20", "-c:a", "aac", "-strict", "-2", "video1-fix.ts"}

// 	cmd := exec.Command("ffmpeg", cmdArguments...)

// 	var out bytes.Buffer
// 	cmd.Stdout = &out
// 	err := cmd.Run()
// 	if err != nil {
// 		response.Success(context, response.CurdStatusOkMsg, "成功")
// 	} else {

// 		response.Fail(context, response.CurdSelectFailCode, err.Error(), "")
// 	}
// }
