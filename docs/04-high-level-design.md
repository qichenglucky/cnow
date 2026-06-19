# 技术概要设计

## 1. 服务边界
### Service Catalog
管理服务、仓库、环境、域名、证书、日志源、监控面板、告警规则。

### Release Orchestrator
管理发布单、审批、回滚、发布时间线和状态机。

### Integration Hub
对接 GitHub/GitLab、CI、K8s、DNS、证书、SLS、监控系统。

### Observability Hub
汇总日志、指标、告警，并关联版本、环境和发布记录。

### AI Copilot
生成方案、配置草案、风险分析、故障解释与复盘摘要。

### Policy & Audit
管理权限、审批、审计、合规。

## 2. 统一状态机
- Service: draft / creating / ready / degraded / archived
- Environment: provisioning / running / updating / unhealthy / deleted
- Release: created / reviewing / approved / deploying / verifying / succeeded / failed / rolled_back
- Approval: pending / approved / rejected / expired

## 3. 事件模型
建议使用统一事件前缀：
- service.*
- build.*
- environment.*
- release.*
- alert.*
- incident.*

