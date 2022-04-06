package service

import (
	"sync"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/mumushuiding/util"
	"go-workflow/workflow-engine/flow"
	"go-workflow/workflow-engine/model"
)

// ProcessReceiver 接收页面传递参数
type ProcessReceiver struct {
	UserID     string             `json:"userId"`
	ProcInstID string             `json:"procInstID"`
	Username   string             `json:"username"`
	Company    string             `json:"company"`
	ProcName   string             `json:"procName"`
	Title      string             `json:"title"`
	Department string             `json:"department"`
	Var        *map[string]string `json:"var"`
}

// ProcessPageReceiver 分页参数
type ProcessPageReceiver struct {
	util.Page
	// 我分管的部门
	Departments []string `json:"departments"`
	// 我所属于的用户组或者角色
	Groups     []string `josn:"groups"`
	UserID     string   `json:"userID"`
	Username   string   `json:"username"`
	Company    string   `json:"company"`
	ProcName   string   `json:"procName"`
	ProcInstID string   `json:"procInstID"`
	IsFinished bool     `json:"is_finished"`
}

var copyLock sync.Mutex

// GetDefaultProcessPageReceiver GetDefaultProcessPageReceiver
func GetDefaultProcessPageReceiver() *ProcessPageReceiver {
	var p = ProcessPageReceiver{}
	p.PageIndex = 1
	p.PageSize = 10
	return &p
}
func findAll(pr *ProcessPageReceiver) ([]*model.ProcInst, int, error) {
	var page = util.Page{}
	page.PageRequest(pr.PageIndex, pr.PageSize)
	return model.FindProcInsts(pr.UserID, pr.ProcName, pr.Company, pr.Groups, pr.Departments, pr.IsFinished, pr.PageIndex, pr.PageSize)
}

// FindProcInstByID FindProcInstByID
func FindProcInstByID(id int) (string, error) {
	data, err := model.FindProcInstByID(id)
	if err != nil {
		return "", err
	}
	return util.ToJSONStr(data)
}

// FindAllPageAsJSON FindAllPageAsJSON
func FindAllPageAsJSON(pr *ProcessPageReceiver) (string, error) {
	datas, count, err := findAll(pr)
	if err != nil {
		return "", err
	}
	return util.ToPageJSON(datas, count, pr.PageIndex, pr.PageSize)
}

// StartProcessInstanceByID 启动流程
func (p *ProcessReceiver) StartProcessInstanceByID(variable *map[string]string) (int, error) {
	// times := time.Now()
	// runtime.GOMAXPROCS(2)
	// 获取流程定义
	node, prodefID, procdefName, err := GetResourceByNameAndCompany(p.ProcName, p.Company)
	if err != nil {
		return 0, err
	}
	// fmt.Printf("获取流程定义耗时：%v", time.Since(times))
	//--------以下需要添加事务-----------------
	step := 0 // 0 为开始节点
	tx := model.GetTx()
	// 新建流程实例
	var procInst = model.ProcInst{
		ProcDefID:     prodefID,
		ProcDefName:   procdefName,
		Title:         p.Title,
		Department:    p.Department,
		StartTime:     util.FormatDate(time.Now(), util.YYYY_MM_DD_HH_MM_SS),
		StartUserID:   p.UserID,
		StartUserName: p.Username,
		Company:       p.Company,
	} //开启事务
	// times = time.Now()
	procInstID, err := CreateProcInstTx(&procInst, tx) // 事务
	// fmt.Printf("启动流程实例耗时：%v", time.Since(times))
	exec := &model.Execution{
		ProcDefID:  prodefID,
		ProcInstID: procInstID,
	}
	task := &model.Task{
		NodeID:        "开始",
		ProcInstID:    procInstID,
		Assignee:      p.UserID,
		IsFinished:    true,
		ClaimTime:     util.FormatDate(time.Now(), util.YYYY_MM_DD_HH_MM_SS),
		Step:          step,
		MemberCount:   1,
		UnCompleteNum: 0,
		ActType:       "or",
		AgreeNum:      1,
	}
	// 生成执行流，一串运行节点
	_, err = GenerateExec(exec, node, p.UserID, variable, tx) //事务
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	// 获取执行流信息
	var nodeinfos []*flow.NodeInfo
	err = util.Str2Struct(exec.NodeInfos, &nodeinfos)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	// fmt.Printf("生成执行流耗时：%v", time.Since(times))
	// -----------------生成新任务-------------
	// times = time.Now()
	if nodeinfos[0].ActType == "and" {
		task.UnCompleteNum = nodeinfos[0].MemberCount
		task.MemberCount = nodeinfos[0].MemberCount
	}
	_, err = NewTaskTx(task, tx)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	// fmt.Printf("生成新任务耗时：%v", time.Since(times))
	//--------------------流转------------------
	// times = time.Now()
	// 流程移动到下一环节
	err = MoveStage(nodeinfos, p.UserID, p.Username, p.Company, "启动流程", "", task.ID, procInstID, step, true, tx)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	// fmt.Printf("流转到下一流程耗时：%v", time.Since(times))
	// fmt.Println("--------------提交事务----------")
	tx.Commit() //结束事务
	return procInstID, err
}

// CreateProcInstByID 新建流程实例
// func CreateProcInstByID(processDefinitionID int, startUserID string) (int, error) {
// 	var procInst = model.ProcInst{
// 		ProcDefID:   processDefinitionID,
// 		StartTime:   util.FormatDate(time.Now(), util.YYYY_MM_DD_HH_MM_SS),
// 		StartUserID: startUserID,
// 	}
// 	return procInst.Save()
// }

// CreateProcInstTx CreateProcInstTx
// 开户事务
func CreateProcInstTx(procInst *model.ProcInst, tx *gorm.DB) (int, error) {

	return procInst.SaveTx(tx)
}

// SetProcInstFinish SetProcInstFinish
// 设置流程结束
func SetProcInstFinish(procInstID int, endTime string, tx *gorm.DB) error {
	var p = &model.ProcInst{}
	p.ID = procInstID
	p.EndTime = endTime
	p.IsFinished = true
	return p.UpdateTx(tx)
}

// StartByMyself 我发起的流程
func StartByMyself(receiver *ProcessPageReceiver) (string, error) {
	var page = util.Page{}
	page.PageRequest(receiver.PageIndex, receiver.PageSize)
	datas, count, err := model.StartByMyself(receiver.UserID, receiver.Company, receiver.PageIndex, receiver.PageSize)
	if err != nil {
		return "", err
	}
	return util.ToPageJSON(datas, count, receiver.PageIndex, receiver.PageSize)
}

// FindProcNotify 查询抄送我的
func FindProcNotify(receiver *ProcessPageReceiver) (string, error) {
	var page = util.Page{}
	page.PageRequest(receiver.PageIndex, receiver.PageSize)
	datas, count, err := model.FindProcNotify(receiver.UserID, receiver.Company, receiver.Groups, receiver.PageIndex, receiver.PageSize)
	if err != nil {
		return "", err
	}
	return util.ToPageJSON(datas, count, receiver.PageIndex, receiver.PageSize)
}

// UpdateProcInst UpdateProcInst
// 更新流程实例
func UpdateProcInst(procInst *model.ProcInst, tx *gorm.DB) error {
	return procInst.UpdateTx(tx)
}

// UpdateProcInstCandidateWithDept 候选人角色添加部门
func UpdateProcInstCandidateWithDept(procInst *model.ProcInst, tx *gorm.DB) error {
	return tx.Exec(`update proc_inst set candidate=concat(department,'-',candidate) where id=?`, procInst.ID).Error
}
