package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"kard/src/dto"
	"net/http"
	"strconv"
	"time"
)

var (
	// client = &http.Client{
	// 	Transport: &http.Transport{
	// 		Proxy: http.ProxyFromEnvironment,
	// 		Dial: (&net.Dialer{
	// 			Timeout:   6 * time.Second,
	// 			Deadline:  time.Now().Add(3 * time.Second),
	// 			KeepAlive: 30 * time.Second,
	// 		}).Dial,
	// 		MaxIdleConns:          100,              //client对与所有host最大空闲连接数总和
	// 		IdleConnTimeout:       90 * time.Second, //空闲连接在连接池中的超时时间
	// 		TLSHandshakeTimeout:   10 * time.Second, //TLS安全连接握手超时时间
	// 		ExpectContinueTimeout: 1 * time.Second,  //发送完请求到接收到响应头的超时时间
	// 	},
	// 	Timeout: 6 * time.Second}
	client = &http.Client{
		// Transport: &http.Transport{

		// 	Dial: (&net.Dialer{
		// 		Timeout:   6 * time.Second,
		// 		Deadline:  time.Now().Add(3 * time.Second),
		// 		KeepAlive: 30 * time.Second,
		// 	}).Dial,
		// },
		Timeout: 6 * time.Second}
)

func loadHtml(urlDto *dto.UrlDto) (*string, []*http.Cookie, error) {
	req, err := getRequest(urlDto)
	if err != nil {
		fmt.Println("create request error", err)
		return nil, nil, err
	}

	var res *http.Response
	res, err = getResponse(req)
	if err != nil {
		fmt.Println("http get error", err)
		return nil, nil, err
	}
	defer res.Body.Close()

	var html *string
	html, err = getHtml(res)
	if err != nil {
		fmt.Println("read html error", err)
		return nil, nil, err
	}

	return html, res.Cookies(), nil
}

func getRequest(urlDto *dto.UrlDto) (*http.Request, error) {

	url := urlDto.DownloadUrl
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 自定义Header
	if len(urlDto.Refers) > 0 {
		req.Header.Set("Referer", urlDto.Refers[len(urlDto.Refers)-1])
	}
	if len(urlDto.Cookies) > 0 {
		for _, cookie := range urlDto.Cookies {
			req.AddCookie(cookie)
		}
	}
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36")

	req.Header.Set("accept-language", "zh-CN,zh;q=0.9")

	return req, nil
}

func getResponse(req *http.Request) (*http.Response, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("http request error:" + err.Error())
	}
	if resp == nil || resp.Body == nil {
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
		return nil, err
	}

	html := string(body)
	return &html, nil
}
