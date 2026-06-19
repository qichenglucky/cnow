import React, { useState } from 'react';
import { Table, Input, Button, Space, Empty, Skeleton } from 'antd';
import { PlusOutlined, SearchOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import dayjs from 'dayjs';
import { useServices } from '../hooks/useServices';
import StatusBadge from '../components/StatusBadge';

const { Search } = Input;

const ServiceListPage: React.FC = () => {
  const navigate = useNavigate();
  const [search, setSearch] = useState('');
  const [page, setPage] = useState(1);
  const pageSize = 10;
  const { data, isLoading } = useServices((page - 1) * pageSize, pageSize, search || undefined);

  const statusFilters = [
    { text: '草稿', value: 'draft' },
    { text: '创建中', value: 'creating' },
    { text: '就绪', value: 'ready' },
    { text: '降级', value: 'degraded' },
    { text: '已归档', value: 'archived' },
  ];

  const techStackFilters = [
    { text: 'Go + gRPC', value: 'Go + gRPC' },
    { text: 'Java + Spring', value: 'Java + Spring' },
    { text: 'Node.js + Nest', value: 'Node.js + Nest' },
    { text: 'Python + FastAPI', value: 'Python + FastAPI' },
    { text: 'React + Next.js', value: 'React + Next.js' },
  ];

  const columns = [
    {
      title: '名称', dataIndex: 'name', key: 'name',
      sorter: (a: { name: string }, b: { name: string }) => a.name.localeCompare(b.name),
      render: (text: string, record: { id: string }) => <a onClick={() => navigate(`/services/${record.id}`)}>{text}</a>,
    },
    {
      title: '状态', dataIndex: 'status', key: 'status',
      filters: statusFilters,
      onFilter: (value: unknown, record: { status: string }) => record.status === value,
      render: (s: string) => <StatusBadge status={s} />,
    },
    {
      title: '技术栈', dataIndex: 'tech_stack', key: 'tech_stack',
      filters: techStackFilters,
      onFilter: (value: unknown, record: { tech_stack: string }) => record.tech_stack === value,
    },
    { title: '负责人', dataIndex: 'owner', key: 'owner' },
    {
      title: '创建时间', dataIndex: 'created_at', key: 'created_at',
      sorter: (a: { created_at: string }, b: { created_at: string }) => dayjs(a.created_at).valueOf() - dayjs(b.created_at).valueOf(),
      render: (t: string) => dayjs(t).format('YYYY-MM-DD HH:mm'),
    },
  ];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Search placeholder="搜索服务名称" allowClear onSearch={(v) => { setSearch(v); setPage(1); }} style={{ width: 300 }} prefix={<SearchOutlined />} />
        <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/services/create')}>创建服务</Button>
      </div>
      {isLoading ? (
        <Skeleton active paragraph={{ rows: 8 }} />
      ) : (
        <Table
          rowKey="id"
          columns={columns}
          dataSource={data?.items ?? []}
          locale={{ emptyText: <Empty description="暂无服务"><Button type="primary" onClick={() => navigate('/services/create')}>创建第一个服务</Button></Empty> }}
          pagination={{
            current: page,
            pageSize,
            total: data?.total ?? 0,
            onChange: setPage,
            showTotal: (total) => `共 ${total} 个服务`,
          }}
          onRow={(record) => ({ onClick: () => navigate(`/services/${record.id}`), style: { cursor: 'pointer' } })}
        />
      )}
    </div>
  );
};

export default ServiceListPage;
