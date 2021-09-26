package main

import (
	"flag"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	_ "kard/src/client"
	"kard/src/global/variable"
	"kard/src/model/dto"
)

var (
	pageVisited sync.Map
	visited     sync.Map

	pageNum         = `<a class="num" href="([^"]+)">.+?</a>`
	fetchPageRegexp = regexp.MustCompile(pageNum)

	lanReg            = `<td class="nobr center">([简繁英日体双语/]*)</td>`
	downloadButtonReg = `\n<td class="nobr center"><a href="(/sub(s)?/\d+.html)" target="_blank"><span class="label label-danger">字幕下载</span></a></td>`
	subtitleReg       = `\n<td class="nobr center">([ASTR/其他]*)</td>`
	fetchListRegexp   = regexp.MustCompile(lanReg + downloadButtonReg + subtitleReg)

	nameReg         = `<div class="md_tt prel">(\n| )*<h1 title=[^>]+>(.+)</h1>(.|\n)+`
	downloadReg     = `<a class="btn btn-info btn-sm" href="([^"]+)"(.|\n)+下载字幕</a>`
	fetchInfoRegexp = regexp.MustCompile(nameReg + downloadReg)

	dx1DownloadReg       = `<a rel="nofollow" href="(.+dx1)"(.|\n)+电信高速下载（一）</a>`
	fetchSelectDx1Regexp = regexp.MustCompile(dx1DownloadReg)

	jsPageDownloadReg    = `location.href="([^"]+)";`
	jsPageDownloadRegexp = regexp.MustCompile(jsPageDownloadReg)
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("main recover error:%s \n", err)
		}
	}()

	q := flag.String("q", "", "useage to search")
	flag.Parse()

	fmt.Printf("q=%s\n", *q)
	work(*q)

	fmt.Println("完美结束")
}

func work(q string) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("work recover error:%s \n", err)
		}
	}()

	var reqUrl string
	if len(q) > 0 {
		v := url.Values{}
		v.Add("q", q)
		v.Add("m", "yes")
		v.Add("f", "_all")
		v.Add("s", "relevance")
		reqUrl = "https://www.zimutiantang.com/search/search.php?" + v.Encode()
	} else {
		reqUrl = "https://www.zimutiantang.com"
	}

	workerQueue := make(chan *dto.UrlDto, 1)

	workerQueue <- &dto.UrlDto{WorkType: variable.FecthPage, DownloadUrl: reqUrl}

	for {
		select {
		case urlDto := <-workerQueue:
			go func() {
				switch urlDto.WorkType {
				case variable.FecthPage:
					fetchPage(urlDto, workerQueue)
				case variable.FecthList:
					fetchList(urlDto, workerQueue)
				case variable.FecthInfo:
					fetchInfo(urlDto, workerQueue)
				case variable.ParseFile:
					parseFile(urlDto, workerQueue)
				}
			}()
		default:
			fmt.Printf("\n等待任务")
			time.Sleep(1 * time.Second)
		}
	}

}

func fetchPage(urlDto *dto.UrlDto, workerQueue chan *dto.UrlDto) {
	if _, ok := pageVisited.Load(urlDto.DownloadUrl); ok {
		return
	} else {
		pageVisited.Store(urlDto.DownloadUrl, &struct{}{})
	}

	html, cookies, err := loadHtml(urlDto)
	if err != nil {
		return
	}

	intoPage := []string{"", urlDto.DownloadUrl}
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
			url = urlJoin(url, "https://www.zimutiantang.com")
		}

		if lastIndex == index {
			urlDto.Cookies = cookies
			urlDto.DownloadUrl = url
			workerQueue <- urlDto
		}

		if _, ok := visited.Load(url); !ok {
			newDto := &dto.UrlDto{WorkType: variable.FecthList, DownloadUrl: url, Cookies: cookies}
			visited.Store(newDto.DownloadUrl, &struct{}{})
			workerQueue <- newDto
		}

	}

}

func fetchList(urlDto *dto.UrlDto, workerQueue chan *dto.UrlDto) {

	html, cookies, err := loadHtml(urlDto)
	if err != nil {
		return
	}

	//获取子页信息
	items := fetchListRegexp.FindAllStringSubmatch(*html, -1)

	for _, item := range items {

		newDto := &dto.UrlDto{WorkType: variable.FecthInfo, Refers: []string{urlDto.DownloadUrl}, DownloadUrl: item[2], Cookies: cookies, Lan: item[1], Subtitles: item[4]}
		if len(strings.Trim(newDto.DownloadUrl, " ")) == 0 {
			continue
		}
		if !strings.HasPrefix(newDto.DownloadUrl, "http:") && !strings.HasPrefix(newDto.DownloadUrl, "https:") {
			newDto.DownloadUrl = urlJoin(newDto.DownloadUrl, "https://www.zimutiantang.com")
		}

		if _, ok := visited.Load(newDto.DownloadUrl); !ok {
			visited.Store(newDto.DownloadUrl, &struct{}{})
			workerQueue <- newDto
		}

	}

}

func fetchInfo(urlDto *dto.UrlDto, workerQueue chan *dto.UrlDto) {
	html, _, err := loadHtml(urlDto)
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

		fmt.Println("匹配失败:" + urlDto.DownloadUrl + items2[0][2] + items3[0][1])
		return
	}

	name := items[0][2]
	url := strings.Trim(strings.Trim(items[0][4], "\n"), " ")
	if !strings.HasPrefix(url, "http:") && !strings.HasPrefix(url, "https:") {
		url = urlJoin(url, "https://www.zimutiantang.com")
	}

	fileName, err := getDownloadFileName(url, nil)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	urlDto.Refers = append(urlDto.Refers, urlDto.DownloadUrl)
	urlDto.DownloadUrl = url
	urlDto.Name = name

	if len(fileName) > 0 {
		download(urlDto, workerQueue)

	} else {
		fetchSelectDx1(urlDto, workerQueue)
	}

}

func fetchSelectDx1(urlDto *dto.UrlDto, workerQueue chan *dto.UrlDto) {
	html, _, err := loadHtml(urlDto)
	if err != nil {
		return
	}

	items := fetchSelectDx1Regexp.FindAllStringSubmatch(*html, -1)

	if items != nil {
		if len(items) != 1 {
			fmt.Println("匹配失败:" + urlDto.DownloadUrl)
			return
		}

		downloadUrl := items[0][1]
		if !strings.HasPrefix(downloadUrl, "http:") && !strings.HasPrefix(downloadUrl, "https:") {
			downloadUrl = urlJoin(downloadUrl, "http://zimuku.org")
		}

		urlDto.Refers = append(urlDto.Refers, urlDto.DownloadUrl)
		urlDto.DownloadUrl = downloadUrl
		download(urlDto, workerQueue)
	} else {

		items = jsPageDownloadRegexp.FindAllStringSubmatch(*html, -1)
		if len(items) == 1 {
			url := items[0][1]

			urlDto.Refers = append(urlDto.Refers, urlDto.DownloadUrl)
			urlDto.DownloadUrl = url
			fetchSelectDx1(urlDto, workerQueue)
		} else if find := strings.Contains(*html, "该字幕不可下载"); !find {
			fmt.Println("匹配失败:" + urlDto.DownloadUrl)
		}
	}

}

func download(dto *dto.UrlDto, workerQueue chan *dto.UrlDto) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("\n download error:%s %s", dto.DownloadUrl, err)
		}
	}()

	err := Download(dto, workerQueue)
	if err != nil {
		// fmt.Printf("\n下载失败：%s %s", dto.DownloadUrl, err.Error())
		fmt.Printf("\n下载失败：%s", err.Error())
	} else {
		//fmt.Printf("\n下载成功：%s", dto.DownloadUrl)

	}
}

func urlJoin(href, base string) string {
	uri, err := url.Parse(href)
	if err != nil {
		return " "
	}
	baseUrl, err := url.Parse(base)
	if err != nil {
		return " "
	}
	return baseUrl.ResolveReference(uri).String()
}
