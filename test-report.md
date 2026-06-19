# CodeNow 后端 API 测试报告

## 1. 测试概览

| 项目 | 内容 |
|------|------|
| 测试日期 | 2026-06-19 |
| 测试环境 | macOS, Docker PostgreSQL 16 (cnow-postgres, port 5433) |
| 后端版本 | Go 1.22, localhost:8080 |
| 数据库 | PostgreSQL 16, database: cnow |
| 测试类型 | 黑盒测试 + 白盒测试 |
| 测试工具 | curl, docker exec psql, 源码审查 |

---

## 2. 黑盒测试用例

### 2.1 正向测试

| 编号 | 类别 | 描述 | 预期结果 | 实际结果 | 状态 |
|------|------|------|----------|----------|------|
| T01 | 正向 | GET /healthz 健康检查 | code=0, status=ok | code=0, status=ok | ✅ 通过 |
| T02 | 正向 | GET /readyz 就绪检查 | code=0, status=ready | code=0, status=ready | ✅ 通过 |
| T03 | 正向 | POST /api/v1/services 创建服务 | code=0, status=draft | code=0, id=42, status=draft | ✅ 通过 |
| T04 | 正向 | GET /api/v1/services/{id} 查询服务 | code=0, 返回服务详情 | code=0, 返回完整服务对象 | ✅ 通过 |
| T05 | 正向 | GET /api/v1/services 分页查询 | code=0, items+total | code=0, items数组+total=3 | ✅ 通过 |
| T06 | 正向 | POST /api/v1/environments 创建环境 | code=0, status=creating | code=0, id=4, status=creating | ✅ 通过 |
| T07 | 正向 | POST /api/v1/releases 创建发布 | code=0, status=created | code=0, id=2, status=created | ✅ 通过 |
| T08 | 正向 | GET /api/v1/releases/{id} 查询发布(含events) | code=0, release+events | code=0, release对象+events数组(1条release_created事件) | ✅ 通过 |
| T09 | 正向 | GET /api/v1/releases 发布列表 | code=0, items+total | code=0, items=[1条], total=1 | ✅ 通过 |

### 2.2 边界测试

| 编号 | 类别 | 描述 | 预期结果 | 实际结果 | 状态 |
|------|------|------|----------|----------|------|
| T10 | 边界 | 创建服务: 空name | code=2001 | code=2001, "name is required" | ✅ 通过 |
| T11 | 边界 | 创建服务: 超长name(129字符) | code=2001 或 DB截断 | code=4001, "value too long for type character varying(128)" | ⚠️ 建议改进 |
| T12 | 边界 | 创建服务: 特殊字符name | 应拒绝或接受 | code=0, 成功创建(name含!@#$%^&*()) | ⚠️ 建议改进 |
| T13 | 边界 | 创建服务: 重复name | code=1002 | code=1002, "service 'test-svc-001' already exists" | ✅ 通过 |
| T14 | 边界 | GET /api/v1/services/99999 不存在的服务 | code=1001 | code=1001, "service 99999 not found" | ✅ 通过 |
| T15 | 边界 | GET /api/v1/releases/99999 不存在的发布 | code=1001 | code=1001, "release 99999 not found" | ✅ 通过 |
| T16 | 边界 | 分页 offset=0&limit=2 | 返回2条记录 | 返回2条记录, total=3, limit=2 | ✅ 通过 |
| T17 | 边界 | 分页 offset=100&limit=10 (超出范围) | 返回空列表 | items=null, total=3, offset=100 | ✅ 通过 |
| T18 | 边界 | 创建发布: 不存在的serviceId | code=1001 | code=1001, "service 99999 not found" | ✅ 通过 |
| T19 | 边界 | 创建发布: 不存在的environmentId | code=1001 | code=1003, "service must be in 'ready' status" (先检查service状态) | ⚠️ 建议改进 |
| T20 | 边界 | 创建发布: 重复version | 应拒绝或允许 | code=0, 成功创建(无唯一约束) | ⚠️ 缺陷 |

### 2.3 异常测试

| 编号 | 类别 | 描述 | 预期结果 | 实际结果 | 状态 |
|------|------|------|----------|----------|------|
| T21 | 异常 | POST 无效JSON body | code=2001 | code=2001, "invalid JSON body" | ✅ 通过 |
| T22 | 异常 | POST 空body | code=2001 | code=2001, "invalid JSON body" | ✅ 通过 |
| T23 | 异常 | 缺少必填字段(name) | code=2001 | code=2001, "name is required" | ✅ 通过 |
| T24 | 异常 | Content-Type: text/plain | 应拒绝或接受 | code=0, 成功创建(未校验Content-Type) | ⚠️ 建议改进 |

---

## 3. 白盒测试 (代码审查)

| 编号 | 类别 | 描述 | 审查结果 | 状态 |
|------|------|------|----------|------|
| W01 | 错误处理 | 所有错误统一包装为AppError | ✅ errors.go定义了完整的AppError体系(1001-4001), 所有app层错误均通过errors.WithDetails/Wrap包装, writeAppError统一映射HTTP状态码 | ✅ 通过 |
| W02 | 事务一致性 | 创建服务时写入审计日志 | ❌ **严重缺陷**: audit_log表schema与代码INSERT语句不匹配。表列(actor_id:bigint, actor_role, detail:jsonb) vs 代码(actor_id:string, actor_type, before_state, after_state, ip_address)。审计日志写入静默失败。 | ❌ 不通过 |
| W03 | 状态机 | Service/Release状态转换完整性 | ✅ Service: 5个状态, 4条转换规则完整。Release: 11个状态, 8条转换规则完整。CanTransitionTo方法实现正确。终态(archived/succeeded/rolled_back)无出边, 符合预期。 | ✅ 通过 |
| W04 | SQL注入 | 是否使用参数化查询 | ✅ 所有SQL查询均使用$1,$2参数化占位符, 无字符串拼接SQL。pgx驱动原生支持参数化。 | ✅ 通过 |
| W05 | 并发安全 | 连接池配置 | ✅ pgxpool配置: MaxConns=20, MinConns=2, MaxConnLifetime=30min, MaxConnIdleTime=5min, HealthCheckPeriod=1min。配置合理。 | ✅ 通过 |
| W06 | 中间件链 | RequestID/CORS/Recovery | ✅ Chain顺序: CORS→Recovery→Logging→RequestID。RequestID支持X-Request-Id传播+UUID生成。Recovery捕获panic返回500。CORS支持OPTIONS预检。 | ✅ 通过 |
| W07 | 幂等性 | 幂等键机制 | ⚠️ idempotency.go已实现完整幂等键表和CheckAndReserve/Complete机制, 但**未在任何handler中集成使用**。CORS headers包含X-Idempotency-Key但未实际处理。 | ⚠️ 未集成 |

---

## 4. 缺陷列表

### 缺陷 #1 [严重] 审计日志写入失败 — schema不匹配

- **严重程度**: 🔴 严重
- **描述**: `audit_log`数据库表schema与`audit.go`中的INSERT语句完全不匹配，导致所有审计日志写入静默失败。
- **表实际schema**: `actor_id(bigint), actor_role(varchar), action, resource_type, resource_id(bigint), request_id, detail(jsonb), result, created_at`
- **代码INSERT**: `actor_id(string), actor_type, action, resource_type, resource_id(string), request_id, ip_address, before_state, after_state, created_at`
- **差异**: 6个列名/类型不匹配, 2个列不存在(ip_address, before_state/after_state), 2个列缺失(detail, result)
- **影响**: 所有操作审计日志丢失, 无法追踪操作历史, 合规风险
- **复现步骤**: 创建服务后查询 `SELECT * FROM audit_log` 返回0行
- **修复建议**: 统一schema, 建议以migration为准修改代码, 或反向修改

### 缺陷 #2 [中等] 发布版本无唯一约束

- **严重程度**: 🟡 中等
- **描述**: 同一service+environment下可以创建相同version的发布记录, 无数据库唯一约束
- **影响**: 可能导致版本混乱, 无法区分同版本的多次发布
- **复现步骤**: 对同一service/environment连续POST两次相同version的release, 均返回成功
- **修复建议**: 添加 `UNIQUE(service_id, environment_id, version)` 约束, 或在应用层校验

### 缺陷 #3 [低] 超长name返回500而非400

- **严重程度**: 🟢 低
- **描述**: 创建服务时name超过128字符, 返回code=4001(internal error)而非2001(invalid param)
- **影响**: 错误码不准确, 前端无法正确提示用户
- **复现步骤**: POST /api/v1/services with name=129个字符
- **修复建议**: 在CreateService中添加name长度校验 (`len(input.Name) > 128`)

### 缺陷 #4 [低] Content-Type未校验

- **严重程度**: 🟢 低
- **描述**: POST请求即使Content-Type为text/plain, 只要body是合法JSON就能成功处理
- **影响**: 不符合RESTful规范, 但不影响功能安全
- **修复建议**: 在handler中添加Content-Type检查中间件

### 缺陷 #5 [低] 特殊字符name未校验

- **严重程度**: 🟢 低
- **描述**: name字段允许包含`!@#$%^&*()`等特殊字符, 缺少格式校验
- **影响**: 可能导致下游系统问题(域名、URL等)
- **修复建议**: 添加name格式校验(建议仅允许 `[a-z0-9][a-z0-9-]*`)

### 缺陷 #6 [低] 幂等键机制未集成

- **严重程度**: 🟢 低
- **描述**: `idempotency.go`已实现完整的幂等键表和逻辑, 但未在任何HTTP handler中调用
- **影响**: 重复提交无法防护
- **修复建议**: 在POST handler中集成幂等键检查

### 缺陷 #7 [信息] 分页返回items=null而非空数组

- **严重程度**: ℹ️ 信息
- **描述**: 当分页查询无结果时, items字段返回null而非[]
- **影响**: 前端需额外处理null情况
- **复现步骤**: offset=100&limit=10 返回 items=null

---

## 5. 代码质量评估

### 5.1 架构设计 ⭐⭐⭐⭐⭐
- 清晰的分层架构: `cmd → platform/http → app → domain → repo`
- 依赖注入模式, 接口驱动(Repository interface)
- 统一的错误体系(AppError)和响应格式(Envelope)

### 5.2 代码规范 ⭐⭐⭐⭐
- Go idioms 基本遵循
- 使用结构化日志(zap)
- 使用pgx连接池, 配置合理
- 优雅关闭(graceful shutdown)已实现

### 5.3 安全性 ⭐⭐⭐⭐
- SQL参数化查询, 无注入风险
- Recovery中间件防panic
- RequestID传播便于链路追踪
- ⚠️ 缺少认证/授权中间件(当前为dev阶段可接受)

### 5.4 可靠性 ⭐⭐⭐
- ✅ 连接池健康检查
- ✅ 状态机转换校验
- ❌ 审计日志静默失败(schema不匹配)
- ⚠️ 事件创建错误被忽略(`_ = a.events.Create(...)`)

### 5.5 测试覆盖 ⭐⭐⭐⭐
- 25个单元/集成测试已通过
- 领域模型测试(models_test.go)覆盖状态机
- 错误处理测试(errors_test.go)覆盖AppError

---

## 6. 测试结论

### 总体评定: ⚠️ 有条件通过

**通过率**: 黑盒 16/24 (67%), 白盒 4/7 (57%)

### 核心功能状态
- ✅ 健康检查: 正常
- ✅ 服务CRUD: 基本正常
- ✅ 环境创建: 正常
- ✅ 发布创建/查询: 正常
- ✅ 分页查询: 正常
- ✅ 错误处理: 基本正常
- ❌ 审计日志: **完全失效**

### 必须修复 (发布前)
1. **审计日志schema不匹配** — 所有审计记录丢失, 合规风险
2. **发布版本唯一约束** — 数据一致性风险

### 建议修复 (可延后)
3. name格式校验(长度+特殊字符)
4. Content-Type校验
5. 幂等键集成
6. 分页空结果返回[]
7. 事件创建错误处理

### 风险项
| 风险 | 等级 | 说明 |
|------|------|------|
| 审计日志丢失 | 高 | 无法追踪操作历史, 不满足合规要求 |
| 重复发布版本 | 中 | 可能导致部署混乱 |
| 无认证机制 | 低(当前阶段) | dev环境可接受, 生产前必须添加 |
