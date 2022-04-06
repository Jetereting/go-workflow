package service

// UserInfo 用户信息
type UserInfo struct {
	Company string `json:"company"`
	// 用户所属部门
	Department string `json:"department"`
	Username   string `json:"username"`
	ID         string `json:"ID"`
	// 用户的角色
	Roles []string `json:"roles"`
	// 用户负责的部门
	Departments []string `json:"departments"`
}
