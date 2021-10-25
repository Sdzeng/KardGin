package crawler

import (
	"flag"
	"fmt"
	"math/rand"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"kard/src/global/helper"
	"kard/src/global/variable"
	"kard/src/model/dto"
	"kard/src/repository"
)

type ZimuCrawler struct {
	// StoreFunc func(dtoSlice []*dto.SubtitlesIndexDto)
	// Wg *sync.WaitGroup
	// helper.Parser
	// helper.Downloader
	Open bool
}

var (
	// pageVisited sync.Map
	// visited     sync.Map

	zmSeedUrlReg    = `.+/index_(\d+).html`
	zmSeedUrlRegexp = regexp.MustCompile(zmSeedUrlReg)

	// zmPageNum         = `<a class="num" href="([^"]+)">(\d+)?</a>`
	zmPageNum         = `<(li|a) class="(num|prev|next|active)"( href="([^"]+)")?>(<span class="current">)?(.+?)(</span>)?</(li|a)>`
	zmFetchPageRegexp = regexp.MustCompile(zmPageNum)

	// zmTitleReg          = `<td class="w75pc">\s*<a href="(/sub(s)?/\d+.html)" target="_blank">(.+)</a>\s*</td>`
	zmTitleReg          = `<td class=.+>\s*<a .+ target="_blank">(.+)</a>\s*</td>`
	zmLanReg            = `(\s|\n)*<td class="nobr center">([简繁英日体双语/]*)</td>`
	zmDownloadButtonReg = `(\s|\n)*<td class="nobr center"><a href="(/sub(s)?/\d+.html)" target="_blank"><span class="label label-danger">字幕下载</span></a></td>`
	zmSubtitleReg       = `(\s|\n)*<td class="nobr center">([ASTR/其他]*)</td>`
	zmFetchListRegexp   = regexp.MustCompile(zmTitleReg + zmLanReg + zmDownloadButtonReg + zmSubtitleReg)

	zmNameReg         = `<div class="md_tt prel">(\n|\s)*<h1 title=[^>]+>(.+)</h1>(.|\n)+`
	zmDownloadReg     = `<a class="btn btn-info btn-sm" href="([^"]+)"(.|\n)+下载字幕</a>`
	zmFetchInfoRegexp = regexp.MustCompile(zmNameReg + zmDownloadReg)

	zmDx1DownloadReg       = `<a rel="nofollow" href="(.+dx1)"(.|\n)+电信高速下载（一）</a>`
	zmFetchSelectDx1Regexp = regexp.MustCompile(zmDx1DownloadReg)

	zmJsPageDownloadReg    = `location.href="([^"]+)";`
	zmJsPageDownloadRegexp = regexp.MustCompile(zmJsPageDownloadReg)
)

func (obj *ZimuCrawler) Work(store func(taskDto *dto.TaskDto)) {
	defer func() {
		if err := recover(); err != nil {
			helper.PrintError("Work", err.(error).Error(), true)
		}
	}()

	obj.search(store)
}

func (obj *ZimuCrawler) search(store func(taskDto *dto.TaskDto)) {

	seedUrl := flag.String("seed-url", "", "useage to search")
	q := flag.String("q", "", "useage to search")
	flag.Parse()

	fmt.Printf("seedUrl=%s\n", *seedUrl)
	fmt.Printf("q=%s\n", *q)
	seedUrlStr := *seedUrl
	qStr := *q

	var reqUrl string
	var pageNum int = 1
	if len(seedUrlStr) > 0 {
		reqUrl = seedUrlStr
		seedPageNumItems := zmSeedUrlRegexp.FindStringSubmatch(reqUrl)
		if len(seedPageNumItems) > 0 {
			var err error
			if pageNum, err = strconv.Atoi(seedPageNumItems[1]); err != nil {
				pageNum = 1
			} else {
				pageNum++
			}
		}

	} else if len(qStr) > 0 {
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

	taskDto := &dto.TaskDto{SearchKeyword: qStr, WorkType: variable.FecthPage, PageNum: pageNum, DownloadUrl: reqUrl, Wg: &sync.WaitGroup{}, StoreFunc: store}

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
	if !obj.Open {
		return
	}

	rand.Seed(time.Now().Unix())
	second := rand.Intn(10)
	fmt.Printf("\n休眠%v秒 后面执行%v", second, newDto.WorkType)
	time.Sleep(time.Duration(second) * time.Second)

	switch newDto.WorkType {
	case variable.FecthPage:
		obj.fetchPage(newDto)
	case variable.FecthList:
		obj.fetchList(newDto)
	case variable.FecthInfo:
		obj.fetchInfo(newDto)
	case variable.Parse:
		obj.parse(newDto)
	}
}

func (obj *ZimuCrawler) fetchPage(taskDto *dto.TaskDto) {

	// if _, ok := pageVisited.Load(taskDto.DownloadUrl); ok {
	// 	return
	// } else {
	// 	pageVisited.Store(taskDto.DownloadUrl, &struct{}{})
	// }

	html, cookies, err := helper.LoadHtml(taskDto)
	if err != nil {
		return
	}

	items := [][]string{}
	pageItems := zmFetchPageRegexp.FindAllStringSubmatch(*html, -1)
	if pageItems != nil {
		items = append(items, pageItems...)
	}

	for index, item := range items {
		aClass := item[2]
		url := item[4]
		pageNum, _ := strconv.Atoi(item[6])

		if aClass == "prev" {
			continue
		}

		if aClass == "active" {
			url = taskDto.DownloadUrl
		}

		if len(strings.Trim(url, " ")) == 0 {
			continue
		}

		if !strings.HasPrefix(url, "http:") && !strings.HasPrefix(url, "https:") {
			url = helper.UrlJoin(url, "https://www.zimutiantang.com")
		}

		if aClass == "next" {
			taskDto.Cookies = cookies
			taskDto.DownloadUrl = url
			taskDto.PageNum, err = strconv.Atoi(items[index-1][6])
			taskDto.PageNum++
			if err != nil {
				fmt.Printf("\n获取最后页码报错:%v", items[index-1][6])
			}
			obj.insertQueue(taskDto)
			break
		}

		if pageNum < taskDto.PageNum {
			continue
		}

		fmt.Printf("\n处理第%v页", pageNum)

		// if _, ok := visited.Load(url); !ok {
		// 	newDto := &dto.TaskDto{SearchKeyword: taskDto.SearchKeyword, WorkType: variable.FecthList, DownloadUrl: url, Cookies: cookies, Wg: taskDto.Wg, StoreFunc: taskDto.StoreFunc}
		// 	visited.Store(newDto.DownloadUrl, &struct{}{})
		// 	obj.insertQueue(newDto)
		// }

		newDto := &dto.TaskDto{SearchKeyword: taskDto.SearchKeyword, WorkType: variable.FecthList, DownloadUrl: url, Cookies: cookies, Wg: taskDto.Wg, StoreFunc: taskDto.StoreFunc}
		obj.insertQueue(newDto)

	}

}

func (obj *ZimuCrawler) fetchList(taskDto *dto.TaskDto) {

	html, cookies, err := helper.LoadHtml(taskDto)
	if err != nil {
		return
	}

	//获取子页信息
	items := zmFetchListRegexp.FindAllStringSubmatch(*html, -1)

	for _, item := range items {
		title := strings.Replace(strings.Replace(item[1], "<em>", "", -1), "</em>", "", -1)
		if len(taskDto.SearchKeyword) > 0 && !taskDto.ContainsKeyword(title) {
			fmt.Printf("\n忽略下载 %v", title)
			// obj.Open = false
			continue
		}

		newDto := &dto.TaskDto{SearchKeyword: taskDto.SearchKeyword, WorkType: variable.FecthInfo, Refers: []string{taskDto.DownloadUrl}, DownloadUrl: item[5], Cookies: cookies, Lan: item[3], SubtitlesType: item[8], Wg: taskDto.Wg, StoreFunc: taskDto.StoreFunc}

		if len(strings.Trim(newDto.DownloadUrl, " ")) == 0 {
			continue
		}
		if !strings.HasPrefix(newDto.DownloadUrl, "http:") && !strings.HasPrefix(newDto.DownloadUrl, "https:") {
			newDto.DownloadUrl = helper.UrlJoin(newDto.DownloadUrl, "https://www.zimutiantang.com")
		}

		// if _, ok := visited.Load(newDto.DownloadUrl); !ok {
		// 	visited.Store(newDto.DownloadUrl, &struct{}{})
		// 	obj.insertQueue(newDto)
		// }
		obj.insertQueue(newDto)
	}

}

func (obj *ZimuCrawler) fetchInfo(taskDto *dto.TaskDto) {
	html, _, err := helper.LoadHtml(taskDto)
	if err != nil {
		return
	}

	//获取子页信息
	items := zmFetchInfoRegexp.FindAllStringSubmatch(*html, -1)

	if len(items) != 1 {
		// rp2 := regexp.MustCompile(nameReg)
		// rp3 := regexp.MustCompile(downloadReg)
		// items2 := rp2.FindAllStringSubmatch(*html, -1)
		// items3 := rp3.FindAllStringSubmatch(*html, -1)

		fmt.Println("fetchInfo匹配失败:" + taskDto.DownloadUrl)
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
		taskDto.WorkType = variable.Parse
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

	items := zmFetchSelectDx1Regexp.FindAllStringSubmatch(*html, -1)

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
		taskDto.WorkType = variable.Parse
		obj.insertQueue(taskDto)
	} else {

		items = zmJsPageDownloadRegexp.FindAllStringSubmatch(*html, -1)
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

func (obj *ZimuCrawler) parse(taskDto *dto.TaskDto) {
	downloadRepository := repository.DownloadFactory()
	if downloadRepository.Exists(taskDto) {
		fmt.Printf("\n跳过已存在数据：%v", taskDto.Name)
		return
	}

	newDto, err := helper.Download(taskDto)
	if err != nil {
		return
	}

	newDto.Wg.Add(1)
	go helper.ParseFile(newDto)
}
