import React from 'react';
import { Layout, Menu } from 'antd';
import { useNavigate, useLocation } from 'react-router-dom';
import {
  DashboardOutlined,
  UnorderedListOutlined,
  RobotOutlined,
  BarChartOutlined,
  SettingOutlined,
} from '@ant-design/icons';

const { Sider } = Layout;

const Sidebar: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();

  const menuItems = [
    {
      key: '/dashboard',
      icon: <DashboardOutlined />,
      label: 'Dashboard',
    },
    {
      key: '/tasks',
      icon: <UnorderedListOutlined />,
      label: '任务管理',
    },
    {
      key: '/models',
      icon: <RobotOutlined />,
      label: '模型管理',
    },
    {
      key: '/statistics',
      icon: <BarChartOutlined />,
      label: '统计分析',
    },
    {
      key: '/system',
      icon: <SettingOutlined />,
      label: '系统管理',
    },
  ];

  const handleMenuClick = (path: string) => {
    navigate(path);
  };

  // 获取当前选中的菜单项
  const getSelectedKey = () => {
    const path = location.pathname;
    if (path === '/' || path === '/dashboard') {
      return '/dashboard';
    }
    
    // 匹配一级路由
    const matched = menuItems.find(item => 
      path.startsWith(item.key) && item.key !== '/dashboard'
    );
    
    return matched ? matched.key : '/dashboard';
  };

  return (
    <Sider
      width={200}
      style={{
        overflow: 'auto',
        height: '100vh',
        position: 'fixed',
        left: 0,
        top: 0,
        zIndex: 100,
      }}
      theme="dark"
    >
      <div className="sidebar-logo">
        <RobotOutlined style={{ marginRight: 8, fontSize: 18 }} />
        LLM Scheduler
      </div>
      <Menu
        theme="dark"
        mode="inline"
        selectedKeys={[getSelectedKey()]}
        style={{ borderRight: 0 }}
        items={menuItems.map(item => ({
          key: item.key,
          icon: item.icon,
          label: item.label,
          onClick: () => handleMenuClick(item.key),
        }))}
      />
    </Sider>
  );
};

export default Sidebar;
