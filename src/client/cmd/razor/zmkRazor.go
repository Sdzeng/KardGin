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

type ZmkRazor struct {
	BaseRazor
}

func NewZmkRazor(seedUrl string) *ZmkRazor {
	return &ZmkRazor{
		BaseRazor{
			Name:        "zmk",
			BaseSeedUrl: "https://zimuku.org",
			BasePage:    1,
			EsIndex:     variable.IndexName,
			Enable:      true,
			SeedUrl:     seedUrl,
		},
	}
}

var (
	zmkPageNum             = `<a class="end" href="(.+)=(\d+)">\d+</a>`
	zmkFetchPageRegexp     = regexp.MustCompile(zmkPageNum)
	zmkLastPageNum         = `</a><span class="current">(\d+)</span>(\s|\n)*<span class="rows">共\s*(\d+)\s*条记录`
	zmkLastFetchPageRegexp = regexp.MustCompile(zmkLastPageNum)

	zmkDownloadButtonReg = `<a href="(/detail/\d+\.html)" target="_blank" `
	zmkTitleReg          = `title="([^"]+?)">.+</a>`
	zmkSubtitleReg       = `(\s|\n)*<span class="label label-info">([ASTRUPIDX\+/]*)</span>`
	zmkLanImgReg         = `(\s|\S)+?((<img .+ alt="[^"]+?"[^>]+?>)+)`

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

func (obj *ZmkRazor) Work(store func(taskDto *dto.TaskDto)) {
	defer func() {
		if err := recover(); err != nil {
			helper.PrintError("Work", err.(error).Error(), true)
		}
	}()

	obj.search(store)
}

func (obj *ZmkRazor) search(store func(taskDto *dto.TaskDto)) {

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
			obj.Page = obj.BasePage
		} else {
			obj.Page, _ = strconv.Atoi(values.Get("p"))
		}
		razorsRepository.CreateOrUpdate(obj.Name, obj.SeedUrl, obj.EsIndex, obj.Page)
	} else {
		raz := razorsRepository.FirstOrCreate(obj.Name, obj.BaseSeedUrl, obj.EsIndex, obj.BasePage)
		obj.SeedUrl = raz.SeedUrl
		obj.Page = raz.Page
	}

	wg := &sync.WaitGroup{}
	obj.fetchPage(wg, store)
	wg.Wait()
}

func (obj *ZmkRazor) fetchPage(wg *sync.WaitGroup, store func(taskDto *dto.TaskDto)) {

	razorsRepository := repository.RazorsFactory()
	taskDto := &dto.TaskDto{Wg: wg, DownloadUrl: obj.SeedUrl, StoreFunc: store}
	html, cookies, err := helper.LoadHtml(taskDto)
	if err != nil {
		return
	}

	pageNum := obj.Page
	pageItems := zmkFetchPageRegexp.FindAllStringSubmatch(*html, -1)
	pathUrl := ""
	endPageNum := 0

	if len(pageItems) <= 0 {
		pageItems = zmkLastFetchPageRegexp.FindAllStringSubmatch(*html, -1)
		pathUrl = "/newsubs?p"
		endPageNum, _ = strconv.Atoi(pageItems[0][1])
	} else {
		pathUrl = pageItems[0][1]

		if !strings.HasPrefix(pathUrl, "http:") && !strings.HasPrefix(pathUrl, "https:") {
			pathUrl = helper.UrlJoin(pathUrl, obj.BaseSeedUrl)
		}
		endPageNum, _ = strconv.Atoi(pageItems[0][2])
	}

	if !strings.HasPrefix(pathUrl, "http:") && !strings.HasPrefix(pathUrl, "https:") {
		pathUrl = helper.UrlJoin(pathUrl, obj.BaseSeedUrl)
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

func (obj *ZmkRazor) insertQueue(newDto *dto.TaskDto) {
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
		helper.Sleep(obj.Name, newDto.WorkType, "m", 18, 35)
		obj.fetchInfo(newDto)
	case variable.Parse:
		helper.Sleep(obj.Name, newDto.WorkType, "s", 1, 5)
		obj.parse(newDto)
	}
}

func (obj *ZmkRazor) fetchList(taskDto *dto.TaskDto) {

	html, cookies, err := helper.LoadHtml(taskDto)
	if err != nil {
		return
	}
	downloadRepository := repository.DownloadsFactory()
	//获取子页信息
	items := zmkFetchListRegexp.FindAllStringSubmatch(*html, -1)

	for _, item := range items {
		title := item[2]

		// if len(taskDto.SearchKeyword) > 0 && !taskDto.ContainsKeyword(title) {
		// 	variable.ZapLog.Sugar().Infof("忽略下载 %v", title)
		// 	// obj.Open = false
		// 	continue
		// }

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

		title = helper.ReplaceTitle(title)
		newDto := &dto.TaskDto{Name: title, WorkType: variable.FecthInfo, Refers: []string{taskDto.DownloadUrl}, DownloadUrl: item[1], Cookies: cookies, Lan: strings.Join(lanSlice, "/"), SubtitlesType: item[4], Wg: taskDto.Wg, StoreFunc: taskDto.StoreFunc, PageNum: taskDto.PageNum}

		if len(strings.Trim(newDto.DownloadUrl, " ")) == 0 {
			continue
		}

		newDto.DownloadUrl = helper.UrlJoin(newDto.DownloadUrl, obj.BaseSeedUrl)
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

func (obj *ZmkRazor) fetchInfo(taskDto *dto.TaskDto) {
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

	url := helper.UrlJoin(items[0][1], obj.BaseSeedUrl)

	taskDto.Refers = append(taskDto.Refers, taskDto.DownloadUrl)
	taskDto.DownloadUrl = url

	obj.fetchSelectDx1(taskDto)

}

func (obj *ZmkRazor) fetchSelectDx1(taskDto *dto.TaskDto) {
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

		downloadUrl := helper.ToUtf8Str(items[0][1])
		if !strings.HasPrefix(downloadUrl, "http:") && !strings.HasPrefix(downloadUrl, "https:") {
			downloadUrl = helper.UrlJoin(downloadUrl, obj.BaseSeedUrl)
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

func (obj *ZmkRazor) parse(taskDto *dto.TaskDto) {

	//清洗数据2
	newDto, err := helper.Download(taskDto)
	if err != nil {
		if err.Error() == "被拦截" {
			obj.Enable = false
		}
		return
	}

	newDto.Wg.Add(1)

	//清洗数据3
	go helper.ParseFile(newDto)
}
