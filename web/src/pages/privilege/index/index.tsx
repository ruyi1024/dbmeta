import { PlusOutlined, FormOutlined, DeleteOutlined } from '@ant-design/icons';
import { Button, Divider, message, Popconfirm } from 'antd';
import React, { useState, useRef } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import ProTable, { ProColumns, ActionType } from '@ant-design/pro-table';
import { TableListItem } from './data.d';
import { query, remove } from './service';
import { useAccess } from 'umi';


/**
 *  删除节点
 * @param selectedRows
 */
const handleRemove = async (id: number) => {
  const hide = message.loading('正在删除');
  try {
    await remove({
      "id": id,
    });
    hide();
    message.success('删除成功，即将刷新');
    return true;
  } catch (error) {
    hide();
    message.error('删除失败，请重试');
    return false;
  }
};


const TableList: React.FC<{}> = () => {
  const actionRef = useRef<ActionType>();
  const access = useAccess();


  const columns: ProColumns<TableListItem>[] = [

    {
      title: '申请账号',
      dataIndex: 'username',
      sorter: true,
    },
    {
      title: '类型',
      dataIndex: 'datasource_type',
      sorter: false,
      search: false,
    },
    {
      title: '授权方式',
      dataIndex: 'grant_type',
      filters: true,
      onFilter: true,
      valueEnum: {
        0: { text: '数据库', status: 'database' },
        1: { text: '数据表', status: 'table' },
      },
      sorter: true,
      search: false,
    },
    {
      title: '数据库',
      dataIndex: 'database_name',
      sorter: false,
    },
    {
      title: '数据表',
      dataIndex: 'table_name',
      sorter: false,
    },
    {
      title: '查询',
      dataIndex: 'do_select',
      valueEnum: {
        0: { text: '', status: 'Default' },
        1: { text: '', status: 'Success' },
      },
      search: false,
    },
    {
      title: '插入',
      dataIndex: 'do_insert',
      valueEnum: {
        0: { text: '', status: 'Default' },
        1: { text: '', status: 'Success' },
      },
      search: false,
    },
    {
      title: '更新',
      dataIndex: 'do_update',
      valueEnum: {
        0: { text: '', status: 'Default' },
        1: { text: '', status: 'Success' },
      },
      search: false,
    },
    {
      title: '删除',
      dataIndex: 'do_delete',
      valueEnum: {
        0: { text: '', status: 'Default' },
        1: { text: '', status: 'Success' },
      },
      search: false,
    },
    {
      title: '结构创建',
      dataIndex: 'do_create',
      valueEnum: {
        0: { text: '', status: 'Default' },
        1: { text: '', status: 'Success' },
      },
      search: false,
    },
    {
      title: '结构变更',
      dataIndex: 'do_alter',
      valueEnum: {
        0: { text: '', status: 'Default' },
        1: { text: '', status: 'Success' },
      },
      search: false,
    },
    {
      title: '查询上限',
      dataIndex: 'max_select',
      search: false,
    },
    {
      title: '更新上限',
      dataIndex: 'max_update',
      search: false,
    },
    {
      title: '删除上限',
      dataIndex: 'max_delete',
      search: false,
    },

    {
      title: '状态',
      dataIndex: 'enable',
      filters: true,
      onFilter: true,
      valueEnum: {
        0: { text: '禁止', status: 'Default' },
        1: { text: '正常', status: 'Success' },
      },
      sorter: true,
      search: false,
    },
    {
      title: '到期日期',
      dataIndex: 'expire_date',
      sorter: true,
      valueType: 'date',
      hideInForm: true,
      search: false,
    },
    {
      title: '授权日期',
      dataIndex: 'gmt_created',
      sorter: true,
      valueType: 'date',
      hideInForm: true,
      hideInTable: false,
      search: false,
    },

  ];

  return (
    <PageContainer>
      <ProTable<TableListItem>
        headerTitle="数据列表"
        actionRef={actionRef}
        rowKey="id"
        search={{
          labelWidth: 100,
        }}
        request={(params, sorter, filter) => query({ ...params, sorter, filter })}
        columns={columns}
      />


    </PageContainer>
  );
};

export default TableList;
