package crawler

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"kard/src/global/helper"
	"kard/src/global/variable"
	"kard/src/model/dto"
	"kard/src/repository"
)

type A4KCrawler struct {
	// StoreFunc func(dtoSlice []*dto.SubtitlesIndexDto)
	// Wg *sync.WaitGroup
	// helper.Parser
	// helper.Downloader
	Open bool
}

var (
	a4kPageNum         = `<a class="item pager__item--last" href="(\?page)=(\d+)"[^>]+?>(\s|\S)+?</a>`
	a4kFetchPageRegexp = regexp.MustCompile(a4kPageNum)

	// a4kTitleReg          = `<td class="w75pc">\s*<a href="(/sub(s)?/\d+.html)" target="_blank">(.+)</a>\s*</td>`
	a4kLanListReg        = `<div class="language">(\s|\n)*<span class="h4">(\s|\n)*(<i class="[^"]+?" data-content="[^"]+?"[^>]+?></i>\s*)+`
	a4kDownloadButtonReg = `(\s|\S)+?<div class="content">(\s|\n)*<h3>(\s|\n)*<a href="([^"]+?)"[^>]+?>`
	a4kTitleReg          = `(.+)</a>(\s|\n)*</h3>`
	a4kFetchListRegexp   = regexp.MustCompile(a4kLanListReg + a4kDownloadButtonReg + a4kTitleReg)

	a4kLanReg    = `<i class="[^"]+?" data-content="([^"]+?)"[^>]+?></i>\s*`
	a4kLanRegexp = regexp.MustCompile(a4kLanReg)

	a4kDownloadReg     = `<div class="download">(\s|\S)+?<a class="ui green button" href="([^"]+?)"(\s|\S)+?下载字幕</a>`
	a4kFetchInfoRegexp = regexp.MustCompile(a4kDownloadReg)
)

func (obj *A4KCrawler) Work(seedUrlStr, qStr string, store func(taskDto *dto.TaskDto)) {
	defer func() {
		if err := recover(); err != nil {
			helper.PrintError("Work", err.(error).Error(), true)
		}
	}()

	obj.search(seedUrlStr, qStr, store)
}

func (obj *A4KCrawler) search(seedUrlStr, qStr string, store func(taskDto *dto.TaskDto)) {

	// seedUrl := flag.String("seed-url", "", "useage to search")
	// q := flag.String("q", "", "useage to search")
	// flag.Parse()

	// fmt.Printf("seedUrl=%s\n", *seedUrl)
	// fmt.Printf("q=%s\n", *q)
	// seedUrlStr := *seedUrl
	// qStr := *q

	var reqUrl string
	var pageNum int = 1
	if len(seedUrlStr) > 0 {
		reqUrl = seedUrlStr
		if values, err := url.ParseQuery(strings.Split(seedUrlStr, "?")[1]); err != nil {
			pageNum = 0
		} else if p, err2 := strconv.Atoi(values.Get("page")); err2 == nil {
			pageNum = p
		}

	} else if len(qStr) > 0 {
		v := url.Values{}
		v.Add("term", qStr)
		reqUrl = "https://www.a4k.net/search?" + v.Encode()
	} else {
		reqUrl = "https://www.a4k.net"
	}

	// workerQueue := make(chan *dto.UrlDto, 1)

	taskDto := &dto.TaskDto{SearchKeyword: qStr, WorkType: variable.FecthPage, PageNum: pageNum, DownloadUrl: reqUrl, Wg: &sync.WaitGroup{}, StoreFunc: store, EsIndex: variable.IndexName, Crawler: "a4k"}

	obj.insertQueue(taskDto)
	taskDto.Wg.Wait()
}

func (obj *A4KCrawler) insertQueue(newDto *dto.TaskDto) {
	if !obj.Open {
		return
	}
	switch newDto.WorkType {
	case variable.FecthPage:
		helper.Sleep(newDto.Crawler, newDto.WorkType, "s", 1, 8)
		obj.fetchPage(newDto)
	case variable.FecthList:
		helper.Sleep(newDto.Crawler, newDto.WorkType, "s", 1, 10)
		obj.fetchList(newDto)
	case variable.FecthInfo:
		helper.WorkClock(newDto.Crawler)
		helper.Sleep(newDto.Crawler, newDto.WorkType, "m", 10, 45)
		obj.fetchInfo(newDto)
	case variable.Parse:
		helper.Sleep(newDto.Crawler, newDto.WorkType, "s", 1, 5)
		obj.parse(newDto)
	}
}

func (obj *A4KCrawler) fetchPage(taskDto *dto.TaskDto) {

	html, cookies, err := helper.LoadHtml(taskDto)
	if err != nil {
		return
	}

	pageItems := a4kFetchPageRegexp.FindAllStringSubmatch(*html, -1)
	pathUrl := ""
	endPageNum := 0
	pageNum := taskDto.PageNum
	if len(pageItems) <= 0 {
		endPageNum = taskDto.PageNum
	} else {
		pathUrl = pageItems[0][1]

		if !strings.HasPrefix(pathUrl, "http:") && !strings.HasPrefix(pathUrl, "https:") {
			pathUrl = helper.UrlJoin(pathUrl, "https://www.a4k.net")
		}

		endPageNum, _ = strconv.Atoi(pageItems[0][2])
	}

	for pageNum <= endPageNum {
		variable.ZapLog.Sugar().Infof("处理第%v页 共%v页", pageNum, endPageNum)
		url := pathUrl + "=" + strconv.Itoa(pageNum)
		newDto := &dto.TaskDto{SearchKeyword: taskDto.SearchKeyword, WorkType: variable.FecthList, DownloadUrl: url, Cookies: cookies, Wg: taskDto.Wg, StoreFunc: taskDto.StoreFunc, EsIndex: taskDto.EsIndex, Crawler: taskDto.Crawler}
		obj.insertQueue(newDto)
		pageNum++
	}

}

func (obj *A4KCrawler) fetchList(taskDto *dto.TaskDto) {

	html, cookies, err := helper.LoadHtml(taskDto)
	if err != nil {
		return
	}

	downloadRepository := repository.DownloadsFactory()

	//获取子页信息
	items := a4kFetchListRegexp.FindAllStringSubmatch(*html, -1)

	for _, item := range items {
		title := item[8]

		if len(taskDto.SearchKeyword) > 0 && !taskDto.ContainsKeyword(title) {
			variable.ZapLog.Sugar().Infof("忽略下载 %v", title)
			// obj.Open = false
			continue
		}

		lanSlice := []string{}
		if strings.Contains(item[3], "双语") || strings.Contains(item[3], "简体") {
			childItems := a4kLanRegexp.FindAllStringSubmatch(item[3], -1)
			for _, childItem := range childItems {
				lanSlice = append(lanSlice, childItem[1])
			}
		} else {
			continue
		}

		title = helper.ReplaceTitle(title)
		newDto := &dto.TaskDto{SearchKeyword: taskDto.SearchKeyword, Name: title, WorkType: variable.FecthInfo, Refers: []string{taskDto.DownloadUrl}, DownloadUrl: item[7], Cookies: cookies, Lan: strings.Join(lanSlice, "/"), SubtitlesType: "", Wg: taskDto.Wg, StoreFunc: taskDto.StoreFunc, EsIndex: taskDto.EsIndex, Crawler: taskDto.Crawler}

		if len(strings.Trim(newDto.DownloadUrl, " ")) == 0 {
			continue
		}

		newDto.DownloadUrl = helper.UrlJoin(newDto.DownloadUrl, "https://www.a4k.net")

		newDto.InfoUrl = newDto.DownloadUrl

		//清洗数据1
		if isCreate, id := downloadRepository.TryCreate(taskDto); !isCreate {
			variable.ZapLog.Sugar().Infof("跳过已存在数据：%v", taskDto.Name)
			continue
		} else {
			taskDto.DownloadId = id
		}

		obj.insertQueue(newDto)
	}

}

func (obj *A4KCrawler) fetchInfo(taskDto *dto.TaskDto) {
	html, _, err := helper.LoadHtml(taskDto)
	if err != nil {
		return
	}

	//获取子页信息
	items := a4kFetchInfoRegexp.FindAllStringSubmatch(*html, -1)

	if len(items) != 1 {
		// rp2 := regexp.MustCompile(nameReg)
		// rp3 := regexp.MustCompile(downloadReg)
		// items2 := rp2.FindAllStringSubmatch(*html, -1)
		// items3 := rp3.FindAllStringSubmatch(*html, -1)

		fmt.Println("fetchInfo匹配失败:" + taskDto.DownloadUrl)
		return
	}

	url := helper.UrlJoin(helper.ToUtf8Str(items[0][2]), "https://www.a4k.net")

	taskDto.Refers = append(taskDto.Refers, taskDto.DownloadUrl)
	taskDto.DownloadUrl = url

	taskDto.WorkType = variable.Parse
	obj.insertQueue(taskDto)

}

func (obj *A4KCrawler) parse(taskDto *dto.TaskDto) {
	// downloadRepository := repository.DownloadsFactory()
	// //清洗数据1
	// if downloadRepository.Exists(taskDto) {
	// 	variable.ZapLog.Sugar().Infof("跳过已存在数据：%v", taskDto.Name)
	// 	return
	// }

	//清洗数据2
	newDto, err := helper.Download(taskDto)
	if err != nil {
		if err.Error() == "被拦截" {
			obj.Open = false
		}
		return
	}

	newDto.Wg.Add(1)

	//清洗数据3
	go helper.ParseFile(newDto)
}
