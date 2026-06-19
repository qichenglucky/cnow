import type { Service, Release, ReleaseEvent, Environment } from '../types/api';

// ---- 10 Mock Services ----

export const mockServices: Service[] = [
  { id: 'svc-001', name: '用户中心', description: '用户注册、登录、权限管理服务', status: 'ready', tech_stack: 'Go + gRPC', owner: '张伟', repo_url: 'https://github.com/codnow/user-center', branch: 'main', language: 'Go', framework: 'gRPC', created_at: '2026-01-15T08:00:00Z', updated_at: '2026-06-10T12:00:00Z' },
  { id: 'svc-002', name: '订单服务', description: '核心交易订单处理', status: 'ready', tech_stack: 'Java + Spring', owner: '李娜', repo_url: 'https://github.com/codnow/order-service', branch: 'main', language: 'Java', framework: 'Spring Boot', created_at: '2026-01-20T09:00:00Z', updated_at: '2026-06-08T15:30:00Z' },
  { id: 'svc-003', name: '支付网关', description: '聚合支付通道网关', status: 'degraded', tech_stack: 'Go + Gin', owner: '王强', repo_url: 'https://github.com/codnow/payment-gw', branch: 'main', language: 'Go', framework: 'Gin', created_at: '2026-02-01T10:00:00Z', updated_at: '2026-06-12T08:00:00Z' },
  { id: 'svc-004', name: '消息推送', description: '站内信、短信、邮件推送', status: 'ready', tech_stack: 'Node.js + Nest', owner: '赵敏', repo_url: 'https://github.com/codnow/msg-push', branch: 'develop', language: 'TypeScript', framework: 'NestJS', created_at: '2026-02-10T11:00:00Z', updated_at: '2026-06-05T09:00:00Z' },
  { id: 'svc-005', name: '数据中台', description: '数据采集、清洗、分析平台', status: 'ready', tech_stack: 'Python + FastAPI', owner: '孙浩', repo_url: 'https://github.com/codnow/data-platform', branch: 'main', language: 'Python', framework: 'FastAPI', created_at: '2026-02-20T08:30:00Z', updated_at: '2026-06-11T14:00:00Z' },
  { id: 'svc-006', name: '内容管理', description: 'CMS 内容发布系统', status: 'draft', tech_stack: 'React + Next.js', owner: '周莉', repo_url: 'https://github.com/codnow/cms', branch: 'main', language: 'TypeScript', framework: 'Next.js', created_at: '2026-03-01T09:00:00Z', updated_at: '2026-03-01T09:00:00Z' },
  { id: 'svc-007', name: '搜索服务', description: '全文搜索引擎', status: 'ready', tech_stack: 'Go + ES', owner: '张伟', repo_url: 'https://github.com/codnow/search-svc', branch: 'main', language: 'Go', framework: 'Elasticsearch', created_at: '2026-03-10T10:00:00Z', updated_at: '2026-06-09T16:00:00Z' },
  { id: 'svc-008', name: '库存服务', description: '商品库存管理', status: 'creating', tech_stack: 'Java + Spring', owner: '李娜', repo_url: 'https://github.com/codnow/inventory', branch: 'main', language: 'Java', framework: 'Spring Boot', created_at: '2026-03-15T08:00:00Z', updated_at: '2026-06-12T10:00:00Z' },
  { id: 'svc-009', name: '通知网关', description: '统一通知出口网关', status: 'archived', tech_stack: 'Rust + Axum', owner: '王强', repo_url: 'https://github.com/codnow/notify-gw', branch: 'main', language: 'Rust', framework: 'Axum', created_at: '2026-01-05T09:00:00Z', updated_at: '2026-04-01T12:00:00Z' },
  { id: 'svc-010', name: 'AI 推荐引擎', description: '个性化推荐服务', status: 'ready', tech_stack: 'Python + PyTorch', owner: '孙浩', repo_url: 'https://github.com/codnow/ai-rec', branch: 'main', language: 'Python', framework: 'PyTorch', created_at: '2026-04-01T08:00:00Z', updated_at: '2026-06-11T11:00:00Z' },
];

// ---- Mock environments: 5 per service ----

const envNames = ['开发环境', '测试环境', '预发布', '灰度环境', '生产环境'];
const envClusters = ['dev-cluster', 'test-cluster', 'staging-cluster', 'gray-cluster', 'prod-cluster'];

export const mockEnvironments: Environment[] = mockServices.flatMap((svc) =>
  envNames.map((name, i) => ({
    id: `env-${svc.id}-${i}`,
    service_id: svc.id,
    name,
    cluster: envClusters[i],
    namespace: `${svc.name.toLowerCase().replace(/\s/g, '-')}-${['dev', 'test', 'staging', 'gray', 'prod'][i]}`,
    replicas: [1, 2, 3, 2, 5][i],
    status: i <= 2 ? 'running' : i === 3 ? 'scaling' : 'running',
    created_at: svc.created_at,
  })),
);

// ---- 20 Mock Releases ----

const releaseStatuses: Array<{ status: Release['status']; message: string }> = [
  { status: 'created', message: '创建发布单' },
  { status: 'reviewing', message: '代码审查中' },
  { status: 'approved', message: '审批通过' },
  { status: 'deploying', message: '部署中' },
  { status: 'verifying', message: '验证中' },
  { status: 'observing', message: '观察中' },
  { status: 'succeeded', message: '发布成功' },
  { status: 'failed', message: '发布失败' },
  { status: 'succeeded', message: '发布成功' },
  { status: 'succeeded', message: '发布成功' },
  { status: 'rollback_pending', message: '等待回滚' },
  { status: 'succeeded', message: '发布成功' },
  { status: 'failed', message: '发布失败' },
  { status: 'succeeded', message: '发布成功' },
  { status: 'succeeded', message: '发布成功' },
  { status: 'succeeded', message: '发布成功' },
  { status: 'succeeded', message: '发布成功' },
  { status: 'succeeded', message: '发布成功' },
  { status: 'succeeded', message: '发布成功' },
  { status: 'succeeded', message: '发布成功' },
];

export const mockReleases: Release[] = Array.from({ length: 20 }, (_, i) => {
  const svcIdx = i % mockServices.length;
  const svc = mockServices[svcIdx];
  const rs = releaseStatuses[i];
  const envIdx = i % 5;
  const events: ReleaseEvent[] = [
    { id: `evt-${i}-0`, release_id: `rel-${String(i + 1).padStart(3, '0')}`, event_type: 'created', message: '发布单已创建', actor: svc.owner, created_at: `2026-06-${String(10 + (i % 10)).padStart(2, '0')}T08:00:00Z` },
    { id: `evt-${i}-1`, release_id: `rel-${String(i + 1).padStart(3, '0')}`, event_type: 'status_change', message: rs.message, actor: 'system', created_at: `2026-06-${String(10 + (i % 10)).padStart(2, '0')}T08:30:00Z` },
  ];
  return {
    id: `rel-${String(i + 1).padStart(3, '0')}`,
    service_id: svc.id,
    service_name: svc.name,
    version: `v1.${i + 1}.0`,
    status: rs.status,
    strategy: (['direct', 'canary', 'blue_green'] as const)[i % 3],
    environment: envNames[envIdx],
    commit_sha: `${String(i + 1).repeat(7)}abcdef0123456789`.slice(0, 40),
    commit_message: `feat: ${svc.name}功能优化 #${i + 1}`,
    image_tag: `registry.cnow.ai/${svc.name}:v1.${i + 1}.0`,
    created_by: svc.owner,
    events,
    created_at: `2026-06-${String(10 + (i % 10)).padStart(2, '0')}T08:00:00Z`,
    updated_at: `2026-06-${String(10 + (i % 10)).padStart(2, '0')}T09:00:00Z`,
  };
});

// ---- Mock metrics data ----

export const mockLatencyData = Array.from({ length: 24 }, (_, i) => ({
  time: `${String(i).padStart(2, '0')}:00`,
  p50: 20 + Math.random() * 30,
  p95: 80 + Math.random() * 60,
  p99: 150 + Math.random() * 100,
}));

export const mockErrorRateData = Array.from({ length: 24 }, (_, i) => ({
  time: `${String(i).padStart(2, '0')}:00`,
  rate: 0.1 + Math.random() * 0.8,
  count: Math.floor(Math.random() * 50),
}));

export const mockAlerts = [
  { id: 'alert-1', level: '严重', service: '支付网关', message: 'P99 延迟超过 500ms', time: '2026-06-12 08:15:00', status: '未处理' },
  { id: 'alert-2', level: '警告', service: '订单服务', message: '错误率 > 1%', time: '2026-06-12 07:30:00', status: '已确认' },
  { id: 'alert-3', level: '信息', service: '用户中心', message: 'CPU 使用率超过 80%', time: '2026-06-12 06:00:00', status: '已解决' },
  { id: 'alert-4', level: '严重', service: '搜索服务', message: 'Pod 重启次数过多', time: '2026-06-11 23:00:00', status: '未处理' },
  { id: 'alert-5', level: '警告', service: '数据中台', message: '磁盘使用率超过 90%', time: '2026-06-11 20:00:00', status: '已确认' },
];

export const mockLogs = Array.from({ length: 50 }, (_, i) => ({
  id: `log-${i}`,
  timestamp: `2026-06-12T${String(8 + Math.floor(i / 10)).padStart(2, '0')}:${String(i % 60).padStart(2, '0')}:00Z`,
  level: ['INFO', 'WARN', 'ERROR', 'DEBUG'][i % 4],
  service: mockServices[i % mockServices.length].name,
  message: [
    'Request processed successfully',
    'Database connection pool exhausted',
    'Cache miss for key user:12345',
    'Health check passed',
  ][i % 4],
  traceId: `trace-${String(i).padStart(6, '0')}`,
}));
