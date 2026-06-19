# 技术架构设计

## 1. 架构目标
- 串起仓库、CI、测试环境、域名、日志、监控、生产发布、回滚
- 所有动作有统一事件模型和状态机
- AI 只生成建议，不直接越权执行

## 2. 总体架构
系统分为五层：
1. 入口层：Web UI、API Gateway、BFF
2. 业务域层：Service Catalog、Release Orchestrator、Integration Hub、Observability Hub、AI Copilot、Policy & Audit
3. 工作流层：Temporal
4. 消息层：Kafka
5. 基础设施层：PostgreSQL、Redis、对象存储、K8s、可观测性系统

## 3. 架构原则
- 业务与执行分离
- 状态与事件分离
- AI 与执行分离
- 外部系统适配隔离

## 4. 已拍板的架构决策
- 主键统一自增 bigint
- 不预留多租户字段
- 配置草案与发布快照分离存储
- 生产发布必须走工作流
- AI 只做建议与分析

