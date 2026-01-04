import { Button, Divider, message } from 'antd';
import React, { useState, useRef } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import ProTable, { ProColumns, ActionType } from '@ant-design/pro-table';
import UpdateForm from './components/UpdateForm';
import { TableListItem } from './data.d';
import { queryDatabase, updateDatabase } from './service';
import { useAccess } from 'umi';

const tableProps = {
  layout: 'horizontal',
  formItemLayout: {
    labelCol: {
      xs: { span: 24 },
      sm: { span: 4 },
    },
    wrapperCol: {
      xs: { span: 24 },
      sm: { span: 20 },
    },
  },
}

/**
 * 更新节点
 * @param fields
 */
const handleUpdate = async (fields: TableListItem, id: number) => {
  const hide = message.loading('正在修改');
  try {
    // 确保is_deleted字段是数字类型
    const updateData = {
      ...fields,
      id: id,
      is_deleted: parseInt(fields.is_deleted as any, 10) || 0
    };
    
    console.log('Sending update request with data:', updateData);
    const response = await updateDatabase(updateData);
    console.log('Update response:', response);
    hide();
    message.success('修改成功');
    return true;
  } catch (error: any) {
    console.error('Update error:', error);
    hide();
    message.error(`修改失败请重试！错误信息: ${error?.message || error}`);
    return false;
  }
};

const formInitValue = {
  id: 0,
  datasource_type: '',
  host: '',
  port: '',
  database_name: '',
  alias_name: '',
  characters: '',
  app_name: '',
  app_desc: '',
  app_owner: '',
  app_owner_email: '',
  app_owner_phone: '',
  is_deleted: 0,
};

const TableList: React.FC<{}> = () => {
  const [updateModalVisible, handleUpdateModalVisible] = useState<boolean>(false);
  const [formValues, setFormValues] = useState(formInitValue);
  const actionRef = useRef<ActionType>();
  const access = useAccess();

  const columns: ProColumns<TableListItem>[] = [
    {
      title: '数据库名',
      dataIndex: 'database_name',
      sorter: true,
    },
    {
      title: '数据库别名',
      dataIndex: 'alias_name',
      hideInSearch: true,
    },
    {
      title: '库字符集',
      dataIndex: 'characters',
      hideInSearch: true,
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
      title: '应用名称',
      dataIndex: 'app_name',
      hideInSearch: true,
    },
    {
      title: '应用描述',
      dataIndex: 'app_desc',
      hideInSearch: true,
    },
    {
      title: '应用负责人',
      dataIndex: 'app_owner',
      hideInSearch: true,
    },
    {
      title: '负责人邮箱',
      dataIndex: 'app_owner_email',
      hideInSearch: true,
    },
    {
      title: '负责人电话',
      dataIndex: 'app_owner_phone',
      hideInSearch: true,
    },
    {
      title: '是否删除',
      dataIndex: 'is_deleted',
      hideInSearch: true,
      valueEnum: {
        0: { text: '否', status: 'Default' },
        1: { text: '是', status: 'Error' },
      },
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
    {
      title: '操作',
      dataIndex: 'option',
      valueType: 'option',
      render: (_, record) => (
        <>
          <a
            onClick={() => {
              handleUpdateModalVisible(true);
              setFormValues(record);
            }}
          >
            修改业务信息
          </a>
        </>
      ),
    },
  ];

  return (
    <PageContainer>
      <ProTable<TableListItem>
        {...tableProps}
        headerTitle="数据库列表"
        actionRef={actionRef}
        rowKey="id"
        search={{
          labelWidth: 120,
        }}
        request={(params, sorter, filter) => queryDatabase({ ...params, sorter, filter: filter as { [key: string]: any[] } })}
        columns={columns}
        pagination={{
          pageSize: 10,
        }}
      />

      <UpdateForm
        onCancel={() => handleUpdateModalVisible(false)}
        updateModalVisible={updateModalVisible}
      >
        <ProTable<TableListItem, TableListItem>
          onSubmit={async (value) => {
            if (!access.canAdmin) {
              message.error('操作权限受限，请联系平台管理员');
              return;
            }
            const success = await handleUpdate(value, formValues.id);
            if (success) {
              handleUpdateModalVisible(false);
              if (actionRef.current) {
                actionRef.current.reload();
              }
            }
          }}
          rowKey="id"
          type="form"
          columns={[
            {
              title: '数据库别名',
              dataIndex: 'alias_name',
              initialValue: formValues.alias_name,
            },
            {
              title: '应用名称',
              dataIndex: 'app_name',
              initialValue: formValues.app_name,
            },
            {
              title: '应用描述',
              dataIndex: 'app_desc',
              initialValue: formValues.app_desc,
            },
            {
              title: '应用负责人',
              dataIndex: 'app_owner',
              initialValue: formValues.app_owner,
            },
            {
              title: '负责人邮箱',
              dataIndex: 'app_owner_email',
              initialValue: formValues.app_owner_email,
            },
            {
              title: '负责人电话',
              dataIndex: 'app_owner_phone',
              initialValue: formValues.app_owner_phone,
            },
            {
              title: '是否删除',
              dataIndex: 'is_deleted',
              initialValue: formValues.is_deleted,
              valueType: 'select',
              valueEnum: {
                0: { text: '否', status: 'Default' },
                1: { text: '是', status: 'Error' },
              },
            },
          ]}
        />
      </UpdateForm>
    </PageContainer>
  );
};

export default TableList;
