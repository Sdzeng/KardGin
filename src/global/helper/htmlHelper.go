package helper

import (
	"fmt"
	"io/ioutil"
	"kard/src/model/dto"
	"net/http"
	"net/url"
)

func LoadHtml(taskDto *dto.TaskDto) (*string, []*http.Cookie, error) {
	req, err := GetRequest(taskDto)
	if err != nil {
		fmt.Println("create request error", err)
		return nil, nil, err
	}

	var res *http.Response
	res, err = GetResponse(req)
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

func getHtml(resp *http.Response) (*string, error) {

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	html := string(body)
	return &html, nil
}

func UrlJoin(href, base string) string {
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
