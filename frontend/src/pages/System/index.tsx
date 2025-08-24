import React, { useState, useEffect } from 'react';
import {
  Card,
  Descriptions,
  Alert,
  Button,
  Tag,
  Typography,
  Row,
  Col,
  Statistic,
  Table,
  Space,
  Spin,
} from 'antd';
import {
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  ReloadOutlined,
  DatabaseOutlined,
  CloudServerOutlined,
  SettingOutlined,
} from '@ant-design/icons';
import { systemApi, statsApi } from '../../services/api';
import { HealthStatus, SystemInfo, DashboardStats } from '../../types';

const { Title, Text } = Typography;

const System: React.FC = () => {
  const [healthStatus, setHealthStatus] = useState<HealthStatus | null>(null);
  const [systemInfo, setSystemInfo] = useState<SystemInfo | null>(null);
  const [dashboardStats, setDashboardStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshLoading, setRefreshLoading] = useState(false);

  // 获取系统健康状态
  const fetchHealthStatus = async () => {
    try {
      const response = await systemApi.health();
      if (response.code === 0) {
        setHealthStatus(response.data!);
      }
    } catch (error) {
      console.error('获取健康状态失败:', error);
    }
  };

  // 获取系统信息
  const fetchSystemInfo = async () => {
    try {
      const response = await systemApi.info();
      if (response.code === 0) {
        setSystemInfo(response.data!);
      }
    } catch (error) {
      console.error('获取系统信息失败:', error);
    }
  };

  // 获取Dashboard统计
  const fetchDashboardStats = async () => {
    try {
      const response = await statsApi.dashboard();
      if (response.code === 0) {
        setDashboardStats(response.data!);
      }
    } catch (error) {
      console.error('获取Dashboard统计失败:', error);
    }
  };

  // 初始化加载
  const loadAllData = async () => {
    try {
      setLoading(true);
      await Promise.all([
        fetchHealthStatus(),
        fetchSystemInfo(),
        fetchDashboardStats(),
      ]);
    } finally {
      setLoading(false);
    }
  };

  // 刷新数据
  const handleRefresh = async () => {
    try {
      setRefreshLoading(true);
      await loadAllData();
    } finally {
      setRefreshLoading(false);
    }
  };

  useEffect(() => {
    loadAllData();
    // 定期刷新健康状态
    const interval = setInterval(fetchHealthStatus, 30000);
    return () => clearInterval(interval);
  }, []);

  // 获取服务状态颜色和图标
  const getServiceStatus = (status: string, error?: string) => {
    if (status === 'ok') {
      return {
        color: 'success',
        icon: <CheckCircleOutlined style={{ color: '#52c41a' }} />,
        text: '正常',
      };
    } else {
      return {
        color: 'error',
        icon: <ExclamationCircleOutlined style={{ color: '#ff4d4f' }} />,
        text: error || '异常',
      };
    }
  };

  // Worker状态表格列
  const workerColumns = [
    {
      title: 'Worker ID',
      dataIndex: 'worker_id',
      key: 'worker_id',
      ellipsis: true,
    },
    {
      title: '模型',
      dataIndex: 'model_name',
      key: 'model_name',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'busy' ? 'processing' : 'success'}>
          {status === 'busy' ? '忙碌' : '空闲'}
        </Tag>
      ),
    },
    {
      title: '当前任务',
      dataIndex: 'current_task_id',
      key: 'current_task_id',
      render: (taskId?: number) => taskId || '-',
    },
    {
      title: '启动时间',
      dataIndex: 'start_time',
      key: 'start_time',
      render: (time: string) => new Date(time).toLocaleString(),
    },
    {
      title: '最后心跳',
      dataIndex: 'last_heartbeat',
      key: 'last_heartbeat',
      render: (time: string) => new Date(time).toLocaleString(),
    },
  ];

  if (loading) {
    return (
      <div className="loading-container">
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Title level={2}>系统管理</Title>
        <Button
          icon={<ReloadOutlined />}
          onClick={handleRefresh}
          loading={refreshLoading}
        >
          刷新
        </Button>
      </div>

      {/* 系统健康状态 */}
      <Card title="系统健康状态" style={{ marginBottom: 16 }}>
        {healthStatus ? (
          <Row gutter={[16, 16]}>
            <Col xs={24} sm={8}>
              <Card size="small">
                <Statistic
                  title="数据库"
                  value={getServiceStatus(healthStatus.database, healthStatus.database_error).text}
                  prefix={<DatabaseOutlined />}
                  valueStyle={{ color: getServiceStatus(healthStatus.database).color === 'success' ? '#52c41a' : '#ff4d4f' }}
                />
              </Card>
            </Col>
            <Col xs={24} sm={8}>
              <Card size="small">
                <Statistic
                  title="缓存服务"
                  value={getServiceStatus(healthStatus.redis, healthStatus.redis_error).text}
                  prefix={<CloudServerOutlined />}
                  valueStyle={{ color: getServiceStatus(healthStatus.redis).color === 'success' ? '#52c41a' : '#ff4d4f' }}
                />
              </Card>
            </Col>
            <Col xs={24} sm={8}>
              <Card size="small">
                <Statistic
                  title="队列服务"
                  value={getServiceStatus(healthStatus.queue, healthStatus.queue_error).text}
                  prefix={<SettingOutlined />}
                  valueStyle={{ color: getServiceStatus(healthStatus.queue).color === 'success' ? '#52c41a' : '#ff4d4f' }}
                />
              </Card>
            </Col>
          </Row>
        ) : (
          <Alert message="无法获取系统健康状态" type="warning" />
        )}
      </Card>

      {/* 系统信息 */}
      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={24} lg={12}>
          <Card title="系统信息" size="small">
            {systemInfo ? (
              <Descriptions column={1} size="small">
                <Descriptions.Item label="版本">{systemInfo.version}</Descriptions.Item>
                <Descriptions.Item label="环境">{systemInfo.environment}</Descriptions.Item>
                {systemInfo.database_stats && (
                  <Descriptions.Item label="数据库连接">
                    活跃: {systemInfo.database_stats.open_connections} / 
                    空闲: {systemInfo.database_stats.idle}
                  </Descriptions.Item>
                )}
                {systemInfo.queue_status && (
                  <Descriptions.Item label="队列长度">
                    总计: {systemInfo.queue_status.total_count}
                  </Descriptions.Item>
                )}
              </Descriptions>
            ) : (
              <Text type="secondary">暂无数据</Text>
            )}
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card title="队列状态" size="small">
            {dashboardStats?.queue_status ? (
              <Space direction="vertical" style={{ width: '100%' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span>高优先级队列</span>
                  <Tag color="error">{dashboardStats.queue_status.high_priority_count}</Tag>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span>中优先级队列</span>
                  <Tag color="warning">{dashboardStats.queue_status.medium_priority_count}</Tag>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span>低优先级队列</span>
                  <Tag color="default">{dashboardStats.queue_status.low_priority_count}</Tag>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span>处理中队列</span>
                  <Tag color="processing">{dashboardStats.queue_status.processing_count}</Tag>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span>延迟队列</span>
                  <Tag color="default">{dashboardStats.queue_status.delayed_count}</Tag>
                </div>
              </Space>
            ) : (
              <Text type="secondary">暂无数据</Text>
            )}
          </Card>
        </Col>
      </Row>

      {/* Worker 状态 */}
      <Card title="Worker 状态" size="small">
        {dashboardStats?.worker_status && dashboardStats.worker_status.length > 0 ? (
          <Table
            dataSource={dashboardStats.worker_status}
            columns={workerColumns}
            pagination={{
              pageSize: 10,
              showSizeChanger: true,
              showQuickJumper: true,
              showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条/共 ${total} 条`,
            }}
            size="small"
            scroll={{ x: 800 }}
            rowKey="worker_id"
          />
        ) : (
          <div style={{ textAlign: 'center', padding: '40px 0' }}>
            <Text type="secondary">当前没有活跃的 Worker</Text>
          </div>
        )}
      </Card>
    </div>
  );
};

export default System;
