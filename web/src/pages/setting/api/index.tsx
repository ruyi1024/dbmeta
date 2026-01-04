import { message } from 'antd';
import React, { useState, useRef } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import ProTable, { ProColumns, ActionType } from '@ant-design/pro-table';
import CreateForm from './components/CreateForm';
import UpdateForm from './components/UpdateForm';
import { TableListItem } from './data.d';
import { query, update, add, remove } from './service';

/**
 * 添加节点
 * @param fields
 */
const handleAdd = async (fields: TableListItem) => {
  const hide = message.loading('正在添加');
  try {
    await add({ ...fields });
    hide();
    message.success('添加成功');
    return true;
  } catch (error) {
    hide();
    message.error('添加失败请重试！');
    return false;
  }
};

/**
 * 更新节点
 * @param fields
 */
const handleUpdate = async (fields: TableListItem, id: number) => {
  const hide = message.loading('正在配置');
  try {
    await update({
      ...fields,
      "id": id,
    });
    hide();
    message.success('修改成功');
    return true;
  } catch (error) {
    hide();
    message.error('修改失败请重试！');
    return false;
  }
};

/**
 * 删除节点
 * @param selectedRows
 */
const handleRemove = async (selectedRows: TableListItem[]) => {
  const hide = message.loading('正在删除');
  if (!selectedRows) return true;
  try {
    await remove({ key: selectedRows.map((row) => row.id) });
    hide();
    message.success('删除成功，即将刷新');
    return true;
  } catch (error) {
    hide();
    message.error('删除失败，请重试');
    return false;
  }
};

const TableList: React.FC = () => {
  const [createModalVisible, handleModalVisible] = useState<boolean>(false);
  const [updateModalVisible, handleUpdateModalVisible] = useState<boolean>(false);
  const [stepFormValues, setStepFormValues] = useState({});
  const actionRef = useRef<ActionType>();

  const columns: ProColumns<TableListItem>[] = [
    {
      title: 'ID',
      dataIndex: 'id',
      tip: 'ID是唯一的 key',
      width: 60,
      hideInSearch: true,
    },
    {
      title: 'API名称',
      dataIndex: 'api_name',
      width: 150,
      ellipsis: true,
      render: (text, record) => (
        <span title={record.api_description}>{text}</span>
      ),
    },
    {
      title: 'API URL',
      dataIndex: 'api_url',
      width: 200,
      ellipsis: true,
      render: (text) => (
        <a href={String(text)} target="_blank" rel="noopener noreferrer" title={String(text)}>
          {String(text)}
        </a>
      ),
    },
    {
      title: '协议',
      dataIndex: 'protocol',
      width: 80,
      valueEnum: {
        HTTP: { text: 'HTTP', status: 'Default' },
        HTTPS: { text: 'HTTPS', status: 'Success' },
      },
      render: (_, record) => (
        <span style={{ 
          color: record.protocol === 'HTTPS' ? '#52c41a' : '#1890ff',
          padding: '2px 8px',
          borderRadius: '4px',
          backgroundColor: record.protocol === 'HTTPS' ? '#f6ffed' : '#e6f7ff',
          border: `1px solid ${record.protocol === 'HTTPS' ? '#b7eb8f' : '#91d5ff'}`
        }}>
          {record.protocol}
        </span>
      ),
    },
    {
      title: '请求方法',
      dataIndex: 'method',
      width: 100,
      valueEnum: {
        GET: { text: 'GET', status: 'Success' },
        POST: { text: 'POST', status: 'Processing' },
        PUT: { text: 'PUT', status: 'Warning' },
        DELETE: { text: 'DELETE', status: 'Error' },
      },
      render: (_, record) => {
        const colorMap = {
          GET: '#52c41a',
          POST: '#1890ff',
          PUT: '#fa8c16',
          DELETE: '#ff4d4f',
        };
        const bgColorMap = {
          GET: '#f6ffed',
          POST: '#e6f7ff',
          PUT: '#fff7e6',
          DELETE: '#fff2f0',
        };
        const borderColorMap = {
          GET: '#b7eb8f',
          POST: '#91d5ff',
          PUT: '#ffd591',
          DELETE: '#ffccc7',
        };
        return (
          <span style={{ 
            color: colorMap[record.method as keyof typeof colorMap],
            padding: '2px 8px',
            borderRadius: '4px',
            backgroundColor: bgColorMap[record.method as keyof typeof bgColorMap],
            border: `1px solid ${borderColorMap[record.method as keyof typeof borderColorMap]}`
          }}>
            {record.method}
          </span>
        );
      },
    },
    {
      title: '认证类型',
      dataIndex: 'auth_type',
      width: 100,
      valueEnum: {
        NONE: { text: '无认证', status: 'Default' },
        BASIC: { text: 'Basic', status: 'Processing' },
        BEARER: { text: 'Bearer', status: 'Success' },
        API_KEY: { text: 'API Key', status: 'Warning' },
      },
      render: (_, record) => {
        const colorMap = {
          NONE: '#d9d9d9',
          BASIC: '#1890ff',
          BEARER: '#52c41a',
          API_KEY: '#fa8c16',
        };
        const bgColorMap = {
          NONE: '#fafafa',
          BASIC: '#e6f7ff',
          BEARER: '#f6ffed',
          API_KEY: '#fff7e6',
        };
        const borderColorMap = {
          NONE: '#d9d9d9',
          BASIC: '#91d5ff',
          BEARER: '#b7eb8f',
          API_KEY: '#ffd591',
        };
        return (
          <span style={{ 
            color: colorMap[record.auth_type as keyof typeof colorMap],
            padding: '2px 8px',
            borderRadius: '4px',
            backgroundColor: bgColorMap[record.auth_type as keyof typeof bgColorMap],
            border: `1px solid ${borderColorMap[record.auth_type as keyof typeof borderColorMap]}`
          }}>
            {record.auth_type}
          </span>
        );
      },
    },
    {
      title: '期望返回码',
      dataIndex: 'expected_codes',
      width: 120,
      hideInSearch: true,
      render: (text) => (
        <span style={{ 
          color: '#1890ff',
          padding: '2px 8px',
          borderRadius: '4px',
          backgroundColor: '#e6f7ff',
          border: '1px solid #91d5ff'
        }}>
          {text}
        </span>
      ),
    },
    {
      title: '超时时间',
      dataIndex: 'timeout',
      width: 100,
      hideInSearch: true,
      render: (text) => `${text}s`,
    },
    {
      title: '重试次数',
      dataIndex: 'retry_count',
      width: 100,
      hideInSearch: true,
      render: (text) => text || 0,
    },
    {
      title: '状态',
      dataIndex: 'enable',
      width: 80,
      valueEnum: {
        0: { text: '禁用', status: 'Error' },
        1: { text: '启用', status: 'Success' },
      },
      render: (_, record) => (
        <span style={{ 
          color: record.enable ? '#52c41a' : '#ff4d4f',
          padding: '2px 8px',
          borderRadius: '4px',
          backgroundColor: record.enable ? '#f6ffed' : '#fff2f0',
          border: `1px solid ${record.enable ? '#b7eb8f' : '#ffccc7'}`
        }}>
          {record.enable ? '启用' : '禁用'}
        </span>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'gmt_created',
      width: 150,
      valueType: 'dateTime',
      hideInSearch: true,
    },
    {
      title: '操作',
      dataIndex: 'option',
      valueType: 'option',
      width: 150,
      render: (_, record) => [
        <a
          key="config"
          onClick={() => {
            setStepFormValues(record);
            handleUpdateModalVisible(true);
          }}
        >
          编辑
        </a>,
        <a
          key="delete"
          style={{ color: 'red' }}
          onClick={() => {
            if (window.confirm('确定要删除这个API配置吗？')) {
              handleRemove([record]);
            }
          }}
        >
          删除
        </a>,
      ],
    },
  ];

  return (
    <PageContainer>
      <ProTable<TableListItem>
        headerTitle="API接口配置"
        actionRef={actionRef}
        rowKey="id"
        search={{
          labelWidth: 120,
        }}
        toolBarRender={() => [
          <button
            key="primary"
            type="button"
            style={{
              backgroundColor: '#1890ff',
              color: 'white',
              border: 'none',
              padding: '8px 16px',
              borderRadius: '6px',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              gap: '4px'
            }}
            onClick={() => {
              console.log('新建按钮被点击');
              setStepFormValues({});
              handleModalVisible(true);
              console.log('createModalVisible应该为true:', true);
            }}
          >
            + 新建API配置
          </button>,
        ]}
        request={async (params, sorter, filter) => {
          const { data, success } = await query({
            currentPage: params.current,
            pageSize: params.pageSize,
            api_name: params.api_name,
            protocol: params.protocol,
            enable: params.enable,
          });
          return {
            data: data?.list || [],
            success,
            total: data?.total || 0,
          };
        }}
        columns={columns}
        rowSelection={{
          onChange: (_, selectedRows) => {
            console.log('selectedRows', selectedRows);
          },
        }}
      />
      <CreateForm
        onSubmit={async (value) => {
          const success = await handleAdd(value);
          if (success) {
            handleModalVisible(false);
            setStepFormValues({});
            if (actionRef.current) {
              actionRef.current.reload();
            }
          }
        }}
        onCancel={() => {
          handleModalVisible(false);
          setStepFormValues({});
        }}
        updateModalVisible={createModalVisible}
        values={stepFormValues}
      />
      <UpdateForm
        onSubmit={async (value) => {
          const success = await handleUpdate(value, (stepFormValues as any).id);
          if (success) {
            handleUpdateModalVisible(false);
            setStepFormValues({});
            if (actionRef.current) {
              actionRef.current.reload();
            }
          }
        }}
        onCancel={() => {
          handleUpdateModalVisible(false);
          setStepFormValues({});
        }}
        updateModalVisible={updateModalVisible}
        values={stepFormValues}
      />
    </PageContainer>
  );
};

export default TableList;