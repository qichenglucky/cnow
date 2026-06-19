# 发布工作流伪代码

## 1. 目标
把生产发布拆成可编排、可补偿、可回放的工作流。

## 2. 生产发布主流程

```text
Start
  -> CreateRelease
  -> AnalyzeRisk
  -> WaitApproval
  -> DeployRelease
  -> VerifyRelease
  -> ObserveWindow
  -> Success or Fail
  -> If Fail: RollbackRelease
  -> End
```

## 3. 活动定义

### CreateRelease
- 输入：serviceId, environmentId, version, strategy, triggeredBy
- 输出：releaseId

### AnalyzeRisk
- 输入：releaseId
- 输出：riskLevel, riskReason, recommendedStrategy

### WaitApproval
- 输入：releaseId
- 输出：approved / rejected / expired

### DeployRelease
- 输入：releaseId
- 输出：deploy result

### VerifyRelease
- 输入：releaseId
- 输出：health result

### ObserveWindow
- 输入：releaseId
- 输出：stable / abnormal

### RollbackRelease
- 输入：releaseId, targetVersion
- 输出：rollback result

## 4. 幂等约束
- CreateRelease 必须幂等
- DeployRelease 必须幂等
- RollbackRelease 必须幂等

