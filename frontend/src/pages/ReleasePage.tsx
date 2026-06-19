import React, { useState } from 'react';
import { Table, Select, Button, Modal, Form, Input, Timeline, Space, Tag, Empty, Skeleton } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import { useReleases, useCreateRelease } from '../hooks/useReleases';
import { mockServices } from '../mock/data';
import StatusBadge from '../components/StatusBadge';

const statusOptions = [
  { value: '', label: '全部状态' },
  { value: 'created', label: '已创建' },
  { value: 'reviewing', label: '审查中' },
  { value: 'approved', label: '已审批' },
  { value: 'deploying', label: '部署中' },
  { value: 'succeeded', label: '成功' },
  { value: 'failed', label: '失败' },
];

const strategyLabels: Record<string, string> = { direct: '直接发布', canary: '金丝雀', blue_green: '蓝绿' };

const ReleasePage: React.FC = () => {
  const [statusFilter, setStatusFilter] = useState('');
  const [serviceFilter, setServiceFilter] = useState('');
  const [page, setPage] = useState(1);
  const [modalOpen, setModalOpen] = useState(false);
  const [detailRelease, setDetailRelease] = useState<string | null>(null);
  const [form] = Form.useForm();
  const createMutation = useCreateRelease();

  const { data, isLoading } = useReleases({
    service_id: serviceFilter || undefined,
    status: statusFilter || undefined,
    offset: (page - 1) * 10,
    limit: 10,
  });

  const columns = [
    { title: '版本', dataIndex: 'version', key: 'version', render: (v: string, r: { id: string }) => <a onClick={() => setDetailRelease(r.id)}>{v}</a> },
    { title: '服务', dataIndex: 'service_name', key: 'service_name' },
    { title: '状态', dataIndex: 'status', key: 'status', render: (s: string) => <StatusBadge status={s} /> },
    { title: '策略', dataIndex: 'strategy', key: 'strategy', render: (s: string) => <Tag>{strategyLabels[s] ?? s}</Tag> },
    { title: '环境', dataIndex: 'environment', key: 'environment' },
    { title: '创建人', dataIndex: 'created_by', key: 'created_by' },
    { title: '创建时间', dataIndex: 'created_at', key: 'created_at', render: (t: string) => dayjs(t).format('YYYY-MM-DD HH:mm') },
  ];

  const selectedRelease = data?.items.find((r) => r.id === detailRelease);

  return (
    <div>
      <div style={{ display: 'flex', gap: 12, marginBottom: 16 }}>
        <Select value={statusFilter} onChange={setStatusFilter} options={statusOptions} style={{ width: 150 }} />
        <Select value={serviceFilter} onChange={setServiceFilter} style={{ width: 180 }} placeholder="选择服务" allowClear
          options={mockServices.map((s) => ({ value: s.id, label: s.name }))} />
        <div style={{ flex: 1 }} />
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>创建发布</Button>
      </div>

      {isLoading ? (
        <Skeleton active paragraph={{ rows: 8 }} />
      ) : (
        <Table rowKey="id" columns={columns} dataSource={data?.items ?? []}
          locale={{ emptyText: <Empty description="暂无发布记录"><Button type="primary" onClick={() => setModalOpen(true)}>创建发布</Button></Empty> }}
          pagination={{ current: page, pageSize: 10, total: data?.total ?? 0, onChange: setPage, showTotal: (t) => `共 ${t} 条` }} />
      )}

      {/* Create release modal */}
      <Modal title="创建发布" open={modalOpen} onCancel={() => setModalOpen(false)}
        onOk={() => { form.validateFields().then((v) => { createMutation.mutate(v, { onSuccess: () => setModalOpen(false) }); }); }}
        confirmLoading={createMutation.isPending}>
        <Form form={form} layout="vertical">
          <Form.Item name="service_id" label="服务" rules={[{ required: true }]}>
            <Select options={mockServices.map((s) => ({ value: s.id, label: s.name }))} />
          </Form.Item>
          <Form.Item name="environment" label="环境" rules={[{ required: true }]}>
            <Select options={['开发环境', '测试环境', '预发布', '灰度环境', '生产环境'].map((e) => ({ value: e, label: e }))} />
          </Form.Item>
          <Form.Item name="strategy" label="策略" initialValue="direct" rules={[{ required: true }]}>
            <Select options={[{ value: 'direct', label: '直接发布' }, { value: 'canary', label: '金丝雀' }, { value: 'blue_green', label: '蓝绿' }]} />
          </Form.Item>
          <Form.Item name="version" label="版本号" rules={[{ required: true }]}>
            <Input placeholder="v1.0.0" />
          </Form.Item>
          <Form.Item name="commit_sha" label="Commit SHA" rules={[{ required: true }]}>
            <Input placeholder="abc123..." />
          </Form.Item>
          <Form.Item name="commit_message" label="提交信息" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="image_tag" label="镜像标签" rules={[{ required: true }]}>
            <Input placeholder="registry/repo:tag" />
          </Form.Item>
        </Form>
      </Modal>

      {/* Release detail modal */}
      <Modal title={`发布详情 - ${selectedRelease?.version ?? ''}`} open={!!detailRelease} onCancel={() => setDetailRelease(null)} footer={null} width={600}>
        {selectedRelease && (
          <>
            <Space style={{ marginBottom: 16 }}>
              <StatusBadge status={selectedRelease.status} />
              <span>策略: {strategyLabels[selectedRelease.strategy]}</span>
              <span>环境: {selectedRelease.environment}</span>
            </Space>
            <Timeline items={(selectedRelease.events ?? []).map((evt) => ({
              children: `${dayjs(evt.created_at).format('HH:mm:ss')} - ${evt.message} (${evt.actor})`,
            }))} />
          </>
        )}
      </Modal>
    </div>
  );
};

export default ReleasePage;
