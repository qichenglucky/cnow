# 数据库表设计文档

## 1. 设计原则
- 主键统一自增 `bigint`
- 不预留多租户字段
- 主表存当前状态，事件表存过程
- 大文本和草案放对象存储或 JSONB
- 高风险动作必须有审计表

## 2. 核心表

### 2.1 `service`
主键：
- `id bigint auto_increment`

字段：
- `name`
- `display_name`
- `description`
- `owner_id`
- `team_id`
- `tech_stack`
- `status`
- `default_repo_id`
- `created_at`
- `updated_at`

### 2.2 `repo`
字段：
- `id`
- `service_id`
- `provider`
- `url`
- `default_branch`
- `visibility`
- `created_at`
- `updated_at`

### 2.3 `pipeline`
字段：
- `id`
- `service_id`
- `type`
- `name`
- `definition_ref`
- `status`
- `last_run_id`
- `created_at`
- `updated_at`

### 2.4 `build`
字段：
- `id`
- `pipeline_id`
- `service_id`
- `commit_sha`
- `branch`
- `triggered_by`
- `artifact_id`
- `status`
- `log_ref`
- `started_at`
- `finished_at`

### 2.5 `environment`
字段：
- `id`
- `service_id`
- `name`
- `type`
- `version`
- `status`
- `domain_id`
- `log_source_id`
- `metric_panel_id`
- `created_at`
- `updated_at`

### 2.6 `domain`
字段：
- `id`
- `service_id`
- `environment_id`
- `domain_name`
- `is_wildcard`
- `protocol`
- `certificate_id`
- `status`
- `created_at`
- `updated_at`

### 2.7 `certificate`
字段：
- `id`
- `domain_id`
- `provider`
- `status`
- `issued_at`
- `expires_at`
- `renewal_policy`
- `created_at`
- `updated_at`

### 2.8 `log_source`
字段：
- `id`
- `service_id`
- `environment_id`
- `provider`
- `project`
- `logstore`
- `status`
- `created_at`
- `updated_at`

### 2.9 `metric_panel`
字段：
- `id`
- `service_id`
- `environment_id`
- `provider`
- `dashboard_url`
- `status`
- `created_at`
- `updated_at`

### 2.10 `alert_rule`
字段：
- `id`
- `service_id`
- `environment_id`
- `name`
- `metric`
- `threshold`
- `severity`
- `enabled`
- `created_at`
- `updated_at`

### 2.11 `release`
字段：
- `id`
- `service_id`
- `environment_id`
- `version`
- `commit_sha`
- `image_tag`
- `strategy`
- `status`
- `triggered_by`
- `approved_by`
- `started_at`
- `finished_at`
- `summary`
- `risk_level`

### 2.12 `approval`
字段：
- `id`
- `release_id`
- `status`
- `approver_id`
- `reason`
- `created_at`
- `updated_at`

### 2.13 `rollback_record`
字段：
- `id`
- `release_id`
- `target_version`
- `triggered_by`
- `reason`
- `status`
- `created_at`
- `finished_at`

### 2.14 `incident`
字段：
- `id`
- `service_id`
- `release_id`
- `severity`
- `title`
- `summary`
- `status`
- `log_ref`
- `metric_ref`
- `created_at`
- `updated_at`

### 2.15 `release_event`
字段：
- `id`
- `release_id`
- `event_type`
- `status_before`
- `status_after`
- `payload`
- `created_by`
- `created_at`

### 2.16 `audit_log`
字段：
- `id`
- `actor_id`
- `actor_role`
- `action`
- `resource_type`
- `resource_id`
- `request_id`
- `detail`
- `result`
- `created_at`

### 2.17 `ai_run`
字段：
- `id`
- `service_id`
- `release_id`
- `run_type`
- `input_ref`
- `output_ref`
- `risk_level`
- `status`
- `created_by`
- `created_at`

