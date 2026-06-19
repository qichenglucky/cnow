# 接口定义文档

## 1. 设计原则
- 按业务域分组接口
- 创建和发布类接口必须支持幂等
- 长流程返回任务状态或工作流标识
- 高风险动作必须带权限与审计上下文
- AI 接口只输出草案和建议，不直接执行生产动作

## 2. 通用约定

### 请求头
- `Authorization: Bearer <token>`
- `X-Request-Id`
- `X-Idempotency-Key`

### 通用返回结构
```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

### 通用错误字段
- `code`
- `message`
- `data`
- `requestId`
- `retryable`

## 3. 核心接口

### 3.1 创建服务
`POST /api/services`

请求：
```json
{
  "name": "user-api",
  "displayName": "用户服务",
  "description": "用户中心 API",
  "techStack": "go",
  "repoSource": {
    "type": "new"
  },
  "environmentTypes": ["test", "prod"],
  "domainPolicy": "wildcard",
  "logPolicy": "sls",
  "metricPolicy": "prometheus"
}
```

响应：
```json
{
  "serviceId": 123,
  "status": "creating",
  "previewId": "pv_456"
}
```

### 3.2 方案预览
`POST /api/services/preview`

用途：
- 根据自然语言或表单生成配置草案

### 3.3 创建或接入仓库
`POST /api/repos`

### 3.4 生成 CI 配置
`POST /api/pipelines/ci/generate`

### 3.5 触发构建
`POST /api/builds`

### 3.6 部署测试环境
`POST /api/environments/test/deploy`

### 3.7 创建发布单
`POST /api/releases`

### 3.8 审批发布
`POST /api/releases/{releaseId}/approvals`

### 3.9 执行发布
`POST /api/releases/{releaseId}/deploy`

### 3.10 回滚
`POST /api/releases/{releaseId}/rollback`

### 3.11 查询服务详情
`GET /api/services/{serviceId}`

### 3.12 查询发布历史
`GET /api/services/{serviceId}/releases`

### 3.13 查询日志
`GET /api/logs`

### 3.14 查询监控
`GET /api/metrics`

### 3.15 AI 方案生成
`POST /api/ai/plan`

### 3.16 AI 风险分析
`POST /api/ai/risk`

### 3.17 AI 故障解释
`POST /api/ai/incident-explain`

## 4. 接口行为约束
- 所有创建接口必须有幂等键
- 所有生产动作必须先过权限检查，再写审计
- 所有长任务必须返回工作流标识或可查询任务 ID
- 所有异常必须返回可追踪的 requestId

