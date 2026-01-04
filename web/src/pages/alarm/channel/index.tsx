import { PlusOutlined, FormOutlined, DeleteOutlined } from '@ant-design/icons';
import { Button, Divider, message, Popconfirm, Select } from 'antd';
import React, { useState, useRef } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import ProTable, { ProColumns, ActionType } from '@ant-design/pro-table';
import CreateForm from './components/CreateForm';
import UpdateForm from './components/UpdateForm';
import { TableListItem } from './data.d';
import { queryChannel, updateChannel, addChannel, removeChannel } from './service';
import { useAccess } from 'umi';

/**
 * 添加节点
 * @param fields
 */
const handleAdd = async (fields: TableListItem) => {
  const hide = message.loading('正在添加');
  try {
    await addChannel({ ...fields });
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
const handleUpdate = async (fields: FormValueType, id: number) => {
  const hide = message.loading('正在配置');
  try {
    await updateChannel({
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
 *  删除节点
 * @param selectedRows
 */
const handleRemove = async (id: number) => {
  const hide = message.loading('正在删除');
  try {
    await removeChannel({
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

function OptionComponent(props: any) {
  return <option {...props}>{props.children}</option>;
}

const formInitValue = { "name": "", "description": "", "enable": "", "webhook_enable": "", "webhook_url": "", "mail_enable": "", "sms_enable": "", "phone_enable": "", "wechat_enable": "", "mail_list": "", "sms_list": "", "phone_list": "", "wechat_list": "" }

const TableList: React.FC<{}> = () => {
  const [createModalVisible, handleModalVisible] = useState<boolean>(false);
  const [updateModalVisible, handleUpdateModalVisible] = useState<boolean>(false);
  const [formValues, setFormValues] = useState(formInitValue);
  const actionRef = useRef<ActionType>();
  const access = useAccess();

  const columns: ProColumns<TableListItem>[] = [
    {
      title: '渠道名称',
      dataIndex: 'name',
      hideInForm: false,
      sorter: true,
      formItemProps: {
        rules: [
          {
            required: true,
            message: '此项为必填项',
          },
        ],
      },
      initialValue: formValues.name,
    },
    {
      title: '描述',
      dataIndex: 'description',
      valueType: 'textarea',
      search: false,
      initialValue: formValues.description,
    },

    {
      title: '邮件通知',
      dataIndex: 'mail_enable',
      filters: true,
      onFilter: true,
      valueEnum: {
        0: { text: '关闭', status: 'Default' },
        1: { text: '开启', status: 'Success' },
      },
      sorter: true,
      initialValue: formValues.mail_enable,
      tip: '开启前请先确保邮件网关配置正确，否则无法收取邮件',
      formItemProps: {
        rules: [
          {
            required: true,
            message: '此项为必填项',
          },
        ],
      },
      renderFormItem: (_, { type, defaultRender, ...rest }, form) => {
        return (
          <Select>
            <OptionComponent value={0}>关闭</OptionComponent>
            <OptionComponent value={1}>开启</OptionComponent>
          </Select>
        );
      },
    },
    {
      title: '邮件地址',
      dataIndex: 'mail_list',
      tip: '多个邮箱号使用英文分号分隔',
      ellipsis: true,
      copyable: true,
      initialValue: formValues.mail_list,
      hideInTable: true,
    },
    {
      title: '短信通知',
      dataIndex: 'sms_enable',
      filters: true,
      onFilter: true,
      valueEnum: {
        0: { text: '关闭', status: 'Default' },
        1: { text: '开启', status: 'Success' },
      },
      sorter: true,
      initialValue: formValues.sms_enable,
      tip: '开启前请先确保阿里云短信网关配置正确，否则无法收取短信',
      formItemProps: {
        rules: [
          {
            required: true,
            message: '此项为必填项',
          },
        ],
      },
      renderFormItem: (_, { type, defaultRender, ...rest }, form) => {
        return (
          <Select>
            <OptionComponent value={0}>关闭</OptionComponent>
            <OptionComponent value={1}>开启</OptionComponent>
          </Select>
        );
      },
    },
    {
      title: '短信地址',
      dataIndex: 'sms_list',
      tip: '多个手机号使用英文分号分隔',
      ellipsis: true,
      copyable: true,
      initialValue: formValues.sms_list,
      hideInTable: true,
    },
    {
      title: '微信通知',
      dataIndex: 'wechat_enable',
      filters: true,
      onFilter: true,
      valueEnum: {
        0: { text: '关闭', status: 'Default' },
        1: { text: '开启', status: 'Success' },
      },
      sorter: true,
      initialValue: formValues.sms_enable,
      tip: '开启前请先确保微信服务号配置正确，否则无法收取微信订阅通知',
      formItemProps: {
        rules: [
          {
            required: true,
            message: '此项为必填项',
          },
        ],
      },
      renderFormItem: (_, { type, defaultRender, ...rest }, form) => {
        return (
          <Select>
            <OptionComponent value={0}>关闭</OptionComponent>
            <OptionComponent value={1}>开启</OptionComponent>
          </Select>
        );
      },
    },
    {
      title: '微信地址',
      dataIndex: 'wechat_list',
      tip: '多个微信号使用英文分号分隔',
      ellipsis: true,
      copyable: true,
      initialValue: formValues.wechat_list,
      hideInTable: true,
    },
    {
      title: '电话通知',
      dataIndex: 'phone_enable',
      filters: true,
      onFilter: true,
      valueEnum: {
        0: { text: '关闭', status: 'Default' },
        1: { text: '开启', status: 'Success' },
      },
      sorter: true,
      initialValue: formValues.phone_enable,
      tip: '开启前请先确保阿里云电话网关配置正确，否则无法接听电话',
      formItemProps: {
        rules: [
          {
            required: true,
            message: '此项为必填项',
          },
        ],
      },
      renderFormItem: (_, { type, defaultRender, ...rest }, form) => {
        return (
          <Select>
            <OptionComponent value={0}>关闭</OptionComponent>
            <OptionComponent value={1}>开启</OptionComponent>
          </Select>
        );
      },
    },
    {
      title: '电话地址',
      dataIndex: 'phone_list',
      tip: '多个手机号使用英文分号分隔',
      ellipsis: true,
      copyable: true,
      initialValue: formValues.phone_list,
      hideInTable: true,
    },
    {
      title: 'WebHook通知',
      dataIndex: 'webhook_enable',
      filters: true,
      onFilter: true,
      valueEnum: {
        0: { text: '关闭', status: 'Default' },
        1: { text: '开启', status: 'Success' },
      },
      sorter: true,
      initialValue: formValues.webhook_enable,
      tip: '开启前请先确保接收通知的WEBURL配置正确，否则无法接收通知',
      formItemProps: {
        rules: [
          {
            required: true,
            message: '此项为必填项',
          },
        ],
      },
      renderFormItem: (_, { type, defaultRender, ...rest }, form) => {
        return (
          <Select>
            <OptionComponent value={0}>关闭</OptionComponent>
            <OptionComponent value={1}>开启</OptionComponent>
          </Select>
        );
      },
    },
    {
      title: 'WebHook地址',
      dataIndex: 'webhook_url',
      tip: '接收POST数据的URL地址',
      ellipsis: true,
      copyable: true,
      initialValue: formValues.webhook_url,
      hideInTable: true,
    },
    {
      title: '状态',
      dataIndex: 'enable',
      filters: true,
      onFilter: true,
      valueEnum: {
        0: { text: '禁用', status: 'Default' },
        1: { text: '启用', status: 'Success' },
      },
      sorter: true,
      initialValue: formValues.enable,
      formItemProps: {
        rules: [
          {
            required: true,
            message: '此项为必填项',
          },
        ],
      },
      renderFormItem: (_, { type, defaultRender, ...rest }, form) => {
        return (
          <Select>
            <OptionComponent value={0}>禁用</OptionComponent>
            <OptionComponent value={1}>启用</OptionComponent>
          </Select>
        );
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
            <FormOutlined />修改
          </a>
          <Divider type="vertical" />
          <Popconfirm
            title={`确认要删除数据【${record.name}】,删除后不可恢复，是否继续？`}
            placement={"left"}
            onConfirm={async () => {
              if (!access.canAdmin) { message.error('操作权限受限，请联系平台管理员'); return }
              const success = await handleRemove(record.id);
              if (success) {
                if (actionRef.current) {
                  actionRef.current.reload();
                }
              }
            }}
          >
            <a><DeleteOutlined />删除</a>
          </Popconfirm>
        </>
      ),
    },
  ];

  return (
    <PageContainer>
      <ProTable<TableListItem>
        headerTitle="数据列表"
        actionRef={actionRef}
        rowKey="id"
        search={{
          labelWidth: 120,
        }}
        toolBarRender={() => [
          <Button type="primary"
            onClick={() => {
              handleModalVisible(true);
              setFormValues(formInitValue);
            }}
          >
            <PlusOutlined /> 新建
          </Button>,
        ]}
        request={(params, sorter, filter) => queryChannel({ ...params, sorter, filter })}
        columns={columns}
        pagination={{
          pageSize: 10,
        }}
      />

      <CreateForm onCancel={() => handleModalVisible(false)} modalVisible={createModalVisible}>
        <ProTable<TableListItem, TableListItem>
          onSubmit={async (value) => {
            if (!access.canAdmin) { message.error('操作权限受限，请联系平台管理员'); return }
            const success = await handleAdd(value);
            if (success) {
              handleModalVisible(false);
              if (actionRef.current) {
                actionRef.current.reload();
              }
            }
          }}
          rowKey="id"
          type="form"
          columns={columns}
        />
      </CreateForm>

      <UpdateForm onCancel={() => handleUpdateModalVisible(false)} updateModalVisible={updateModalVisible}>
        <ProTable<TableListItem, TableListItem>
          onSubmit={async (value) => {
            if (!access.canAdmin) { message.error('操作权限受限，请联系平台管理员'); return }
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
          columns={columns}
        />
      </UpdateForm>

    </PageContainer>
  );
};

export default TableList;
