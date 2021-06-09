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
	"strings"

	"github.com/mholt/archiver"
	"github.com/nwaples/rardecode"
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
	refer := dto.Refers[len(dto.Refers)-1]
	req, err := getRequest(dto.DownloadUrl, refer)
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
	checkErrorName(fileName)
	dto.FilePaths = downloadFiles(dto.FileName, res.Body, dto.DownloadUrl)

	for _, v := range dto.FilePaths {
		checkErrorName(v)

	}
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

func checkErrorName(fileName string) {

	if strings.Contains(fileName, "fffd") || strings.Contains(fileName, "\\x") {

		_ = "出错"
	}
}

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
		//defer zipReader.Close()

		for _, f := range zipReader.File {
			//fpath := filepath.Join(destDir, f.Name)
			if f.FileInfo().IsDir() {
				// rc2, err := f.Open()
				// if err != nil {
				// 	break
				// }

				// body2, err := ioutil.ReadAll(rc2)
				// if err != nil {
				// 	break
				// }

				// zipReader2, err := zip.NewReader(bytes.NewReader(body2), int64(len(body2)))
				// if err != nil {
				// 	break
				// }
				// for _, f2 := range zipReader2.File {
				// 	fmt.Printf("%v", f2)
				// }

				continue
			}

			inFile, err := f.Open()
			if err != nil {
				continue
			}
			defer inFile.Close()

			fileName = f.Name
			//checkErrorName(fileName)
			if strings.Contains(fileName, "/") || strings.Contains(fileName, "\\") {

				fileName = getFileName(fileName)
			}

			childFilePath := downloadFiles(fileName, inFile, "")
			result = append(result, childFilePath...)

			// outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			// if err != nil {
			// 	return err
			// }
			// defer outFile.Close()

			// _, err = io.Copy(outFile, inFile)
			// if err != nil {
			// 	return err
			// }

		}

	case "7z":

	case "rar":
		// body, err := ioutil.ReadAll(rc)
		// if err != nil {
		// 	break
		// }

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

	wc := &WriterCounter{FileName: fileName}
	_, err = io.Copy(out, io.TeeReader(reader, wc))
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
