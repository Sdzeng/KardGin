package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"kard/src/client/web/response"
	"kard/src/global/variable"
	"kard/src/model/dto"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
)

type HomeController struct {
}

var (
	hitReplacer *strings.Replacer
)

func init() {

	replaceKeywords := []string{
		"<em>", "",
		"</em>", "",
	}

	hitReplacer = strings.NewReplacer(replaceKeywords...)
}

func (c *HomeController) Index(context *gin.Context) {
	json := make(map[string]interface{})
	context.BindJSON(&json)

	pageCount := 15
	if json["page_count"] != nil {
		pageCount = int(json["page_count"].(float64))
	}

	res := getIndexData(pageCount)
	if res == nil {
		response.Success(context, variable.CurdStatusOkMsg, nil)
		return
	}

	data := buildIndexResult(res)
	if data != nil {
		response.Success(context, variable.CurdStatusOkMsg, data)
	} else {
		response.Fail(context, variable.CurdSelectFailCode, variable.CurdSelectFailMsg, "")
	}
}

func (c *HomeController) ScrollIndex(context *gin.Context) {
	json := make(map[string]interface{})
	context.BindJSON(&json)

	if json["scroll_id"] == nil {
		response.Fail(context, variable.CurdSelectFailCode, variable.CurdSelectFailMsg, "")
		return
	}
	scrollId := json["scroll_id"].(string)

	res := getScrollData(scrollId)
	if res == nil {
		response.Success(context, variable.CurdStatusOkMsg, nil)
		return
	}

	data := buildIndexResult(res)
	if data != nil {
		response.Success(context, variable.CurdStatusOkMsg, data)
	} else {
		response.Fail(context, variable.CurdSelectFailCode, variable.CurdSelectFailMsg, "")
	}
}

func getIndexData(pageCount int) *elastic.SearchResult {

	es_index := variable.IndexName

	esq := elastic.NewTermQuery("part_id", 3)

	fsc := elastic.NewFetchSourceContext(true).Include("path_id", "title", "subtitle", "texts", "start_at", "lan", "pic_path", "create_time")

	scroll := variable.ES.Scroll(es_index).
		Scroll("5m").
		Query(esq).
		FetchSourceContext(fsc).
		Size(pageCount).
		TrackTotalHits(true).
		FilterPath("hits.total", "hits.hits._source").
		Sort("path_id", false)

	res, err := scroll.Do(context.TODO())
	if err == io.EOF {
		return nil
	}
	if err != nil {
		fmt.Print(err)
		return nil
	}
	if res == nil {
		fmt.Printf("expected results != nil; got nil")
		return nil
	}
	if res.Hits == nil {
		fmt.Printf("expected results.Hits != nil; got nil")
		return nil
	}

	return res
}

func buildIndexResult(res *elastic.SearchResult) *dto.EsResultDto {
	esResultDto := new(dto.EsResultDto)

	dtos := []*dto.SubtitlesIndexDto{}
	for _, hit := range res.Hits.Hits {
		dto := new(dto.SubtitlesIndexDto)
		if err := json.Unmarshal(hit.Source, dto); err != nil {
			continue
		}
		dtos = append(dtos, dto)
	}

	esResultDto.ScrollId = res.ScrollId
	esResultDto.TookInMillis = res.TookInMillis
	esResultDto.Total = res.Hits.TotalHits.Value
	esResultDto.SearchHits = dtos

	return esResultDto
}

func (c *HomeController) Search(context *gin.Context) {
	json := make(map[string]interface{})
	context.BindJSON(&json)

	pageCount := 15
	if json["page_count"] != nil {
		pageCount = int(json["page_count"].(float64))
	}

	if json["search_word"] == nil {
		response.Fail(context, variable.ValidatorParamsCheckFailCode, variable.ValidatorParamsCheckFailMsg, "")
		return
	}
	searchWord := json["search_word"].(string)
	// indexName := variable.IndexName //+ time.Now().Format("20060102")

	res := getSearchData(pageCount, searchWord)
	if res == nil {
		response.Success(context, variable.CurdStatusOkMsg, nil)
		return
	}

	data := buildSearchResult(res)
	if data != nil {
		response.Success(context, variable.CurdStatusOkMsg, data)
	} else {
		response.Fail(context, variable.CurdSelectFailCode, variable.CurdSelectFailMsg, "")
	}
}

func (c *HomeController) ScrollSearch(context *gin.Context) {
	json := make(map[string]interface{})
	context.BindJSON(&json)
	if json["scroll_id"] == nil {
		response.Fail(context, variable.CurdSelectFailCode, variable.CurdSelectFailMsg, "")
		return
	}
	scrollId := json["scroll_id"].(string)

	res := getScrollData(scrollId)
	if res == nil {
		response.Success(context, variable.CurdStatusOkMsg, nil)
		return
	}

	data := buildSearchResult(res)
	if data != nil {
		response.Success(context, variable.CurdStatusOkMsg, data)
	} else {
		response.Fail(context, variable.CurdSelectFailCode, variable.CurdSelectFailMsg, "")
	}
}

func getSearchData(pageCount int, search_word string) *elastic.SearchResult {

	// p := (page - 1) * li
	// collapsedata := elastic.NewCollapseBuilder("texts")
	es_index := variable.IndexName
	esq := elastic.NewBoolQuery()

	// // esq = esq.Should(
	// // 	elastic.NewWildcardQuery("title", search_word),
	// // 	elastic.NewWildcardQuery("subtitle", search_word),
	// // 	elastic.NewWildcardQuery("texts", search_word),
	// // )
	words := strings.Fields(search_word)
	for _, word := range words {
		esq = esq.Should(
			elastic.NewMatchPhraseQuery("title", word),
			elastic.NewMatchPhraseQuery("subtitle", word),
			elastic.NewMatchPhraseQuery("texts", word),
		)
	}

	// esq.QueryName()
	// esq := elastic.NewMultiMatchQuery(search_word, "title", "subtitle", "texts").Type("phrase")

	fsc := elastic.NewFetchSourceContext(true).Include("path_id", "part_id", "title", "subtitle", "texts", "start_at", "lan", "pic_path", "create_time")

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
		// Pretty(true).
		Sort("_score", false)

	res, err := scroll.Do(context.TODO())
	if err == io.EOF {
		return nil
	}
	if err != nil {
		fmt.Print(err)
		return nil
	}
	if res == nil {
		fmt.Printf("expected results != nil; got nil")
		return nil
	}
	if res.Hits == nil {
		fmt.Printf("expected results.Hits != nil; got nil")
		return nil
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

	return res
}

func getScrollData(scrollId string) *elastic.SearchResult {

	//游标（缓存）有效期延长*分钟
	scroll := variable.ES.Scroll().Scroll("2m").ScrollId(scrollId)

	res, err := scroll.Do(context.TODO())
	if err == io.EOF {
		return nil
	}
	if err != nil {
		fmt.Print(err)
		return nil
	}
	if res == nil {
		fmt.Printf("expected results != nil; got nil")
		return nil
	}
	if res.Hits == nil {
		fmt.Printf("expected results.Hits != nil; got nil")
		return nil
	}

	// err = scroll.Clear(context.TODO())
	// if err != nil {
	// 	fmt.Print(err)
	// 	return searchResultDto
	// }

	return res
}

func buildSearchResult(res *elastic.SearchResult) *dto.EsResultDto {
	esResultDto := new(dto.EsResultDto)

	dtos := []*dto.SubtitlesIndexDto{}
	for _, hit := range res.Hits.Hits {
		dto := new(dto.SubtitlesIndexDto)
		if err := json.Unmarshal(hit.Source, dto); err != nil {
			continue
		}

		for key, highlight := range hit.Highlight {
			switch key {
			case "title":
				dto.Title = highlight[0]
			case "subtitle":
				dto.SubTitle = highlight[0]
			case "texts":
				for _, hl := range highlight {
					t := hitReplacer.Replace(hl)
					for index, text := range dto.Texts {
						if text == t {
							dto.Texts[index] = hl
							break
						}

						if strings.Contains(text, t) {
							dto.Texts[index] = strings.ReplaceAll(text, t, hl)
							break
						}
					}
				}
			}
		}

		dtos = append(dtos, dto)
	}

	esResultDto.ScrollId = res.ScrollId
	esResultDto.TookInMillis = res.TookInMillis
	esResultDto.Total = res.Hits.TotalHits.Value
	esResultDto.SearchHits = dtos

	return esResultDto
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
