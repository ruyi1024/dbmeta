import React, { useState } from 'react';
import { Card } from 'antd';
import ProTable, { ProColumns } from '@ant-design/pro-table';

// 数据表容量数据类型
interface TableCapacity {
  id: number;
  databaseName: string;
  tableName: string;
  datasourceType: string;
  host?: string;
  port?: string;
  dataSize: string;
  dataSizeBytes: number;
  rowCount: number;
  dataSizeIncr: string;
  dataSizeIncrBytes: number;
  rowCountIncr: number;
}

const TableCapacity: React.FC = () => {
  const [loading] = useState<boolean>(false);

  // 数据表容量信息查询表格列定义
  const tableCapacityColumns: ProColumns<TableCapacity>[] = [
    {
      title: '数据库名',
      dataIndex: 'databaseName',
      width: 180,
      ellipsis: true,
      sorter: true,
      fieldProps: {
        placeholder: '请输入数据库名',
      },
    },
    {
      title: '表名',
      dataIndex: 'tableName',
      width: 200,
      ellipsis: true,
      sorter: true,
      fieldProps: {
        placeholder: '请输入表名',
      },
    },
    {
      title: '数据库类型',
      dataIndex: 'datasourceType',
      width: 120,
      sorter: true,
      fieldProps: {
        placeholder: '请输入数据库类型',
      },
    },
    {
      title: '主机',
      dataIndex: 'host',
      width: 150,
      ellipsis: true,
      fieldProps: {
        placeholder: '请输入主机',
      },
    },
    {
      title: '端口',
      dataIndex: 'port',
      width: 80,
      fieldProps: {
        placeholder: '请输入端口',
      },
    },
    {
      title: '数据存储大小',
      dataIndex: 'dataSize',
      width: 130,
      sorter: true,
      render: (text) => <span style={{ color: '#1890ff', fontWeight: 500 }}>{text}</span>,
      search: false,
    },
    {
      title: '数据记录条数',
      dataIndex: 'rowCount',
      width: 130,
      sorter: true,
      render: (text) => {
        const count = Number(text) || 0;
        if (count === 0) return '0';
        if (count >= 1000000) {
          return `${(count / 1000000).toFixed(2)}M`;
        } else if (count >= 1000) {
          return `${(count / 1000).toFixed(2)}K`;
        } else {
          return count.toLocaleString();
        }
      },
      search: false,
    },
    {
      title: '数据存储日增长',
      dataIndex: 'dataSizeIncr',
      width: 140,
      sorter: true,
      render: (text, record) => {
        const bytes = record.dataSizeIncrBytes || 0;
        if (bytes === 0) return <span>0 B</span>;
        const icon = bytes > 0 ? <span style={{ color: '#52c41a', marginRight: 4, fontSize: 12 }}>↑</span> : <span style={{ color: '#ff4d4f', marginRight: 4, fontSize: 12 }}>↓</span>;
        return <span style={{ color: '#000', fontWeight: 500 }}>{icon}{text}</span>;
      },
      search: false,
    },
    {
      title: '数据记录日增长',
      dataIndex: 'rowCountIncr',
      width: 140,
      sorter: true,
      render: (text) => {
        const count = Number(text) || 0;
        if (count === 0) return '0';
        const icon = count > 0 ? <span style={{ color: '#52c41a', marginRight: 4, fontSize: 12 }}>↑</span> : <span style={{ color: '#ff4d4f', marginRight: 4, fontSize: 12 }}>↓</span>;
        const formatted = Math.abs(count) >= 1000000 
          ? `${(count / 1000000).toFixed(2)}M` 
          : count >= 1000 
          ? `${(count / 1000).toFixed(2)}K` 
          : count.toString();
        return <span style={{ color: '#000', fontWeight: 500 }}>{icon}{count > 0 ? '+' : ''}{formatted}</span>;
      },
      search: false,
    },
  ];

  // 获取数据表容量数据（用于表格，支持分页、搜索、排序）
  const fetchTableCapacity = async (params: any, sorter: any) => {
    try {
      // 构建查询参数
      const queryParams = new URLSearchParams();
      queryParams.append('current', params.current || '1');
      queryParams.append('pageSize', params.pageSize || '10');
      
      // 添加搜索条件
      if (params.databaseName) {
        queryParams.append('databaseName', params.databaseName);
      }
      if (params.tableName) {
        queryParams.append('tableName', params.tableName);
      }
      if (params.datasourceType) {
        queryParams.append('datasourceType', params.datasourceType);
      }
      if (params.host) {
        queryParams.append('host', params.host);
      }
      if (params.port) {
        queryParams.append('port', params.port);
      }

      // 处理排序
      if (sorter && Object.keys(sorter).length > 0) {
        const sortField = Object.keys(sorter)[0];
        const sortOrder = sorter[sortField] === 'ascend' ? 'asc' : 'desc';
        queryParams.append('sortField', sortField);
        queryParams.append('sortOrder', sortOrder);
      }

      const response = await fetch(`/api/v1/pumpkin/capacity/table/growth?${queryParams.toString()}`);
      const json = await response.json();
      if (json.success && json.data) {
        const data: TableCapacity[] = json.data.map((item: any) => {
          const dataSizeBytes = typeof item.dataSizeBytes === 'number' ? item.dataSizeBytes : 0;
          const dataSizeIncrBytes = typeof item.dataSizeIncrBytes === 'number' ? item.dataSizeIncrBytes : 0;
          return {
            id: item.id,
            databaseName: item.databaseName,
            tableName: item.tableName,
            datasourceType: item.datasourceType,
            host: item.host || '',
            port: item.port || '',
            dataSize: item.dataSize,
            dataSizeBytes: dataSizeBytes,
            rowCount: item.rowCount || 0,
            dataSizeIncr: item.dataSizeIncr || '0 B',
            dataSizeIncrBytes: dataSizeIncrBytes,
            rowCountIncr: item.rowCountIncr || 0,
          };
        });
        return {
          data,
          success: true,
          total: json.total || 0,
        };
      }
    } catch (error) {
      console.error('获取数据表容量数据失败:', error);
    }
    return {
      data: [],
      success: true,
      total: 0,
    };
  };

  return (
    <div>
      <Card
        title="数据表容量信息查询"
        style={{ marginBottom: 24 }}
      >
        <ProTable<TableCapacity>
          headerTitle={false}
          search={{
            labelWidth: 120,
          }}
          toolBarRender={false}
          rowKey="id"
          loading={loading}
          request={(params, sorter) => fetchTableCapacity(params, sorter)}
          columns={tableCapacityColumns}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total: number) => `共 ${total} 条`,
          }}
          size="middle"
        />
      </Card>
    </div>
  );
};

export default TableCapacity;

