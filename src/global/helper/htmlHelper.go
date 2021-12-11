package helper

import (
	"fmt"
	"io/ioutil"
	"kard/src/model/dto"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var (
	titleReplacer *strings.Replacer
	seasonFmt     = `(s|S)([0-9]+)(e|E)([0-9]+)`
	seasonRegexp  = regexp.MustCompile(seasonFmt)
)

func init() {

	replaceKeywords := []string{
		".WEBDL", "",
		"FIX字幕侠", "",
		"擦枪字幕组", "",
		"擦枪组", "",
		"YYeTs组", "",
		"WEB调轴", "",
		"双语", "",
		"特效", "",
		"蓝光", "",
		"碟机", "",
		"官方", "",
		"译本", "",
		"外挂", "",
		"对照", "",
		"BD原盘", "",
		"-加长版", "",
		"加长版", "",
		"精译版", "",
		"中英文", "",
		"简繁英", "",
		"日版", "",
		"双字", "",
		"简体", "",
		"简中", "",
		"简英", "",
		"中英", "",
		"中文", "",
		"中字", "",
		"简繁", "",
		"英文", "",
		"机翻", "",
		"字幕", "",
		"下载", "",
		"H264-", "",
		"h264-", "",
		"X264", "",
		"x264", "",
		".zip", "",
		".rar", "",
		".7z", "",
		".srt", "",
		".ssa", "",
		".ass", "",
		".stl", "",
		".ts", "",
		".ttml", "",
		".vtt", "",
		"&amp", "",
		"amp", "",
		".1080p", "",
		"1080p", "",
		"1080P", "",
		".chs", "",
		"chs", "",
		".eng", "",
		"eng", "",
		"en", "",

		"[", " ",
		"]", " ",
		"【", " ",
		"】", " ",
		"(", " ",
		")", " ",
		"/", " ",
		";", " ",
		".", " ",
		"&", " ",
	}

	titleReplacer = strings.NewReplacer(replaceKeywords...)
}

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

// func getHtml(resp *http.Response) (*string, error) {
// 	r := bufio.NewReader(resp.Body)
// 	e := determineEncoding(r)
// 	newReader := transform.NewReader(resp.Body, e.NewDecoder())
// 	result, err := ioutil.ReadAll(newReader)
// 	if err != nil {

// 		return nil, err
// 	}
// 	html := string(result)
// 	return &html, nil
// }

// func determineEncoding(r *bufio.Reader) encoding.Encoding {
// 	bytes, err := r.Peek(1024)
// 	if err != nil {
// 		log.Printf("Fetch error: %v", err)
// 		return unicode.UTF8
// 	}
// 	encode, _, _ := charset.DetermineEncoding(bytes, "")
// 	return encode
// }

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

	// source = strings.ReplaceAll(source, ".WEBDL", "")
	// source = strings.ReplaceAll(source, "FIX字幕侠", "")
	// source = strings.ReplaceAll(source, "擦枪组", "")
	// source = strings.ReplaceAll(source, "YYeTs组", "")
	// source = strings.ReplaceAll(source, "WEB调轴", "")
	// source = strings.ReplaceAll(source, "双语", "")
	// source = strings.ReplaceAll(source, "特效", "")
	// source = strings.ReplaceAll(source, "蓝光", "")
	// source = strings.ReplaceAll(source, "官方", "")
	// source = strings.ReplaceAll(source, "译本", "")
	// source = strings.ReplaceAll(source, "对照", "")
	// source = strings.ReplaceAll(source, "BD原盘", "")
	// source = strings.ReplaceAll(source, "-加长版", "")
	// source = strings.ReplaceAll(source, "加长版", "")
	// source = strings.ReplaceAll(source, "精译版", "")
	// source = strings.ReplaceAll(source, "日版", "")
	// source = strings.ReplaceAll(source, "双字", "")
	// source = strings.ReplaceAll(source, "简体", "")
	// source = strings.ReplaceAll(source, "简中", "")
	// source = strings.ReplaceAll(source, "简英", "")
	// source = strings.ReplaceAll(source, "中英", "")
	// source = strings.ReplaceAll(source, "中文", "")
	// source = strings.ReplaceAll(source, "简繁", "")
	// source = strings.ReplaceAll(source, "简繁英", "")
	// source = strings.ReplaceAll(source, "英文", "")
	// source = strings.ReplaceAll(source, "机翻", "")
	// source = strings.ReplaceAll(source, "字幕", "")
	// source = strings.ReplaceAll(source, "下载", "")
	// source = strings.ReplaceAll(source, ".zip", "")
	// source = strings.ReplaceAll(source, ".rar", "")
	// source = strings.ReplaceAll(source, ".7z", "")
	// source = strings.ReplaceAll(source, ".srt", "")
	// source = strings.ReplaceAll(source, ".ssa", "")
	// source = strings.ReplaceAll(source, ".ass", "")
	// source = strings.ReplaceAll(source, ".stl", "")
	// source = strings.ReplaceAll(source, ".ts", "")
	// source = strings.ReplaceAll(source, ".ttml", "")
	// source = strings.ReplaceAll(source, ".vtt", "")
	// source = strings.ReplaceAll(source, "&amp", "")
	// source = strings.ReplaceAll(source, "amp", "")
	// source = strings.ReplaceAll(source, ".1080p", "")
	// source = strings.ReplaceAll(source, ".chs.eng", "")

	// source = strings.ReplaceAll(source, "[", " ")
	// source = strings.ReplaceAll(source, "]", " ")
	// source = strings.ReplaceAll(source, "【", " ")
	// source = strings.ReplaceAll(source, "】", " ")
	// source = strings.ReplaceAll(source, "(", " ")
	// source = strings.ReplaceAll(source, ")", " ")
	// source = strings.ReplaceAll(source, "/", " ")
	source = titleReplacer.Replace(source)
	source = ReplaceSeason(source)
	source = MergerOfSpace(source)
	return source
}

func ReplaceSeason(source string) string {
	if strings.Contains(source, "季") && strings.Contains(source, "集") {
		return source
	}

	items := seasonRegexp.FindStringSubmatch(source)
	if len(items) <= 0 {
		return source
	}

	return strings.Replace(source, items[0], "第"+items[2]+"季第"+items[4]+"集", -1)
}

func MergerOfSpace(source string) string {
	if len(source) <= 0 {
		return source
	}

	arr := strings.Fields(source)
	// newArr := []string{}

	// for _, str := range arr {
	// 	if len(str) > 0 {
	// 		newArr = append(newArr, str)
	// 	}
	// }

	return strings.Join(arr, ".")
}
