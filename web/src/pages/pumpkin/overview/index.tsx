import React, { useState, useMemo, useEffect } from 'react';
import { Row, Col, Card } from 'antd';
import { StatisticCard } from '@ant-design/pro-components';
import { 
  DatabaseOutlined, 
  TableOutlined, 
  HddOutlined, 
  RiseOutlined 
} from '@ant-design/icons';
import { Column } from '@ant-design/plots';

// 数据容量统计数据类型
interface CapacityStats {
  totalDatabases: number;
  totalTables: number;
  totalDataSize: string;
  totalRows: number;
  dailyGrowth: string;
  dailyGrowthRows: number;
}

// 数据库容量Top10数据类型
interface DatabaseCapacity {
  id: number;
  databaseName: string;
  datasourceType: string;
  host?: string;
  port?: string;
  dataSize: string;
  dataSizeBytes: number;
  tableCount: number;
  rowCount: number;
  indexSize: string;
  indexSizeBytes: number;
}

// 数据表容量Top10数据类型
interface TableCapacity {
  id: number;
  tableName: string;
  databaseName: string;
  datasourceType: string;
  dataSize: string;
  dataSizeBytes: number;
  rowCount: number;
  indexSize: string;
  indexSizeBytes: number;
  avgRowLength: string;
}

const Overview: React.FC = () => {
  const [databaseChartDataState, setDatabaseChartDataState] = useState<DatabaseCapacity[]>([]);
  const [tableData, setTableData] = useState<TableCapacity[]>([]);
  const [tableFragmentationData, setTableFragmentationData] = useState<any[]>([]);
  const [tableRowsData, setTableRowsData] = useState<any[]>([]);
  const [statsData, setStatsData] = useState<CapacityStats>({
    totalDatabases: 0,
    totalTables: 0,
    totalDataSize: '0 B',
    totalRows: 0,
    dailyGrowth: '0 B',
    dailyGrowthRows: 0,
  });

  // 将字节数转换为GB数值（用于图表显示）
  const bytesToGB = (bytes: number): number => {
    return bytes / (1024 * 1024 * 1024);
  };

  // 准备数据库容量图表数据
  const databaseChartData = useMemo(() => {
    if (!databaseChartDataState || databaseChartDataState.length === 0) {
      return [];
    }
    const sortedData = [...databaseChartDataState].sort((a, b) => (b.dataSizeBytes || 0) - (a.dataSizeBytes || 0));
    const chartData = sortedData.slice(0, 10).map((item) => {
      const bytes = item.dataSizeBytes || 0;
      return {
        name: item.databaseName.length > 10 ? item.databaseName.substring(0, 10) + '...' : item.databaseName,
        value: bytes > 0 ? bytesToGB(bytes) : 0,
        valueBytes: bytes,
        fullName: item.databaseName,
        datasourceType: item.datasourceType,
        host: item.host || '',
        port: item.port || '',
        dataSize: item.dataSize,
      };
    });
    return chartData;
  }, [databaseChartDataState]);

  // 准备数据表容量图表数据
  const tableChartData = useMemo(() => {
    if (!tableData || tableData.length === 0) {
      return [];
    }
    const sortedData = [...tableData].sort((a, b) => (b.dataSizeBytes || 0) - (a.dataSizeBytes || 0));
    const chartData = sortedData.slice(0, 10).map((item) => {
      const bytes = item.dataSizeBytes || 0;
      return {
        name: item.tableName.length > 10 ? item.tableName.substring(0, 10) + '...' : item.tableName,
        value: bytes > 0 ? bytesToGB(bytes) : 0,
        valueBytes: bytes,
        fullName: item.tableName,
      };
    });
    return chartData;
  }, [tableData]);

  // 准备表碎片率图表数据
  const tableFragmentationChartData = useMemo(() => {
    if (!tableFragmentationData || tableFragmentationData.length === 0) {
      return [];
    }
    const sortedData = [...tableFragmentationData].sort((a, b) => (b.fragmentationRateValue || 0) - (a.fragmentationRateValue || 0));
    return sortedData.slice(0, 10).map((item: any) => ({
      name: item.tableName.length > 10 ? item.tableName.substring(0, 10) + '...' : item.tableName,
      value: item.fragmentationRateValue || 0,
      fullName: item.tableName,
      databaseName: item.databaseName,
      datasourceType: item.datasourceType,
      host: item.host || '',
      port: item.port || '',
      fragmentationRate: item.fragmentationRate,
    }));
  }, [tableFragmentationData]);

  // 准备表记录数图表数据
  const tableRowsChartData = useMemo(() => {
    if (!tableRowsData || tableRowsData.length === 0) {
      return [];
    }
    const sortedData = [...tableRowsData].sort((a, b) => (b.rowCountValue || 0) - (a.rowCountValue || 0));
    return sortedData.slice(0, 10).map((item: any) => ({
      name: item.tableName.length > 10 ? item.tableName.substring(0, 10) + '...' : item.tableName,
      value: item.rowCountValue || 0,
      fullName: item.tableName,
      databaseName: item.databaseName,
      datasourceType: item.datasourceType,
      host: item.host || '',
      port: item.port || '',
      rowCount: item.rowCount,
    }));
  }, [tableRowsData]);

  // 获取统计数据
  const fetchStats = async () => {
    try {
      const response = await fetch('/api/v1/pumpkin/capacity/stats');
      if (!response.ok) {
        console.error('获取统计数据失败: HTTP', response.status, response.statusText);
        return;
      }
      const json = await response.json();
      if (json.success && json.data) {
        setStatsData({
          totalDatabases: Number(json.data.totalDatabases) || 0,
          totalTables: Number(json.data.totalTables) || 0,
          totalDataSize: json.data.totalDataSize || '0 B',
          totalRows: Number(json.data.totalRows) || 0,
          dailyGrowth: json.data.dailyGrowth || '0 B',
          dailyGrowthRows: Number(json.data.dailyGrowthRows) || 0,
        });
      } else {
        console.error('统计数据格式错误:', json);
      }
    } catch (error) {
      console.error('获取统计数据失败:', error);
    }
  };

  // 获取数据库容量Top10数据（用于图表）
  const fetchDatabaseCapacityChart = async () => {
    try {
      const response = await fetch('/api/v1/pumpkin/capacity/database/top10/chart');
      const json = await response.json();
      if (json.success && json.data) {
        const data: DatabaseCapacity[] = json.data.map((item: any, index: number) => {
          const dataSizeBytes = typeof item.dataSizeBytes === 'number' ? item.dataSizeBytes : 0;
          return {
            id: index + 1,
            databaseName: item.databaseName,
            datasourceType: item.datasourceType || '',
            host: item.host || '',
            port: item.port || '',
            dataSize: item.dataSize,
            dataSizeBytes: dataSizeBytes,
            tableCount: 0,
            rowCount: 0,
            indexSize: '',
            indexSizeBytes: 0,
          };
        });
        setDatabaseChartDataState(data);
      }
    } catch (error) {
      console.error('获取数据库容量图表数据失败:', error);
      setDatabaseChartDataState([]);
    }
  };

  // 获取数据表容量Top10数据
  const fetchTableCapacity = async () => {
    try {
      const response = await fetch('/api/v1/pumpkin/capacity/table/top10');
      const json = await response.json();
      if (json.success && json.data) {
        const data: TableCapacity[] = json.data.map((item: any, index: number) => {
          const dataSizeBytes = typeof item.dataSizeBytes === 'number' ? item.dataSizeBytes : 0;
          const indexSizeBytes = typeof item.indexSizeBytes === 'number' ? item.indexSizeBytes : 0;
          return {
            id: index + 1,
            tableName: item.tableName,
            databaseName: item.databaseName,
            datasourceType: item.datasourceType,
            dataSize: item.dataSize,
            dataSizeBytes: dataSizeBytes,
            rowCount: item.rowCount,
            indexSize: item.indexSize,
            indexSizeBytes: indexSizeBytes,
            avgRowLength: item.avgRowLength,
          };
        });
        setTableData(data);
      }
    } catch (error) {
      console.error('获取数据表容量数据失败:', error);
      setTableData([]);
    }
  };

  // 获取表碎片率Top10数据
  const fetchTableFragmentation = async () => {
    try {
      const response = await fetch('/api/v1/pumpkin/capacity/table/fragmentation/top10');
      const json = await response.json();
      if (json.success && json.data) {
        setTableFragmentationData(json.data);
      }
    } catch (error) {
      console.error('获取表碎片率数据失败:', error);
      setTableFragmentationData([]);
    }
  };

  // 获取表记录数Top10数据
  const fetchTableRows = async () => {
    try {
      const response = await fetch('/api/v1/pumpkin/capacity/table/rows/top10');
      const json = await response.json();
      if (json.success && json.data) {
        setTableRowsData(json.data);
      }
    } catch (error) {
      console.error('获取表记录数数据失败:', error);
      setTableRowsData([]);
    }
  };

  // 初始化数据
  useEffect(() => {
    const initData = async () => {
      await fetchStats();
      await fetchDatabaseCapacityChart();
      await fetchTableCapacity();
      await fetchTableFragmentation();
      await fetchTableRows();
    };
    initData();
  }, []);

  return (
    <div>
      {/* 数据统计卡片 */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} md={4}>
          <Card>
            <StatisticCard
              statistic={{
                title: '数据库数量',
                value: statsData.totalDatabases,
                prefix: <DatabaseOutlined style={{ color: '#1890ff' }} />,
                valueStyle: { color: '#1890ff' },
              }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={4}>
          <Card>
            <StatisticCard
              statistic={{
                title: '数据表数量',
                value: statsData.totalTables,
                prefix: <TableOutlined style={{ color: '#52c41a' }} />,
                valueStyle: { color: '#52c41a' },
              }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={4}>
          <Card>
            <StatisticCard
              statistic={{
                title: '总数据容量',
                value: statsData.totalDataSize,
                prefix: <HddOutlined style={{ color: '#faad14' }} />,
                valueStyle: { color: '#faad14' },
              }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={4}>
          <Card>
            <StatisticCard
              statistic={{
                title: '总数据记录',
                value: statsData.totalRows.toLocaleString(),
                prefix: <TableOutlined style={{ color: '#722ed1' }} />,
                valueStyle: { color: '#722ed1' },
              }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={4}>
          <Card>
            <StatisticCard
              statistic={{
                title: '天增长数据量',
                value: statsData.dailyGrowth,
                prefix: <RiseOutlined style={{ color: '#f5222d' }} />,
                valueStyle: { color: '#f5222d' },
              }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={4}>
          <Card>
            <StatisticCard
              statistic={{
                title: '天增长记录数',
                value: statsData.dailyGrowthRows.toLocaleString(),
                prefix: <RiseOutlined style={{ color: '#52c41a' }} />,
                valueStyle: { color: '#52c41a' },
              }}
            />
          </Card>
        </Col>
      </Row>

      {/* 容量排行柱状图 */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} lg={12}>
          <Card title="数据库容量TOP排行">
            <Column
              data={databaseChartData}
              xField="name"
              yField="value"
              columnStyle={{ radius: [4, 4, 0, 0] }}
              color="#1890ff"
              label={{
                position: 'top',
                formatter: (datum: any) => {
                  const value = datum?.value ?? 0;
                  if (value === 0 || isNaN(value)) return '0.00 GB';
                  return `${value.toFixed(2)} GB`;
                },
              }}
              tooltip={{
                customContent: (title, items) => {
                  if (!items || items.length === 0) return '';
                  const item = items[0];
                  const data = item?.data as any;
                  return `<div style="padding: 8px;">
                    <div style="margin-bottom: 6px; font-weight: 500; font-size: 14px;">${data?.fullName || title}</div>
                    <div style="margin-bottom: 4px;">数据库类型: <span style="color: #1890ff;">${data?.datasourceType || '-'}</span></div>
                    <div style="margin-bottom: 4px;">主机: <span style="color: #1890ff;">${data?.host || '-'}</span></div>
                    <div style="margin-bottom: 4px;">端口: <span style="color: #1890ff;">${data?.port || '-'}</span></div>
                    <div style="margin-top: 6px; padding-top: 6px; border-top: 1px solid #e8e8e8;">
                      容量: <span style="color: #1890ff; font-weight: 500; font-size: 14px;">${data?.dataSize || '-'}</span>
                    </div>
                  </div>`;
                },
              }}
              xAxis={{
                label: {
                  autoRotate: true,
                  autoHide: true,
                  autoEllipsis: true,
                },
              }}
              yAxis={{
                label: {
                  formatter: (v: string) => {
                    const num = parseFloat(v);
                    if (isNaN(num)) return '0.00 GB';
                    return `${num.toFixed(2)} GB`;
                  },
                },
                title: {
                  text: '容量 (GB)',
                  style: {
                    fontSize: 12,
                  },
                },
              }}
              height={300}
            />
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card title="数据表容量TOP排行">
            <Column
              data={tableChartData}
              xField="name"
              yField="value"
              columnStyle={{ radius: [4, 4, 0, 0] }}
              color="#52c41a"
              label={{
                position: 'top',
                formatter: (datum: any) => {
                  const value = datum?.value ?? 0;
                  if (value === 0 || isNaN(value)) return '0.00 GB';
                  return `${value.toFixed(2)} GB`;
                },
              }}
              tooltip={{
                customContent: (title, items) => {
                  if (!items || items.length === 0) return '';
                  const item = items[0];
                  const data = item?.data as any;
                  const bytes = data?.valueBytes || (data?.value * 1024 * 1024 * 1024);
                  const formatBytes = (b: number): string => {
                    if (b < 1024) return `${b} B`;
                    if (b < 1024 * 1024) return `${(b / 1024).toFixed(2)} KB`;
                    if (b < 1024 * 1024 * 1024) return `${(b / (1024 * 1024)).toFixed(2)} MB`;
                    if (b < 1024 * 1024 * 1024 * 1024) return `${(b / (1024 * 1024 * 1024)).toFixed(2)} GB`;
                    return `${(b / (1024 * 1024 * 1024 * 1024)).toFixed(2)} TB`;
                  };
                  return `<div style="padding: 8px;">
                    <div style="margin-bottom: 4px;">${data?.fullName || title}</div>
                    <div>数据大小: <span style="color: #52c41a; font-weight: 500;">${formatBytes(bytes)}</span></div>
                  </div>`;
                },
              }}
              xAxis={{
                label: {
                  autoRotate: true,
                  autoHide: true,
                  autoEllipsis: true,
                },
              }}
              yAxis={{
                label: {
                  formatter: (v: string) => {
                    const num = parseFloat(v);
                    if (isNaN(num)) return '0.00 GB';
                    return `${num.toFixed(2)} GB`;
                  },
                },
                title: {
                  text: '容量 (GB)',
                  style: {
                    fontSize: 12,
                  },
                },
              }}
              height={300}
            />
          </Card>
        </Col>
      </Row>

      {/* 表碎片率和记录数排行柱状图 */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} lg={12}>
          <Card title="表碎片率TOP排行">
            <Column
              data={tableFragmentationChartData}
              xField="name"
              yField="value"
              columnStyle={{ radius: [4, 4, 0, 0] }}
              color="#faad14"
              label={{
                position: 'top',
                formatter: (datum: any) => {
                  const value = datum?.value ?? 0;
                  if (value === 0 || isNaN(value)) return '0.00%';
                  return `${value.toFixed(2)}%`;
                },
              }}
              tooltip={{
                customContent: (title, items) => {
                  if (!items || items.length === 0) return '';
                  const item = items[0];
                  const data = item?.data as any;
                  return `<div style="padding: 8px;">
                    <div style="margin-bottom: 6px; font-weight: 500; font-size: 14px;">${data?.fullName || title}</div>
                    <div style="margin-bottom: 4px;">所属数据库: <span style="color: #faad14;">${data?.databaseName || '-'}</span></div>
                    <div style="margin-bottom: 4px;">数据库类型: <span style="color: #faad14;">${data?.datasourceType || '-'}</span></div>
                    <div style="margin-bottom: 4px;">主机: <span style="color: #faad14;">${data?.host || '-'}</span></div>
                    <div style="margin-bottom: 4px;">端口: <span style="color: #faad14;">${data?.port || '-'}</span></div>
                    <div style="margin-top: 6px; padding-top: 6px; border-top: 1px solid #e8e8e8;">
                      碎片率: <span style="color: #faad14; font-weight: 500; font-size: 14px;">${data?.fragmentationRate || '-'}</span>
                    </div>
                  </div>`;
                },
              }}
              xAxis={{
                label: {
                  autoRotate: true,
                  autoHide: true,
                  autoEllipsis: true,
                },
              }}
              yAxis={{
                label: {
                  formatter: (v: string) => {
                    const num = parseFloat(v);
                    if (isNaN(num)) return '0.00%';
                    return `${num.toFixed(2)}%`;
                  },
                },
                title: {
                  text: '碎片率 (%)',
                  style: {
                    fontSize: 12,
                  },
                },
              }}
              height={300}
            />
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card title="表记录数TOP排行">
            <Column
              data={tableRowsChartData}
              xField="name"
              yField="value"
              columnStyle={{ radius: [4, 4, 0, 0] }}
              color="#722ed1"
              label={{
                position: 'top',
                formatter: (datum: any) => {
                  const value = datum?.value ?? 0;
                  if (value === 0 || isNaN(value)) return '0';
                  if (value >= 1000000000) return `${(value / 1000000000).toFixed(2)}B`;
                  if (value >= 1000000) return `${(value / 1000000).toFixed(2)}M`;
                  if (value >= 1000) return `${(value / 1000).toFixed(2)}K`;
                  return `${value}`;
                },
              }}
              tooltip={{
                customContent: (title, items) => {
                  if (!items || items.length === 0) return '';
                  const item = items[0];
                  const data = item?.data as any;
                  return `<div style="padding: 8px;">
                    <div style="margin-bottom: 6px; font-weight: 500; font-size: 14px;">${data?.fullName || title}</div>
                    <div style="margin-bottom: 4px;">所属数据库: <span style="color: #722ed1;">${data?.databaseName || '-'}</span></div>
                    <div style="margin-bottom: 4px;">数据库类型: <span style="color: #722ed1;">${data?.datasourceType || '-'}</span></div>
                    <div style="margin-bottom: 4px;">主机: <span style="color: #722ed1;">${data?.host || '-'}</span></div>
                    <div style="margin-bottom: 4px;">端口: <span style="color: #722ed1;">${data?.port || '-'}</span></div>
                    <div style="margin-top: 6px; padding-top: 6px; border-top: 1px solid #e8e8e8;">
                      记录数: <span style="color: #722ed1; font-weight: 500; font-size: 14px;">${data?.rowCount || '-'}</span>
                    </div>
                  </div>`;
                },
              }}
              xAxis={{
                label: {
                  autoRotate: true,
                  autoHide: true,
                  autoEllipsis: true,
                },
              }}
              yAxis={{
                label: {
                  formatter: (v: string) => {
                    const num = parseFloat(v);
                    if (isNaN(num)) return '0';
                    if (num >= 1000000000) return `${(num / 1000000000).toFixed(2)}B`;
                    if (num >= 1000000) return `${(num / 1000000).toFixed(2)}M`;
                    if (num >= 1000) return `${(num / 1000).toFixed(2)}K`;
                    return `${num}`;
                  },
                },
                title: {
                  text: '记录数',
                  style: {
                    fontSize: 12,
                  },
                },
              }}
              height={300}
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Overview;

