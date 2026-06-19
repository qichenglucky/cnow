import React, { useState } from 'react';
import { Drawer, Input, Button, List, Typography, Space, Avatar, Spin } from 'antd';
import { RobotOutlined, UserOutlined, SendOutlined, LoadingOutlined } from '@ant-design/icons';
import { useAppStore } from '../store/app';
import { requestAIPlan, requestRiskAnalysis } from '../api/ai';

const { Paragraph } = Typography;
const { TextArea } = Input;

interface Message {
  role: 'user' | 'assistant';
  content: string;
}

const defaultGreeting = '我是即码 AI 助手，可以帮你分析服务状态、规划发布流程、评估风险。请问有什么可以帮助你的？';

function formatAIResponse(data: Record<string, unknown>): string {
  const lines: string[] = [];
  if (data.risk_level) lines.push(`风险等级: ${data.risk_level}`);
  if (data.risk_score !== undefined) lines.push(`风险评分: ${data.risk_score}`);
  if (data.confidence !== undefined) lines.push(`置信度: ${(data.confidence as number * 100).toFixed(0)}%`);
  if (data.summary) lines.push(`\n${data.summary}`);
  if (data.reason) lines.push(`\n原因: ${data.reason}`);
  if (data.steps && Array.isArray(data.steps)) {
    lines.push('\n执行步骤:');
    (data.steps as Array<Record<string, unknown>>).forEach((step) => {
      lines.push(`  ${step.order}. [${step.action}] ${step.description} (${step.estimated_duration})`);
    });
  }
  if (data.factors && Array.isArray(data.factors)) {
    lines.push(`\n风险因素: ${(data.factors as string[]).join(', ')}`);
  }
  if (data.recommendations && Array.isArray(data.recommendations)) {
    lines.push('\n建议:');
    (data.recommendations as string[]).forEach((r) => lines.push(`  - ${r}`));
  }
  return lines.join('\n') || JSON.stringify(data, null, 2);
}

const AIAssistantSidebar: React.FC = () => {
  const { aiDrawerOpen, toggleAiDrawer } = useAppStore();
  const [messages, setMessages] = useState<Message[]>([
    { role: 'assistant', content: defaultGreeting },
  ]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);

  const send = async () => {
    if (!input.trim() || loading) return;
    const userMsg: Message = { role: 'user', content: input };
    setMessages((prev) => [...prev, userMsg]);
    setInput('');
    setLoading(true);

    try {
      let responseText: string;
      const isRiskQuery = input.includes('风险') || input.includes('risk');
      if (isRiskQuery) {
        const data = await requestRiskAnalysis({ service_id: '', environment: 'production', version: '', prompt: input });
        responseText = formatAIResponse(data as unknown as Record<string, unknown>);
      } else {
        const data = await requestAIPlan(input);
        responseText = formatAIResponse(data as unknown as Record<string, unknown>);
      }
      setMessages((prev) => [...prev, { role: 'assistant', content: responseText }]);
    } catch {
      // Fallback to local mock on API failure
      const fallbackMap: Record<string, string> = {
        '发布': '根据当前服务状态，我建议使用金丝雀发布策略。步骤如下：\n1. 先在灰度环境部署 10% 流量\n2. 监控 5 分钟，确认无异常\n3. 逐步扩大到 50%、100%\n4. 全量发布后观察 30 分钟',
        '风险': '基于当前发布配置的风险评估：\n- 风险等级：中等\n- 主要风险：新版本未经过充分的压力测试\n- 建议：先在测试环境运行基准测试，确认性能指标正常后再发布到生产环境',
        '回滚': '检测到当前版本存在异常，建议回滚操作：\n1. 立即停止灰度发布\n2. 将流量切回上一个稳定版本\n3. 通知相关开发人员排查问题\n预计回滚时间：2-3 分钟',
      };
      const key = Object.keys(fallbackMap).find((k) => input.includes(k));
      const fallback = key ? fallbackMap[key] : '抱歉，AI 服务暂时不可用，请稍后再试。';
      setMessages((prev) => [...prev, { role: 'assistant', content: fallback }]);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Drawer
      title="AI 助手"
      placement="right"
      width={420}
      open={aiDrawerOpen}
      onClose={toggleAiDrawer}
      styles={{ body: { padding: 0, display: 'flex', flexDirection: 'column', height: 'calc(100vh - 55px)' } }}
    >
      <div style={{ flex: 1, overflow: 'auto', padding: 16 }}>
        <List
          dataSource={messages}
          renderItem={(msg) => (
            <List.Item style={{ border: 'none', padding: '8px 0' }}>
              <Space align="start">
                <Avatar icon={msg.role === 'assistant' ? <RobotOutlined /> : <UserOutlined />}
                  style={{ backgroundColor: msg.role === 'assistant' ? '#1677ff' : '#87d068' }} />
                <div style={{ background: msg.role === 'assistant' ? '#f0f5ff' : '#f6ffed', padding: '8px 12px', borderRadius: 8, maxWidth: 300 }}>
                  <Paragraph style={{ margin: 0, whiteSpace: 'pre-wrap' }}>{msg.content}</Paragraph>
                </div>
              </Space>
            </List.Item>
          )}
        />
        {loading && (
          <div style={{ textAlign: 'center', padding: 16 }}>
            <Spin indicator={<LoadingOutlined style={{ fontSize: 20 }} spin />} />
          </div>
        )}
      </div>
      <div style={{ padding: 16, borderTop: '1px solid #f0f0f0', display: 'flex', gap: 8 }}>
        <TextArea
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="输入问题，例如：如何发布这个服务？"
          autoSize={{ minRows: 1, maxRows: 3 }}
          onPressEnter={(e) => { if (!e.shiftKey) { e.preventDefault(); send(); } }}
          disabled={loading}
        />
        <Button type="primary" icon={<SendOutlined />} onClick={send} loading={loading} />
      </div>
    </Drawer>
  );
};

export default AIAssistantSidebar;
