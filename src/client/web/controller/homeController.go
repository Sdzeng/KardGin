package controller

import (
	"bytes"
	"kard/src/client/web/response"
	"kard/src/repository"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type HomeController struct {
}

func (c *HomeController) GetCover(context *gin.Context) {

	today := time.Now()
	// userName := context.GetString(consts.ValidatorPrefix + "user_name")
	// page := context.GetFloat64(consts.ValidatorPrefix + "page")
	// limits := context.GetFloat64(consts.ValidatorPrefix + "limits")
	// limitStart := (page - 1) * limits
	videoDto, _ := repository.CreateVideoFactory().GetCover(today)
	if videoDto != nil {
		response.Success(context, response.CurdStatusOkMsg, videoDto)
	} else {
		response.Fail(context, response.CurdSelectFailCode, response.CurdSelectFailMsg, "")
	}
}

func (c *HomeController) ExtractSubtitles(context *gin.Context) {

	var txtName = time.Now().Second()
	cmdStr := "-i output.mkv -an -vn -bsf:s mov2textsub -scodec copy -f rawvideo " + strconv.Itoa(txtName) + ".txt"
	cmdArguments := strings.Split(cmdStr, " ")
	// []string{"-i", "divx.avi", "-c:v", "libx264", "-crf", "20", "-c:a", "aac", "-strict", "-2", "video1-fix.ts"}

	cmd := exec.Command("ffmpeg", cmdArguments...)

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		response.Success(context, response.CurdStatusOkMsg, "成功")
	} else {

		response.Fail(context, response.CurdSelectFailCode, err.Error(), "")
	}
}
