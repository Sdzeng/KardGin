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
	"strconv"
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
	detector         *chardet.Detector
	fileNameReplacer *strings.Replacer
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
	}

	fileNameReplacer = strings.NewReplacer(replaceKeywords...)
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

func Download(taskDto *dto.TaskDto) (*dto.TaskDto, error) {
	//请求资源
	req, err := GetRequest(taskDto)
	if err != nil {
		return nil, err
	}
	var res *http.Response
	res, err = GetResponse(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	//拷贝
	fileName, err := GetDownloadFileName(taskDto.DownloadUrl, res)
	if err != nil {

		html, err := getHtml(res)
		if err != nil {
			variable.ZapLog.Sugar().Infof("read html error %v", err)
			return nil, err
		}

		if strings.Contains(*html, "已超出字幕下载个数限制，涉嫌恶意采集") {
			return nil, errors.New("被拦截")
		}

		return nil, err
	}

	if len(fileName) == 0 {
		return nil, errors.New("GetDownloadFileName:获取不到文件名")
	}

	// fileName = ToUtf8Str(fileName)
	// md5Seed := ToUtf8Str(taskDto.DownloadUrl)
	//checkErrorName(fileName)
	taskDto.SubtitlesFiles = downloadFiles(taskDto.DownloadUrl, fileName, &res.Body)

	// taskDto.SubtitlesFiles = make([]*dto.SubtitlesFileDto, 0)
	// for _, filePath := range filePaths {
	// 	taskDto.SubtitlesFiles = append(taskDto.SubtitlesFiles, &dto.SubtitlesFileDto{FilePath: filePath})
	// }

	// for _, v := range dto.FilePaths {
	// 	checkErrorName(v)

	// }

	return taskDto, nil
}

func GetDownloadFileName(url string, resp *http.Response) (string, error) {
	fileName := getFileName(url)
	fileNameLower := strings.ToLower(fileName)
	urlRp := regexp.MustCompile(`\d+.html`)

	if strings.HasSuffix(fileNameLower, ".rar") || strings.HasSuffix(fileNameLower, ".zip") || strings.HasSuffix(fileNameLower, ".ass") || strings.HasSuffix(fileNameLower, ".srt") || strings.HasSuffix(fileNameLower, ".7z") {
		return fileName, nil

	} else if fileNameLower == "dx1" {
		//attachment; filename="[zmk.pw]海贼王15周年纪念特别篇幻之篇章「3D2Y跨越艾斯之死！与路飞伙伴的誓言[720p].Chs.rar""
		contentDisposition := resp.Header.Get("Content-Disposition")
		filenameReg := `attachment; filename="([^"]+)"`
		headerRp := regexp.MustCompile(filenameReg)

		items := headerRp.FindAllStringSubmatch(contentDisposition, -1)

		if len(items) == 1 {
			fileName = items[0][1]
			return fileName, nil
		}

	} else if fileNameLower == "target.php" {
		return "", nil
	} else if items := urlRp.FindAllStringSubmatch(fileNameLower, -1); items != nil {
		return "", nil
	}

	return "", errors.New("不支持下载的文件" + fileName)
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

func downloadFiles(md5Seed, fileName string, rc *io.ReadCloser) []*dto.SubtitlesFileDto {
	result := []*dto.SubtitlesFileDto{}

	fileNameExtension := getFileNameExtension(fileName)
	itemDtos := []*dto.FileItemFilterDto{}
	switch fileNameExtension {
	case "zip":
		body, err := ioutil.ReadAll(*rc)
		if err != nil {
			break
		}

		zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
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

				childFileName = strconv.Itoa(int(f.Flags)) + "_" + ToUtf8Str(childFileName)

			} else {
				childFileName = "2048_" + childFileName
			}

			if strings.Contains(childFileName, "/") || strings.Contains(childFileName, "\\") {

				childFileName = getFileName(childFileName)
			}

			inFile, err := f.Open()
			if err != nil {
				continue
			}
			defer inFile.Close()

			itemDtos = append(itemDtos, &dto.FileItemFilterDto{Md5Seed: md5Seed + "/" + fileName, FileName: childFileName, FilePointer: &inFile})
		}

		result = append(result, greate(itemDtos)...)

	case "7z":

		//todo
		body, err := ioutil.ReadAll(*rc)
		if err != nil {
			break
		}
		go7zReader, err := go7z.NewReader(bytes.NewReader(body), int64(len(body)))
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

			itemDtos = append(itemDtos, &dto.FileItemFilterDto{Md5Seed: md5Seed + "/" + fileName, FileName: childFileName, FilePointer: &newrc})

			// childFilePath := downloadFiles(md5Seed+"/"+fileName, childFileName, &newrc)
			// result = append(result, childFilePath...)

			// f, err := os.Create(hdr.Name)
			// if err != nil {
			// 	panic(err)
			// }
			// defer f.Close()

			// if _, err := io.Copy(f, go7zReader); err != nil {
			// 	panic(err)
			// }
		}

		result = append(result, greate(itemDtos)...)

	case "rar":
		r := archiver.NewRar()
		err := r.Open(*rc, 0)
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
			itemDtos = append(itemDtos, &dto.FileItemFilterDto{Md5Seed: md5Seed + "/" + fileName, FileName: childFileName, FilePointer: &f.ReadCloser})
			// childFilePath := downloadFiles(md5Seed+"/"+fileName, childFileName, &f.ReadCloser)
			// result = append(result, childFilePath...)

			// err = f.Close()
			// if err != nil {
			// 	PrintError("f.Close", err.Error(), false)
			// }
		}

		result = append(result, greate(itemDtos)...)

	default:
		fn := strings.ToLower(fileName)
		if strings.HasSuffix(fn, ".srt") || strings.HasSuffix(fn, ".ssa") || strings.HasSuffix(fn, ".ass") || strings.HasSuffix(fn, ".stl") || strings.HasSuffix(fn, ".ts") || strings.HasSuffix(fn, ".ttml") || strings.HasSuffix(fn, ".vtt") {
			filePtah, content := ChangeCharset(md5Seed, fileName, rc)
			if len(filePtah) > 0 {
				fileNameWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
				result = append(result, &dto.SubtitlesFileDto{FilePath: filePtah, FileName: fileNameWithoutExt, Content: content})
			}
		}

	}
	return result
}

func greate(itemDtos []*dto.FileItemFilterDto) []*dto.SubtitlesFileDto {
	//清洗数据2
	result := []*dto.SubtitlesFileDto{}
	fileMap := map[string]*dto.FileFilterDto{}
	for _, itemDto := range itemDtos {

		fn := strings.ToLower(itemDto.FileName)
		mapKey := fileNameReplacer.Replace(fn)

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
			level += 70
		} else if strings.HasSuffix(fn, ".ssa") {
			level += 60
		} else if strings.HasSuffix(fn, ".ass") {
			level += 50
		} else if strings.HasSuffix(fn, ".stl") {
			level += 40
		} else if strings.HasSuffix(fn, ".ts") {
			level += 30
		} else if strings.HasSuffix(fn, ".ttml") {
			level += 20
		} else if strings.HasSuffix(fn, ".vtt") {
			level += 10
		} else if strings.HasSuffix(fn, ".zip") {
			level += 3
		} else if strings.HasSuffix(fn, ".rar") {
			level += 2
		} else if strings.HasSuffix(fn, ".7z") {
			level += 1
		} else {
			continue
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
				childFilePath := downloadFiles(item.Md5Seed, item.FileName, item.FilePointer)
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

func ChangeCharset(md5Seed, fileName string, reader *io.ReadCloser) (string, *string) {

	md5Str := StrMd5(md5Seed)
	filePath := md5Str + `\` + fileName

	bytes, err := ioutil.ReadAll(*reader)
	if err != nil {
		return "", nil
	}

	charset, err := detector.DetectBest(bytes)
	if err != nil {
		panic(err)
	}

	targetStr := ""
	if charset.Charset == "UTF-8" {
		targetStr = string(bytes)
	} else {
		targetStr = Convert(string(bytes), charset.Charset, "utf-8")
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
	if now.Hour() < 8 || now.Hour() > 24 {
		next := now
		if now.Hour() > 20 {
			next = now.Add(time.Hour * 24)
		}
		next = time.Date(next.Year(), next.Month(), next.Day(), 8, 0, 0, 0, now.Location())
		// 5.初始化全局日志句柄，并载入日志钩子处理函数
		variable.ZapLog.Sugar().Infof("%v 现在是%v 休眠到%v", name, now.Format("2006-01-02 15:04:05"), next.Format("2006-01-02 15:04:05"))
		// variable.ZapLog.Sugar().Infof("%v 现在是%v 休眠到%v", name, now.Format("2006-01-02 15:04:05"), next.Format("2006-01-02 15:04:05"))
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
