import React, { useState, useEffect, useRef } from 'react';
import { request } from "@@/plugin-request/request";

const ProcessListMonitor: React.FC = () => {
  const [datasources, setDatasources] = useState<any[]>([]);
  const [selectedDatasource, setSelectedDatasource] = useState<number>();
  const [processList, setProcessList] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(false);
  const [refreshInterval, setRefreshInterval] = useState(5);
  const intervalRef = useRef<NodeJS.Timeout>();

  // 获取数据源列表
  const fetchDatasources = async () => {
    try {
      const response = await request('/api/v1/datasource/list');
      if (response.success) {
        setDatasources(response.data);
      }
    } catch (error) {
      console.error('获取数据源列表失败:', error);
    }
  };

  // 获取进程列表
  const fetchProcessList = async () => {
    if (!selectedDatasource) return;
    
    setLoading(true);
    try {
      const response = await request('/api/v1/monitor/processlist', {
        method: 'POST',
        data: { datasource_id: selectedDatasource }
      });
      
      if (response.success) {
        setProcessList(response.data);
      } else {
        console.error('获取进程列表失败:', response.msg);
      }
    } catch (error) {
      console.error('获取进程列表失败:', error);
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

  const getCommandColor = (command: string) => {
    const colorMap: { [key: string]: string } = {
      'Query': '#1890ff',
      'Sleep': '#52c41a',
      'Connect': '#fa8c16',
      'Binlog Dump': '#722ed1',
    };
    return colorMap[command] || '#d9d9d9';
  };

  const getTimeColor = (time: number) => {
    if (time > 60) return '#f5222d';
    if (time > 10) return '#fa8c16';
    return '#52c41a';
  };

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ 
        background: '#fff', 
        padding: '24px', 
        borderRadius: '6px', 
        boxShadow: '0 1px 3px rgba(0,0,0,0.12)',
        marginBottom: '16px'
      }}>
        <h2 style={{ margin: '0 0 16px 0' }}>MySQL进程实时监控</h2>
        
        <div style={{ marginBottom: '16px' }}>
          <label style={{ marginRight: '8px' }}>数据库实例：</label>
          <select 
            style={{ 
              width: '200px', 
              padding: '8px', 
              marginRight: '10px',
              border: '1px solid #d9d9d9',
              borderRadius: '4px'
            }}
            value={selectedDatasource || ''}
            onChange={(e) => setSelectedDatasource(Number(e.target.value))}
          >
            <option value="">请选择数据库实例</option>
            {datasources.map((ds: any) => (
              <option key={ds.id} value={ds.id}>
                {ds.name} ({ds.host}:{ds.port})
              </option>
            ))}
          </select>
          
          <button
            style={{
              padding: '8px 16px',
              marginRight: '10px',
              backgroundColor: '#1890ff',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: loading || !selectedDatasource ? 'not-allowed' : 'pointer',
              opacity: loading || !selectedDatasource ? 0.6 : 1
            }}
            onClick={fetchProcessList}
            disabled={loading || !selectedDatasource}
          >
            {loading ? '加载中...' : '刷新'}
          </button>
          
          <button
            style={{
              padding: '8px 16px',
              marginRight: '10px',
              backgroundColor: autoRefresh ? '#d9d9d9' : '#52c41a',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: !selectedDatasource ? 'not-allowed' : 'pointer',
              opacity: !selectedDatasource ? 0.6 : 1
            }}
            onClick={() => setAutoRefresh(!autoRefresh)}
            disabled={!selectedDatasource}
          >
            {autoRefresh ? '停止自动刷新' : '开始自动刷新'}
          </button>
          
          {autoRefresh && (
            <select
              style={{ 
                width: '120px', 
                padding: '8px',
                border: '1px solid #d9d9d9',
                borderRadius: '4px'
              }}
              value={refreshInterval}
              onChange={(e) => setRefreshInterval(Number(e.target.value))}
            >
              <option value={5}>5秒</option>
              <option value={10}>10秒</option>
              <option value={30}>30秒</option>
              <option value={60}>60秒</option>
            </select>
          )}
        </div>
        
        <div style={{ 
          overflow: 'auto',
          border: '1px solid #d9d9d9',
          borderRadius: '4px'
        }}>
          <table style={{ 
            width: '100%', 
            borderCollapse: 'collapse',
            fontSize: '14px'
          }}>
            <thead>
              <tr style={{ backgroundColor: '#f5f5f5' }}>
                <th style={{ padding: '8px', border: '1px solid #d9d9d9', textAlign: 'left' }}>ID</th>
                <th style={{ padding: '8px', border: '1px solid #d9d9d9', textAlign: 'left' }}>用户</th>
                <th style={{ padding: '8px', border: '1px solid #d9d9d9', textAlign: 'left' }}>主机</th>
                <th style={{ padding: '8px', border: '1px solid #d9d9d9', textAlign: 'left' }}>数据库</th>
                <th style={{ padding: '8px', border: '1px solid #d9d9d9', textAlign: 'left' }}>命令</th>
                <th style={{ padding: '8px', border: '1px solid #d9d9d9', textAlign: 'left' }}>时间(秒)</th>
                <th style={{ padding: '8px', border: '1px solid #d9d9d9', textAlign: 'left' }}>状态</th>
                <th style={{ padding: '8px', border: '1px solid #d9d9d9', textAlign: 'left' }}>SQL语句</th>
              </tr>
            </thead>
            <tbody>
              {processList.map((process: any) => (
                <tr key={process.id}>
                  <td style={{ padding: '8px', border: '1px solid #d9d9d9' }}>{process.id}</td>
                  <td style={{ padding: '8px', border: '1px solid #d9d9d9' }}>{process.user}</td>
                  <td style={{ padding: '8px', border: '1px solid #d9d9d9' }}>{process.host}</td>
                  <td style={{ padding: '8px', border: '1px solid #d9d9d9' }}>{process.db || 'NULL'}</td>
                  <td style={{ padding: '8px', border: '1px solid #d9d9d9' }}>
                    <span style={{
                      padding: '2px 6px',
                      borderRadius: '3px',
                      color: 'white',
                      fontSize: '12px',
                      backgroundColor: getCommandColor(process.command)
                    }}>
                      {process.command}
                    </span>
                  </td>
                  <td style={{ padding: '8px', border: '1px solid #d9d9d9' }}>
                    <span style={{
                      padding: '2px 6px',
                      borderRadius: '3px',
                      color: 'white',
                      fontSize: '12px',
                      backgroundColor: getTimeColor(process.time)
                    }}>
                      {process.time}
                    </span>
                  </td>
                  <td style={{ 
                    padding: '8px', 
                    border: '1px solid #d9d9d9',
                    maxWidth: '200px',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                    title: process.state
                  }}>
                    {process.state || 'NULL'}
                  </td>
                  <td style={{ 
                    padding: '8px', 
                    border: '1px solid #d9d9d9',
                    maxWidth: '300px',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                    title: process.info
                  }}>
                    {process.info || 'NULL'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        
        {processList.length > 0 && (
          <div style={{ marginTop: '16px', color: '#666' }}>
            共 {processList.length} 个进程
          </div>
        )}
      </div>
    </div>
  );
};

export default ProcessListMonitor;
