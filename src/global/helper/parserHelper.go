package helper

import (
	"io"
	"kard/src/model/dto"
	"path/filepath"
	"strings"

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

	//dtoSlice := []*dto.SubtitlesIndexDto{}
	// sysFilePath := ""
	for _, subtitlesFile := range taskDto.SubtitlesFiles {
		subtitlesFile.SubtitleItems = []*dto.SubtitlesItemDto{}
		// subtitlesFile.FileName = getPathFileName(subtitlesFile.FilePath)

		// sysFilePath = variable.BasePath + `\client\cmd\assert\` + subtitlesFile.FilePath
		subtitles, err := open(subtitlesFile.FileName, subtitlesFile.Reader)
		if err != nil {
			continue
		}

		lineTextSlice := []string{}
		for _, item := range subtitles.Items {
			for _, line := range item.Lines {
				//lineText := line.VoiceName + "："
				for _, lineItem := range line.Items {
					if len(strings.Trim(lineItem.Text, " ")) <= 0 {
						continue
					}

					lineTextSlice = append(lineTextSlice, lineItem.Text)
				}
			}

			itemDto := &dto.SubtitlesItemDto{
				Text:    lineTextSlice,
				StartAt: item.StartAt,
			}

			subtitlesFile.SubtitleItems = append(subtitlesFile.SubtitleItems, itemDto)

			lineTextSlice = []string{} //(lineTextSlice)[0:0]

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

func open(fileName string, reader io.Reader) (s *astisub.Subtitles, err error) {
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
