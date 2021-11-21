package helper

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"kard/src/global/variable"
	"kard/src/model/dto"
	"log"
	"math/rand"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/axgle/mahonia"
	"github.com/mholt/archiver"
	"github.com/nwaples/rardecode"
	"github.com/saintfish/chardet"
	"github.com/saracen/go7z"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	//"github.com/sony/sonyflake"
)

var (
	detector       *chardet.Detector
	mapKeyReplacer *strings.Replacer
	// nameReplacer   *strings.Replacer
)

func init() {
	detector = chardet.NewTextDetector()
	replaceKeywords := []string{
		"简体&英文", "",
		"繁体&英文", "",
		"简体", "",
		"英文", "",
		"繁体", "",
		".srt", "",
		".ssa", "",
		".ass", "",
		".stl", "",
		".ts", "",
		".ttml", "",
		".vtt", "",
		".zip", "",
		".rar", "",
		".7z", "",
		".", "",
	}

	mapKeyReplacer = strings.NewReplacer(replaceKeywords...)

	// nameReplaceKeywords := []string{
	// 	"简体&英文", "",
	// 	"繁体&英文", "",
	// 	"简体", "",
	// 	"英文", "",
	// 	"繁体", "",
	// 	".", " ",
	// 	".", " ",
	// }

	// nameReplacer = strings.NewReplacer(nameReplaceKeywords...)
}

type WriterCounter struct {
	//taskDto   *dto.UrlDto
	FileName string
	Total    uint64
}

//var flake *sonyflake.Sonyflake

func (wc *WriterCounter) PrintProgress() {
	//variable.ZapLog.Sugar().Infof("%s", strings.Repeat(" ", 50))
	//variable.ZapLog.Sugar().Infof("Downloading... %s complete %s(%s)", humanize.Bytes(wc.Total), wc.taskDto.name, wc.taskDto.fileName)

}

func (wc *WriterCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func Download(taskDto *dto.TaskDto) *dto.TaskDto {
	//请求资源
	req, err := GetRequest(taskDto)
	if err != nil {
		taskDto.Error = err
		return taskDto
	}
	var res *http.Response
	res, err = GetResponse(req)
	if err != nil {
		taskDto.Error = err
		return taskDto
	}
	defer res.Body.Close()

	//拷贝
	fileName, err := GetDownloadFileName(taskDto.DownloadUrl, res)
	if err != nil {
		taskDto.Error = err
		return taskDto
	}

	fileBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		taskDto.Error = err
		return taskDto
	}
	defer res.Body.Close()

	taskDto.SubtitlesFiles = downloadFiles(taskDto.DownloadUrl, fileName, fileBytes)
	if len(taskDto.SubtitlesFiles) <= 0 {
		taskDto.Error = errors.New("没有可下载的字幕文件")
	}

	return taskDto
}

func GetDownloadFileName(url string, resp *http.Response) (string, error) {
	fileName := getFileName(url)
	fileNameLower := strings.ToLower(fileName)
	// urlRp := regexp.MustCompile(`\d+.html`)
	result := ""
	if strings.HasSuffix(fileNameLower, ".rar") || strings.HasSuffix(fileNameLower, ".zip") || strings.HasSuffix(fileNameLower, ".7z") || strings.HasSuffix(fileNameLower, ".ass") || strings.HasSuffix(fileNameLower, ".ssa") || strings.HasSuffix(fileNameLower, ".srt") || strings.HasSuffix(fileNameLower, ".stl") || strings.HasSuffix(fileNameLower, ".ts") || strings.HasSuffix(fileNameLower, ".ttml") || strings.HasSuffix(fileNameLower, ".vtt") || strings.HasSuffix(fileNameLower, ".sup") {
		result = fileName
	} else if fileNameLower == "dx1" {
		//attachment; filename="[zmk.pw]海贼王15周年纪念特别篇幻之篇章「3D2Y跨越艾斯之死！与路飞伙伴的誓言[720p].Chs.rar""
		contentDisposition := resp.Header.Get("Content-Disposition")
		filenameReg := `attachment; filename="([^"]+)"`
		headerRp := regexp.MustCompile(filenameReg)

		items := headerRp.FindAllStringSubmatch(contentDisposition, -1)

		if len(items) == 1 {
			result = items[0][1]
		}
	} else {
		html, err := getHtml(resp)
		if err == nil && strings.Contains(*html, "已超出字幕下载个数限制，涉嫌恶意采集") {
			return "", errors.New("被拦截")
		}
	}

	if len(result) == 0 {
		result = fileName
	}
	// else if fileNameLower == "target.php" {
	// 	return "", nil
	// } else if items := urlRp.FindAllStringSubmatch(fileNameLower, -1); items != nil {
	// 	return "", nil
	// }

	return result, nil
}

// func checkErrorName(fileName string) {

// 	if strings.Contains(fileName, "fffd") || strings.Contains(fileName, "\\x") {

// 		_ = "出错"
// 	}
// }

func getFileName(url string) string {
	reqPathSlice := strings.Split(url, "/")
	fileUrl := reqPathSlice[len(reqPathSlice)-1]
	fileName := strings.Split(fileUrl, "?")[0]
	return fileName
}

func getFileNameExtension(fileName string) string {
	fileNameArr := strings.Split(fileName, ".")
	fileNameExtension := strings.ToLower(fileNameArr[len(fileNameArr)-1])
	return fileNameExtension
}

func downloadFiles(md5Seed, fileName string, sourceFileBytes []byte) []*dto.SubtitlesFileDto {
	result := []*dto.SubtitlesFileDto{}

	fileNameExtension := getFileNameExtension(fileName)
	itemDtos := []*dto.FileItemFilterDto{}
	switch fileNameExtension {
	case "zip":
		zipReader, err := zip.NewReader(bytes.NewReader(sourceFileBytes), int64(len(sourceFileBytes)))
		if err != nil {
			break
		}

		for _, f := range zipReader.File {

			if f.FileInfo().IsDir() {
				continue
			}

			childFileName := f.Name

			//如果标致位是0  则是默认的本地编码   默认为gbk
			//如果标志为是 1 << 11也就是 2048  则是utf-8编码

			if f.Flags != 2048 {
				// i := bytes.NewReader([]byte(f.Name))
				// decoder := transform.NewReader(i, simplifiedchinese.GB18030.NewDecoder())
				// content, _ := ioutil.ReadAll(decoder)

				childFileName = ToUtf8Str(childFileName) //strconv.Itoa(int(f.Flags)) + "_" + ToUtf8Str(childFileName)

			}
			// else {
			// 	childFileName = "2048_" + childFileName
			// }

			if strings.Contains(childFileName, "/") || strings.Contains(childFileName, "\\") {

				childFileName = getFileName(childFileName)
			}

			inFile, err := f.Open()
			if err != nil {
				variable.ZapLog.Sugar().Errorf("%v文件解压失败：%v", fileNameExtension, err)
				continue
			}
			defer inFile.Close()

			fileBytes, err := ioutil.ReadAll(inFile)
			if err != nil {
				variable.ZapLog.Sugar().Errorf("%v文件解压失败：%v", fileNameExtension, err)
				continue
			}

			itemDtos = append(itemDtos, &dto.FileItemFilterDto{Md5Seed: md5Seed + "/" + fileName, FileName: childFileName, FileBytes: fileBytes})
		}

		result = append(result, greate(itemDtos)...)

	case "7z":
		go7zReader, err := go7z.NewReader(bytes.NewReader(sourceFileBytes), int64(len(sourceFileBytes)))
		if err != nil {
			panic(err)
		}

		for {
			hdr, err := go7zReader.Next()
			if err == io.EOF {
				break // End of archive
			}
			if err != nil {
				panic(err)
			}

			// If empty stream (no contents) and isn't specifically an empty file...
			// then it's a directory.
			if hdr.IsEmptyStream && !hdr.IsEmptyFile {
				// if err := os.MkdirAll(hdr.Name, os.ModePerm); err != nil {
				// 	panic(err)
				// }
				continue
			}

			childFileName := hdr.Name
			if strings.Contains(childFileName, "/") || strings.Contains(childFileName, "\\") {

				childFileName = getFileName(childFileName)
			}

			newrc := io.NopCloser(go7zReader)
			defer newrc.Close()

			fileBytes, err := ioutil.ReadAll(newrc)
			if err != nil {
				variable.ZapLog.Sugar().Errorf("%v文件解压失败：%v", fileNameExtension, err)
				continue
			}
			itemDtos = append(itemDtos, &dto.FileItemFilterDto{Md5Seed: md5Seed + "/" + fileName, FileName: childFileName, FileBytes: fileBytes})
		}

		result = append(result, greate(itemDtos)...)

	case "rar":
		r := archiver.NewRar()
		err := r.Open(bytes.NewReader(sourceFileBytes), 0)
		if err != nil {
			break
		}
		defer r.Close()

		for {
			f, err := r.Read()
			//err为io.EOF 代表结束，没文件可读取
			if err == io.EOF {
				break
			}

			if err != nil {
				if r.ContinueOnError {
					log.Printf("[ERROR] Opening next file: %v", err)
					continue
				}
				break
			}

			defer f.Close()

			rh, ok := f.Header.(*rardecode.FileHeader)
			if !ok {
				continue
			}

			if f.FileInfo.IsDir() {
				continue
			}

			childFileName := rh.Name
			//checkErrorName(fileName)
			if strings.Contains(childFileName, "/") || strings.Contains(childFileName, "\\") {

				childFileName = getFileName(childFileName)
			}

			fileBytes, err := ioutil.ReadAll(f.ReadCloser)
			defer f.ReadCloser.Close()
			if err != nil {
				variable.ZapLog.Sugar().Errorf("%v文件解压失败：%v", fileNameExtension, err)
				continue
			}

			itemDtos = append(itemDtos, &dto.FileItemFilterDto{Md5Seed: md5Seed + "/" + fileName, FileName: childFileName, FileBytes: fileBytes})
		}

		result = append(result, greate(itemDtos)...)

	default:
		// fn := strings.ToLower(fileName)
		// if strings.HasSuffix(fn, ".srt") || strings.HasSuffix(fn, ".ssa") || strings.HasSuffix(fn, ".ass") || strings.HasSuffix(fn, ".stl") || strings.HasSuffix(fn, ".ts") || strings.HasSuffix(fn, ".ttml") || strings.HasSuffix(fn, ".vtt") {
		filePtah, content := ChangeCharset(md5Seed, fileName, sourceFileBytes)
		if len(filePtah) > 0 {
			name := ReplaceTitle(strings.TrimSuffix(fileName, filepath.Ext(fileName)))
			result = append(result, &dto.SubtitlesFileDto{FilePath: filePtah, Name: name, FileName: fileName, Content: content})
		}
		// }

	}
	return result
}

func greate(itemDtos []*dto.FileItemFilterDto) []*dto.SubtitlesFileDto {
	//清洗数据2
	result := []*dto.SubtitlesFileDto{}
	fileMap := map[string]*dto.FileFilterDto{}
	for _, itemDto := range itemDtos {

		fn := strings.ToLower(itemDto.FileName)
		mapKey := mapKeyReplacer.Replace(fn)

		level := 0
		if strings.Contains(fn, "繁体") {
			continue
		} else if strings.Contains(fn, "简体&英文") {
			level = 300
		} else if strings.Contains(fn, "简体") || strings.Contains(fn, "英文") {
			level = 200
		} else {
			level = 100
		}

		if strings.HasSuffix(fn, ".srt") {
			level += 80
		} else if strings.HasSuffix(fn, ".ssa") {
			level += 70
		} else if strings.HasSuffix(fn, ".ass") {
			level += 60
		} else if strings.HasSuffix(fn, ".stl") {
			level += 50
		} else if strings.HasSuffix(fn, ".ts") {
			level += 40
		} else if strings.HasSuffix(fn, ".ttml") {
			level += 30
		} else if strings.HasSuffix(fn, ".vtt") {
			level += 20
		} else if strings.HasSuffix(fn, ".zip") {
			level += 4
		} else if strings.HasSuffix(fn, ".rar") {
			level += 3
		} else if strings.HasSuffix(fn, ".7z") {
			level += 2
		} else {
			level += 1
		}

		itemDto.Level = level

		if fileMap[mapKey] == nil {
			fileMap[mapKey] = &dto.FileFilterDto{Level: level, Files: []*dto.FileItemFilterDto{itemDto}}

		} else {
			if fileMap[mapKey].Level < level {
				fileMap[mapKey].Level = level
			}

			fileMap[mapKey].Files = append(fileMap[mapKey].Files, itemDto)
		}
	}

	for _, ffd := range fileMap {
		for _, item := range ffd.Files {
			if item.Level == ffd.Level {
				childFilePath := downloadFiles(item.Md5Seed, item.FileName, item.FileBytes)
				result = append(result, childFilePath...)
			}
		}
	}

	return result
}

// func SaveFile(md5Seed, fileName string, reader io.Reader) string {

// 	//文件是否已经存在
// 	md5Str := StrMd5(md5Seed)
// 	filePath := `subtitles\` + md5Str + `\` + fileName
// 	sysFilePath := variable.BasePath + `\client\cmd\assert\` + filePath

// 	_, err := os.Stat(sysFilePath)
// 	if err == nil {
// 		variable.ZapLog.Sugar().Infof("跳过已下载文件：%v", fileName)
// 		return ""
// 	} else if !os.IsNotExist(err) {
// 		variable.ZapLog.Sugar().Infof("判断文件是否存在发生异常：%v", fileName)
// 		return ""
// 	}

// 	//创建文件夹
// 	dirPath := filepath.Dir(sysFilePath)
// 	err = os.MkdirAll(dirPath, os.ModePerm)
// 	if err != nil {
// 		return ""
// 	}
// 	//生成文件
// 	out, err := os.Create(sysFilePath)
// 	if err != nil {
// 		return ""
// 	}
// 	defer out.Close()

// 	bytes, err := ioutil.ReadAll(reader)
// 	if err != nil {
// 		return ""
// 	}

// 	detector := chardet.NewTextDetector()
// 	charset, err := detector.DetectBest(bytes)
// 	if err != nil {
// 		panic(err)
// 	}
// 	// println(charset.Charset)
// 	// println(charset.Language)

// 	targetStr := ""
// 	if charset.Charset == "UTF-8" {
// 		targetStr = string(bytes)
// 	} else {
// 		targetStr = Convert(string(bytes), charset.Charset, "utf-8")
// 	}
// 	_, err = out.WriteString(targetStr)
// 	// _, err = io.Copy(out, decoder.NewReader(reader))
// 	if err != nil {
// 		return ""
// 	}
// 	return filePath
// }

func ChangeCharset(md5Seed, fileName string, fileBytes []byte) (string, *string) {

	md5Str := StrMd5(md5Seed)
	filePath := md5Str + `\` + fileName

	// bytes, err := ioutil.ReadAll(*reader)
	// if err != nil {
	// 	return "", nil
	// }

	charset, err := detector.DetectBest(fileBytes)
	if err != nil {
		panic(err)
	}

	targetStr := ""
	if charset.Charset == "UTF-8" {
		targetStr = string(fileBytes)
	} else {
		targetStr = Convert(string(fileBytes), charset.Charset, "utf-8")
	}

	return filePath, &targetStr
}

func ToUtf8Str(fileName string) string {
	i := bytes.NewReader([]byte(fileName))
	decoder := transform.NewReader(i, simplifiedchinese.GB18030.NewDecoder())
	content, _ := ioutil.ReadAll(decoder)
	return string(content)
}

// func genSonyflake() uint64 {

// 	id, err := flake.NextID()
// 	if err != nil {
// 		fmt.Printf("flake.NextID() failed with %s\n", err)
// 	}
// 	return id
// }

func Convert(src string, srcCode string, tagCode string) string {
	srcCoder := mahonia.NewDecoder(srcCode)
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder(tagCode)
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
	result := string(cdata)
	return result
}

// func Convert(src []byte, srcCode string) string {
// 	tagCoder := mahonia.NewDecoder(srcCode)
// 	result := tagCoder.ConvertString(string(src))
// 	return result
// }

// func isGBK(data []byte) bool {
// 	length := len(data)
// 	var i int = 0
// 	for i < length {
// 		if data[i] <= 0x7f {
// 			//编码0~127,只有一个字节的编码，兼容ASCII码
// 			i++
// 			continue
// 		} else {
// 			//大于127的使用双字节编码，落在gbk编码范围内的字符
// 			if data[i] >= 0x81 &&
// 				data[i] <= 0xfe &&
// 				data[i+1] >= 0x40 &&
// 				data[i+1] <= 0xfe &&
// 				data[i+1] != 0xf7 {
// 				i += 2
// 				continue
// 			} else {
// 				return false
// 			}
// 		}
// 	}
// 	return true
// }

// func isUtf8(data []byte) bool {
// 	i := 0
// 	for i < len(data) {
// 		if (data[i] & 0x80) == 0x00 {
// 			// 0XXX_XXXX
// 			i++
// 			continue
// 		} else if num := preNUm(data[i]); num > 2 {
// 			// 110X_XXXX 10XX_XXXX
// 			// 1110_XXXX 10XX_XXXX 10XX_XXXX
// 			// 1111_0XXX 10XX_XXXX 10XX_XXXX 10XX_XXXX
// 			// 1111_10XX 10XX_XXXX 10XX_XXXX 10XX_XXXX 10XX_XXXX
// 			// 1111_110X 10XX_XXXX 10XX_XXXX 10XX_XXXX 10XX_XXXX 10XX_XXXX
// 			// preNUm() 返回首个字节的8个bits中首个0bit前面1bit的个数，该数量也是该字符所使用的字节数
// 			i++
// 			for j := 0; j < num-1; j++ {
// 				//判断后面的 num - 1 个字节是不是都是10开头
// 				if (data[i] & 0xc0) != 0x80 {
// 					return false
// 				}
// 				i++
// 			}
// 		} else {
// 			//其他情况说明不是utf-8
// 			return false
// 		}
// 	}
// 	return true
// }

// func preNUm(data byte) int {
// 	var mask byte = 0x80
// 	var num int = 0
// 	//8bit中首个0bit前有多少个1bits
// 	for i := 0; i < 8; i++ {
// 		if (data & mask) == mask {
// 			num++
// 			mask = mask >> 1
// 		} else {
// 			break
// 		}
// 	}
// 	return num
// }

func WorkClock(name string) {
	now := time.Now()
	if now.Hour() < 8 || now.Hour() > 23 {
		next := now
		if now.Hour() > 20 {
			next = now.Add(time.Hour * 24)
		}
		next = time.Date(next.Year(), next.Month(), next.Day(), 8, 0, 0, 0, now.Location())
		// 5.初始化全局日志句柄，并载入日志钩子处理函数
		variable.ZapLog.Sugar().Infof("%v 现在是%v 休眠到%v", name, now.Format(variable.TimeFormat), next.Format(variable.TimeFormat))
		// variable.ZapLog.Sugar().Infof("%v 现在是%v 休眠到%v", name, now.Format(variable.TimeFormat), next.Format(variable.TimeFormat))
		time.Sleep(next.Sub(now))

		variable.ZapLog.Sugar().Infof("%v 开始工作...", name)
	}
}

func Sleep(name, workType, timeType string, min, max int) {

	sleepTime := RandInt(min, max)
	variable.ZapLog.Sugar().Infof("%v 休眠%v%v 后面执行%v", name, sleepTime, timeType, workType)
	switch timeType {
	case "m":
		time.Sleep(time.Duration(sleepTime) * time.Minute)
	case "s":
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}
}

func RandInt(min, max int) int {
	if min >= max || min == 0 || max == 0 {
		return max
	}
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}
