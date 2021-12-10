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

type ZmksRazor struct {
	BaseRazor
}

func NewZmksRazor(seedUrl string) *ZmksRazor {
	// if len(seedUrl) <= 0 {
	// 	seedUrl = "https://zimuku.org/search?q=&p=1"
	// }
	return &ZmksRazor{
		BaseRazor{
			Name:     "zmks",
			Domain:   "https://zimuku.org",
			InitPage: 1,
			EsIndex:  variable.IndexName,
			Enable:   true,
			SeedUrl:  seedUrl,
		},
	}
}

var (
	zmksPageNum             = `<a class="end" href="(.+)=(\d+)">\d+</a>`
	zmksFetchPageRegexp     = regexp.MustCompile(zmksPageNum)
	zmksPathUrlNum          = `(.+\?.+)=(\d+)`
	zmksPathUrlRegexp       = regexp.MustCompile(zmksPathUrlNum)
	zmksLastPageNum         = `</a><span class="current">(\d+)</span>(\s|\n)*<span class="rows">共\s*(\d+)\s*条记录`
	zmksLastFetchPageRegexp = regexp.MustCompile(zmksLastPageNum)

	// zmksDownloadButtonReg = `<a href="(/detail/\d+\.html)" target="_blank" `
	// zmksTitleReg          = `title="([^"]+?)">.+</a>`
	// zmksSubtitleReg       = `(\s|\n)*<span class="label label-info">([ASTRUPIDX\+/]*)</span>`
	// zmksLanImgReg         = `(\s|\S)+?((<img .+ alt="[^"]+?"[^>]+?>)+)`
	zmksSubtitleReg = `<tr class="(odd|even)">(\s|\S)+?<img .+ alt="([^"]+?)"[^>]+?>(\s|\S)+?<a href="(/detail/\d+\.html)" target="_blank" title="[^"]+?">.+</a>(\s|\n)*<span class="label label-info">([ASTRUPIDX\+/]*)</span>(\s|\S)+?</tr>(\s|\n)*`

	// zmksFetchListRegexp = regexp.MustCompile(zmksDownloadButtonReg + zmksTitleReg + zmksSubtitleReg + zmksLanImgReg)
	zmksFetchListRegexp = regexp.MustCompile(zmksSubtitleReg)

	// zmksLanReg    = `<img .+? alt="([^"]+?)"[^>]+?>`
	// zmksLanRegexp = regexp.MustCompile(zmksLanReg)

	zmksDownloadReg     = `<a id="\w+" href="([^"]+)" target="_blank" rel="nofollow">(\s|\S)*?下载字幕(\s|\S)*?</a>`
	zmksFetchInfoRegexp = regexp.MustCompile(zmksDownloadReg)

	zmksDx1DownloadReg       = `<a rel="nofollow" href="(.+dx1)"(.|\n)+电信高速下载（一）</a>`
	zmksFetchSelectDx1Regexp = regexp.MustCompile(zmksDx1DownloadReg)

	zmksJsPageDownloadReg    = `location.href="([^"]+)";`
	zmksJsPageDownloadRegexp = regexp.MustCompile(zmksJsPageDownloadReg)
)

func (obj *ZmksRazor) Work(storeFunc func(taskDto *dto.TaskDto)) {
	defer func() {
		if err := recover(); err != nil {
			helper.PrintError("Work", err.(error).Error(), true)
		}
	}()

	obj.search(storeFunc)
}

func (obj *ZmksRazor) search(storeFunc func(taskDto *dto.TaskDto)) {

	// var reqUrl string
	// var pageNum int = 1
	// if len(seedUrlStr) > 0 {
	// 	reqUrl = seedUrlStr
	// 	if values, err := url.ParseQuery(strings.Split(seedUrlStr, "?")[1]); err != nil {
	// 		pageNum = 1
	// 	} else if p, err2 := strconv.Atoi(values.Get("p")); err2 == nil {
	// 		pageNum = p
	// 	}
	// } else if len(qStr) > 0 {
	// 	v := url.Values{}
	// 	v.Add("q", qStr)
	// 	reqUrl = "https://zimuku.org/search?" + v.Encode()
	// } else {
	// 	reqUrl = "https://zimuku.org/"
	// }

	razorsRepository := repository.RazorsFactory()
	if len(obj.SeedUrl) > 0 {
		if values, err := url.ParseQuery(strings.Split(obj.SeedUrl, "?")[1]); err != nil {
			obj.Page = obj.InitPage
		} else {
			obj.Page, _ = strconv.Atoi(values.Get("p"))
		}
		razorsRepository.CreateOrUpdate(obj.Name, obj.SeedUrl, obj.EsIndex, obj.Page)
	} else {
		raz := razorsRepository.FirstOrCreate(obj.Name, obj.Domain, obj.EsIndex, obj.InitPage)
		obj.SeedUrl = raz.SeedUrl
		obj.Page = raz.Page
	}

	item := zmksPathUrlRegexp.FindStringSubmatch(strings.ReplaceAll(obj.SeedUrl, obj.Domain, ""))
	obj.SeedPathUrl = item[1]

	wg := &sync.WaitGroup{}
	obj.fetchPage(wg, storeFunc)
	wg.Wait()
}

func (obj *ZmksRazor) fetchPage(wg *sync.WaitGroup, storeFunc func(taskDto *dto.TaskDto)) {

	razorsRepository := repository.RazorsFactory()
	taskDto := &dto.TaskDto{Wg: wg, DownloadUrl: obj.SeedUrl, StoreFunc: storeFunc}
	html, cookies, err := helper.LoadHtml(taskDto)
	if err != nil {
		return
	}

	pageNum := obj.Page
	pageItems := zmksFetchPageRegexp.FindAllStringSubmatch(*html, -1)
	pathUrl := ""
	endPageNum := 0

	if len(pageItems) <= 0 {
		pageItems = zmksLastFetchPageRegexp.FindAllStringSubmatch(*html, -1)
		pathUrl = obj.SeedPathUrl
		endPageNum, _ = strconv.Atoi(pageItems[0][1])
	} else {
		pathUrl = pageItems[0][1]
		endPageNum, _ = strconv.Atoi(pageItems[0][2])
	}

	if !strings.HasPrefix(pathUrl, "http:") && !strings.HasPrefix(pathUrl, "https:") {
		pathUrl = helper.UrlJoin(pathUrl, obj.Domain)
	}

	for pageNum <= endPageNum {
		variable.ZapLog.Sugar().Infof("处理第%v页 共%v页", pageNum, endPageNum)
		url := pathUrl + "=" + strconv.Itoa(pageNum)

		razorsRepository.Update(obj.Name, url, obj.EsIndex, pageNum)

		newDto := &dto.TaskDto{WorkType: variable.FecthList, DownloadUrl: url, Cookies: cookies, Wg: taskDto.Wg, StoreFunc: taskDto.StoreFunc, PageNum: pageNum}
		obj.insertQueue(newDto)
		pageNum++
	}

}

func (obj *ZmksRazor) insertQueue(newDto *dto.TaskDto) {
	if !obj.Enable {
		return
	}

	switch newDto.WorkType {
	// case variable.FecthPage:
	// 	helper.Sleep(obj.Name, newDto.WorkType, "s", 1, 10)
	// 	obj.fetchPage(newDto)
	case variable.FecthList:
		helper.Sleep(obj.Name, newDto.WorkType, "s", 1, 10)
		obj.fetchList(newDto)
	case variable.FecthInfo:
		helper.WorkClock(obj.Name)
		helper.Sleep(obj.Name, newDto.WorkType, "s", 5, 10)
		obj.fetchInfo(newDto)
	case variable.Parse:
		helper.Sleep(obj.Name, newDto.WorkType, "s", 1, 5)
		obj.parse(newDto)
	}
}

func (obj *ZmksRazor) CompletionData(storeFunc func(taskDto *dto.TaskDto), downloadIds ...int32) {
	downloadRepository := repository.DownloadsFactory()
	downloadRefersRepository := repository.DownloadRefersFactory()
	downloadPathsRepository := repository.DownloadPathsFactory()

	wg := &sync.WaitGroup{}
	for _, downloadId := range downloadIds {
		download := downloadRepository.KFirst(downloadId)
		refers := downloadRefersRepository.KFind(downloadId)
		referArr := []string{}
		for _, refer := range refers {
			if refer.Refer == download.InfoUrl {
				break
			}
			referArr = append(referArr, refer.Refer)
		}
		delDownloadPathIds := downloadPathsRepository.KFindIdByDownloadId(downloadId)

		taskDto := &dto.TaskDto{WorkType: variable.FecthInfo, DownloadId: downloadId,
			InfoUrl:     download.InfoUrl,
			DownloadUrl: download.InfoUrl, Name: download.Name, Lan: download.Lan, Refers: referArr, Wg: wg, StoreFunc: storeFunc, PageNum: download.Page, DelDownloadPathIds: delDownloadPathIds}

		downloadPathsRepository.KDelete(downloadId)
		downloadRefersRepository.KDelete(downloadId)

		obj.insertQueue(taskDto)
	}
	wg.Wait()
}

func (obj *ZmksRazor) fetchList(taskDto *dto.TaskDto) {

	html, cookies, err := helper.LoadHtml(taskDto)
	if err != nil {
		return
	}
	downloadRepository := repository.DownloadsFactory()
	//获取子页信息
	items := zmksFetchListRegexp.FindAllStringSubmatch(*html, -1)

	for _, item := range items {
		title := item[2]

		// if len(taskDto.SearchKeyword) > 0 && !taskDto.ContainsKeyword(title) {
		// 	variable.ZapLog.Sugar().Infof("忽略下载 %v", title)
		// 	// obj.Open = false
		// 	continue
		// }

		lanSlice := []string{}
		if strings.Contains(item[6], "双语") || strings.Contains(item[6], "简体") {
			childItems := strings.Split(item[6], "&nbsp;")
			for _, childItem := range childItems {
				switch childItem {
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

		title = helper.ReplaceTitle(title)
		newDto := &dto.TaskDto{Name: title, WorkType: variable.FecthInfo, Refers: []string{taskDto.DownloadUrl}, DownloadUrl: item[1], Cookies: cookies, Lan: strings.Join(lanSlice, "/"), SubtitlesType: item[4], Wg: taskDto.Wg, StoreFunc: taskDto.StoreFunc, PageNum: taskDto.PageNum}

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

func (obj *ZmksRazor) fetchInfo(taskDto *dto.TaskDto) {
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
	items := zmksFetchInfoRegexp.FindAllStringSubmatch(*html, -1)

	if len(items) != 1 {
		// rp2 := regexp.MustCompile(nameReg)
		// rp3 := regexp.MustCompile(downloadReg)
		// items2 := rp2.FindAllStringSubmatch(*html, -1)
		// items3 := rp3.FindAllStringSubmatch(*html, -1)

		fmt.Println("fetchInfo匹配失败:" + taskDto.DownloadUrl)
		return
	}

	url := helper.UrlJoin(items[0][1], obj.Domain)

	taskDto.Refers = append(taskDto.Refers, taskDto.DownloadUrl)
	taskDto.DownloadUrl = url

	obj.fetchSelectDx1(taskDto)

}

func (obj *ZmksRazor) fetchSelectDx1(taskDto *dto.TaskDto) {
	html, _, err := helper.LoadHtml(taskDto)
	if err != nil {
		return
	}

	items := zmksFetchSelectDx1Regexp.FindAllStringSubmatch(*html, -1)

	if items != nil {
		if len(items) != 1 {
			fmt.Println("匹配失败:" + taskDto.DownloadUrl)
			return
		}

		downloadUrl := helper.ToUtf8Str(items[0][1])
		if !strings.HasPrefix(downloadUrl, "http:") && !strings.HasPrefix(downloadUrl, "https:") {
			downloadUrl = helper.UrlJoin(downloadUrl, obj.Domain)
		}

		taskDto.Refers = append(taskDto.Refers, taskDto.DownloadUrl)
		taskDto.DownloadUrl = downloadUrl
		taskDto.WorkType = variable.Parse
		obj.insertQueue(taskDto)
	} else {

		items = zmksJsPageDownloadRegexp.FindAllStringSubmatch(*html, -1)
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

func (obj *ZmksRazor) parse(taskDto *dto.TaskDto) {

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
