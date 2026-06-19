# 技术详细设计

## 1. 核心对象
- Service
- Repo
- Pipeline
- Build
- Environment
- Domain
- Certificate
- LogSource
- MetricPanel
- AlertRule
- Release
- Approval
- Incident
- RollbackRecord
- ReleaseEvent
- AuditLog
- AiRun

## 2. 关键设计
### 服务创建
创建服务主记录 -> 创建或绑定仓库 -> 生成 CI -> 创建测试环境 -> 绑定域名 -> 申请证书 -> 接入日志和监控 -> 写入事件。

### 发布
创建发布单 -> 风险分析 -> 审批 -> 部署 -> 验证 -> 观察 -> 成功或回滚。

### 日志与监控
日志和指标作为服务详情页的统一观测能力，必须支持按服务、环境、版本和时间窗口查询。

### AI
AI 仅用于生成草案、分析风险、解释故障；所有输出必须可预览、可编辑、可审计。

## 3. 数据层原则
- 主表存当前状态
- 事件表存过程
- 大文本和草案落对象存储或 JSONB
- 关键流程必须保留事件链

