package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"

	_ "kard/src/client"
	"kard/src/dto"
	"kard/src/global/variable"
)

var pageVisited sync.Map
var visited sync.Map
var client = &http.Client{}

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("recover error:")
			fmt.Println(err)
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
			fmt.Println("work异常:")
			fmt.Println(err)
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

	workerQueue := make(chan *dto.UrlDto, 100)

	workerQueue <- &dto.UrlDto{WorkType: variable.FecthPage, DownloadUrl: reqUrl}
	for urlDto := range workerQueue {

		switch urlDto.WorkType {
		case variable.FecthPage:
			go fetchPage(urlDto, workerQueue)
		case variable.FecthList:
			go fetchList(urlDto, workerQueue)
		case variable.FecthInfo:
			go fetchInfo(urlDto, workerQueue)
		case variable.ParseFile:
			go parseFile(urlDto, workerQueue)
		}

	}
}

func fetchPage(urlDto *dto.UrlDto, workerQueue chan *dto.UrlDto) {
	if _, ok := pageVisited.Load(urlDto.DownloadUrl); ok {
		return
	} else {
		pageVisited.Store(urlDto.DownloadUrl, &struct{}{})
	}

	req, err := getRequest(urlDto.DownloadUrl, "")
	if err != nil {
		fmt.Println("create request error", err)
		return
	}
	var res *http.Response
	res, err = getResponse(req)
	if err != nil {
		fmt.Println("http get error", err)
		return
	}
	var html *string
	html, err = getHtml(res)
	if err != nil {
		fmt.Println("read html error", err)
		return
	}

	page := `<a class="num" href="([^"]+)">.+?</a>`
	rp := regexp.MustCompile(page)

	intoPage := []string{"", urlDto.DownloadUrl}
	items := [][]string{intoPage}
	pageItems := rp.FindAllStringSubmatch(*html, -1)
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
			urlDto.DownloadUrl = url
			workerQueue <- urlDto
		}

		if _, ok := visited.Load(url); !ok {
			newDto := &dto.UrlDto{WorkType: variable.FecthList, DownloadUrl: url}
			visited.Store(newDto.DownloadUrl, &struct{}{})
			workerQueue <- newDto
		}

	}

}

func fetchList(urlDto *dto.UrlDto, workerQueue chan *dto.UrlDto) {

	req, err := getRequest(urlDto.DownloadUrl, "")
	if err != nil {
		fmt.Println("create request error", err)
		return
	}
	var res *http.Response
	res, err = getResponse(req)
	if err != nil {
		fmt.Println("http get error", err)
		return
	}
	var html *string
	html, err = getHtml(res)
	if err != nil {
		fmt.Println("read html error", err)
		return
	}

	//获取子页信息
	lan := `<td class="nobr center">([简繁英日体双语/]*)</td>`
	downloadPage := `\n<td class="nobr center"><a href="(/sub(s)?/\d+.html)" target="_blank"><span class="label label-danger">字幕下载</span></a></td>`
	subtitles := `\n<td class="nobr center">([ASTR/其他]*)</td>`

	rp := regexp.MustCompile(lan + downloadPage + subtitles)

	items := rp.FindAllStringSubmatch(*html, -1)

	for _, item := range items {

		newDto := &dto.UrlDto{WorkType: variable.FecthInfo, Refers: []string{urlDto.DownloadUrl}, DownloadUrl: item[2], Lan: item[1], Subtitles: item[4]}
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
	req, err := getRequest(urlDto.DownloadUrl, urlDto.Refers[len(urlDto.Refers)-1])
	if err != nil {
		fmt.Println("create request error", err)
		return
	}
	var res *http.Response
	res, err = getResponse(req)
	if err != nil {
		fmt.Println("http get error", err)
		return
	}
	var html *string
	html, err = getHtml(res)
	if err != nil {
		fmt.Println("read html error", err)
		return
	}

	//获取子页信息
	nameReg := `<div class="md_tt prel">(\n| )*<h1 title=[^>]+>(.+)</h1>(.|\n)+`
	downloadReg := `<a class="btn btn-info btn-sm" href="([^"]+)"(.|\n)+下载字幕</a>`
	rp := regexp.MustCompile(nameReg + downloadReg)

	items := rp.FindAllStringSubmatch(*html, -1)

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

func fetchSelectDx1(dto *dto.UrlDto, workerQueue chan *dto.UrlDto) {
	req, err := getRequest(dto.DownloadUrl, dto.Refers[len(dto.Refers)-1])
	if err != nil {
		fmt.Println("create request error", err)
		return
	}
	var res *http.Response
	res, err = getResponse(req)
	if err != nil {
		fmt.Println("http get error", err)
		return
	}
	var html *string
	html, err = getHtml(res)
	if err != nil {
		fmt.Println("read html error", err)
		return
	}
	downloadReg := `<a rel="nofollow" href="(.+dx1)"(.|\n)+电信高速下载（一）</a>`
	rp := regexp.MustCompile(downloadReg)

	items := rp.FindAllStringSubmatch(*html, -1)

	if items != nil {
		if len(items) != 1 {
			fmt.Println("匹配失败:" + dto.DownloadUrl)
			return
		}

		downloadUrl := items[0][1]
		if !strings.HasPrefix(downloadUrl, "http:") && !strings.HasPrefix(downloadUrl, "https:") {
			downloadUrl = urlJoin(downloadUrl, "http://zimuku.org")
		}

		dto.Refers = append(dto.Refers, dto.DownloadUrl)
		dto.DownloadUrl = downloadUrl
		download(dto, workerQueue)
	} else {
		downloadReg = `location.href="([^"]+)";`
		rp = regexp.MustCompile(downloadReg)

		items = rp.FindAllStringSubmatch(*html, -1)
		if len(items) == 1 {
			url := items[0][1]

			dto.Refers = append(dto.Refers, dto.DownloadUrl)
			dto.DownloadUrl = url
			fetchSelectDx1(dto, workerQueue)
		} else if find := strings.Contains(*html, "该字幕不可下载"); !find {
			fmt.Println("匹配失败:" + dto.DownloadUrl)
		}
	}

}

// func fetchSource(dto *dto.UrlDto, workerQueue chan *dto.UrlDto) {

// 	req := getRequest(dto.Url, dto.Refers[len(dto.Refers)-1])
// 	res, err := getResponse(req)
// 	if err != nil {
// 		fmt.Println("http get error", err)
// 		return
// 	}
// 	defer res.Body.Close()
// 	location := res.Header.Get("Location")

// 	if len(location) == 0 {
// 		return
// 	}

// 	dto.WorkType = variable.Download
// 	dto.Refers = append(dto.Refers, dto.Url)
// 	dto.Url = location
// 	workerQueue <- dto

// }

func download(dto *dto.UrlDto, workerQueue chan *dto.UrlDto) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("\n下载异常：" + dto.DownloadUrl + ":")
			fmt.Println(err)
		}
	}()

	err := Download(dto, workerQueue)
	if err != nil {
		fmt.Printf("\n下载失败：" + dto.DownloadUrl + " " + err.Error())
	} else {

		fmt.Printf("\n下载成功：%s(%s)", dto.Name, dto.FileName)
	}
}

func getRequest(url, refer string) (*http.Request, error) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	// 自定义Header
	if len(refer) > 0 {
		req.Header.Set("Referer", refer)
	}
	req.Header.Set("cookie", `__cfduid=d363ed555c72ebcef60b1aaeaf2e30b361620608582; __51cke__=; __gads=ID=81fcef958ed24ce0-22cb596c12c800e6:T=1620608584:RT=1620608584:S=ALNI_MZlepN6HPtflySi9nlVoOx07OFGVg; Hm_lvt_22aa5d46d8019d57f41bac7a4d290998=1620608586,1620609452; Hm_lpvt_22aa5d46d8019d57f41bac7a4d290998=1620609490; __tins__19749253={"sid": 1620608585546, "vd": 11, "expires": 1620611296279}; __51laig__=11`)

	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36")

	req.Header.Set("accept-language", "zh-CN,zh;q=0.9")
	// req.Header.Set("accept-encoding", "gzip, deflate, br")

	return req, nil
}

func getResponse(req *http.Request) (*http.Response, error) {
	// defer func() {
	// 	if err := recover(); err != nil {
	// 		fmt.Println("recover error:")
	// 		fmt.Println(err)
	// 		fmt.Println(req)
	// 	}
	// }()
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("http request error", err)
		return nil, err
	}
	if resp == nil || resp.Body == nil {
		fmt.Println("response is nil")
		return nil, errors.New("response is nil")
	}
	if resp.StatusCode != http.StatusOK {

		return nil, errors.New("http status code is " + strconv.Itoa(resp.StatusCode))
	}

	return resp, nil
}

func getHtml(resp *http.Response) (*string, error) {

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("Read error", err)
		return nil, err
	}
	html := string(body)
	return &html, nil
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
