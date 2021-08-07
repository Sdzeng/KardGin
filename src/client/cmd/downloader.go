package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"kard/src/dto"
	"kard/src/global/variable"
	"kard/src/repository"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/axgle/mahonia"
	"github.com/mholt/archiver"
	"github.com/nwaples/rardecode"
	"github.com/saintfish/chardet"
	"github.com/saracen/go7z"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	//"github.com/sony/sonyflake"
)

var downloadFileRepository = &repository.DownloadFileRepository{}

type WriterCounter struct {
	//urlDto   *dto.UrlDto
	FileName string
	Total    uint64
}

//var flake *sonyflake.Sonyflake

func (wc *WriterCounter) PrintProgress() {
	//fmt.Printf("\n%s", strings.Repeat(" ", 50))
	//fmt.Printf("\nDownloading... %s complete %s(%s)", humanize.Bytes(wc.Total), wc.urlDto.name, wc.urlDto.fileName)

}

func (wc *WriterCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func Download(dto *dto.UrlDto, workerQueue chan *dto.UrlDto) error {
	//请求资源
	req, err := getRequest(dto)
	if err != nil {
		return err
	}
	var res *http.Response
	res, err = getResponse(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	//拷贝
	fileName, err := getDownloadFileName(dto.DownloadUrl, res)
	if err != nil || len(fileName) == 0 {
		return errors.New("getDownloadFileName:获取不到文件名")
	}

	dto.FileName = ToUtf8Str(fileName)
	//checkErrorName(fileName)
	dto.FilePaths = downloadFiles(dto.FileName, res.Body, dto.DownloadUrl)

	// for _, v := range dto.FilePaths {
	// 	checkErrorName(v)

	// }
	err = downloadFileRepository.Save(dto)
	if err != nil {
		return err
	}

	dto.WorkType = variable.ParseFile
	workerQueue <- dto
	return nil
}

func getDownloadFileName(url string, resp *http.Response) (string, error) {
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

func downloadFiles(fileName string, rc io.ReadCloser, du string) []string {
	result := []string{}

	fileNameExtension := getFileNameExtension(fileName)
	switch fileNameExtension {
	case "zip":
		body, err := ioutil.ReadAll(rc)
		if err != nil {
			break
		}

		zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
		if err != nil {
			break
		}

		for _, f := range zipReader.File {
			//fpath := filepath.Join(destDir, f.Name)
			if f.FileInfo().IsDir() {
				continue
			}

			//如果标致位是0  则是默认的本地编码   默认为gbk
			//如果标志为是 1 << 11也就是 2048  则是utf-8编码
			if f.Flags != 2048 {
				i := bytes.NewReader([]byte(f.Name))
				decoder := transform.NewReader(i, simplifiedchinese.GB18030.NewDecoder())
				content, _ := ioutil.ReadAll(decoder)
				fileName = strconv.Itoa(int(f.Flags)) + "_" + string(content)

			} else {
				fileName = "2048_" + f.Name
			}

			if strings.Contains(fileName, "/") || strings.Contains(fileName, "\\") {

				fileName = getFileName(fileName)
			}

			inFile, err := f.Open()
			if err != nil {
				continue
			}
			defer inFile.Close()

			childFilePath := downloadFiles(fileName, inFile, "")
			result = append(result, childFilePath...)

		}

	case "7z":

		//todo
		body, err := ioutil.ReadAll(rc)
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

			// Create file
			fileName := hdr.Name
			if strings.Contains(fileName, "/") || strings.Contains(fileName, "\\") {

				fileName = getFileName(fileName)
			}

			rc := io.NopCloser(go7zReader)
			childFilePath := downloadFiles(fileName, rc, "")
			result = append(result, childFilePath...)

			// f, err := os.Create(hdr.Name)
			// if err != nil {
			// 	panic(err)
			// }
			// defer f.Close()

			// if _, err := io.Copy(f, go7zReader); err != nil {
			// 	panic(err)
			// }
		}

		break
	case "rar":
		r := archiver.NewRar()
		err := r.Open(rc, 0)
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
			defer f.Close()

			if err != nil {
				if r.ContinueOnError {
					log.Printf("[ERROR] Opening next file: %v", err)
					continue
				}
				break
			}

			rh, ok := f.Header.(*rardecode.FileHeader)
			if !ok {
				continue
			}

			if f.FileInfo.IsDir() {

				continue
			}

			fileName = rh.Name
			//checkErrorName(fileName)
			if strings.Contains(fileName, "/") || strings.Contains(fileName, "\\") {

				fileName = getFileName(fileName)
			}
			childFilePath := downloadFiles(fileName, f.ReadCloser, "")
			result = append(result, childFilePath...)
		}

	default:
		filePtah := SaveFile(fileName, rc)
		if len(filePtah) > 0 {
			result = append(result, filePtah)
		}

	}
	return result
}

func SaveFile(fileName string, reader io.Reader) string {

	//文件是否已经存在
	filePath := variable.BasePath + `\client\cmd\assert\subtitles\` + fileName
	if _, err := os.Stat(filePath); err != nil && os.IsExist(err) {

		return ""

	}

	//生成文件
	out, err := os.Create(filePath)
	if err != nil {
		return ""
	}
	defer out.Close()

	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return ""
	}

	detector := chardet.NewTextDetector()
	charset, err := detector.DetectBest(bytes)
	if err != nil {
		panic(err)
	}
	// println(charset.Charset)
	// println(charset.Language)

	targetStr := ""
	if charset.Charset == "UTF-8" {
		targetStr = string(bytes)
	} else {
		targetStr = Convert(string(bytes), charset.Charset, "utf-8")
	}
	_, err = out.WriteString(targetStr)
	// _, err = io.Copy(out, decoder.NewReader(reader))
	if err != nil {
		return ""
	}
	return filePath
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
