import React, { useState, useRef, useEffect } from 'react';
import { PageContainer } from '@ant-design/pro-components';
import { Tabs, message, Alert, Row, Col, Card, Statistic } from 'antd';
import { Column } from '@ant-design/plots';

import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { ProTable } from '@ant-design/pro-components';

// 发布清单记录数据类型
interface ReleaseRecord {
  id: string;
  title: string;
  version: string;
  status: string;
  environment: string;
  startTime: string;
  endTime: string;
  creator: string;
  description: string;
  priority: string;
  risk: string;
  releaseType: string;
  affectedSystems: string;
  rollbackPlan: string;
}

// 运维变更工单记录数据类型
interface WorkOrderRecord {
  id: string;
  title: string;
  workOrderType: string;
  status: string;
  environment: string;
  startTime: string;
  endTime: string;
  creator: string;
  assignee: string;
  description: string;
  priority: string;
  risk: string;
  category: string;
  affectedServices: string;
  solution: string;
}

// 自动化变更记录数据类型
interface AutoChangeRecord {
  id: string;
  title: string;
  automationType: string;
  status: string;
  environment: string;
  startTime: string;
  endTime: string;
  creator: string;
  description: string;
  priority: string;
  risk: string;
  scriptName: string;
  executionMode: string;
  targetServers: string;
  successRate: number;
}

// 查询请求参数类型
interface QueryParams {
  startTime?: string;
  endTime?: string;
  status?: string;
  environment?: string;
  page?: number;
  pageSize?: number;
  searchKeyword?: string;
}

// 趋势数据类型
interface TrendData {
  date: string;
  count: number;
  type: string;
}

const ChangeQuery: React.FC = () => {
  const [activeTab, setActiveTab] = useState('release');
  const releaseActionRef = useRef<ActionType>();
  const operationChangeActionRef = useRef<ActionType>();
  const autoChangeActionRef = useRef<ActionType>();

  // API调用函数
  const fetchChangeData = async (params: QueryParams, apiEndpoint: string) => {
    try {
      console.log(`正在调用API: /api/v1/change/${apiEndpoint}`, params);
      const response = await fetch(`/api/v1/change/${apiEndpoint}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify(params),
      });

      console.log(`API响应状态: ${response.status}`);
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      console.log(`API响应数据:`, data);
      return data;
    } catch (error) {
      console.error('API调用失败:', error);
      message.error('数据获取失败');
      return { success: false, data: [], total: 0 };
    }
  };

  // 获取仪表板数据
  const fetchDashboardData = async () => {
    try {
      console.log('正在获取仪表板数据...');
      const response = await fetch('/api/v1/change/dashboard', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
      });

      console.log(`仪表板API响应状态: ${response.status}`);
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      console.log('仪表板API响应数据:', data);
      if (data.success) {
        setDashboardData({
          todayRelease: data.data.today.releaseCount,
          todayWorkOrder: data.data.today.workOrderCount,
          todayAutoChange: data.data.today.autoChangeCount,
          todayFault: data.data.today.faultCount,
          yearRelease: data.data.annual.releaseCount,
          yearWorkOrder: data.data.annual.workOrderCount,
          yearAutoChange: data.data.annual.autoChangeCount,
          yearFault: data.data.annual.faultCount,
          releaseTrend: data.data.trends.releaseTrend,
          workOrderTrend: data.data.trends.workOrderTrend,
          autoChangeTrend: data.data.trends.autoChangeTrend,
          faultTrend: data.data.trends.faultTrend,
        });
      }
    } catch (error) {
      console.error('获取仪表板数据失败:', error);
      message.error('仪表板数据获取失败');
    }
  };

  // 统计数据状态
  const [dashboardData, setDashboardData] = useState<{
    todayRelease: number;
    todayWorkOrder: number;
    todayAutoChange: number;
    todayFault: number;
    yearRelease: number;
    yearWorkOrder: number;
    yearAutoChange: number;
    yearFault: number;
    releaseTrend: Array<{ date: string; count: number }>;
    workOrderTrend: Array<{ date: string; count: number }>;
    autoChangeTrend: Array<{ date: string; count: number }>;
    faultTrend: Array<{ date: string; count: number }>;
  }>({
    todayRelease: 0,
    todayWorkOrder: 0,
    todayAutoChange: 0,
    todayFault: 0,
    yearRelease: 0,
    yearWorkOrder: 0,
    yearAutoChange: 0,
    yearFault: 0,
    releaseTrend: [],
    workOrderTrend: [],
    autoChangeTrend: [],
    faultTrend: []
  });
  const [loading, setLoading] = useState(false);

  // 处理趋势数据，转换为图表格式
  const processTrendData = (trendData: Array<{ date: string; count: number }>, type: string): TrendData[] => {
    return trendData.map(item => ({
      date: item.date,
      count: item.count,
      type: type
    }));
  };

  // 获取所有趋势数据用于图表
  const getAllTrendData = () => {
    const allData: TrendData[] = [];
    
    // 添加发布趋势
    allData.push(...processTrendData(dashboardData.releaseTrend, '发布'));
    
    // 添加工单趋势
    allData.push(...processTrendData(dashboardData.workOrderTrend, '工单'));
    
    // 添加自动化变更趋势
    allData.push(...processTrendData(dashboardData.autoChangeTrend, '自动化变更'));
    
    // 添加故障趋势
    allData.push(...processTrendData(dashboardData.faultTrend, '故障'));
    
    return allData;
  };

  // 图表配置
  const chartConfig = {
    xField: 'date',
    yField: 'count',
    seriesField: 'type',
    isGroup: true,
    columnStyle: {
      radius: [4, 4, 0, 0],
    },
    label: {
      position: 'top',
    },
    legend: {
      position: 'top',
    },
    xAxis: {
      label: {
        autoRotate: true,
        autoHide: true,
        autoEllipsis: true,
      },
    },
    yAxis: {
      label: {
        formatter: (v: string) => `${v}`,
      },
    },
  };

  // 发布清单列定义
  const releaseColumns: ProColumns<ReleaseRecord>[] = [
    {
      title: '发布ID',
      dataIndex: 'id',
      key: 'id',
      width: 120,
      ellipsis: true,
    },
    {
      title: '发布标题',
      dataIndex: 'title',
      key: 'title',
      width: 200,
      ellipsis: true,
    },
    {
      title: '版本号',
      dataIndex: 'version',
      key: 'version',
      width: 100,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      valueType: 'select',
      valueEnum: {
        '进行中': { text: '进行中', status: 'processing' },
        '已完成': { text: '已完成', status: 'success' },
        '已取消': { text: '已取消', status: 'default' },
        '待审核': { text: '待审核', status: 'warning' },
      },
    },
    {
      title: '环境',
      dataIndex: 'environment',
      key: 'environment',
      width: 100,
      valueType: 'select',
      valueEnum: {
        '生产环境': { text: '生产环境', status: 'error' },
        '测试环境': { text: '测试环境', status: 'warning' },
        '开发环境': { text: '开发环境', status: 'default' },
      },
    },
    {
      title: '开始时间',
      dataIndex: 'startTime',
      key: 'startTime',
      width: 150,
      valueType: 'dateTime',
      sorter: true,
    },
    {
      title: '结束时间',
      dataIndex: 'endTime',
      key: 'endTime',
      width: 150,
      valueType: 'dateTime',
      render: (text) => text || '-',
    },
    {
      title: '创建人',
      dataIndex: 'creator',
      key: 'creator',
      width: 100,
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      key: 'priority',
      width: 80,
      valueType: 'select',
      valueEnum: {
        '高': { text: '高', status: 'error' },
        '中': { text: '中', status: 'warning' },
        '低': { text: '低', status: 'default' },
      },
    },
    {
      title: '风险等级',
      dataIndex: 'risk',
      key: 'risk',
      width: 100,
      valueType: 'select',
      valueEnum: {
        '高风险': { text: '高风险', status: 'error' },
        '中风险': { text: '中风险', status: 'warning' },
        '低风险': { text: '低风险', status: 'default' },
      },
    },
    {
      title: '发布类型',
      dataIndex: 'releaseType',
      key: 'releaseType',
      width: 120,
    },
    {
      title: '影响系统',
      dataIndex: 'affectedSystems',
      key: 'affectedSystems',
      width: 150,
      ellipsis: true,
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      fixed: 'right',
      render: (_, record) => (
        <div>
          <a onClick={() => handleView(record)}>
            查看
          </a>
        </div>
      ),
    },
  ];

  // 运维变更工单列定义
  const workOrderColumns: ProColumns<WorkOrderRecord>[] = [
    {
      title: '工单ID',
      dataIndex: 'id',
      key: 'id',
      width: 120,
      ellipsis: true,
    },
    {
      title: '工单标题',
      dataIndex: 'title',
      key: 'title',
      width: 200,
      ellipsis: true,
    },
    {
      title: '工单类型',
      dataIndex: 'workOrderType',
      key: 'workOrderType',
      width: 120,
      valueType: 'select',
      valueEnum: {
        '配置变更': { text: '配置变更', status: 'processing' },
        '权限申请': { text: '权限申请', status: 'warning' },
        '故障处理': { text: '故障处理', status: 'error' },
        '系统维护': { text: '系统维护', status: 'default' },
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      valueType: 'select',
      valueEnum: {
        '进行中': { text: '进行中', status: 'processing' },
        '已完成': { text: '已完成', status: 'success' },
        '已取消': { text: '已取消', status: 'default' },
        '待审核': { text: '待审核', status: 'warning' },
      },
    },
    {
      title: '环境',
      dataIndex: 'environment',
      key: 'environment',
      width: 100,
      valueType: 'select',
      valueEnum: {
        '生产环境': { text: '生产环境', status: 'error' },
        '测试环境': { text: '测试环境', status: 'warning' },
        '开发环境': { text: '开发环境', status: 'default' },
      },
    },
    {
      title: '开始时间',
      dataIndex: 'startTime',
      key: 'startTime',
      width: 150,
      valueType: 'dateTime',
      sorter: true,
    },
    {
      title: '结束时间',
      dataIndex: 'endTime',
      key: 'endTime',
      width: 150,
      valueType: 'dateTime',
      render: (text) => text || '-',
    },
    {
      title: '创建人',
      dataIndex: 'creator',
      key: 'creator',
      width: 100,
    },
    {
      title: '负责人',
      dataIndex: 'assignee',
      key: 'assignee',
      width: 100,
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      key: 'priority',
      width: 80,
      valueType: 'select',
      valueEnum: {
        '高': { text: '高', status: 'error' },
        '中': { text: '中', status: 'warning' },
        '低': { text: '低', status: 'default' },
      },
    },
    {
      title: '风险等级',
      dataIndex: 'risk',
      key: 'risk',
      width: 100,
      valueType: 'select',
      valueEnum: {
        '高风险': { text: '高风险', status: 'error' },
        '中风险': { text: '中风险', status: 'warning' },
        '低风险': { text: '低风险', status: 'default' },
      },
    },
    {
      title: '工单分类',
      dataIndex: 'category',
      key: 'category',
      width: 100,
    },
    {
      title: '影响服务',
      dataIndex: 'affectedServices',
      key: 'affectedServices',
      width: 150,
      ellipsis: true,
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      fixed: 'right',
      render: (_, record) => (
        <div>
          <a onClick={() => handleView(record)}>
            查看
          </a>
        </div>
      ),
    },
  ];

  // 自动化变更列定义
  const autoChangeColumns: ProColumns<AutoChangeRecord>[] = [
    {
      title: '变更ID',
      dataIndex: 'id',
      key: 'id',
      width: 120,
      ellipsis: true,
    },
    {
      title: '变更标题',
      dataIndex: 'title',
      key: 'title',
      width: 200,
      ellipsis: true,
    },
    {
      title: '自动化类型',
      dataIndex: 'automationType',
      key: 'automationType',
      width: 120,
      valueType: 'select',
      valueEnum: {
        '脚本执行': { text: '脚本执行', status: 'processing' },
        '配置推送': { text: '配置推送', status: 'warning' },
        '服务重启': { text: '服务重启', status: 'error' },
        '数据同步': { text: '数据同步', status: 'default' },
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      valueType: 'select',
      valueEnum: {
        '进行中': { text: '进行中', status: 'processing' },
        '已完成': { text: '已完成', status: 'success' },
        '已取消': { text: '已取消', status: 'default' },
        '待审核': { text: '待审核', status: 'warning' },
      },
    },
    {
      title: '环境',
      dataIndex: 'environment',
      key: 'environment',
      width: 100,
      valueType: 'select',
      valueEnum: {
        '生产环境': { text: '生产环境', status: 'error' },
        '测试环境': { text: '测试环境', status: 'warning' },
        '开发环境': { text: '开发环境', status: 'default' },
      },
    },
    {
      title: '开始时间',
      dataIndex: 'startTime',
      key: 'startTime',
      width: 150,
      valueType: 'dateTime',
      sorter: true,
    },
    {
      title: '结束时间',
      dataIndex: 'endTime',
      key: 'endTime',
      width: 150,
      valueType: 'dateTime',
      render: (text) => text || '-',
    },
    {
      title: '创建人',
      dataIndex: 'creator',
      key: 'creator',
      width: 100,
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      key: 'priority',
      width: 80,
      valueType: 'select',
      valueEnum: {
        '高': { text: '高', status: 'error' },
        '中': { text: '中', status: 'warning' },
        '低': { text: '低', status: 'default' },
      },
    },
    {
      title: '风险等级',
      dataIndex: 'risk',
      key: 'risk',
      width: 100,
      valueType: 'select',
      valueEnum: {
        '高风险': { text: '高风险', status: 'error' },
        '中风险': { text: '中风险', status: 'warning' },
        '低风险': { text: '低风险', status: 'default' },
      },
    },
    {
      title: '脚本名称',
      dataIndex: 'scriptName',
      key: 'scriptName',
      width: 150,
    },
    {
      title: '执行模式',
      dataIndex: 'executionMode',
      key: 'executionMode',
      width: 100,
    },
    {
      title: '目标服务器',
      dataIndex: 'targetServers',
      key: 'targetServers',
      width: 150,
      ellipsis: true,
    },
    {
      title: '成功率',
      dataIndex: 'successRate',
      key: 'successRate',
      width: 100,
      render: (text) => text ? `${text}%` : '-',
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      fixed: 'right',
      render: (_, record) => (
        <div>
          <a onClick={() => handleView(record)}>
            查看
          </a>
        </div>
      ),
    },
  ];

  // 处理查看详情
  const handleView = (record: any) => {
    message.info(`查看详情: ${record.title}`);
  };

  // 获取统计数据
  const loadDashboardData = async () => {
    setLoading(true);
    try {
      await fetchDashboardData();
    } catch (error) {
      console.error('获取统计数据失败:', error);
      message.error('获取统计数据失败');
    } finally {
      setLoading(false);
    }
  };

  // 当切换到dashboard tab时加载数据
  useEffect(() => {
    if (activeTab === 'dashboard') {
      loadDashboardData();
    }
  }, [activeTab]);

  // 发布清单API请求
  const fetchReleaseData = async (params: any) => {
    try {
      const queryParams: QueryParams = {
        startTime: params.startTime,
        endTime: params.endTime,
        status: params.status,
        environment: params.environment,
        page: params.current,
        pageSize: params.pageSize,
        searchKeyword: params.title || params.creator,
      };

      const result = await fetchChangeData(queryParams, 'release/list');
      return {
        data: result.data || [],
        success: result.success,
        total: result.total || 0,
      };
    } catch (error) {
      console.error('获取发布清单数据失败:', error);
      return {
        data: [],
        success: false,
        total: 0,
      };
    }
  };

  // 运维变更API请求
  const fetchOperationChangeData = async (params: any) => {
    try {
      const queryParams: QueryParams = {
        startTime: params.startTime,
        endTime: params.endTime,
        status: params.status,
        environment: params.environment,
        page: params.current,
        pageSize: params.pageSize,
        searchKeyword: params.title || params.creator,
      };

      const result = await fetchChangeData(queryParams, 'workorder/list');
      return {
        data: result.data || [],
        success: result.success,
        total: result.total || 0,
      };
    } catch (error) {
      console.error('获取运维变更数据失败:', error);
      return {
        data: [],
        success: false,
        total: 0,
      };
    }
  };

  // 自动化变更API请求
  const fetchAutoChangeData = async (params: any) => {
    try {
      const queryParams: QueryParams = {
        startTime: params.startTime,
        endTime: params.endTime,
        status: params.status,
        environment: params.environment,
        page: params.current,
        pageSize: params.pageSize,
        searchKeyword: params.title || params.creator,
      };

      const result = await fetchChangeData(queryParams, 'autochange/list');
      return {
        data: result.data || [],
        success: result.success,
        total: result.total || 0,
      };
    } catch (error) {
      console.error('获取自动化变更数据失败:', error);
      return {
        data: [],
        success: false,
        total: 0,
      };
    }
  };

  const tabItems = [
    {
      key: 'dashboard',
      label: '变更发布统计',
      children: (
        <div>
          <Alert
            message="统计信息"
            description="显示今日和全年的变更统计数据，以及最近7天的趋势分析"
            type="info"
            showIcon
            style={{ marginBottom: '16px' }}
          />

          {/* 今日统计 */}
          <Row gutter={16} style={{ marginBottom: '16px' }}>
            <Col span={6}>
              <Card>
                <Statistic
                  title="今日发布数"
                  value={dashboardData.todayRelease}
                  valueStyle={{ color: '#1890ff' }}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic
                  title="今日工单数"
                  value={dashboardData.todayWorkOrder}
                  valueStyle={{ color: '#52c41a' }}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic
                  title="今日自动化变更数"
                  value={dashboardData.todayAutoChange}
                  valueStyle={{ color: '#722ed1' }}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic
                  title="今日故障数"
                  value={dashboardData.todayFault}
                  valueStyle={{ color: '#fa8c16' }}
                />
              </Card>
            </Col>
          </Row>

          {/* 全年统计 */}
          <Row gutter={16} style={{ marginBottom: '16px' }}>
            <Col span={6}>
              <Card>
                <Statistic
                  title="全年发布数"
                  value={dashboardData.yearRelease}
                  valueStyle={{ color: '#1890ff' }}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic
                  title="全年工单数"
                  value={dashboardData.yearWorkOrder}
                  valueStyle={{ color: '#52c41a' }}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic
                  title="全年自动化变更数"
                  value={dashboardData.yearAutoChange}
                  valueStyle={{ color: '#722ed1' }}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic
                  title="全年故障数"
                  value={dashboardData.yearFault}
                  valueStyle={{ color: '#fa8c16' }}
                />
              </Card>
            </Col>
          </Row>

          {/* 趋势图表 */}
          <Card title="近7日变更趋势" style={{ marginBottom: '16px' }}>
            <div style={{ height: '400px' }}>
              <Column {...chartConfig} data={getAllTrendData()} />
            </div>
          </Card>

          {/* 各类型趋势图表 */}
          <Row gutter={16}>
            <Col span={12}>
              <Card title="发布趋势">
                <div style={{ height: '300px' }}>
                  <Column 
                    {...chartConfig} 
                    data={processTrendData(dashboardData.releaseTrend, '发布')}
                    seriesField="type"
                    isGroup={false}
                  />
                </div>
              </Card>
            </Col>
            <Col span={12}>
              <Card title="工单趋势">
                <div style={{ height: '300px' }}>
                  <Column 
                    {...chartConfig} 
                    data={processTrendData(dashboardData.workOrderTrend, '工单')}
                    seriesField="type"
                    isGroup={false}
                  />
                </div>
              </Card>
            </Col>
          </Row>

          <Row gutter={16} style={{ marginTop: '16px' }}>
            <Col span={12}>
              <Card title="自动化变更趋势">
                <div style={{ height: '300px' }}>
                  <Column 
                    {...chartConfig} 
                    data={processTrendData(dashboardData.autoChangeTrend, '自动化变更')}
                    seriesField="type"
                    isGroup={false}
                  />
                </div>
              </Card>
            </Col>
            <Col span={12}>
              <Card title="故障趋势">
                <div style={{ height: '300px' }}>
                  <Column 
                    {...chartConfig} 
                    data={processTrendData(dashboardData.faultTrend, '故障')}
                    seriesField="type"
                    isGroup={false}
                  />
                </div>
              </Card>
            </Col>
          </Row>
        </div>
      ),
    },
    {
      key: 'release',
      label: '发布清单查询',
      children: (
        <div>
          <Alert
            message="发布清单查询"
            description="查询发布类型的变更记录，支持按时间、状态、环境等条件筛选"
            type="info"
            showIcon
            style={{ marginBottom: '16px' }}
          />
          <ProTable<ReleaseRecord>
            headerTitle="发布清单"
            actionRef={releaseActionRef}
            rowKey="id"
            search={{
              labelWidth: 120,
            }}
            toolBarRender={() => [
              <a key="export" onClick={() => message.info('导出功能')}>
                导出
              </a>,
            ]}
            request={fetchReleaseData}
            columns={releaseColumns}
            pagination={{
              pageSize: 10,
              showSizeChanger: true,
              showQuickJumper: true,
            }}
          />
        </div>
      ),
    },
    {
      key: 'operationChange',
      label: '运维变更查询',
      children: (
        <div>
          <Alert
            message="运维变更查询"
            description="查询工单类型的变更记录，支持按时间、状态、环境等条件筛选"
            type="info"
            showIcon
            style={{ marginBottom: '16px' }}
          />
          <ProTable<WorkOrderRecord>
            headerTitle="运维变更"
            actionRef={operationChangeActionRef}
            rowKey="id"
            search={{
              labelWidth: 120,
            }}
            toolBarRender={() => [
              <a key="export" onClick={() => message.info('导出功能')}>
                导出
              </a>,
            ]}
            request={fetchOperationChangeData}
            columns={workOrderColumns}
            pagination={{
              pageSize: 10,
              showSizeChanger: true,
              showQuickJumper: true,
            }}
          />
        </div>
      ),
    },
    {
      key: 'autoChange',
      label: '自动化变更查询',
      children: (
        <div>
          <Alert
            message="自动化变更查询"
            description="查询自动化变更类型的记录，支持按时间、状态、环境等条件筛选"
            type="info"
            showIcon
            style={{ marginBottom: '16px' }}
          />
          <ProTable<AutoChangeRecord>
            headerTitle="自动化变更"
            actionRef={autoChangeActionRef}
            rowKey="id"
            search={{
              labelWidth: 120,
            }}
            toolBarRender={() => [
              <a key="export" onClick={() => message.info('导出功能')}>
                导出
              </a>,
            ]}
            request={fetchAutoChangeData}
            columns={autoChangeColumns}
            pagination={{
              pageSize: 10,
              showSizeChanger: true,
              showQuickJumper: true,
            }}
          />
        </div>
      ),
    },
  ];

  return (
    <PageContainer>
      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        items={tabItems}
        style={{ backgroundColor: '#fff', padding: '16px' }}
      />
    </PageContainer>
  );
};

export default ChangeQuery; 