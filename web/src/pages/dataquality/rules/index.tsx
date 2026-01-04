import React, { useRef, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import ProTable, { ProColumns, ActionType } from '@ant-design/pro-table';
import { Button, Tag, Space, message, Popconfirm, Switch } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, SettingOutlined } from '@ant-design/icons';
import { queryRules, createRule, updateRule, deleteRule } from './service';
import type { RuleListItem } from './data.d';
import CreateForm from './components/CreateForm';
import UpdateForm from './components/UpdateForm';

const RuleList: React.FC<{}> = () => {
  const actionRef = useRef<ActionType>();
  const [createModalVisible, handleCreateModalVisible] = useState<boolean>(false);
  const [updateModalVisible, handleUpdateModalVisible] = useState<boolean>(false);
  const [formValues, setFormValues] = useState<Partial<RuleListItem>>({});

  const handleAdd = async (fields: RuleListItem) => {
    try {
      const response = await createRule(fields);
      if (response.code === 200) {
        message.success('创建成功');
        handleCreateModalVisible(false);
        if (actionRef.current) {
          actionRef.current.reload();
        }
        return true;
      } else {
        message.error(response.msg || '创建失败');
        return false;
      }
    } catch (error) {
      message.error('创建失败');
      return false;
    }
  };

  const handleUpdate = async (fields: RuleListItem) => {
    try {
      const response = await updateRule(fields);
      if (response.code === 200) {
        message.success('更新成功');
        handleUpdateModalVisible(false);
        if (actionRef.current) {
          actionRef.current.reload();
        }
        return true;
      } else {
        message.error(response.msg || '更新失败');
        return false;
      }
    } catch (error) {
      message.error('更新失败');
      return false;
    }
  };

  const handleDelete = async (id: number) => {
    try {
      const response = await deleteRule(id);
      if (response.code === 200) {
        message.success('删除成功');
        if (actionRef.current) {
          actionRef.current.reload();
        }
      } else {
        message.error(response.msg || '删除失败');
      }
    } catch (error) {
      message.error('删除失败');
    }
  };

  const handleToggleEnabled = async (record: RuleListItem, enabled: boolean) => {
    try {
      const response = await updateRule({
        ...record,
        id: record.id,
        enabled: enabled ? 1 : 0,
      });
      if (response.code === 200) {
        message.success(enabled ? '已启用' : '已禁用');
        if (actionRef.current) {
          actionRef.current.reload();
        }
      } else {
        message.error(response.msg || '操作失败');
      }
    } catch (error) {
      message.error('操作失败');
    }
  };

  const columns: ProColumns<RuleListItem>[] = [
    {
      title: '规则名称',
      dataIndex: 'ruleName',
      width: 200,
      sorter: true,
    },
    {
      title: '规则类型',
      dataIndex: 'ruleType',
      width: 120,
      valueEnum: {
        '完整性': { text: '完整性' },
        '准确性': { text: '准确性' },
        '唯一性': { text: '唯一性' },
        '一致性': { text: '一致性' },
        '及时性': { text: '及时性' },
      },
    },
    {
      title: '规则描述',
      dataIndex: 'ruleDesc',
      ellipsis: true,
      hideInSearch: true,
    },
    {
      title: '阈值',
      dataIndex: 'threshold',
      width: 100,
      hideInSearch: true,
      render: (text: number) => `${text}%`,
    },
    {
      title: '严重程度',
      dataIndex: 'severity',
      width: 100,
      valueEnum: {
        'high': { text: '高', status: 'Error' },
        'medium': { text: '中', status: 'Warning' },
        'low': { text: '低', status: 'Default' },
      },
      render: (_, record) => {
        const color = record.severity === 'high' ? 'red' : record.severity === 'medium' ? 'orange' : 'blue';
        const text = record.severity === 'high' ? '高' : record.severity === 'medium' ? '中' : '低';
        return <Tag color={color}>{text}</Tag>;
      },
    },
    {
      title: '是否启用',
      dataIndex: 'enabled',
      width: 100,
      valueEnum: {
        1: { text: '启用', status: 'Success' },
        0: { text: '禁用', status: 'Default' },
      },
      render: (_, record) => (
        <Switch
          checked={record.enabled === 1}
          onChange={(checked) => handleToggleEnabled(record, checked)}
        />
      ),
    },
    {
      title: '创建人',
      dataIndex: 'createdBy',
      width: 100,
      hideInSearch: true,
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      width: 180,
      valueType: 'dateTime',
      hideInSearch: true,
      sorter: true,
    },
    {
      title: '操作',
      valueType: 'option',
      width: 150,
      render: (_, record) => [
        <Button
          key="edit"
          type="link"
          size="small"
          icon={<EditOutlined />}
          onClick={() => {
            setFormValues(record);
            handleUpdateModalVisible(true);
          }}
        >
          编辑
        </Button>,
        <Popconfirm
          key="delete"
          title="确定要删除这条规则吗？"
          onConfirm={() => handleDelete(record.id)}
        >
          <Button
            type="link"
            danger
            size="small"
            icon={<DeleteOutlined />}
          >
            删除
          </Button>
        </Popconfirm>,
      ],
    },
  ];

  return (
    <PageContainer>
      <ProTable<RuleListItem>
        headerTitle="质量规则配置"
        actionRef={actionRef}
        rowKey="id"
        search={{
          labelWidth: 120,
        }}
        request={async (params) => {
          const response = await queryRules(params);
          return {
            data: response.data?.list || [],
            success: response.code === 200,
            total: response.data?.total || 0,
          };
        }}
        columns={columns}
        pagination={{
          pageSize: 10,
          showSizeChanger: true,
        }}
        toolBarRender={() => [
          <Button
            type="primary"
            key="primary"
            onClick={() => {
              setFormValues({});
              handleCreateModalVisible(true);
            }}
          >
            <PlusOutlined /> 新建规则
          </Button>,
        ]}
      />

      <CreateForm
        onCancel={() => handleCreateModalVisible(false)}
        modalVisible={createModalVisible}
        onSubmit={async (value) => {
          const success = await handleAdd(value as RuleListItem);
          if (success) {
            handleCreateModalVisible(false);
            if (actionRef.current) {
              actionRef.current.reload();
            }
          }
        }}
      />

      <UpdateForm
        onCancel={() => {
          handleUpdateModalVisible(false);
          setFormValues({});
        }}
        modalVisible={updateModalVisible}
        onSubmit={async (value) => {
          const success = await handleUpdate({ ...formValues, ...value } as RuleListItem);
          if (success) {
            handleUpdateModalVisible(false);
            setFormValues({});
            if (actionRef.current) {
              actionRef.current.reload();
            }
          }
        }}
        values={formValues}
      />
    </PageContainer>
  );
};

export default RuleList;

