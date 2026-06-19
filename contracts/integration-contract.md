# 联调约定

## 1. 服务端约定
- 返回统一 JSON 结构
- 创建和发布接口必须支持幂等键
- 长流程返回 workflowId 或可查询任务状态

## 2. 前端约定
- 页面首次进入优先使用 mock 数据渲染骨架
- 所有危险按钮必须显示状态
- 所有错误必须显示 requestId 和 retryable

## 3. 接口联调顺序
1. 服务列表
2. 服务创建
3. 发布列表
4. 发布创建
5. AI 方案生成
6. AI 风险分析

## 4. Mock 数据约定
- `contracts/mock/service-list.json`
- `contracts/mock/release-detail.json`

## 5. 统一错误展示
- 短提示
- 原因摘要
- 可操作建议
- requestId

