import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Card,
  Descriptions,
  Tag,
  Button,
  Space,
  Typography,
  Spin,
  Alert,
  Timeline,
  Modal,
  message,
} from 'antd';
import {
  ArrowLeftOutlined,
  PlayCircleOutlined,
  StopOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons';
import { taskApi } from '../../services/api';
import { Task } from '../../types';
import dayjs from 'dayjs';

const { Title, Text } = Typography;

const TaskDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  
  const [task, setTask] = useState<Task | null>(null);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 获取任务详情
  const fetchTaskDetail = async () => {
    if (!id) return;
    
    try {
      setLoading(true);
      setError(null);
      const response = await taskApi.get(parseInt(id));
      if (response.code === 0) {
        setTask(response.data!);
      } else {
        setError(response.message);
      }
    } catch (err) {
      setError('获取任务详情失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTaskDetail();
  }, [id]);

  // 取消任务
  const handleCancelTask = async () => {
    if (!task) return;
    
    try {
      setActionLoading(true);
      const response = await taskApi.cancel(task.id);
      if (response.code === 0) {
        message.success('任务已取消');
        fetchTaskDetail();
      }
    } catch (error) {
      // 错误已在拦截器中处理
    } finally {
      setActionLoading(false);
    }
  };

  // 重试任务
  const handleRetryTask = async () => {
    if (!task) return;
    
    try {
      setActionLoading(true);
      const response = await taskApi.retry(task.id);
      if (response.code === 0) {
        message.success('任务已重新提交');
        fetchTaskDetail();
      }
    } catch (error) {
      // 错误已在拦截器中处理
    } finally {
      setActionLoading(false);
    }
  };

  // 任务状态标签
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
      <Tag color={config.color} icon={config.icon} style={{ fontSize: '14px', padding: '4px 8px' }}>
        {config.text}
      </Tag>
    );
  };

  // 优先级标签
  const getPriorityTag = (priority: number) => {
    const priorityMap = {
      1: { color: 'default', text: '低' },
      2: { color: 'warning', text: '中' },
      3: { color: 'error', text: '高' },
    };
    
    const config = priorityMap[priority as keyof typeof priorityMap] || priorityMap[1];
    
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  // 日志级别颜色
  const getLogLevelColor = (level: string) => {
    const levelMap = {
      debug: 'default',
      info: 'blue',
      warn: 'orange',
      error: 'red',
    };
    return levelMap[level as keyof typeof levelMap] || 'default';
  };

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
          <Space>
            <Button size="small" onClick={fetchTaskDetail}>
              重试
            </Button>
            <Button size="small" onClick={() => navigate('/tasks')}>
              返回列表
            </Button>
          </Space>
        }
      />
    );
  }

  if (!task) {
    return <div>任务不存在</div>;
  }

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Button
          icon={<ArrowLeftOutlined />}
          onClick={() => navigate('/tasks')}
          style={{ marginRight: 16 }}
        >
          返回列表
        </Button>
        <Title level={2} style={{ display: 'inline-block', margin: 0 }}>
          任务详情 #{task.id}
        </Title>
      </div>

      {/* 基本信息 */}
      <Card title="基本信息" style={{ marginBottom: 16 }}>
        <Descriptions column={2} bordered>
          <Descriptions.Item label="任务ID">{task.id}</Descriptions.Item>
          <Descriptions.Item label="任务类型">
            <Tag>{task.type}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="状态">
            {getStatusTag(task.status)}
          </Descriptions.Item>
          <Descriptions.Item label="优先级">
            {getPriorityTag(task.priority)}
          </Descriptions.Item>
          <Descriptions.Item label="关联模型">
            {task.model ? (
              <Space>
                <Text>{task.model.name}</Text>
                <Tag color="blue">{task.model.type}</Tag>
              </Space>
            ) : (
              '未知'
            )}
          </Descriptions.Item>
          <Descriptions.Item label="重试次数">
            {task.retry_count} / {task.max_retries}
          </Descriptions.Item>
          <Descriptions.Item label="创建时间">
            {dayjs(task.created_at).format('YYYY-MM-DD HH:mm:ss')}
          </Descriptions.Item>
          <Descriptions.Item label="开始时间">
            {task.started_at ? dayjs(task.started_at).format('YYYY-MM-DD HH:mm:ss') : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="完成时间">
            {task.completed_at ? dayjs(task.completed_at).format('YYYY-MM-DD HH:mm:ss') : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="处理时长">
            {task.started_at && task.completed_at
              ? `${dayjs(task.completed_at).diff(dayjs(task.started_at), 'second')} 秒`
              : '-'
            }
          </Descriptions.Item>
        </Descriptions>

        {/* 操作按钮 */}
        <div style={{ marginTop: 16 }}>
          <Space>
            {(task.status === 'pending' || task.status === 'running') && (
              <Button
                type="primary"
                danger
                icon={<StopOutlined />}
                loading={actionLoading}
                onClick={() => {
                  Modal.confirm({
                    title: '确认取消任务？',
                    content: '取消后任务将无法恢复执行',
                    onOk: handleCancelTask,
                  });
                }}
              >
                取消任务
              </Button>
            )}
            {task.status === 'failed' && task.retry_count < task.max_retries && (
              <Button
                type="primary"
                icon={<PlayCircleOutlined />}
                loading={actionLoading}
                onClick={handleRetryTask}
              >
                重试任务
              </Button>
            )}
            <Button onClick={fetchTaskDetail}>刷新</Button>
          </Space>
        </div>
      </Card>

      {/* 输入内容 */}
      <Card title="输入内容" style={{ marginBottom: 16 }}>
        <Text code style={{ fontSize: '12px', lineHeight: 1.6, whiteSpace: 'pre-wrap' }}>
          {task.input}
        </Text>
      </Card>

      {/* 输出结果 */}
      {task.output && (
        <Card title="输出结果" style={{ marginBottom: 16 }}>
          <Text code style={{ fontSize: '12px', lineHeight: 1.6, whiteSpace: 'pre-wrap' }}>
            {task.output}
          </Text>
        </Card>
      )}

      {/* 错误信息 */}
      {task.error_message && (
        <Card title="错误信息" style={{ marginBottom: 16 }}>
          <Alert
            message="任务执行失败"
            description={task.error_message}
            type="error"
            showIcon
          />
        </Card>
      )}

      {/* 执行日志 */}
      {task.logs && task.logs.length > 0 && (
        <Card title="执行日志">
          <Timeline>
            {task.logs.map((log, index) => (
              <Timeline.Item
                key={index}
                color={getLogLevelColor(log.level)}
                dot={<Tag color={getLogLevelColor(log.level)}>{log.level.toUpperCase()}</Tag>}
              >
                <div style={{ marginTop: 4 }}>
                  <Text type="secondary" style={{ fontSize: '12px' }}>
                    {dayjs(log.created_at).format('YYYY-MM-DD HH:mm:ss')}
                  </Text>
                  <div style={{ marginTop: 4 }}>
                    <Text>{log.message}</Text>
                  </div>
                  {log.data && Object.keys(log.data).length > 0 && (
                    <details style={{ marginTop: 8 }}>
                      <summary style={{ cursor: 'pointer', color: '#1890ff' }}>
                        查看详细数据
                      </summary>
                      <pre style={{ 
                        marginTop: 8, 
                        fontSize: '12px', 
                        background: '#f5f5f5', 
                        padding: '8px',
                        borderRadius: '4px',
                        overflow: 'auto'
                      }}>
                        {JSON.stringify(log.data, null, 2)}
                      </pre>
                    </details>
                  )}
                </div>
              </Timeline.Item>
            ))}
          </Timeline>
        </Card>
      )}
    </div>
  );
};

export default TaskDetail;
