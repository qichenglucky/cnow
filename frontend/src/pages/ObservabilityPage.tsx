import React from 'react';
import { Tabs, Table, Tag, Spin } from 'antd';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { useObsLogs, useObsMetrics, useObsAlerts } from '../hooks/useObservability';

const ObservabilityPage: React.FC = () => {
  const { data: logsData, isLoading: logsLoading } = useObsLogs();
  const { data: metricsData, isLoading: metricsLoading } = useObsMetrics();
  const { data: alertsData, isLoading: alertsLoading } = useObsAlerts();

  const logColumns = [
    { title: '时间', dataIndex: 'timestamp', key: 'timestamp', render: (t: string) => t.replace('T', ' ').replace('Z', '') },
    { title: '级别', dataIndex: 'level', key: 'level', render: (l: string) => {
      const colorMap: Record<string, string> = { INFO: 'blue', WARN: 'orange', ERROR: 'red', DEBUG: 'default' };
      const labelMap: Record<string, string> = { INFO: '信息', WARN: '警告', ERROR: '错误', DEBUG: '调试' };
      return <Tag color={colorMap[l] ?? 'default'}>{labelMap[l] ?? l}</Tag>;
    }},
    { title: '服务', dataIndex: 'service', key: 'service' },
    { title: '消息', dataIndex: 'message', key: 'message', ellipsis: true },
    { title: 'TraceId', dataIndex: 'traceId', key: 'traceId', render: (t: string) => <Tag>{t}</Tag> },
  ];

  const alertColumns = [
    { title: '级别', dataIndex: 'level', key: 'level', render: (l: string) => {
      const colorMap: Record<string, string> = { critical: 'red', warning: 'orange', info: 'blue', '严重': 'red', '警告': 'orange', '信息': 'blue' };
      return <Tag color={colorMap[l] ?? 'default'}>{l}</Tag>;
    }},
    { title: '服务', dataIndex: 'service', key: 'service' },
    { title: '消息', dataIndex: 'message', key: 'message' },
    { title: '时间', dataIndex: 'time', key: 'time' },
    { title: '状态', dataIndex: 'status', key: 'status', render: (s: string) => {
      const colorMap: Record<string, string> = { firing: 'red', acknowledged: 'orange', resolved: 'green' };
      const labelMap: Record<string, string> = { firing: '告警中', acknowledged: '已确认', resolved: '已解决' };
      return <Tag color={colorMap[s] ?? 'default'}>{labelMap[s] ?? s}</Tag>;
    }},
  ];

  const latencyData = metricsData?.latency ?? [];
  const errorRateData = metricsData?.error_rate ?? [];

  return (
    <Tabs items={[
      {
        key: 'metrics',
        label: '指标',
        children: metricsLoading ? <Spin style={{ display: 'block', margin: '80px auto' }} /> : (
          <div style={{ display: 'flex', gap: 24, flexWrap: 'wrap' }}>
            <div style={{ flex: 1, minWidth: 400, height: 350 }}>
              <h3>请求延迟 (ms)</h3>
              <ResponsiveContainer width="100%" height={300}>
                <LineChart data={latencyData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="time" />
                  <YAxis />
                  <Tooltip />
                  <Legend />
                  <Line type="monotone" dataKey="p50" stroke="#1677ff" name="P50" />
                  <Line type="monotone" dataKey="p95" stroke="#faad14" name="P95" />
                  <Line type="monotone" dataKey="p99" stroke="#ff4d4f" name="P99" />
                </LineChart>
              </ResponsiveContainer>
            </div>
            <div style={{ flex: 1, minWidth: 400, height: 350 }}>
              <h3>错误率 (%)</h3>
              <ResponsiveContainer width="100%" height={300}>
                <LineChart data={errorRateData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="time" />
                  <YAxis />
                  <Tooltip />
                  <Legend />
                  <Line type="monotone" dataKey="rate" stroke="#ff4d4f" name="错误率 %" />
                  <Line type="monotone" dataKey="count" stroke="#52c41a" name="错误数" yAxisId={0} />
                </LineChart>
              </ResponsiveContainer>
            </div>
          </div>
        ),
      },
      {
        key: 'logs',
        label: '日志',
        children: <Table rowKey="id" columns={logColumns} dataSource={logsData?.items ?? []} loading={logsLoading} pagination={{ pageSize: 15 }} size="small" />,
      },
      {
        key: 'alerts',
        label: '告警',
        children: <Table rowKey="id" columns={alertColumns} dataSource={alertsData?.items ?? []} loading={alertsLoading} pagination={false} />,
      },
    ]} />
  );
};

export default ObservabilityPage;
