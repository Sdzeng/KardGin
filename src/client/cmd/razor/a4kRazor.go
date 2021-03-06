package razor

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

type A4KRazor struct {
	BaseRazor
}

func NewA4KRazor(seedUrl string) *A4KRazor {
	return &A4KRazor{
		BaseRazor{
			Name:    "a4k",
			EsIndex: variable.IndexName,
			Enable:  true,
			SeedUrl: seedUrl,
		},
	}
}

var (
	a4kPageNum             = `<a class="item pager__item--last" href="(\?page=)(\d+)"[^>]+?>(\s|\S)+?</a>`
	a4kFetchPageRegexp     = regexp.MustCompile(a4kPageNum)
	a4kPathUrlNum          = `(\?page=)(\d+)`
	a4kPathUrlRegexp       = regexp.MustCompile(a4kPathUrlNum)
	a4kLastPageNum         = `<a class="item active" href="(\?page)=(\d+)"[^>]+?>(\s|\n)*<span(\s|\S)+?</span>(\s|\n)*(\d+)</a>`
	a4kLastFetchPageRegexp = regexp.MustCompile(a4kLastPageNum)
	// a4kTitleReg          = `<td class="w75pc">\s*<a href="(/sub(s)?/\d+.html)" target="_blank">(.+)</a>\s*</td>`
	a4kLanListReg        = `<div class="language">(\s|\n)*<span class="h4">(\s|\n)*((<i class="[^"]+?" data-content="[^"]+?"[^>]+?></i>\s*)+)`
	a4kDownloadButtonReg = `(\s|\S)+?<div class="content">(\s|\n)*<h3>(\s|\n)*<a href="([^"]+?)"[^>]+?>`
	a4kTitleReg          = `(.+)</a>(\s|\n)*</h3>`
	a4kFetchListRegexp   = regexp.MustCompile(a4kLanListReg + a4kDownloadButtonReg + a4kTitleReg)

	a4kLanReg    = `<i class="[^"]+?" data-content="([^"]+?)"[^>]+?></i>\s*`
	a4kLanRegexp = regexp.MustCompile(a4kLanReg)

	a4kDownloadReg     = `<div class="download">(\s|\S)+?<a class="ui green button" href="([^"]+?)"(\s|\S)+?下载字幕</a>`
	a4kFetchInfoRegexp = regexp.MustCompile(a4kDownloadReg)
)

func (obj *A4KRazor) Work(storeFunc func(taskDto *dto.TaskDto)) {
	defer func() {
		if err := recover(); err != nil {
			helper.PrintError("Work", err.(error).Error(), true)
		}
	}()

	obj.search(storeFunc)
}

func (obj *A4KRazor) search(storeFunc func(taskDto *dto.TaskDto)) {

	// seedUrl := flag.String("seed-url", "", "useage to search")
	// q := flag.String("q", "", "useage to search")
	// flag.Parse()

	// fmt.Printf("seedUrl=%s\n", *seedUrl)
	// fmt.Printf("q=%s\n", *q)
	// seedUrlStr := *seedUrl
	// qStr := *q

	// var reqUrl string
	// var pageNum int = 1
	// if len(seedUrlStr) > 0 {
	// 	reqUrl = seedUrlStr
	// 	if values, err := url.ParseQuery(strings.Split(seedUrlStr, "?")[1]); err != nil {
	// 		pageNum = 0
	// 	} else if p, err2 := strconv.Atoi(values.Get("page")); err2 == nil {
	// 		pageNum = p
	// 	}

	// } else if len(qStr) > 0 {
	// 	v := url.Values{}
	// 	v.Add("term", qStr)
	// 	reqUrl = "https://www.a4k.net/search?" + v.Encode()
	// } else {
	// 	reqUrl = "https://www.a4k.net"
	// }

	// workerQueue := make(chan *dto.UrlDto, 1)

	razorsRepository := repository.RazorsFactory()
	if len(obj.SeedUrl) > 0 {
		if u, err := url.Parse(obj.SeedUrl); err == nil {
			obj.Domain = u.Scheme + "://" + u.Host
			item := a4kPathUrlRegexp.FindStringSubmatch(u.Path + "?" + u.RawQuery)
			obj.PathUrl = item[1]
			obj.Page, _ = strconv.Atoi(item[2])
			razorsRepository.CreateOrUpdate(obj.EsIndex, obj.Name, obj.Domain, obj.PathUrl, obj.Page)
		}
	} else {
		raz := razorsRepository.KFirst(obj.Name, obj.EsIndex)
		obj.Domain = raz.Domain
		obj.PathUrl = raz.PathUrl
		obj.Page = raz.Page
	}

	wg := &sync.WaitGroup{}
	obj.fetchPage(wg, storeFunc)
	wg.Wait()
}

func (obj *A4KRazor) fetchPage(wg *sync.WaitGroup, storeFunc func(taskDto *dto.TaskDto)) {
	razorsRepository := repository.RazorsFactory()
	url := helper.UrlJoin(obj.PathUrl, obj.Domain) + strconv.Itoa(obj.Page)
	taskDto := &dto.TaskDto{Wg: wg, DownloadUrl: url, StoreFunc: storeFunc}

	html, cookies, err := helper.LoadHtml(taskDto)
	if err != nil {
		return
	}

	pageNum := obj.Page
	pageItems := a4kFetchPageRegexp.FindAllStringSubmatch(*html, -1)
	pathUrl := ""
	endPageNum := 0

	if len(pageItems) <= 0 {
		pageItems = a4kLastFetchPageRegexp.FindAllStringSubmatch(*html, -1)
		pathUrl = obj.PathUrl
		endPageNum, _ = strconv.Atoi(pageItems[0][2])
	} else {
		pathUrl = pageItems[0][1]
		endPageNum, _ = strconv.Atoi(pageItems[0][2])
	}

	if strings.HasPrefix(pathUrl, "http:") || strings.HasPrefix(pathUrl, "https:") {
		pathUrl = strings.ReplaceAll(pathUrl, obj.Domain, "")
	}
	// endPageNum += 1

	for pageNum <= endPageNum {
		variable.ZapLog.Sugar().Infof("处理第%v页 共%v页", pageNum, endPageNum)

		razorsRepository.Update(obj.Name, obj.EsIndex, obj.Domain, pathUrl, pageNum)

		url = helper.UrlJoin(pathUrl, obj.Domain) + strconv.Itoa(pageNum)

		newDto := &dto.TaskDto{WorkType: variable.FecthList, DownloadUrl: url, Cookies: cookies, Wg: taskDto.Wg, StoreFunc: taskDto.StoreFunc, PageNum: pageNum}
		obj.insertQueue(newDto)
		pageNum++
	}

}

func (obj *A4KRazor) insertQueue(newDto *dto.TaskDto) {
	if !obj.Enable {
		return
	}
	switch newDto.WorkType {
	// case variable.FecthPage:
	// 	helper.Sleep(obj.Name, newDto.WorkType, "s", 1, 8)
	// 	obj.fetchPage(newDto)
	case variable.FecthList:
		helper.Sleep(obj.Name, newDto.WorkType, "s", 1, 10)
		obj.fetchList(newDto)
	case variable.FecthInfo:
		helper.WorkClock(obj.Name)
		helper.Sleep(obj.Name, newDto.WorkType, "m", 1, 3)
		obj.fetchInfo(newDto)
	case variable.Parse:
		helper.Sleep(obj.Name, newDto.WorkType, "s", 1, 5)
		obj.parse(newDto)
	}
}

func (obj *A4KRazor) CompletionData(storeFunc func(taskDto *dto.TaskDto), downloadIds ...int32) {
	downloadRepository := repository.DownloadsFactory()
	downloadRefersRepository := repository.DownloadRefersFactory()
	downloadPathsRepository := repository.DownloadPathsFactory()

	wg := &sync.WaitGroup{}
	for _, downloadId := range downloadIds {
		download := downloadRepository.KFirst(downloadId)
		delDownloadPathIds := downloadPathsRepository.KFindIdByDownloadId(downloadId)
		refers := downloadRefersRepository.KFind(downloadId)
		referArr := []string{}
		for _, refer := range refers {
			if refer.Refer == download.InfoUrl {
				break
			}
			referArr = append(referArr, refer.Refer)
		}

		download.Name = helper.ReplaceTitle(download.Name)
		downloadRepository.Update(download.Id, download.Name)

		taskDto := &dto.TaskDto{WorkType: variable.FecthInfo, DownloadId: downloadId,
			InfoUrl:     download.InfoUrl,
			DownloadUrl: download.InfoUrl, Name: download.Name, Lan: download.Lan, Refers: referArr, Wg: wg, StoreFunc: storeFunc, PageNum: download.Page, DelDownloadPathIds: delDownloadPathIds}

		downloadPathsRepository.KDelete(downloadId)
		downloadRefersRepository.KDelete(downloadId)

		obj.insertQueue(taskDto)
	}
	wg.Wait()
}

func (obj *A4KRazor) fetchList(taskDto *dto.TaskDto) {

	html, cookies, err := helper.LoadHtml(taskDto)
	if err != nil {
		return
	}

	downloadRepository := repository.DownloadsFactory()

	//获取子页信息
	items := a4kFetchListRegexp.FindAllStringSubmatch(*html, -1)

	for _, item := range items {
		title := item[9]

		// if len(taskDto.SearchKeyword) > 0 && !taskDto.ContainsKeyword(title) {
		// 	variable.ZapLog.Sugar().Infof("忽略下载 %v", title)
		// 	// obj.Open = false
		// 	continue
		// }

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
		newDto := &dto.TaskDto{Name: title, WorkType: variable.FecthInfo, Refers: []string{taskDto.DownloadUrl}, DownloadUrl: item[8], Cookies: cookies, Lan: strings.Join(lanSlice, "/"), SubtitlesType: "", Wg: taskDto.Wg, StoreFunc: taskDto.StoreFunc, PageNum: taskDto.PageNum}

		if len(strings.Trim(newDto.DownloadUrl, " ")) == 0 {
			continue
		}

		newDto.DownloadUrl = helper.UrlJoin(newDto.DownloadUrl, obj.Domain)
		newDto.InfoUrl = newDto.DownloadUrl

		//清洗数据1
		isCreate, id, err := downloadRepository.TryCreate(obj.EsIndex, obj.Name, newDto)
		if err != nil {
			variable.ZapLog.Sugar().Errorf("跳过插入失败的数据：%v %v", newDto.Name, err)
			continue
		} else if !isCreate {
			variable.ZapLog.Sugar().Infof("跳过已存在数据：%v", newDto.Name)
			continue
		} else {
			newDto.DownloadId = id
		}

		obj.insertQueue(newDto)
	}

}

func (obj *A4KRazor) fetchInfo(taskDto *dto.TaskDto) {
	defer func() {
		if err := recover(); err != nil {
			helper.PrintError("fetchInfo", err.(error).Error(), true)
		}
	}()

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

	url := helper.UrlJoin(helper.ToUtf8Str(items[0][2]), obj.Domain)

	taskDto.Refers = append(taskDto.Refers, taskDto.DownloadUrl)
	taskDto.DownloadUrl = url

	taskDto.WorkType = variable.Parse
	obj.insertQueue(taskDto)

}

func (obj *A4KRazor) parse(taskDto *dto.TaskDto) {

	//清洗数据2
	newDto := helper.Download(taskDto)
	if newDto.Error != nil && newDto.Error.Error() == "被拦截" {
		obj.Enable = false
	}

	newDto.Wg.Add(1)

	//清洗数据3
	go func(nd *dto.TaskDto) {
		defer func(dto *dto.TaskDto) {
			dto.Wg.Done()

			if err := recover(); err != nil {
				helper.PrintError("ParseFile", err.(error).Error(), true)
			}
		}(nd)

		if nd.Error == nil {
			helper.ParseFile(nd)
		}

		taskDto.StoreFunc(nd)
	}(newDto)
}
