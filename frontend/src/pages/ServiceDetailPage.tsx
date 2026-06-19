import React, { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, Descriptions, Tabs, Table, Button, Space, Popconfirm, Modal, Form, Input, Select, Empty, Skeleton, message } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import { useService, useEnvironments, useUpdateService, useDeleteService } from '../hooks/useServices';
import { useReleases } from '../hooks/useReleases';
import StatusBadge from '../components/StatusBadge';

const { TextArea } = Input;

const ServiceDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data: svc, isLoading } = useService(id!);
  const { data: envs } = useEnvironments(id!);
  const { data: releases } = useReleases({ service_id: id, limit: 50 });
  const [editOpen, setEditOpen] = useState(false);
  const [editForm] = Form.useForm();
  const updateMutation = useUpdateService(id!);
  const deleteMutation = useDeleteService();

  const handleEdit = () => {
    editForm.setFieldsValue({
      name: svc?.name,
      description: svc?.description,
      owner: svc?.owner,
      tech_stack: svc?.tech_stack,
    });
    setEditOpen(true);
  };

  const handleDelete = async () => {
    try {
      await deleteMutation.mutateAsync(id!);
      message.success('服务已删除');
      navigate('/services');
    } catch { /* handled by interceptor */ }
  };

  const envColumns = [
    { title: '环境', dataIndex: 'name', key: 'name' },
    { title: '集群', dataIndex: 'cluster', key: 'cluster' },
    { title: '命名空间', dataIndex: 'namespace', key: 'namespace' },
    { title: '副本数', dataIndex: 'replicas', key: 'replicas' },
    { title: '状态', dataIndex: 'status', key: 'status', render: (s: string) => <StatusBadge status={s} /> },
  ];

  const releaseColumns = [
    { title: '版本', dataIndex: 'version', key: 'version' },
    { title: '状态', dataIndex: 'status', key: 'status', render: (s: string) => <StatusBadge status={s} /> },
    { title: '策略', dataIndex: 'strategy', key: 'strategy' },
    { title: '环境', dataIndex: 'environment', key: 'environment' },
    { title: '创建人', dataIndex: 'created_by', key: 'created_by' },
    { title: '创建时间', dataIndex: 'created_at', key: 'created_at', render: (t: string) => dayjs(t).format('YYYY-MM-DD HH:mm') },
  ];

  if (isLoading) {
    return <Skeleton active paragraph={{ rows: 10 }} />;
  }

  return (
    <div>
      <Card title={svc?.name ?? '加载中...'} extra={
        <Space>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/releases')}>创建发布</Button>
          <Button icon={<EditOutlined />} onClick={handleEdit}>编辑</Button>
          <Popconfirm title="确认删除此服务？" onConfirm={handleDelete} okText="确认" cancelText="取消">
            <Button danger icon={<DeleteOutlined />}>删除</Button>
          </Popconfirm>
        </Space>
      }>
        <Descriptions column={2} bordered size="small">
          <Descriptions.Item label="状态"><StatusBadge status={svc?.status ?? ''} /></Descriptions.Item>
          <Descriptions.Item label="负责人">{svc?.owner}</Descriptions.Item>
          <Descriptions.Item label="技术栈">{svc?.tech_stack}</Descriptions.Item>
          <Descriptions.Item label="仓库地址"><a href={svc?.repo_url} target="_blank" rel="noreferrer">{svc?.repo_url}</a></Descriptions.Item>
          <Descriptions.Item label="描述" span={2}>{svc?.description}</Descriptions.Item>
          <Descriptions.Item label="创建时间">{dayjs(svc?.created_at).format('YYYY-MM-DD HH:mm:ss')}</Descriptions.Item>
          <Descriptions.Item label="更新时间">{dayjs(svc?.updated_at).format('YYYY-MM-DD HH:mm:ss')}</Descriptions.Item>
        </Descriptions>
      </Card>

      <Card style={{ marginTop: 16 }}>
        <Tabs items={[
          { key: 'envs', label: '环境列表', children: <Table rowKey="id" columns={envColumns} dataSource={envs ?? []} pagination={false} /> },
          { key: 'releases', label: '发布历史', children: <Table rowKey="id" columns={releaseColumns} dataSource={releases?.items ?? []} pagination={{ pageSize: 10 }} /> },
          { key: 'settings', label: '设置', children: <Empty description="服务设置功能开发中" /> },
        ]} />
      </Card>

      <Modal title="编辑服务" open={editOpen} onCancel={() => setEditOpen(false)}
        onOk={() => { editForm.validateFields().then((v) => { updateMutation.mutate(v, { onSuccess: () => { setEditOpen(false); message.success('更新成功'); } }); }); }}
        confirmLoading={updateMutation.isPending}>
        <Form form={editForm} layout="vertical">
          <Form.Item name="name" label="服务名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="description" label="描述" rules={[{ required: true }]}>
            <TextArea rows={3} />
          </Form.Item>
          <Form.Item name="owner" label="负责人" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="tech_stack" label="技术栈">
            <Select options={[
              { value: 'Go + gRPC', label: 'Go + gRPC' },
              { value: 'Java + Spring', label: 'Java + Spring' },
              { value: 'Node.js + Nest', label: 'Node.js + NestJS' },
              { value: 'Python + FastAPI', label: 'Python + FastAPI' },
              { value: 'React + Next.js', label: 'React + Next.js' },
            ]} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default ServiceDetailPage;
