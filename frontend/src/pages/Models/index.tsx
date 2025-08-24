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
  Typography,
  Progress,
  InputNumber,
} from 'antd';
import {
  PlusOutlined,
  ReloadOutlined,
  EditOutlined,
  DeleteOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  RobotOutlined,
} from '@ant-design/icons';
import { modelApi } from '../../services/api';
import { Model, ModelStats } from '../../types';

const { Title } = Typography;
const { Option } = Select;
const { TextArea } = Input;

const Models: React.FC = () => {
  const [form] = Form.useForm();
  
  const [models, setModels] = useState<ModelStats[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [modalLoading, setModalLoading] = useState(false);
  const [editingModel, setEditingModel] = useState<Model | null>(null);

  // 获取模型列表
  const fetchModels = async () => {
    try {
      setLoading(true);
      const response = await modelApi.stats();
      if (response.code === 0) {
        setModels(response.data || []);
      }
    } catch (error) {
      message.error('获取模型列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchModels();
  }, []);

  // 创建/更新模型
  const handleSubmitModel = async (values: any) => {
    try {
      setModalLoading(true);
      
      // 处理配置JSON
      let config = {};
      if (values.configJson) {
        try {
          config = JSON.parse(values.configJson);
        } catch (error) {
          message.error('配置JSON格式错误');
          return;
        }
      }

      const modelData = {
        ...values,
        config,
      };

      let response;
      if (editingModel) {
        response = await modelApi.update(editingModel.id, modelData);
      } else {
        response = await modelApi.create(modelData);
      }

      if (response.code === 0) {
        message.success(editingModel ? '模型更新成功' : '模型创建成功');
        setModalVisible(false);
        form.resetFields();
        setEditingModel(null);
        fetchModels();
      }
    } catch (error) {
      // 错误已在拦截器中处理
    } finally {
      setModalLoading(false);
    }
  };

  // 删除模型
  const handleDeleteModel = async (id: number) => {
    try {
      const response = await modelApi.delete(id);
      if (response.code === 0) {
        message.success('模型删除成功');
        fetchModels();
      }
    } catch (error) {
      // 错误已在拦截器中处理
    }
  };

  // 更新模型状态
  const handleUpdateStatus = async (id: number, status: string) => {
    try {
      const response = await modelApi.updateStatus(id, status);
      if (response.code === 0) {
        message.success('状态更新成功');
        fetchModels();
      }
    } catch (error) {
      // 错误已在拦截器中处理
    }
  };

  // 编辑模型
  const handleEditModel = (model: Model) => {
    setEditingModel(model);
    form.setFieldsValue({
      ...model,
      configJson: JSON.stringify(model.config, null, 2),
    });
    setModalVisible(true);
  };

  // 状态标签
  const getStatusTag = (status: string) => {
    const statusMap = {
      online: { color: 'success', text: '在线' },
      offline: { color: 'default', text: '离线' },
      maintenance: { color: 'warning', text: '维护中' },
    };
    
    const config = statusMap[status as keyof typeof statusMap] || statusMap.offline;
    
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  // 类型标签
  const getTypeTag = (type: string) => {
    const typeMap = {
      openai: { color: 'blue', text: 'OpenAI' },
      local: { color: 'green', text: '本地模型' },
      custom: { color: 'purple', text: '自定义' },
    };
    
    const config = typeMap[type as keyof typeof typeMap] || typeMap.custom;
    
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  // 表格列定义
  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '模型名称',
      dataIndex: 'name',
      key: 'name',
      width: 150,
      render: (name: string) => (
        <Space>
          <RobotOutlined />
          {name}
        </Space>
      ),
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      width: 100,
      render: (type: string) => getTypeTag(type),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => getStatusTag(status),
    },
    {
      title: 'Worker使用情况',
      key: 'workers',
      width: 200,
      render: (_, record: ModelStats) => (
        <div>
          <Progress
            percent={(record.current_workers / record.max_workers) * 100}
            format={() => `${record.current_workers}/${record.max_workers}`}
            size="small"
            status={record.current_workers === record.max_workers ? 'active' : 'normal'}
          />
        </div>
      ),
    },
    {
      title: '请求统计',
      key: 'requests',
      width: 120,
      render: (_, record: ModelStats) => (
        <div style={{ fontSize: '12px' }}>
          <div>总计: {record.total_requests}</div>
          <div>成功: {record.success_requests}</div>
          <div style={{ color: record.success_rate > 90 ? '#52c41a' : '#faad14' }}>
            成功率: {record.success_rate.toFixed(1)}%
          </div>
        </div>
      ),
    },
    {
      title: '任务统计',
      key: 'tasks',
      width: 120,
      render: (_, record: ModelStats) => (
        <div style={{ fontSize: '12px' }}>
          <div>待处理: {record.pending_tasks}</div>
          <div>运行中: {record.running_tasks}</div>
          <div>平均响应: {record.avg_response_ms}ms</div>
        </div>
      ),
    },
    {
      title: '操作',
      key: 'actions',
      width: 200,
      fixed: 'right' as const,
      render: (_, record: Model) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEditModel(record)}
          >
            编辑
          </Button>
          {record.status === 'online' ? (
            <Button
              type="link"
              size="small"
              icon={<PauseCircleOutlined />}
              onClick={() => handleUpdateStatus(record.id, 'offline')}
            >
              下线
            </Button>
          ) : (
            <Button
              type="link"
              size="small"
              icon={<PlayCircleOutlined />}
              onClick={() => handleUpdateStatus(record.id, 'online')}
            >
              上线
            </Button>
          )}
          <Button
            type="link"
            size="small"
            danger
            icon={<DeleteOutlined />}
            onClick={() => {
              Modal.confirm({
                title: '确认删除模型？',
                content: '删除后无法恢复，请确保没有相关任务正在执行',
                onOk: () => handleDeleteModel(record.id),
              });
            }}
          >
            删除
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Title level={2}>模型管理</Title>
      
      <Card>
        <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => {
              setEditingModel(null);
              form.resetFields();
              setModalVisible(true);
            }}
          >
            添加模型
          </Button>
          <Button
            icon={<ReloadOutlined />}
            onClick={fetchModels}
            loading={loading}
          >
            刷新
          </Button>
        </div>
        
        <Table
          dataSource={models}
          columns={columns}
          loading={loading}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条/共 ${total} 条`,
          }}
          scroll={{ x: 1000 }}
          rowKey="id"
        />
      </Card>

      {/* 创建/编辑模型模态框 */}
      <Modal
        title={editingModel ? '编辑模型' : '添加模型'}
        open={modalVisible}
        onOk={() => form.submit()}
        onCancel={() => {
          setModalVisible(false);
          form.resetFields();
          setEditingModel(null);
        }}
        confirmLoading={modalLoading}
        width={700}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmitModel}
        >
          <Form.Item
            label="模型名称"
            name="name"
            rules={[{ required: true, message: '请输入模型名称' }]}
          >
            <Input placeholder="请输入模型名称" />
          </Form.Item>
          
          <Form.Item
            label="模型类型"
            name="type"
            rules={[{ required: true, message: '请选择模型类型' }]}
          >
            <Select placeholder="选择模型类型">
              <Option value="openai">OpenAI</Option>
              <Option value="local">本地模型</Option>
              <Option value="custom">自定义</Option>
            </Select>
          </Form.Item>
          
          <Form.Item
            label="最大Worker数"
            name="max_workers"
            rules={[{ required: true, message: '请设置最大Worker数' }]}
            initialValue={1}
          >
            <InputNumber min={1} max={20} style={{ width: '100%' }} />
          </Form.Item>
          
          <Form.Item
            label="状态"
            name="status"
            initialValue="offline"
          >
            <Select>
              <Option value="online">在线</Option>
              <Option value="offline">离线</Option>
              <Option value="maintenance">维护中</Option>
            </Select>
          </Form.Item>
          
          <Form.Item
            label="配置信息"
            name="configJson"
            rules={[
              { required: true, message: '请输入配置信息' },
              {
                validator: (_, value) => {
                  if (!value) return Promise.resolve();
                  try {
                    JSON.parse(value);
                    return Promise.resolve();
                  } catch {
                    return Promise.reject(new Error('请输入有效的JSON格式'));
                  }
                }
              }
            ]}
          >
            <TextArea
              rows={8}
              placeholder={`请输入JSON格式的配置信息，例如：
{
  "api_key": "your-api-key",
  "base_url": "https://api.openai.com/v1",
  "model": "gpt-3.5-turbo",
  "max_tokens": 4096,
  "temperature": 0.7
}`}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default Models;
