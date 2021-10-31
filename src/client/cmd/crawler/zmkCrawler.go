package crawler

import (
	"flag"
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

type ZmkCrawler struct {
	// StoreFunc func(dtoSlice []*dto.SubtitlesIndexDto)
	// Wg *sync.WaitGroup
	// helper.Parser
	// helper.Downloader
	Open bool
}

var (
	// pageVisited sync.Map
	// visited     sync.Map

	// zmkPageNum         = `<a class="num" href="([^"]+)">(\d+)?</a>`
	zmkPageNum         = `<a class="end" href="(.+)=(\d+)">\d+</a>`
	zmkFetchPageRegexp = regexp.MustCompile(zmkPageNum)

	zmkDownloadButtonReg = `<a href="(/detail/\d+\.html)" target="_blank" `
	zmkTitleReg          = `title="(.+)">.+</a>`
	zmkSubtitleReg       = `(\s|\n)*<span class="label label-info">([ASTRUPIDX\+/]*)</span>`
	zmkLanImgReg         = `(\s|\S)+?(<img .+ alt="[^"]+?"[^>]+?>)+`

	zmkFetchListRegexp = regexp.MustCompile(zmkDownloadButtonReg + zmkTitleReg + zmkSubtitleReg + zmkLanImgReg)

	zmkLanReg    = `<img .+? alt="([^"]+?)"[^>]+?>`
	zmkLanRegexp = regexp.MustCompile(zmkLanReg)

	zmkDownloadReg     = `<a id="\w+" href="([^"]+)" target="_blank" rel="nofollow">(\s|\S)*?下载字幕(\s|\S)*?</a>`
	zmkFetchInfoRegexp = regexp.MustCompile(zmkDownloadReg)

	zmkDx1DownloadReg       = `<a rel="nofollow" href="(.+dx1)"(.|\n)+电信高速下载（一）</a>`
	zmkFetchSelectDx1Regexp = regexp.MustCompile(zmkDx1DownloadReg)

	zmkJsPageDownloadReg    = `location.href="([^"]+)";`
	zmkJsPageDownloadRegexp = regexp.MustCompile(zmkJsPageDownloadReg)
)

func (obj *ZmkCrawler) Work(store func(taskDto *dto.TaskDto)) {
	defer func() {
		if err := recover(); err != nil {
			helper.PrintError("Work", err.(error).Error(), true)
		}
	}()

	obj.search(store)
}

func (obj *ZmkCrawler) search(store func(taskDto *dto.TaskDto)) {

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
		if values, err := url.ParseQuery(strings.Split(seedUrlStr, "?")[1]); err != nil {
			pageNum = 1
		} else if p, err2 := strconv.Atoi(values.Get("p")); err2 == nil {
			pageNum = p
		}
	} else if len(qStr) > 0 {
		v := url.Values{}
		v.Add("q", qStr)
		reqUrl = "https://zimuku.org/search?" + v.Encode()
	} else {
		reqUrl = "https://zimuku.org/"
	}

	taskDto := &dto.TaskDto{SearchKeyword: qStr, WorkType: variable.FecthPage, PageNum: pageNum, DownloadUrl: reqUrl, Wg: &sync.WaitGroup{}, StoreFunc: store, EsIndex: variable.IndexName}

	obj.insertQueue(taskDto)
	taskDto.Wg.Wait()
}

func (obj *ZmkCrawler) insertQueue(newDto *dto.TaskDto) {
	if !obj.Open {
		return
	}

	switch newDto.WorkType {
	case variable.FecthPage:
		helper.Sleep(newDto.WorkType, "s", 1, 10)
		obj.fetchPage(newDto)
	case variable.FecthList:
		helper.Sleep(newDto.WorkType, "s", 1, 10)
		obj.fetchList(newDto)
	case variable.FecthInfo:
		helper.WorkClock()
		helper.Sleep(newDto.WorkType, "m", 30, 50)
		obj.fetchInfo(newDto)
	case variable.Parse:
		helper.Sleep(newDto.WorkType, "s", 1, 5)
		obj.parse(newDto)
	}
}

func (obj *ZmkCrawler) fetchPage(taskDto *dto.TaskDto) {

	html, cookies, err := helper.LoadHtml(taskDto)
	if err != nil {
		return
	}

	pageItems := zmkFetchPageRegexp.FindAllStringSubmatch(*html, -1)
	pathUrl := ""
	endPageNum := 0
	pageNum := taskDto.PageNum
	if len(pageItems) <= 0 {
		endPageNum = taskDto.PageNum
	} else {
		pathUrl = pageItems[0][1]

		if !strings.HasPrefix(pathUrl, "http:") && !strings.HasPrefix(pathUrl, "https:") {
			pathUrl = helper.UrlJoin(pathUrl, "https://zimuku.org")
		}

		endPageNum, _ = strconv.Atoi(pageItems[0][2])
	}

	for pageNum <= endPageNum {
		fmt.Printf("\n处理第%v页 共%v页", pageNum, endPageNum)
		url := pathUrl + "=" + strconv.Itoa(pageNum)
		newDto := &dto.TaskDto{SearchKeyword: taskDto.SearchKeyword, WorkType: variable.FecthList, DownloadUrl: url, Cookies: cookies, Wg: taskDto.Wg, StoreFunc: taskDto.StoreFunc, EsIndex: taskDto.EsIndex}
		obj.insertQueue(newDto)
		pageNum++
	}

}

func (obj *ZmkCrawler) fetchList(taskDto *dto.TaskDto) {

	html, cookies, err := helper.LoadHtml(taskDto)
	if err != nil {
		return
	}

	//获取子页信息
	items := zmkFetchListRegexp.FindAllStringSubmatch(*html, -1)

	for _, item := range items {
		title := item[2]

		if len(taskDto.SearchKeyword) > 0 && !taskDto.ContainsKeyword(title) {
			fmt.Printf("\n忽略下载 %v", title)
			// obj.Open = false
			continue
		}

		lanSlice := []string{}
		if strings.Contains(item[6], "双语") || strings.Contains(item[6], "简体") {
			childItems := zmkLanRegexp.FindAllStringSubmatch(item[6], -1)
			for _, childItem := range childItems {
				switch childItem[1] {
				case "简体中文字幕":
					lanSlice = append(lanSlice, "简体")
				// case "English字幕":
				// 	lanSlice = append(lanSlice, "英文")
				case "双语字幕":
					lanSlice = append(lanSlice, "双语")
				}
			}
		} else {
			continue
		}

		newDto := &dto.TaskDto{SearchKeyword: taskDto.SearchKeyword, Name: title, WorkType: variable.FecthInfo, Refers: []string{taskDto.DownloadUrl}, DownloadUrl: item[1], Cookies: cookies, Lan: strings.Join(lanSlice, "/"), SubtitlesType: item[4], Wg: taskDto.Wg, StoreFunc: taskDto.StoreFunc, EsIndex: taskDto.EsIndex}

		if len(strings.Trim(newDto.DownloadUrl, " ")) == 0 {
			continue
		}

		newDto.DownloadUrl = helper.UrlJoin(newDto.DownloadUrl, "https://zimuku.org")

		obj.insertQueue(newDto)
	}

}

func (obj *ZmkCrawler) fetchInfo(taskDto *dto.TaskDto) {
	html, _, err := helper.LoadHtml(taskDto)
	if err != nil {
		return
	}

	//获取子页信息
	items := zmkFetchInfoRegexp.FindAllStringSubmatch(*html, -1)

	if len(items) != 1 {
		// rp2 := regexp.MustCompile(nameReg)
		// rp3 := regexp.MustCompile(downloadReg)
		// items2 := rp2.FindAllStringSubmatch(*html, -1)
		// items3 := rp3.FindAllStringSubmatch(*html, -1)

		fmt.Println("fetchInfo匹配失败:" + taskDto.DownloadUrl)
		return
	}

	url := helper.UrlJoin(items[0][1], "https://zimuku.org")

	taskDto.Refers = append(taskDto.Refers, taskDto.DownloadUrl)
	taskDto.DownloadUrl = url

	obj.fetchSelectDx1(taskDto)

}

func (obj *ZmkCrawler) fetchSelectDx1(taskDto *dto.TaskDto) {
	html, _, err := helper.LoadHtml(taskDto)
	if err != nil {
		return
	}

	items := zmkFetchSelectDx1Regexp.FindAllStringSubmatch(*html, -1)

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

		items = zmkJsPageDownloadRegexp.FindAllStringSubmatch(*html, -1)
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

func (obj *ZmkCrawler) parse(taskDto *dto.TaskDto) {
	downloadRepository := repository.DownloadsFactory()
	//清洗数据1
	if downloadRepository.Exists(taskDto) {
		fmt.Printf("\n跳过已存在数据：%v", taskDto.Name)
		return
	}

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
