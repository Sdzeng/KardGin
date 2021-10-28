package helper

import (
	"kard/src/model/dto"
	"path/filepath"
	"strings"
	"time"

	"github.com/asticode/go-astisub"
)

// func ParseFile(taskDto *dto.TaskDto) {
// 	defer func(dto *dto.TaskDto) {
// 		dto.Wg.Done()
// 	}(taskDto)

// 	batchNum := 10
// 	//dtoSlice := []*dto.SubtitlesIndexDto{}

// 	for _, subtitlesFile := range taskDto.SubtitlesFiles {
// 		subtitlesFile.SubtitleItems = []*dto.SubtitlesItemDto{}
// 		subtitlesFile.FileName = getPathFileName(subtitlesFile.FilePath)

// 		subtitles, err := astisub.Open(astisub.Options{Filename: subtitlesFile.FilePath})
// 		if err != nil {
// 			continue
// 		}

// 		lineTextSlice := []string{}
// 		lineTextSliceLen := 0
// 		var startAt time.Duration

// 		for _, item := range subtitles.Items {
// 			for _, line := range item.Lines {
// 				//lineText := line.VoiceName + "："
// 				for _, lineItem := range line.Items {
// 					if len(strings.Trim(lineItem.Text, " ")) <= 0 {
// 						continue
// 					}

// 					lineTextSlice = append(lineTextSlice, lineItem.Text)
// 					lineTextSliceLen := len(lineTextSlice)

// 					if (lineTextSliceLen-1)%batchNum == 0 {
// 						startAt = item.StartAt //.Seconds().String()
// 					}

// 					if lineTextSliceLen%batchNum == 0 {

// 						itemDto := &dto.SubtitlesItemDto{
// 							Text:    lineTextSlice,
// 							StartAt: startAt,
// 						}

// 						subtitlesFile.SubtitleItems = append(subtitlesFile.SubtitleItems, itemDto)

// 						lineTextSlice = []string{} //(lineTextSlice)[0:0]
// 						startAt = 0 * time.Second
// 					}

// 				}
// 				//fmt.Println(lineText)
// 			}
// 		}

// 		lineTextSliceLen = len(lineTextSlice)
// 		if lineTextSliceLen > 0 {
// 			itemDto := &dto.SubtitlesItemDto{
// 				Text:    lineTextSlice,
// 				StartAt: startAt,
// 			}

// 			subtitlesFile.SubtitleItems = append(subtitlesFile.SubtitleItems, itemDto)
// 		}

// 	}

// 	taskDto.StoreFunc(taskDto)
// }

func ParseFile(taskDto *dto.TaskDto) {
	defer func(dto *dto.TaskDto) {
		dto.Wg.Done()

		if err := recover(); err != nil {
			PrintError("ParseFile", err.(error).Error(), true)
		}
	}(taskDto)

	batchNum := 10
	startAt := 0 * time.Second
	texts := []string{}
	for _, subtitlesFile := range taskDto.SubtitlesFiles {
		subtitlesFile.SubtitleItems = []*dto.SubtitlesItemDto{}

		subtitles, err := open(subtitlesFile.FileName, subtitlesFile.Content)
		if err != nil {
			continue
		}

		for _, item := range subtitles.Items {
			for _, line := range item.Lines {
				//lineText := line.VoiceName + "："
				for _, lineItem := range line.Items {
					if len(strings.Trim(lineItem.Text, " ")) <= 0 {
						continue
					}

					texts = append(texts, lineItem.Text)
					//分批
					if (len(texts)-1)%batchNum == 0 {
						startAt = item.StartAt
					}
					if len(texts)%batchNum == 0 {
						itemDto := &dto.SubtitlesItemDto{
							Texts:   texts,
							StartAt: item.StartAt,
						}

						subtitlesFile.SubtitleItems = append(subtitlesFile.SubtitleItems, itemDto)
						texts = []string{} //(lineTextSlice)[0:0]
						startAt = 0 * time.Second
					}
				}
			}
		}

		//分批
		if len(texts) > 0 {
			itemDto := &dto.SubtitlesItemDto{
				Texts:   texts,
				StartAt: startAt,
			}

			subtitlesFile.SubtitleItems = append(subtitlesFile.SubtitleItems, itemDto)
			texts = []string{} //(lineTextSlice)[0:0]
			startAt = 0 * time.Second
		}

		//生成文件md5
		if len(subtitlesFile.SubtitleItems) > 0 {
			md5Seed := ""
			for _, text := range subtitlesFile.SubtitleItems[0].Texts {
				md5Seed += strings.Trim(text, " ")
			}
			subtitlesFile.FileSum = StrMd5(md5Seed)
		}

	}

	taskDto.StoreFunc(taskDto)
}

// func getPathFileName(filePath string) string {
// 	filePathSlice := strings.Split(filePath, "\\")
// 	fileFullName := filePathSlice[len(filePathSlice)-1]
// 	fileSuffix := path.Ext(fileFullName)

// 	return strings.TrimSuffix(fileFullName, fileSuffix)
// }

func open(fileName string, content *string) (s *astisub.Subtitles, err error) {
	reader := strings.NewReader(*content)
	o := astisub.Options{Filename: fileName}

	// Parse the content
	switch filepath.Ext(strings.ToLower(o.Filename)) {
	case ".srt":
		s, err = astisub.ReadFromSRT(reader)
	case ".ssa", ".ass":
		s, err = astisub.ReadFromSSA(reader)
	case ".stl":
		s, err = astisub.ReadFromSTL(reader)
	case ".ts":
		s, err = astisub.ReadFromTeletext(reader, o.Teletext)
	case ".ttml":
		s, err = astisub.ReadFromTTML(reader)
	case ".vtt":
		s, err = astisub.ReadFromWebVTT(reader)
	default:
		err = astisub.ErrInvalidExtension
	}
	return
}
