import React, { useState, useEffect } from 'react';
import {
  Card,
  Row,
  Col,
  Select,
  Typography,
  Spin,
  Table,
} from 'antd';
import ReactECharts from 'echarts-for-react';
import { statsApi } from '../../services/api';
import dayjs from 'dayjs';

const { Title } = Typography;
const { Option } = Select;

const Statistics: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [dateRangeData, setDateRangeData] = useState<any[]>([]);
  const [modelStatsData, setModelStatsData] = useState<any[]>([]);
  const [typeStatsData, setTypeStatsData] = useState<any[]>([]);
  const [selectedDays, setSelectedDays] = useState(7);

  // 获取按日期统计数据
  const fetchDateRangeStats = async (days: number) => {
    try {
      setLoading(true);
      const response = await statsApi.tasksByDate(days);
      if (response.code === 0) {
        setDateRangeData(response.data || []);
      }
    } catch (error) {
      console.error('获取日期统计失败:', error);
    } finally {
      setLoading(false);
    }
  };

  // 获取按模型统计数据
  const fetchModelStats = async () => {
    try {
      const response = await statsApi.tasksByModel();
      if (response.code === 0) {
        setModelStatsData(response.data || []);
      }
    } catch (error) {
      console.error('获取模型统计失败:', error);
    }
  };

  // 获取按类型统计数据
  const fetchTypeStats = async () => {
    try {
      const response = await statsApi.tasksByType();
      if (response.code === 0) {
        setTypeStatsData(response.data || []);
      }
    } catch (error) {
      console.error('获取类型统计失败:', error);
    }
  };

  useEffect(() => {
    fetchDateRangeStats(selectedDays);
    fetchModelStats();
    fetchTypeStats();
  }, [selectedDays]);

  // 日期统计图表配置
  const getDateStatsOption = () => {
    const dates = dateRangeData.map(item => dayjs(item.date).format('MM-DD'));
    const totalData = dateRangeData.map(item => item.total);
    const completedData = dateRangeData.map(item => item.completed);
    const failedData = dateRangeData.map(item => item.failed);

    return {
      title: {
        text: '任务趋势统计',
        left: 'center',
      },
      tooltip: {
        trigger: 'axis',
        axisPointer: {
          type: 'cross',
        },
      },
      legend: {
        data: ['总任务数', '成功任务', '失败任务'],
        top: 30,
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        containLabel: true,
      },
      xAxis: {
        type: 'category',
        boundaryGap: false,
        data: dates,
      },
      yAxis: {
        type: 'value',
      },
      series: [
        {
          name: '总任务数',
          type: 'line',
          data: totalData,
          smooth: true,
          itemStyle: { color: '#1890ff' },
        },
        {
          name: '成功任务',
          type: 'line',
          data: completedData,
          smooth: true,
          itemStyle: { color: '#52c41a' },
        },
        {
          name: '失败任务',
          type: 'line',
          data: failedData,
          smooth: true,
          itemStyle: { color: '#ff4d4f' },
        },
      ],
    };
  };

  // 模型统计饼图配置
  const getModelStatsOption = () => {
    const data = modelStatsData.map(item => ({
      name: item.model_name,
      value: item.total_tasks,
    }));

    return {
      title: {
        text: '按模型统计',
        left: 'center',
      },
      tooltip: {
        trigger: 'item',
        formatter: '{a} <br/>{b}: {c} ({d}%)',
      },
      legend: {
        orient: 'vertical',
        left: 'left',
        top: 50,
      },
      series: [
        {
          name: '任务数量',
          type: 'pie',
          radius: '50%',
          data: data,
          emphasis: {
            itemStyle: {
              shadowBlur: 10,
              shadowOffsetX: 0,
              shadowColor: 'rgba(0, 0, 0, 0.5)',
            },
          },
        },
      ],
    };
  };

  // 类型统计柱状图配置
  const getTypeStatsOption = () => {
    const types = typeStatsData.map(item => item.type);
    const totalData = typeStatsData.map(item => item.total_tasks);
    const completedData = typeStatsData.map(item => item.completed_tasks);
    const failedData = typeStatsData.map(item => item.failed_tasks);

    return {
      title: {
        text: '按任务类型统计',
        left: 'center',
      },
      tooltip: {
        trigger: 'axis',
        axisPointer: {
          type: 'shadow',
        },
      },
      legend: {
        data: ['总任务', '成功任务', '失败任务'],
        top: 30,
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        containLabel: true,
      },
      xAxis: {
        type: 'category',
        data: types,
      },
      yAxis: {
        type: 'value',
      },
      series: [
        {
          name: '总任务',
          type: 'bar',
          data: totalData,
          itemStyle: { color: '#1890ff' },
        },
        {
          name: '成功任务',
          type: 'bar',
          data: completedData,
          itemStyle: { color: '#52c41a' },
        },
        {
          name: '失败任务',
          type: 'bar',
          data: failedData,
          itemStyle: { color: '#ff4d4f' },
        },
      ],
    };
  };

  // 模型统计表格列
  const modelColumns = [
    {
      title: '模型名称',
      dataIndex: 'model_name',
      key: 'model_name',
    },
    {
      title: '模型类型',
      dataIndex: 'model_type',
      key: 'model_type',
    },
    {
      title: '总任务数',
      dataIndex: 'total_tasks',
      key: 'total_tasks',
      sorter: (a: any, b: any) => a.total_tasks - b.total_tasks,
    },
    {
      title: '成功任务',
      dataIndex: 'completed_tasks',
      key: 'completed_tasks',
      sorter: (a: any, b: any) => a.completed_tasks - b.completed_tasks,
    },
    {
      title: '失败任务',
      dataIndex: 'failed_tasks',
      key: 'failed_tasks',
      sorter: (a: any, b: any) => a.failed_tasks - b.failed_tasks,
    },
    {
      title: '成功率',
      dataIndex: 'success_rate',
      key: 'success_rate',
      render: (rate: number) => `${rate?.toFixed(1)}%`,
      sorter: (a: any, b: any) => a.success_rate - b.success_rate,
    },
    {
      title: '平均处理时间',
      dataIndex: 'avg_processing_ms',
      key: 'avg_processing_ms',
      render: (ms: number) => ms ? `${Math.round(ms)}ms` : '-',
      sorter: (a: any, b: any) => (a.avg_processing_ms || 0) - (b.avg_processing_ms || 0),
    },
  ];

  // 类型统计表格列
  const typeColumns = [
    {
      title: '任务类型',
      dataIndex: 'type',
      key: 'type',
    },
    {
      title: '总任务数',
      dataIndex: 'total_tasks',
      key: 'total_tasks',
      sorter: (a: any, b: any) => a.total_tasks - b.total_tasks,
    },
    {
      title: '成功任务',
      dataIndex: 'completed_tasks',
      key: 'completed_tasks',
      sorter: (a: any, b: any) => a.completed_tasks - b.completed_tasks,
    },
    {
      title: '失败任务',
      dataIndex: 'failed_tasks',
      key: 'failed_tasks',
      sorter: (a: any, b: any) => a.failed_tasks - b.failed_tasks,
    },
    {
      title: '成功率',
      dataIndex: 'success_rate',
      key: 'success_rate',
      render: (rate: number) => `${rate?.toFixed(1)}%`,
      sorter: (a: any, b: any) => a.success_rate - b.success_rate,
    },
    {
      title: '平均处理时间',
      dataIndex: 'avg_processing_ms',
      key: 'avg_processing_ms',
      render: (ms: number) => ms ? `${Math.round(ms)}ms` : '-',
      sorter: (a: any, b: any) => (a.avg_processing_ms || 0) - (b.avg_processing_ms || 0),
    },
  ];

  return (
    <div>
      <Title level={2}>统计分析</Title>
      
      {/* 日期范围选择 */}
      <Card style={{ marginBottom: 16 }}>
        <div style={{ marginBottom: 16 }}>
          时间范围：
          <Select
            value={selectedDays}
            onChange={setSelectedDays}
            style={{ marginLeft: 8, width: 120 }}
          >
            <Option value={7}>最近7天</Option>
            <Option value={14}>最近14天</Option>
            <Option value={30}>最近30天</Option>
            <Option value={90}>最近90天</Option>
          </Select>
        </div>
        
        <Spin spinning={loading}>
          <ReactECharts
            option={getDateStatsOption()}
            style={{ height: '400px' }}
          />
        </Spin>
      </Card>

      <Row gutter={[16, 16]}>
        {/* 模型统计 */}
        <Col xs={24} lg={12}>
          <Card title="模型统计分布">
            <ReactECharts
              option={getModelStatsOption()}
              style={{ height: '400px' }}
            />
          </Card>
        </Col>

        {/* 任务类型统计 */}
        <Col xs={24} lg={12}>
          <Card title="任务类型分布">
            <ReactECharts
              option={getTypeStatsOption()}
              style={{ height: '400px' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        {/* 模型详细统计表格 */}
        <Col xs={24} lg={12}>
          <Card title="模型详细统计" size="small">
            <Table
              dataSource={modelStatsData}
              columns={modelColumns}
              pagination={{ pageSize: 10 }}
              size="small"
              scroll={{ x: 800 }}
              rowKey="model_name"
            />
          </Card>
        </Col>

        {/* 类型详细统计表格 */}
        <Col xs={24} lg={12}>
          <Card title="类型详细统计" size="small">
            <Table
              dataSource={typeStatsData}
              columns={typeColumns}
              pagination={{ pageSize: 10 }}
              size="small"
              scroll={{ x: 800 }}
              rowKey="type"
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Statistics;
