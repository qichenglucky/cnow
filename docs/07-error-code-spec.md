# 接口错误码与异常处理规范

## 1. 目标
统一前后端、测试、运维对错误的理解，要求每个错误都能说明：
- 谁失败了
- 为什么失败
- 是否可重试
- 是否可补偿
- 是否影响业务状态

## 2. 错误分层
- 1xxxx 参数与校验错误
- 2xxxx 认证与权限错误
- 3xxxx 资源与状态错误
- 4xxxx 外部系统错误
- 5xxxx 平台内部错误

## 3. 错误响应格式
```json
{
  "code": 30002,
  "message": "release state not allowed",
  "data": {
    "resourceType": "release",
    "resourceId": 123,
    "currentStatus": "deploying"
  },
  "requestId": "req_abc123",
  "retryable": false
}
```

## 4. 核心规则
- 状态错误不重试
- 外部系统 5xx 可重试
- 所有长流程都要返回 requestId
- 所有生产动作都要审计
- AI 错误不能影响主流程存活

