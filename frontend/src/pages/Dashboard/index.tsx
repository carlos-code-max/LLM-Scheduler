import React, { useState, useEffect } from 'react';
import {
  Card,
  Row,
  Col,
  Statistic,
  Typography,
  Spin,
  Alert,
  Progress,
  Tag,
  Table,
  Space,
} from 'antd';
import {
  ClockCircleOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  PlayCircleOutlined,
  StopOutlined,
  RobotOutlined,
} from '@ant-design/icons';
import ReactECharts from 'echarts-for-react';
import { statsApi } from '../../services/api';
import { DashboardStats, Task } from '../../types';
import dayjs from 'dayjs';

const { Title } = Typography;

const Dashboard: React.FC = () => {
  const [dashboardData, setDashboardData] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // 获取 Dashboard 数据
  const fetchDashboardData = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await statsApi.dashboard();
      if (response.code === 0) {
        setDashboardData(response.data!);
      } else {
        setError(response.message);
      }
    } catch (err) {
      setError('获取 Dashboard 数据失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDashboardData();
    // 每30秒刷新一次数据
    const interval = setInterval(fetchDashboardData, 30000);
    return () => clearInterval(interval);
  }, []);

  // 任务状态标签映射
  const getStatusTag = (status: string) => {
    const statusMap = {
      pending: { color: 'default', icon: <ClockCircleOutlined />, text: '待处理' },
      running: { color: 'processing', icon: <PlayCircleOutlined />, text: '运行中' },
      completed: { color: 'success', icon: <CheckCircleOutlined />, text: '已完成' },
      failed: { color: 'error', icon: <ExclamationCircleOutlined />, text: '失败' },
      cancelled: { color: 'default', icon: <StopOutlined />, text: '已取消' },
    };
    
    const config = statusMap[status as keyof typeof statusMap] || statusMap.pending;
    
    return (
      <Tag color={config.color} icon={config.icon}>
        {config.text}
      </Tag>
    );
  };

  // 优先级标签映射
  const getPriorityTag = (priority: number) => {
    const priorityMap = {
      1: { color: 'default', text: '低' },
      2: { color: 'warning', text: '中' },
      3: { color: 'error', text: '高' },
    };
    
    const config = priorityMap[priority as keyof typeof priorityMap] || priorityMap[1];
    
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  // 最近任务表格列定义
  const recentTaskColumns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '任务类型',
      dataIndex: 'type',
      key: 'type',
      width: 120,
    },
    {
      title: '模型',
      dataIndex: ['model', 'name'],
      key: 'model_name',
      width: 120,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => getStatusTag(status),
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      key: 'priority',
      width: 80,
      render: (priority: number) => getPriorityTag(priority),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 150,
      render: (time: string) => dayjs(time).format('MM-DD HH:mm'),
    },
  ];

  if (loading) {
    return (
      <div className="loading-container">
        <Spin size="large" />
      </div>
    );
  }

  if (error) {
    return (
      <Alert
        message="加载失败"
        description={error}
        type="error"
        showIcon
        action={
          <span 
            style={{ cursor: 'pointer', textDecoration: 'underline' }}
            onClick={fetchDashboardData}
          >
            重试
          </span>
        }
      />
    );
  }

  if (!dashboardData) {
    return <div>暂无数据</div>;
  }

  const { task_stats, model_stats, queue_status, recent_tasks } = dashboardData;

  return (
    <div>
      <Title level={2}>Dashboard</Title>
      
      {/* 任务统计卡片 */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="总任务数"
              value={task_stats.total_tasks}
              valueStyle={{ color: '#1890ff' }}
              prefix={<ClockCircleOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="运行中"
              value={task_stats.running_tasks}
              valueStyle={{ color: '#52c41a' }}
              prefix={<PlayCircleOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="已完成"
              value={task_stats.completed_tasks}
              valueStyle={{ color: '#52c41a' }}
              prefix={<CheckCircleOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="成功率"
              value={task_stats.success_rate}
              precision={1}
              suffix="%"
              valueStyle={{ 
                color: task_stats.success_rate > 90 ? '#52c41a' : '#faad14' 
              }}
              prefix={<ExclamationCircleOutlined />}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        {/* 队列状态 */}
        <Col xs={24} lg={8}>
          <Card title="队列状态" size="small">
            <Space direction="vertical" style={{ width: '100%' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <span>高优先级</span>
                <Tag color="error">{queue_status.high_priority_count}</Tag>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <span>中优先级</span>
                <Tag color="warning">{queue_status.medium_priority_count}</Tag>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <span>低优先级</span>
                <Tag color="default">{queue_status.low_priority_count}</Tag>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <span>处理中</span>
                <Tag color="processing">{queue_status.processing_count}</Tag>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <span>延迟队列</span>
                <Tag color="default">{queue_status.delayed_count}</Tag>
              </div>
            </Space>
          </Card>
        </Col>

        {/* 模型状态 */}
        <Col xs={24} lg={16}>
          <Card title="模型状态" size="small">
            <Row gutter={[12, 12]}>
              {model_stats.map((model) => (
                <Col xs={24} sm={12} key={model.id}>
                  <Card size="small" style={{ backgroundColor: '#fafafa' }}>
                    <Space direction="vertical" style={{ width: '100%' }}>
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <span><RobotOutlined /> {model.name}</span>
                        <Tag color={model.status === 'online' ? 'success' : 'default'}>
                          {model.status}
                        </Tag>
                      </div>
                      <Progress
                        percent={(model.current_workers / model.max_workers) * 100}
                        format={() => `${model.current_workers}/${model.max_workers}`}
                        size="small"
                        status={model.current_workers === model.max_workers ? 'active' : 'normal'}
                      />
                      <div style={{ fontSize: '12px', color: '#666' }}>
                        成功率: {model.success_rate.toFixed(1)}% | 
                        待处理: {model.pending_tasks}
                      </div>
                    </Space>
                  </Card>
                </Col>
              ))}
            </Row>
          </Card>
        </Col>
      </Row>

      {/* 最近任务 */}
      <Card title="最近任务" style={{ marginTop: 16 }} size="small">
        <Table
          dataSource={recent_tasks}
          columns={recentTaskColumns}
          pagination={false}
          size="small"
          scroll={{ x: 800 }}
          rowKey="id"
        />
      </Card>
    </div>
  );
};

export default Dashboard;
