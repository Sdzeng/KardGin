package controller

import (
	"context"
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

func search(es_index string, page int, li int, search_word string) interface{} {
	p := (page - 1) * li
	// collapsedata := elastic.NewCollapseBuilder("texts")
	esq := elastic.NewBoolQuery()
	esq.Should(elastic.NewMatchQuery("texts", search_word))
	esq.Should(elastic.NewMatchQuery("title", search_word))
	esq.Should(elastic.NewMatchQuery("subtitle", search_word))
	search := variable.ES.Search().
		Index(es_index).
		From(p).Size(li).
		Query(esq).
		// Collapse(collapsedata).
		Pretty(true)
	searchResult, err := search.Do(context.Background())
	if err != nil {
		panic(err)
	} else {
		return searchResult
	}
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
