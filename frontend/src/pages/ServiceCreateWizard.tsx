import React, { useState } from 'react';
import { Steps, Form, Input, Select, Button, Card, message, Space } from 'antd';
import { useNavigate } from 'react-router-dom';
import { useCreateService } from '../hooks/useServices';
import type { CreateServiceInput } from '../types/api';

const { TextArea } = Input;

const steps = ['基本信息', '仓库配置', '确认提交'];

const ServiceCreateWizard: React.FC = () => {
  const [current, setCurrent] = useState(0);
  const [form] = Form.useForm<CreateServiceInput>();
  const navigate = useNavigate();
  const createMutation = useCreateService();

  const next = async () => {
    try {
      if (current === 0) await form.validateFields(['name', 'description', 'owner', 'tech_stack']);
      if (current === 1) await form.validateFields(['repo_url', 'branch', 'language', 'framework']);
      setCurrent(current + 1);
    } catch { /* validation failed */ }
  };

  const submit = async () => {
    try {
      const values = form.getFieldsValue(true);
      await createMutation.mutateAsync(values);
      message.success('服务创建成功');
      navigate('/services');
    } catch { /* handled by interceptor */ }
  };

  const formItems: Record<number, React.ReactNode> = {
    0: (
      <>
        <Form.Item name="name" label="服务名称" rules={[{ required: true, message: '请输入服务名称' }]}>
          <Input placeholder="例如: 用户中心" />
        </Form.Item>
        <Form.Item name="description" label="服务描述" rules={[{ required: true, message: '请输入描述' }]}>
          <TextArea rows={3} placeholder="简要描述服务功能" />
        </Form.Item>
        <Form.Item name="owner" label="负责人" rules={[{ required: true, message: '请输入负责人' }]}>
          <Input placeholder="负责人姓名" />
        </Form.Item>
        <Form.Item name="tech_stack" label="技术栈" rules={[{ required: true, message: '请选择技术栈' }]}>
          <Select placeholder="选择技术栈" options={[
            { value: 'Go + gRPC', label: 'Go + gRPC' },
            { value: 'Java + Spring', label: 'Java + Spring' },
            { value: 'Node.js + Nest', label: 'Node.js + NestJS' },
            { value: 'Python + FastAPI', label: 'Python + FastAPI' },
            { value: 'React + Next.js', label: 'React + Next.js' },
          ]} />
        </Form.Item>
      </>
    ),
    1: (
      <>
        <Form.Item name="repo_url" label="仓库地址" rules={[{ required: true, message: '请输入仓库地址' }]}>
          <Input placeholder="https://github.com/org/repo" />
        </Form.Item>
        <Form.Item name="branch" label="默认分支" initialValue="main" rules={[{ required: true }]}>
          <Input />
        </Form.Item>
        <Form.Item name="language" label="编程语言" rules={[{ required: true }]}>
          <Select options={[
            { value: 'Go', label: 'Go' },
            { value: 'Java', label: 'Java' },
            { value: 'TypeScript', label: 'TypeScript' },
            { value: 'Python', label: 'Python' },
            { value: 'Rust', label: 'Rust' },
          ]} />
        </Form.Item>
        <Form.Item name="framework" label="框架" rules={[{ required: true }]}>
          <Input placeholder="例如: Spring Boot" />
        </Form.Item>
      </>
    ),
    2: null,
  };

  const values = Form.useWatch([], form);

  return (
    <Card title="创建服务" style={{ maxWidth: 700, margin: '0 auto' }}>
      <Steps current={current} onChange={(v) => { if (v < current) setCurrent(v); }} items={steps.map((t) => ({ title: t }))} style={{ marginBottom: 24 }} />
      <Form form={form} layout="vertical">
        {formItems[current]}
        {current === 2 && (
          <Card type="inner" title="确认信息">
            <p><strong>名称：</strong>{values?.name}</p>
            <p><strong>描述：</strong>{values?.description}</p>
            <p><strong>负责人：</strong>{values?.owner}</p>
            <p><strong>技术栈：</strong>{values?.tech_stack}</p>
            <p><strong>仓库：</strong>{values?.repo_url}</p>
            <p><strong>分支：</strong>{values?.branch}</p>
            <p><strong>语言：</strong>{values?.language}</p>
            <p><strong>框架：</strong>{values?.framework}</p>
          </Card>
        )}
      </Form>
      <div style={{ marginTop: 24, display: 'flex', justifyContent: 'flex-end' }}>
        <Space>
          {current > 0 && <Button onClick={() => setCurrent(current - 1)}>上一步</Button>}
          {current < 2 && <Button type="primary" onClick={next}>下一步</Button>}
          {current === 2 && <Button type="primary" onClick={submit} loading={createMutation.isPending}>提交</Button>}
          <Button onClick={() => navigate('/services')}>取消</Button>
        </Space>
      </div>
    </Card>
  );
};

export default ServiceCreateWizard;
