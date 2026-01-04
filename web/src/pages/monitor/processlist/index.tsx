import React, { useState, useEffect, useRef } from 'react';
import { Card, Table, Select, Button, Space, message, Tag, Tooltip } from 'antd';
import { ReloadOutlined, PauseCircleOutlined, PlayCircleOutlined } from '@ant-design/icons';
import { getDatasourceList, getProcessList } from './service';

// 类型定义
interface ProcessItem {
  id: number;
  user: string;
  host: string;
  db: string;
  command: string;
  time: number;
  state: string;
  info: string;
}

interface DatasourceItem {
  id: number;
  name: string;
  host: string;
  port: string;
}

const ProcessListMonitor: React.FC = () => {
  const [datasources, setDatasources] = useState<DatasourceItem[]>([]);
  const [selectedDatasource, setSelectedDatasource] = useState<number>();
  const [processList, setProcessList] = useState<ProcessItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(false);
  const [refreshInterval, setRefreshInterval] = useState(5); // 秒
  const intervalRef = useRef<NodeJS.Timeout>();

  // 获取数据源列表
  const fetchDatasources = async () => {
    try {
      const response = await getDatasourceList();
      if (response.success) {
        setDatasources(response.data);
      }
    } catch (error) {
      message.error('获取数据源列表失败');
    }
  };

  // 获取进程列表
  const fetchProcessList = async () => {
    if (!selectedDatasource) return;
    
    setLoading(true);
    try {
      const response = await getProcessList({ datasource_id: selectedDatasource });
      
      if (response.success) {
        setProcessList(response.data);
      } else {
        message.error(response.msg || '获取进程列表失败');
      }
    } catch (error) {
      message.error('获取进程列表失败');
    } finally {
      setLoading(false);
    }
  };

  // 自动刷新
  useEffect(() => {
    if (autoRefresh && selectedDatasource) {
      intervalRef.current = setInterval(fetchProcessList, refreshInterval * 1000);
    } else {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    }
    
    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    };
  }, [autoRefresh, selectedDatasource, refreshInterval]);

  // 组件挂载时获取数据源
  useEffect(() => {
    fetchDatasources();
  }, []);

  // 数据源选择变化时获取数据
  useEffect(() => {
    if (selectedDatasource) {
      fetchProcessList();
    }
  }, [selectedDatasource]);

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '用户',
      dataIndex: 'user',
      key: 'user',
      width: 120,
    },
    {
      title: '主机',
      dataIndex: 'host',
      key: 'host',
      width: 150,
    },
    {
      title: '数据库',
      dataIndex: 'db',
      key: 'db',
      width: 120,
    },
    {
      title: '命令',
      dataIndex: 'command',
      key: 'command',
      width: 100,
      render: (command: string) => {
        const colorMap: { [key: string]: string } = {
          'Query': 'blue',
          'Sleep': 'green',
          'Connect': 'orange',
          'Binlog Dump': 'purple',
        };
        return <Tag color={colorMap[command] || 'default'}>{command}</Tag>;
      },
    },
    {
      title: '时间(秒)',
      dataIndex: 'time',
      key: 'time',
      width: 100,
      render: (time: number) => {
        const color = time > 60 ? 'red' : time > 10 ? 'orange' : 'green';
        return <Tag color={color}>{time}</Tag>;
      },
    },
    {
      title: '状态',
      dataIndex: 'state',
      key: 'state',
      width: 200,
      render: (state: string) => (
        <Tooltip title={state}>
          <span style={{ 
            display: 'inline-block', 
            maxWidth: '180px', 
            overflow: 'hidden', 
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap'
          }}>
            {state || 'NULL'}
          </span>
        </Tooltip>
      ),
    },
    {
      title: 'SQL语句',
      dataIndex: 'info',
      key: 'info',
      render: (info: string) => (
        <Tooltip title={info}>
          <span style={{ 
            display: 'inline-block', 
            maxWidth: '300px', 
            overflow: 'hidden', 
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap'
          }}>
            {info || 'NULL'}
          </span>
        </Tooltip>
      ),
    },
  ];

  return (
    <div style={{ padding: '24px' }}>
      <Card title="MySQL进程实时监控" style={{ marginBottom: 16 }}>
        <Space style={{ marginBottom: 16 }}>
          <span>数据库实例：</span>
          <Select
            style={{ width: 200 }}
            placeholder="请选择数据库实例"
            value={selectedDatasource}
            onChange={setSelectedDatasource}
            options={datasources.map((ds: any) => ({
              label: `${ds.name} (${ds.host}:${ds.port})`,
              value: ds.id
            }))}
          />
          
          <Button
            icon={<ReloadOutlined />}
            onClick={fetchProcessList}
            loading={loading}
            disabled={!selectedDatasource}
          >
            刷新
          </Button>
          
          <Button
            icon={autoRefresh ? <PauseCircleOutlined /> : <PlayCircleOutlined />}
            type={autoRefresh ? 'default' : 'primary'}
            onClick={() => setAutoRefresh(!autoRefresh)}
            disabled={!selectedDatasource}
          >
            {autoRefresh ? '停止自动刷新' : '开始自动刷新'}
          </Button>
          
          {autoRefresh && (
            <Select
              style={{ width: 120 }}
              value={refreshInterval}
              onChange={setRefreshInterval}
              options={[
                { label: '5秒', value: 5 },
                { label: '10秒', value: 10 },
                { label: '30秒', value: 30 },
                { label: '60秒', value: 60 },
              ]}
            />
          )}
        </Space>
        
        <Table
          columns={columns}
          dataSource={processList}
          rowKey="id"
          loading={loading}
          pagination={{
            pageSize: 50,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 个进程`,
          }}
          scroll={{ x: 1200 }}
          size="small"
        />
      </Card>
    </div>
  );
};

export default ProcessListMonitor;
