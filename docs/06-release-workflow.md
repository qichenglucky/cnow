# 发布工作流设计文档

## 1. 目标
发布工作流要把发布拆成可编排、可追踪、可补偿、可回放的标准流程。

## 2. 工作流类型
- service.create
- ci.generate
- test.deploy
- prod.release
- rollback.execute
- domain.bind
- certificate.issue
- logsource.attach
- metricpanel.attach

## 3. 生产发布状态机
- created
- risk_analyzing
- reviewing
- approved
- deploying
- verifying
- observing
- succeeded
- failed
- rollback_pending
- rolling_back
- rolled_back

## 4. 核心规则
- 长流程必须走工作流
- 所有外部系统调用必须可重试、可幂等
- 人工审批是流程节点
- 回滚是正式流程，不是临时分支

## 5. 超时与重试
- 外部调用支持指数退避重试
- 4xx 业务错误不重试
- 生产部署、观察窗口、证书申请、域名绑定都要定义超时

