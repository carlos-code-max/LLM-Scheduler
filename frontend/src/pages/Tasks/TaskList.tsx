import React, { useState, useEffect } from 'react';
import {
  Card,
  Table,
  Button,
  Tag,
  Space,
  Modal,
  Form,
  Input,
  Select,
  message,
  Tooltip,
  Typography,
  Row,
  Col,
  Statistic,
} from 'antd';
import {
  PlusOutlined,
  ReloadOutlined,
  PlayCircleOutlined,
  StopOutlined,
  EyeOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { taskApi, modelApi } from '../../services/api';
import { Task, TaskCreateRequest, Model, TaskListParams, PaginationParams, TaskStats } from '../../types';
import dayjs from 'dayjs';

const { Title } = Typography;
const { Option } = Select;
const { TextArea } = Input;

const TaskList: React.FC = () => {
  const navigate = useNavigate();
  const [form] = Form.useForm();
  
  const [tasks, setTasks] = useState<Task[]>([]);
  const [models, setModels] = useState<Model[]>([]);
  const [taskStats, setTaskStats] = useState<TaskStats | null>(null);
  const [loading, setLoading] = useState(false);
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [createLoading, setCreateLoading] = useState(false);
  const [pagination, setPagination] = useState<PaginationParams>({
    current: 1,
    pageSize: 20,
    total: 0,
  });
  const [filters, setFilters] = useState<TaskListParams>({});

  // 获取任务列表
  const fetchTasks = async (params: TaskListParams = {}) => {
    try {
      setLoading(true);
      const response = await taskApi.list({
        page: pagination.current,
        page_size: pagination.pageSize,
        ...filters,
        ...params,
      });
      
      if (response.code === 0) {
        setTasks(response.data || []);
        setPagination(prev => ({
          ...prev,
          total: response.total,
          current: response.page,
          pageSize: response.size,
        }));
      }
    } catch (error) {
      message.error('获取任务列表失败');
    } finally {
      setLoading(false);
    }
  };

  // 获取模型列表
  const fetchModels = async () => {
    try {
      const response = await modelApi.list();
      if (response.code === 0) {
        setModels(response.data || []);
      }
    } catch (error) {
      message.error('获取模型列表失败');
    }
  };

  // 获取任务统计
  const fetchTaskStats = async () => {
    try {
      const response = await taskApi.stats();
      if (response.code === 0) {
        setTaskStats(response.data!);
      }
    } catch (error) {
      console.error('获取任务统计失败:', error);
    }
  };

  useEffect(() => {
    fetchTasks();
    fetchModels();
    fetchTaskStats();
  }, []);

  // 创建任务
  const handleCreateTask = async (values: TaskCreateRequest) => {
    try {
      setCreateLoading(true);
      const response = await taskApi.create(values);
      if (response.code === 0) {
        message.success('任务创建成功');
        setCreateModalVisible(false);
        form.resetFields();
        fetchTasks();
        fetchTaskStats();
      }
    } catch (error) {
      // 错误已在拦截器中处理
    } finally {
      setCreateLoading(false);
    }
  };

  // 取消任务
  const handleCancelTask = async (id: number) => {
    try {
      const response = await taskApi.cancel(id);
      if (response.code === 0) {
        message.success('任务已取消');
        fetchTasks();
        fetchTaskStats();
      }
    } catch (error) {
      // 错误已在拦截器中处理
    }
  };

  // 重试任务
  const handleRetryTask = async (id: number) => {
    try {
      const response = await taskApi.retry(id);
      if (response.code === 0) {
        message.success('任务已重新提交');
        fetchTasks();
        fetchTaskStats();
      }
    } catch (error) {
      // 错误已在拦截器中处理
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
      <Tag color={config.color} icon={config.icon}>
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

  // 表格列定义
  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
      sorter: true,
    },
    {
      title: '任务类型',
      dataIndex: 'type',
      key: 'type',
      width: 120,
      filters: [
        { text: '文本生成', value: 'text-generation' },
        { text: '翻译', value: 'translation' },
        { text: '摘要', value: 'summarization' },
        { text: '向量化', value: 'embedding' },
      ],
    },
    {
      title: '模型',
      dataIndex: ['model', 'name'],
      key: 'model_name',
      width: 120,
    },
    {
      title: '输入内容',
      dataIndex: 'input',
      key: 'input',
      ellipsis: { showTitle: false },
      render: (text: string) => (
        <Tooltip title={text}>
          {text?.length > 50 ? `${text.substring(0, 50)}...` : text}
        </Tooltip>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      filters: [
        { text: '待处理', value: 'pending' },
        { text: '运行中', value: 'running' },
        { text: '已完成', value: 'completed' },
        { text: '失败', value: 'failed' },
        { text: '已取消', value: 'cancelled' },
      ],
      render: (status: string) => getStatusTag(status),
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      key: 'priority',
      width: 80,
      filters: [
        { text: '高', value: 3 },
        { text: '中', value: 2 },
        { text: '低', value: 1 },
      ],
      render: (priority: number) => getPriorityTag(priority),
    },
    {
      title: '重试次数',
      dataIndex: 'retry_count',
      key: 'retry_count',
      width: 80,
      render: (count: number, record: Task) => `${count}/${record.max_retries}`,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 150,
      sorter: true,
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '操作',
      key: 'actions',
      width: 200,
      fixed: 'right' as const,
      render: (_, record: Task) => (
        <Space size="small" className="table-actions">
          <Button
            type="link"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => navigate(`/tasks/detail/${record.id}`)}
          >
            查看
          </Button>
          {(record.status === 'pending' || record.status === 'running') && (
            <Button
              type="link"
              size="small"
              danger
              icon={<StopOutlined />}
              onClick={() => {
                Modal.confirm({
                  title: '确认取消任务？',
                  content: '取消后任务将无法恢复执行',
                  onOk: () => handleCancelTask(record.id),
                });
              }}
            >
              取消
            </Button>
          )}
          {record.status === 'failed' && record.retry_count < record.max_retries && (
            <Button
              type="link"
              size="small"
              icon={<PlayCircleOutlined />}
              onClick={() => handleRetryTask(record.id)}
            >
              重试
            </Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Title level={2}>任务管理</Title>
      
      {/* 统计卡片 */}
      {taskStats && (
        <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
          <Col xs={24} sm={12} md={6}>
            <Card size="small">
              <Statistic
                title="总任务数"
                value={taskStats.total_tasks}
                prefix={<ClockCircleOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card size="small">
              <Statistic
                title="运行中"
                value={taskStats.running_tasks}
                valueStyle={{ color: '#52c41a' }}
                prefix={<PlayCircleOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card size="small">
              <Statistic
                title="已完成"
                value={taskStats.completed_tasks}
                valueStyle={{ color: '#52c41a' }}
                prefix={<CheckCircleOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card size="small">
              <Statistic
                title="成功率"
                value={taskStats.success_rate}
                precision={1}
                suffix="%"
                valueStyle={{ 
                  color: taskStats.success_rate > 90 ? '#52c41a' : '#faad14' 
                }}
              />
            </Card>
          </Col>
        </Row>
      )}

      {/* 任务列表 */}
      <Card>
        <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Space>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => setCreateModalVisible(true)}
            >
              创建任务
            </Button>
          </Space>
          <Button
            icon={<ReloadOutlined />}
            onClick={() => fetchTasks()}
            loading={loading}
          >
            刷新
          </Button>
        </div>
        
        <Table
          dataSource={tasks}
          columns={columns}
          loading={loading}
          pagination={{
            ...pagination,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条/共 ${total} 条`,
          }}
          onChange={(paginationConfig, filtersConfig, sorterConfig) => {
            const newPagination = {
              current: paginationConfig.current || 1,
              pageSize: paginationConfig.pageSize || 20,
              total: paginationConfig.total || 0,
            };
            setPagination(newPagination);
            
            const newFilters: TaskListParams = {};
            Object.keys(filtersConfig).forEach(key => {
              const filterValue = filtersConfig[key];
              if (filterValue && filterValue.length > 0) {
                newFilters[key as keyof TaskListParams] = filterValue[0] as any;
              }
            });
            setFilters(newFilters);
            
            fetchTasks({
              page: newPagination.current,
              page_size: newPagination.pageSize,
              ...newFilters,
            });
          }}
          scroll={{ x: 1200 }}
          rowKey="id"
        />
      </Card>

      {/* 创建任务模态框 */}
      <Modal
        title="创建任务"
        open={createModalVisible}
        onOk={() => form.submit()}
        onCancel={() => {
          setCreateModalVisible(false);
          form.resetFields();
        }}
        confirmLoading={createLoading}
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleCreateTask}
        >
          <Form.Item
            label="任务类型"
            name="type"
            rules={[{ required: true, message: '请选择任务类型' }]}
          >
            <Select placeholder="选择任务类型">
              <Option value="text-generation">文本生成</Option>
              <Option value="translation">翻译</Option>
              <Option value="summarization">摘要</Option>
              <Option value="embedding">向量化</Option>
              <Option value="custom">自定义</Option>
            </Select>
          </Form.Item>
          
          <Form.Item
            label="模型"
            name="model_id"
            rules={[{ required: true, message: '请选择模型' }]}
          >
            <Select placeholder="选择模型">
              {models.map(model => (
                <Option key={model.id} value={model.id}>
                  {model.name} ({model.type})
                </Option>
              ))}
            </Select>
          </Form.Item>
          
          <Form.Item
            label="优先级"
            name="priority"
            initialValue={2}
          >
            <Select>
              <Option value={3}>高</Option>
              <Option value={2}>中</Option>
              <Option value={1}>低</Option>
            </Select>
          </Form.Item>
          
          <Form.Item
            label="输入内容"
            name="input"
            rules={[{ required: true, message: '请输入内容' }]}
          >
            <TextArea
              rows={6}
              placeholder="请输入任务的输入内容..."
              maxLength={2000}
              showCount
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default TaskList;
