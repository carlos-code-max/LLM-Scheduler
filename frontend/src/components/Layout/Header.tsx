import React, { useState, useEffect } from 'react';
import { Layout, Badge, Tooltip, Typography, Space } from 'antd';
import { BellOutlined, ApiOutlined } from '@ant-design/icons';
import { systemApi } from '../../services/api';
import { HealthStatus } from '../../types';

const { Header: AntHeader } = Layout;
const { Text } = Typography;

const Header: React.FC = () => {
  const [healthStatus, setHealthStatus] = useState<HealthStatus | null>(null);
  const [loading, setLoading] = useState(false);

  // 获取系统健康状态
  const fetchHealthStatus = async () => {
    try {
      setLoading(true);
      const response = await systemApi.health();
      if (response.code === 0) {
        setHealthStatus(response.data!);
      }
    } catch (error) {
      console.error('Failed to fetch health status:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchHealthStatus();
    // 定期检查健康状态（每30秒）
    const interval = setInterval(fetchHealthStatus, 30000);
    return () => clearInterval(interval);
  }, []);

  // 获取健康状态颜色
  const getHealthStatusColor = () => {
    if (!healthStatus) return 'default';
    return healthStatus.status === 'ok' ? 'success' : 'error';
  };

  // 获取健康状态文本
  const getHealthStatusText = () => {
    if (loading) return '检查中...';
    if (!healthStatus) return '未知';
    return healthStatus.status === 'ok' ? '正常' : '异常';
  };

  // 获取健康状态详情
  const getHealthStatusDetail = () => {
    if (!healthStatus) return '系统状态未知';
    
    const details = [];
    if (healthStatus.database !== 'ok') {
      details.push(`数据库: ${healthStatus.database_error || '异常'}`);
    }
    if (healthStatus.redis !== 'ok') {
      details.push(`缓存: ${healthStatus.redis_error || '异常'}`);
    }
    if (healthStatus.queue !== 'ok') {
      details.push(`队列: ${healthStatus.queue_error || '异常'}`);
    }
    
    if (details.length === 0) {
      return '所有服务运行正常';
    }
    
    return details.join('\n');
  };

  return (
    <AntHeader
      style={{
        background: '#fff',
        padding: '0 24px',
        borderBottom: '1px solid #f0f0f0',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        position: 'sticky',
        top: 0,
        zIndex: 50,
      }}
    >
      <div className="header-title">
        <Text strong style={{ fontSize: 16 }}>
          LLM 调度管理平台
        </Text>
      </div>
      
      <div className="header-actions">
        <Space size="large">
          {/* 系统状态指示器 */}
          <Tooltip title={getHealthStatusDetail()} placement="bottomRight">
            <Space>
              <ApiOutlined style={{ fontSize: 16 }} />
              <Badge status={getHealthStatusColor()} text={getHealthStatusText()} />
            </Space>
          </Tooltip>

          {/* 通知图标（预留） */}
          <Tooltip title="通知">
            <Badge count={0} size="small">
              <BellOutlined style={{ fontSize: 16 }} />
            </Badge>
          </Tooltip>

          {/* 系统信息 */}
          <Text type="secondary" style={{ fontSize: 12 }}>
            v1.0.0
          </Text>
        </Space>
      </div>
    </AntHeader>
  );
};

export default Header;
