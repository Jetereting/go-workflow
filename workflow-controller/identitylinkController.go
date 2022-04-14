package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/mumushuiding/util"
	"go-workflow/workflow-engine/service"
)

// FindParticipantByProcInstID 根据流程id查询流程参与者
func FindParticipantByProcInstID(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" {
		util.ResponseErr(writer, "只支持get方法！！")
		return
	}
	request.ParseForm()
	if len(request.Form["procInstID"]) == 0 {
		util.ResponseErr(writer, "流程 procInstID 不能为空")
		return
	}
	procInstID, err := strconv.Atoi(request.Form["procInstID"][0])
	if err != nil {
		util.ResponseErr(writer, err)
		return
	}
	result, err := service.FindParticipantByProcInstID(procInstID)
	if err != nil {
		util.ResponseErr(writer, err)
		return
	}
	fmt.Fprintf(writer, result)

}

// FindUserFlow 查询用户参与的流程
func FindUserFlow(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" {
		util.ResponseErr(writer, "只支持get方法！！")
		return
	}
	request.ParseForm()
	if len(request.Form["user_id"]) == 0 {
		util.ResponseErr(writer, "用户 userID 不能为空")
		return
	}
	userId := request.Form["user_id"][0]
	if userId == "" {
		util.ResponseErr(writer, "流程 user_id 不能为空")
		return
	}
	isLauncher := false
	if len(request.Form["is_launcher"]) != 0 {
		isLauncher, _ = strconv.ParseBool(request.Form["is_launcher"][0])
	}

	instIDs, err := service.FindUserFlow(userId, isLauncher)
	if err != nil {
		util.ResponseErr(writer, err)
		return
	}
	result, err := service.FindProcInstByIDs(instIDs)
	if err != nil {
		util.ResponseErr(writer, err)
		return
	}
	fmt.Fprintf(writer, result)

}
