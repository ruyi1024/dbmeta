import React, { useState, useRef, useEffect } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import ProTable, { ProColumns, ActionType } from '@ant-design/pro-table';
import { TableListItem } from './data.d';
import { query } from './service';


const TableList: React.FC<{}> = () => {

  const actionRef = useRef<ActionType>();

  const columns: ProColumns<TableListItem>[] = [
    {
      title: '实例类型',
      dataIndex: 'datasource_type',
      sorter: true,
    },

    {
      title: '主机',
      dataIndex: 'host',
      sorter: false,
    },
    {
      title: '端口',
      dataIndex: 'port',
      sorter: false,
    },
    {
      title: '数据库',
      dataIndex: 'database_name',
    },
    {
      title: '数据表',
      dataIndex: 'table_name',
    },
    {
      title: '表备注',
      dataIndex: 'table_comment',
      search: false,
    },
    {
      title: '数据列',
      dataIndex: 'column_name',
    },
    {
      title: '列备注',
      dataIndex: 'column_comment',
      search: false,
    },
    {
      title: '采集类型',
      dataIndex: 'rule_type',
    },
    {
      title: '敏感规则',
      dataIndex: 'rule_key',
    },
    {
      title: '敏感说明',
      dataIndex: 'rule_name',
      search: false,
    },
    {
      title: '级别',
      dataIndex: 'level',
      filters: true,
      onFilter: true,
      valueEnum: {
        0: { text: '低敏', status: 'warning' },
        1: { text: '高敏', status: 'error' },
      },
      sorter: true,
      search: false,
      formItemProps: {
        rules: [
          {
            required: true,
            message: '此项为必填项',
          },
        ],
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      filters: true,
      onFilter: true,
      valueEnum: {
        '-1': { text: '疑似敏感', status: 'warning' },
        0: { text: '非敏感', status: 'default' },
        1: { text: '确认敏感', status: 'success' },
      },
      sorter: true,
      search: false,
      formItemProps: {
        rules: [
          {
            required: true,
            message: '此项为必填项',
          },
        ],
      },
    },
    {
      title: '初采日期',
      dataIndex: 'gmt_created',
      sorter: true,
      valueType: 'date',
      hideInForm: true,
      search: false,
    },
    {
      title: '复采日期',
      dataIndex: 'gmt_updated',
      sorter: true,
      valueType: 'date',
      hideInForm: true,
      search: false,
    },

  ];

  return (
    <PageContainer>
      <ProTable<TableListItem>
        headerTitle=""
        actionRef={actionRef}
        rowKey="id"
        search={{
          labelWidth: 120,
        }}
        request={(params, sorter, filter) => query({ ...params, sorter, filter })}
        columns={columns}
        pagination={{
          pageSize: 15,
        }}
      />
    </PageContainer>
  );
};

export default TableList;
