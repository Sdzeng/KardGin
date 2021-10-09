package helper

import (
	"kard/src/model/dto"
	"path"
	"strconv"
	"strings"

	"github.com/asticode/go-astisub"
)

var (
	ai *AutoInc
)

func init() {
	ai = NewAi(0, 1)
}

func ParseFile(taskDto *dto.TaskDto) {
	defer func(dto *dto.TaskDto) {
		dto.Wg.Done()
	}(taskDto)

	batchNum := 10
	dtoSlice := []*dto.SubtitlesIndexDto{}

	for _, filePath := range taskDto.FilePaths {
		subtitles, err := astisub.Open(astisub.Options{Filename: filePath})
		if err != nil {
			continue
		}

		subTitle := getPathFileName(filePath)

		lineTextSlice := []string{}
		lineTextSliceLen := 0
		timeDuration := ""

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
						timeDuration = item.StartAt.String()
					}

					if lineTextSliceLen%batchNum == 0 {

						indexDto := &dto.SubtitlesIndexDto{
							IndexId:      strconv.Itoa(ai.Id()),
							Title:        taskDto.Name,
							SubTitle:     subTitle,
							Text:         lineTextSlice,
							TimeDuration: timeDuration,
							Lan:          taskDto.Lan,
						}

						dtoSlice = append(dtoSlice, indexDto)

						lineTextSlice = []string{} //(lineTextSlice)[0:0]
						timeDuration = ""
					}

				}
				//fmt.Println(lineText)
			}
		}

		lineTextSliceLen = len(lineTextSlice)
		if lineTextSliceLen > 0 {

			indexDto := &dto.SubtitlesIndexDto{
				IndexId:      strconv.Itoa(ai.Id()),
				Title:        taskDto.Name,
				SubTitle:     subTitle,
				Text:         lineTextSlice,
				TimeDuration: timeDuration,
				Lan:          taskDto.Lan,
			}

			dtoSlice = append(dtoSlice, indexDto)
		}

	}

	if len(dtoSlice) <= 0 {
		return
	}

	taskDto.StoreFunc(dtoSlice)

}

func getPathFileName(filePath string) string {
	filePathSlice := strings.Split(filePath, "\\")
	fileFullName := filePathSlice[len(filePathSlice)-1]
	fileSuffix := path.Ext(fileFullName)

	return strings.TrimSuffix(fileFullName, fileSuffix)
}
