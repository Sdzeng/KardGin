package helper

import (
	"fmt"
	"io/ioutil"
	"kard/src/model/dto"
	"net/http"
	"net/url"
	"strings"
)

func LoadHtml(taskDto *dto.TaskDto) (*string, []*http.Cookie, error) {
	req, err := GetRequest(taskDto)
	if err != nil {
		fmt.Printf("\ncreate request error %v", err)
		return nil, nil, err
	}

	var res *http.Response
	res, err = GetResponse(req)
	if err != nil {
		fmt.Printf("\nhttp get error %v", err)
		return nil, nil, err
	}
	defer res.Body.Close()

	var html *string
	html, err = getHtml(res)
	if err != nil {
		fmt.Printf("\nread html error %v", err)
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

func ReplaceTitle(source string) string {
	source = strings.ReplaceAll(source, ".WEBDL.FIX字幕侠", "")
	source = strings.ReplaceAll(source, "双语", "")
	source = strings.ReplaceAll(source, "特效", "")
	source = strings.ReplaceAll(source, "蓝光", "")
	source = strings.ReplaceAll(source, "官方", "")
	source = strings.ReplaceAll(source, "译本", "")
	source = strings.ReplaceAll(source, "对照", "")
	source = strings.ReplaceAll(source, "简英", "")
	source = strings.ReplaceAll(source, "中英", "")
	source = strings.ReplaceAll(source, "中文", "")
	source = strings.ReplaceAll(source, "简繁", "")
	source = strings.ReplaceAll(source, "简繁英", "")
	source = strings.ReplaceAll(source, "机翻", "")
	source = strings.ReplaceAll(source, "字幕", "")
	source = strings.ReplaceAll(source, "下载", "")
	source = strings.ReplaceAll(source, " ]", "")
	source = strings.ReplaceAll(source, "【", " ")
	source = strings.ReplaceAll(source, "】", " ")
	source = strings.ReplaceAll(source, "/", " ")
	source = MergerOfSpace(source)
	return source
}

func MergerOfSpace(source string) string {
	if len(source) <= 0 {
		return source
	}

	arr := strings.Fields(source)
	newArr := []string{}

	for _, str := range arr {
		if len(str) > 0 {
			newArr = append(newArr, str)
		}
	}

	return strings.Join(newArr, " ")
}
