package crawler

import (
	"flag"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"kard/src/global/helper"
	"kard/src/global/variable"
	"kard/src/model/dto"
)

type ZimuCrawler struct {
	// StoreFunc func(dtoSlice []*dto.SubtitlesIndexDto)
	// Wg *sync.WaitGroup
	// helper.Parser
	// helper.Downloader
	Open bool
}

var (
	pageVisited sync.Map
	visited     sync.Map

	pageNum         = `<a class="num" href="([^"]+)">.+?</a>`
	fetchPageRegexp = regexp.MustCompile(pageNum)

	titleReg          = `<td class="w75pc">\s*<a href="(/sub(s)?/\d+.html)" target="_blank">(.+)</a>\s*</td>`
	lanReg            = `\n<td class="nobr center">([简繁英日体双语/]*)</td>`
	downloadButtonReg = `\n<td class="nobr center"><a href="(/sub(s)?/\d+.html)" target="_blank"><span class="label label-danger">字幕下载</span></a></td>`
	subtitleReg       = `\n<td class="nobr center">([ASTR/其他]*)</td>`
	fetchListRegexp   = regexp.MustCompile(titleReg + lanReg + downloadButtonReg + subtitleReg)

	nameReg         = `<div class="md_tt prel">(\n| )*<h1 title=[^>]+>(.+)</h1>(.|\n)+`
	downloadReg     = `<a class="btn btn-info btn-sm" href="([^"]+)"(.|\n)+下载字幕</a>`
	fetchInfoRegexp = regexp.MustCompile(nameReg + downloadReg)

	dx1DownloadReg       = `<a rel="nofollow" href="(.+dx1)"(.|\n)+电信高速下载（一）</a>`
	fetchSelectDx1Regexp = regexp.MustCompile(dx1DownloadReg)

	jsPageDownloadReg    = `location.href="([^"]+)";`
	jsPageDownloadRegexp = regexp.MustCompile(jsPageDownloadReg)
)

func (obj *ZimuCrawler) Work(store func(dtoSlice []*dto.SubtitlesIndexDto)) {
	obj.search(store)
}

func (obj *ZimuCrawler) search(store func(dtoSlice []*dto.SubtitlesIndexDto)) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("main recover error:%s \n", err)
		}
	}()

	q := flag.String("q", "", "useage to search")
	flag.Parse()

	fmt.Printf("q=%s\n", *q)
	qStr := *q

	var reqUrl string
	if len(qStr) > 0 {
		v := url.Values{}
		v.Add("q", qStr)
		v.Add("m", "yes")
		// v.Add("f", "_all")
		v.Add("s", "relevance")
		reqUrl = "https://www.zimutiantang.com/search/?" + v.Encode()
	} else {
		reqUrl = "https://www.zimutiantang.com"
	}

	// workerQueue := make(chan *dto.UrlDto, 1)

	taskDto := &dto.TaskDto{SearchKeyword: qStr, WorkType: variable.FecthPage, DownloadUrl: reqUrl, Wg: &sync.WaitGroup{}, StoreFunc: store}

	// for {
	// 	select {
	// 	case taskDto := <-workerQueue:
	// 		go func(dto *dto.UrlDto, queue chan *dto.UrlDto) {
	// 			switch dto.WorkType {
	// 			case variable.FecthPage:
	// 				fetchPage(dto, queue)
	// 			case variable.FecthList:
	// 				fetchList(dto, queue)
	// 			case variable.FecthInfo:
	// 				fetchInfo(dto, queue)
	// 			case variable.ParseFile:
	// 				parseFile(dto, queue)
	// 			}
	// 		}(taskDto, workerQueue)
	// 	default:
	// 		fmt.Printf("\n等待任务")
	// 		time.Sleep(1 * time.Second)
	// 	}
	// }

	// f := func(wg *sync.WaitGroup, taskDto *dto.UrlDto, queue chan *dto.UrlDto) {
	// 	defer func(wd *sync.WaitGroup) {
	// 		wd.Done()
	// 	}(wg)

	// 	// fmt.Printf("\n执行任务：%s", taskDto.WorkType)
	// 	wg.Add(1)

	// 	time.Sleep(1 * time.Second)

	// 	switch taskDto.WorkType {
	// 	case variable.FecthPage:
	// 		fetchPage(taskDto)
	// 	case variable.FecthList:
	// 		fetchList(taskDto)
	// 	case variable.FecthInfo:
	// 		fetchInfo(taskDto)
	// 	case variable.ParseFile:
	// 		parseFile(taskDto)
	// 	}
	// }

	// wg := &sync.WaitGroup{}
	// wg.Add(1)
	// go func(wg *sync.WaitGroup, queue chan *dto.UrlDto) {
	// 	defer func(w *sync.WaitGroup) {
	// 		w.Done()
	// 	}(wg)

	// 	for taskDto := range queue {
	// 		go f(wg, taskDto, queue)
	// 	}
	// }(wg, workerQueue)

	// wg.Wait()

	obj.insertQueue(taskDto)
	taskDto.Wg.Wait()
}

func (obj *ZimuCrawler) insertQueue(newDto *dto.TaskDto) {

	// fmt.Printf("v2=%p", &obj)
	if !obj.Open {
		return
	}

	switch newDto.WorkType {
	case variable.FecthPage:
		obj.fetchPage(newDto)
	case variable.FecthList:
		obj.fetchList(newDto)
	case variable.FecthInfo:
		obj.fetchInfo(newDto)
	case variable.Store:
		obj.store(newDto)
	}
}

func (obj *ZimuCrawler) fetchPage(taskDto *dto.TaskDto) {

	if _, ok := pageVisited.Load(taskDto.DownloadUrl); ok {
		return
	} else {
		pageVisited.Store(taskDto.DownloadUrl, &struct{}{})
	}

	html, cookies, err := helper.LoadHtml(taskDto)
	if err != nil {
		return
	}

	intoPage := []string{"", taskDto.DownloadUrl}
	items := [][]string{intoPage}
	pageItems := fetchPageRegexp.FindAllStringSubmatch(*html, -1)
	if pageItems != nil {
		items = append(items, pageItems...)
	}
	lastIndex := len(items) - 1
	for index, item := range items {
		url := item[1]
		if len(strings.Trim(url, " ")) == 0 {
			continue
		}
		if !strings.HasPrefix(url, "http:") && !strings.HasPrefix(url, "https:") {
			url = helper.UrlJoin(url, "https://www.zimutiantang.com")
		}

		if lastIndex == index {
			taskDto.Cookies = cookies
			taskDto.DownloadUrl = url
			obj.insertQueue(taskDto)
		}

		if _, ok := visited.Load(url); !ok {
			newDto := &dto.TaskDto{SearchKeyword: taskDto.SearchKeyword, WorkType: variable.FecthList, DownloadUrl: url, Cookies: cookies, Wg: taskDto.Wg, StoreFunc: taskDto.StoreFunc}
			visited.Store(newDto.DownloadUrl, &struct{}{})
			obj.insertQueue(newDto)
		}

	}

}

func (obj *ZimuCrawler) fetchList(taskDto *dto.TaskDto) {
	// defer func(d *dto.UrlDto) {
	// 	d.Wg.Done()
	// }(taskDto)

	html, cookies, err := helper.LoadHtml(taskDto)
	if err != nil {
		return
	}

	//获取子页信息
	items := fetchListRegexp.FindAllStringSubmatch(*html, -1)

	for _, item := range items {
		title := strings.Replace(strings.Replace(item[3], "<em>", "", -1), "</em>", "", -1)
		if len(taskDto.SearchKeyword) > 0 && !strings.Contains(title, taskDto.SearchKeyword) {
			fmt.Printf("\n忽略下载 %v", title)
			// o := &obj
			fmt.Printf("v1=%p", &obj)
			obj.Open = false
			return
		}

		newDto := &dto.TaskDto{SearchKeyword: taskDto.SearchKeyword, WorkType: variable.FecthInfo, Refers: []string{taskDto.DownloadUrl}, DownloadUrl: item[5], Cookies: cookies, Lan: item[4], Subtitles: item[7], Wg: taskDto.Wg, StoreFunc: taskDto.StoreFunc}

		if len(strings.Trim(newDto.DownloadUrl, " ")) == 0 {
			continue
		}
		if !strings.HasPrefix(newDto.DownloadUrl, "http:") && !strings.HasPrefix(newDto.DownloadUrl, "https:") {
			newDto.DownloadUrl = helper.UrlJoin(newDto.DownloadUrl, "https://www.zimutiantang.com")
		}

		if _, ok := visited.Load(newDto.DownloadUrl); !ok {
			visited.Store(newDto.DownloadUrl, &struct{}{})
			obj.insertQueue(newDto)
		}

	}

}

func (obj *ZimuCrawler) fetchInfo(taskDto *dto.TaskDto) {
	html, _, err := helper.LoadHtml(taskDto)
	if err != nil {
		return
	}

	//获取子页信息
	items := fetchInfoRegexp.FindAllStringSubmatch(*html, -1)

	if len(items) != 1 {
		rp2 := regexp.MustCompile(nameReg)
		rp3 := regexp.MustCompile(downloadReg)
		items2 := rp2.FindAllStringSubmatch(*html, -1)
		items3 := rp3.FindAllStringSubmatch(*html, -1)

		fmt.Println("匹配失败:" + taskDto.DownloadUrl + items2[0][2] + items3[0][1])
		return
	}

	name := items[0][2]
	url := strings.Trim(strings.Trim(items[0][4], "\n"), " ")
	if !strings.HasPrefix(url, "http:") && !strings.HasPrefix(url, "https:") {
		url = helper.UrlJoin(url, "https://www.zimutiantang.com")
	}

	fileName, err := helper.GetDownloadFileName(url, nil)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	taskDto.Refers = append(taskDto.Refers, taskDto.DownloadUrl)
	taskDto.DownloadUrl = url
	taskDto.Name = name

	if len(fileName) > 0 {
		taskDto.WorkType = variable.Store
		obj.insertQueue(taskDto)

	} else {
		obj.fetchSelectDx1(taskDto)
	}

}

func (obj *ZimuCrawler) fetchSelectDx1(taskDto *dto.TaskDto) {
	html, _, err := helper.LoadHtml(taskDto)
	if err != nil {
		return
	}

	items := fetchSelectDx1Regexp.FindAllStringSubmatch(*html, -1)

	if items != nil {
		if len(items) != 1 {
			fmt.Println("匹配失败:" + taskDto.DownloadUrl)
			return
		}

		downloadUrl := items[0][1]
		if !strings.HasPrefix(downloadUrl, "http:") && !strings.HasPrefix(downloadUrl, "https:") {
			downloadUrl = helper.UrlJoin(downloadUrl, "http://zimuku.org")
		}

		taskDto.Refers = append(taskDto.Refers, taskDto.DownloadUrl)
		taskDto.DownloadUrl = downloadUrl
		taskDto.WorkType = variable.Store
		obj.insertQueue(taskDto)
	} else {

		items = jsPageDownloadRegexp.FindAllStringSubmatch(*html, -1)
		if len(items) == 1 {
			url := items[0][1]

			taskDto.Refers = append(taskDto.Refers, taskDto.DownloadUrl)
			taskDto.DownloadUrl = url
			obj.fetchSelectDx1(taskDto)
		} else if find := strings.Contains(*html, "该字幕不可下载"); !find {
			fmt.Println("匹配失败:" + taskDto.DownloadUrl)
		}
	}

}

func (obj *ZimuCrawler) store(taskDto *dto.TaskDto) {

	newDto, err := helper.Download(taskDto)
	if err != nil {
		return
	}

	newDto.Wg.Add(1)
	go helper.ParseFile(newDto)
}
