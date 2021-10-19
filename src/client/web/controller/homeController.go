package controller

import (
	"context"
	"fmt"
	"io"
	"kard/src/client/web/response"
	"kard/src/global/variable"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
)

type HomeController struct {
}

func (c *HomeController) Search(context *gin.Context) {
	json := make(map[string]interface{})
	context.BindJSON(&json)
	searchWord := json["search_word"].(string)

	indexName := "subtitles_20060102" //+ time.Now().Format("20060102")
	data := search(indexName, 1, 100, searchWord)

	if data != nil {
		response.Success(context, variable.CurdStatusOkMsg, data)
	} else {
		response.Fail(context, variable.CurdSelectFailCode, variable.CurdSelectFailMsg, "")
	}
}

func search(es_index string, page int, li int, search_word string) []*elastic.SearchHit {
	// p := (page - 1) * li
	// collapsedata := elastic.NewCollapseBuilder("texts")
	esq := elastic.NewBoolQuery()
	esq = esq.Should(
		elastic.NewWildcardQuery("title.keyword", search_word),
		elastic.NewWildcardQuery("subtitle.keyword", search_word),
		elastic.NewWildcardQuery("texts.keyword", search_word),
	)

	fsc := elastic.NewFetchSourceContext(true).Include("path_id", "title", "subtitle", "texts", "lan")

	hl := elastic.NewHighlight().Fields(
		elastic.NewHighlighterField("title.keyword"),
		elastic.NewHighlighterField("subtitle.keyword"),
		elastic.NewHighlighterField("texts.keyword"),
	)

	// search := variable.ES.Search().
	// 	Index(es_index).
	// 	Highlight(hl).
	// 	Query(esq).
	// 	FetchSourceContext(fsc).
	// 	// Collapse(collapsedata).
	// 	Pretty(true)

	scroll := variable.ES.Scroll(es_index).
		Query(esq).
		Highlight(hl).
		FetchSourceContext(fsc).
		Size(20).
		TrackTotalHits(true).
		// // Collapse(collapsedata).
		Pretty(true)

	result := []*elastic.SearchHit{}
	for {
		res, err := scroll.Do(context.TODO())
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Print(err)
		}
		if res == nil {
			fmt.Printf("expected results != nil; got nil")
			continue
		}
		if res.Hits == nil {
			fmt.Printf("expected results.Hits != nil; got nil")
			continue
		}
		// if want, have := int64(3), res.TotalHits(); want != have {
		// 	fmt.Printf("expected results.TotalHits() = %d; got %d", want, have)
		// 	continue
		// }
		// if want, have := 1, len(res.Hits.Hits); want != have {
		// 	fmt.Printf("expected len(results.Hits.Hits) = %d; got %d", want, have)
		// 	continue
		// }

		result = append(result, res.Hits.Hits...)
	}

	err := scroll.Clear(context.TODO())
	if err != nil {
		fmt.Print(err)
	}

	return result
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
