package helper

import (
	"kard/src/model/dto"
	"path"
	"strings"

	"github.com/asticode/go-astisub"
)

func ParseFile(taskDto *dto.TaskDto) {
	defer func(dto *dto.TaskDto) {
		dto.Wg.Done()
	}(taskDto)

	batchNum := 10
	//dtoSlice := []*dto.SubtitlesIndexDto{}

	for _, subtitlesFile := range taskDto.SubtitlesFiles {
		subtitlesFile.SubtitleItems = []*dto.SubtitlesItemDto{}
		subtitlesFile.FileName = getPathFileName(subtitlesFile.FilePath)

		subtitles, err := astisub.Open(astisub.Options{Filename: subtitlesFile.FilePath})
		if err != nil {
			continue
		}

		lineTextSlice := []string{}
		lineTextSliceLen := 0
		startAt := ""

		for _, item := range subtitles.Items {
			for _, line := range item.Lines {
				//lineText := line.VoiceName + "："
				for _, lineItem := range line.Items {
					if len(strings.Trim(lineItem.Text, " ")) <= 0 {
						continue
					}

					lineTextSlice = append(lineTextSlice, lineItem.Text)
					lineTextSliceLen := len(lineTextSlice)

					if (lineTextSliceLen-1)%batchNum == 0 {
						startAt = item.StartAt.String()
					}

					if lineTextSliceLen%batchNum == 0 {

						itemDto := &dto.SubtitlesItemDto{
							Text:    lineTextSlice,
							StartAt: startAt,
						}

						subtitlesFile.SubtitleItems = append(subtitlesFile.SubtitleItems, itemDto)

						lineTextSlice = []string{} //(lineTextSlice)[0:0]
						startAt = ""
					}

				}
				//fmt.Println(lineText)
			}
		}

		lineTextSliceLen = len(lineTextSlice)
		if lineTextSliceLen > 0 {
			itemDto := &dto.SubtitlesItemDto{
				Text:    lineTextSlice,
				StartAt: startAt,
			}

			subtitlesFile.SubtitleItems = append(subtitlesFile.SubtitleItems, itemDto)
		}

	}

	taskDto.StoreFunc(taskDto)
}

func getPathFileName(filePath string) string {
	filePathSlice := strings.Split(filePath, "\\")
	fileFullName := filePathSlice[len(filePathSlice)-1]
	fileSuffix := path.Ext(fileFullName)

	return strings.TrimSuffix(fileFullName, fileSuffix)
}
