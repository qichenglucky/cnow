import React from 'react';
import { Layout as AntLayout, Menu, Breadcrumb, theme } from 'antd';
import {
  AppstoreOutlined,
  CloudUploadOutlined,
  LineChartOutlined,
  RobotOutlined,
} from '@ant-design/icons';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { useAppStore } from '../store/app';
import AIAssistantSidebar from '../pages/AIAssistantSidebar';

const { Header, Sider, Content } = AntLayout;

const menuItems = [
  { key: '/services', icon: <AppstoreOutlined />, label: '服务管理' },
  { key: '/releases', icon: <CloudUploadOutlined />, label: '发布中心' },
  { key: '/observability', icon: <LineChartOutlined />, label: '可观测' },
  { key: 'ai', icon: <RobotOutlined />, label: 'AI 助手' },
];

const breadcrumbMap: Record<string, string> = {
  '/services': '服务管理',
  '/services/create': '创建服务',
  '/releases': '发布中心',
  '/observability': '可观测',
};

const Layout: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { sidebarCollapsed, toggleSidebar, toggleAiDrawer } = useAppStore();
  const { token: { colorBgContainer, borderRadiusLG } } = theme.useToken();

  const selectedKey = location.pathname.startsWith('/services')
    ? '/services'
    : location.pathname.startsWith('/releases')
    ? '/releases'
    : location.pathname.startsWith('/observability')
    ? '/observability'
    : '';

  const breadcrumbItems = [{ title: '首页' }];
  const parts = location.pathname.split('/').filter(Boolean);
  let path = '';
  for (const part of parts) {
    path += `/${part}`;
    const label = breadcrumbMap[path] ?? (part.length === 36 ? '详情' : part);
    breadcrumbItems.push({ title: label });
  }

  return (
    <AntLayout style={{ minHeight: '100vh' }}>
      <Sider
        collapsible
        collapsed={sidebarCollapsed}
        onCollapse={toggleSidebar}
        theme="light"
        style={{ borderRight: '1px solid #f0f0f0' }}
      >
        <div style={{ height: 48, display: 'flex', alignItems: 'center', justifyContent: 'center', fontWeight: 700, fontSize: sidebarCollapsed ? 14 : 18, color: '#1677ff' }}>
          {sidebarCollapsed ? '即码' : '即码 CodeNow.ai'}
        </div>
        <Menu
          mode="inline"
          selectedKeys={[selectedKey]}
          items={menuItems}
          onClick={({ key }) => {
            if (key === 'ai') {
              toggleAiDrawer();
            } else {
              navigate(key);
            }
          }}
        />
      </Sider>
      <AntLayout>
        <Header style={{ background: colorBgContainer, padding: '0 24px', display: 'flex', alignItems: 'center', borderBottom: '1px solid #f0f0f0' }}>
          <Breadcrumb items={breadcrumbItems} />
        </Header>
        <Content style={{ margin: 24, padding: 24, background: colorBgContainer, borderRadius: borderRadiusLG, minHeight: 280 }}>
          <Outlet />
        </Content>
      </AntLayout>
      <AIAssistantSidebar />
    </AntLayout>
  );
};

export default Layout;
