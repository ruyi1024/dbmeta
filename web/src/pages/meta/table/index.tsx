import React, { useState, useRef, useEffect } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import ProTable, { ProColumns, ActionType } from '@ant-design/pro-table';
import { Badge, Button, message, Popconfirm } from 'antd';
import { CheckOutlined, CloseOutlined } from '@ant-design/icons';
import { TableListItem } from './data.d';
import { queryTable, batchUpdateAiFixed } from './service';

const TableList: React.FC<{}> = () => {
  const actionRef = useRef<ActionType>();
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [selectedRows, setSelectedRows] = useState<TableListItem[]>([]);

  // 批量操作：不应用AI注释
  const handleBatchNotApply = async () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请先选择要操作的表');
      return;
    }

    try {
      const ids = selectedRowKeys.map(key => Number(key));
      await batchUpdateAiFixed({ ids, ai_fixed: 1 });
      message.success(`成功将 ${selectedRowKeys.length} 个表的AI注释状态设置为"不应用"`);
      setSelectedRowKeys([]);
      setSelectedRows([]);
      if (actionRef.current) {
        actionRef.current.reload();
      }
    } catch (error) {
      message.error('批量操作失败，请重试');
      console.error('Batch update error:', error);
    }
  };

  // 批量操作：应用AI注释
  const handleBatchApply = async () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请先选择要操作的表');
      return;
    }

    try {
      const ids = selectedRowKeys.map(key => Number(key));
      await batchUpdateAiFixed({ ids, ai_fixed: 2 });
      message.success(`成功将 ${selectedRowKeys.length} 个表的AI注释状态设置为"待应用"`);
      setSelectedRowKeys([]);
      setSelectedRows([]);
      if (actionRef.current) {
        actionRef.current.reload();
      }
    } catch (error) {
      message.error('批量操作失败，请重试');
      console.error('Batch update error:', error);
    }
  };

  const rowSelection = {
    selectedRowKeys,
    onChange: (keys: React.Key[], rows: TableListItem[]) => {
      setSelectedRowKeys(keys);
      setSelectedRows(rows);
    },
    getCheckboxProps: (record: TableListItem) => ({
      // 只有有AI注释的记录才能被选择
      disabled: !record.ai_comment || record.ai_comment === '',
    }),
  };

  const columns: ProColumns<TableListItem>[] = [
    {
      title: '数据表名',
      dataIndex: 'table_name',
      sorter: true,
    },
    {
      title: '表类型',
      dataIndex: 'table_type',
      hideInSearch: true,
    },
    {
      title: '表字符集',
      dataIndex: 'characters',
      hideInSearch: true,
    },
    {
      title: '表备注',
      dataIndex: 'table_comment',
      hideInSearch: true,
    },
    {
      title: 'AI注释生成',
      dataIndex: 'ai_comment',
      hideInSearch: true,
      render: (text) => {
        if (!text || text === '') {
          return <span style={{ color: '#999' }}>暂无AI注释</span>;
        }
        return text;
      },
    },
    {
      title: 'AI注释应用',
      dataIndex: 'ai_fixed',
      hideInSearch: true,
      valueType: 'select',
      valueEnum: {
        0: { text: '待审核', status: 'Default' },
        1: { text: '不应用', status: 'Error' },
        2: { text: '待应用', status: 'Warning' },
        3: { text: '已应用', status: 'Success' },
      },
      render: (text, record) => {
        if (record.ai_fixed === 0) {
          return <Badge status="default" text="待审核" />;
        } else if (record.ai_fixed === 1) {
          return <Badge status="error" text="不应用" />;
        } else if (record.ai_fixed === 2) {
          return <Badge status="warning" text="待应用" />;
        } else if (record.ai_fixed === 3) {
          return <Badge status="success" text="已应用" />;
        } else {
          return <Badge status="default" text="未知" />;
        }
      },
    },
    {
      title: '所属数据库',
      dataIndex: 'database_name',
      sorter: true,
    },
    {
      title: '数据库类型',
      dataIndex: 'datasource_type',
      sorter: true,
    },
    {
      title: '所属主机',
      dataIndex: 'host',
    },
    {
      title: '所属端口',
      dataIndex: 'port',
    },
    {
      title: '创建时间',
      dataIndex: 'gmt_created',
      sorter: true,
      valueType: 'dateTime',
      hideInForm: true,
      search: false,
    },
    {
      title: '修改时间',
      dataIndex: 'gmt_updated',
      sorter: true,
      valueType: 'dateTime',
      hideInForm: true,
      search: false,
    },
  ];

  return (
    <PageContainer>
      <ProTable<TableListItem>
        headerTitle="数据表列表"
        actionRef={actionRef}
        rowKey="id"
        search={{
          labelWidth: 120,
        }}
        request={(params, sorter, filter) => queryTable({ ...params, sorter, filter: filter as { [key: string]: any[] } })}
        columns={columns}
        pagination={{
          pageSize: 10,
        }}
        rowSelection={rowSelection}
        toolBarRender={() => [
          <Button
            type="default"
            danger
            onClick={handleBatchNotApply}
            disabled={selectedRowKeys.length === 0}
            icon={<CloseOutlined />}
          >
            不应用AI注释 ({selectedRowKeys.length})
          </Button>,
          <Button
            type="primary"
            onClick={handleBatchApply}
            disabled={selectedRowKeys.length === 0}
            icon={<CheckOutlined />}
          >
            应用AI注释 ({selectedRowKeys.length})
          </Button>,
        ]}
      />
    </PageContainer>
  );
};

export default TableList;
