import { Tag } from 'antd';
import React from 'react';

const statusMap: Record<string, { color: string; label: string }> = {
  // Service statuses
  draft: { color: 'default', label: '草稿' },
  creating: { color: 'processing', label: '创建中' },
  ready: { color: 'success', label: '就绪' },
  degraded: { color: 'warning', label: '降级' },
  archived: { color: 'default', label: '已归档' },
  // Release statuses
  created: { color: 'default', label: '已创建' },
  reviewing: { color: 'processing', label: '审查中' },
  approved: { color: 'cyan', label: '已审批' },
  deploying: { color: 'processing', label: '部署中' },
  verifying: { color: 'processing', label: '验证中' },
  observing: { color: 'processing', label: '观察中' },
  succeeded: { color: 'success', label: '成功' },
  failed: { color: 'error', label: '失败' },
  rollback_pending: { color: 'warning', label: '待回滚' },
  rolling_back: { color: 'orange', label: '回滚中' },
  rolled_back: { color: 'default', label: '已回滚' },
};

interface Props {
  status: string;
}

const StatusBadge: React.FC<Props> = ({ status }) => {
  const info = statusMap[status] ?? { color: 'default', label: status };
  return <Tag color={info.color}>{info.label}</Tag>;
};

export default StatusBadge;
