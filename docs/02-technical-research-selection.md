# 技术调研与选型

## 1. 选型原则
- 成熟优先
- 可观测优先
- 可运维优先
- AI 必须可控
- 外部系统必须通过适配层接入

## 2. 推荐选型
| 领域 | 选型 |
|---|---|
| 前端 | React + TypeScript + Ant Design Pro |
| 后端 | Go + Gin |
| 工作流编排 | Temporal |
| 事件总线 | Kafka |
| 主数据库 | PostgreSQL |
| 缓存/锁 | Redis |
| 对象存储 | S3/MinIO |
| 部署目标 | Kubernetes + Helm |
| 渐进发布 | Argo Rollouts |
| 域名/证书 | ExternalDNS + cert-manager + Ingress/Gateway API |
| 可观测性 | OpenTelemetry + Prometheus + Grafana |
| AI 接入 | 统一 OpenAI-compatible Provider Gateway |
| 认证授权 | OIDC + RBAC |
| IaC | Terraform |

## 3. 关键理由
- Temporal 适合长流程、审批和补偿
- K8s + Rollouts 适合标准化发布
- Kafka 适合作为全局事件时间线
- Go 适合轻量编排与集成服务
- React + AntD 适合企业后台与大量表单/表格

