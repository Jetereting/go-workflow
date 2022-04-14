# 其他平台接入：
1.发起流程
```
POST /api/v1/workflow/process/start
{"procName":"请假A","title":"请假-小明","userId":"21","username":"小明","department":"技术中心","company":"A公司"}
返回的message 为实例ID 更新到业务表
```

2.查询需要我审批的
```
POST /api/v1/workflow/process/findTask
{"userID":"22","groups":["技术中心-主任","主任"],"company":"A公司","pageIndex":1,"pageSize":10}
```

3.审批
```
POST /api/v1/workflow/task/complete
{"taskID":2,"pass":"true","userID":"22","username":"王主任","company":"A公司","comment": "拒绝"}
//如果拒绝 写入到业务表
//如果同意 查询 /api/v1/workflow/process/findById?id=1 如果返回完成，写同意到业务表
```

4.我审批的
```
GET /api/v1/workflow/identitylink/findUserFlow?user_id=22&is_launcher=false
```


5.我发起的
```
GET /api/v1/workflow/identitylink/findUserFlow?user_id=21&is_launcher=true
```

6.审批过程
```
GET /api/v1/workflow/identitylink/findParticipant?procInstID=1
```

