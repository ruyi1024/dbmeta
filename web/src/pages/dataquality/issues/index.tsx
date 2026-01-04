import React, { useRef } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import ProTable, { ProColumns, ActionType } from '@ant-design/pro-table';
import { Button, Tag, Space, message, Modal, Form, Input, Select } from 'antd';
import { 
  TableOutlined, 
  ColumnHeightOutlined, 
  RobotOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  EditOutlined
} from '@ant-design/icons';
import { queryIssues, updateIssueStatus } from './service';
import type { IssueListItem } from './data.d';

const { TextArea } = Input;

const IssueList: React.FC<{}> = () => {
  const actionRef = useRef<ActionType>();
  const [form] = Form.useForm();
  const [modalVisible, setModalVisible] = React.useState(false);
  const [currentRecord, setCurrentRecord] = React.useState<IssueListItem | null>(null);

  const handleUpdateStatus = async (values: any) => {
    try {
      const response = await updateIssueStatus({
        id: currentRecord?.key,
        ...values,
      });
      if (response.code === 200) {
        message.success('更新成功');
        setModalVisible(false);
        form.resetFields();
        if (actionRef.current) {
          actionRef.current.reload();
        }
      } else {
        message.error(response.msg || '更新失败');
      }
    } catch (error) {
      message.error('更新失败');
    }
  };

  const columns: ProColumns<IssueListItem>[] = [
    {
      title: '数据库名',
      dataIndex: 'databaseName',
      width: 150,
      hideInSearch: true,
    },
    {
      title: '表名',
      dataIndex: 'tableName',
      width: 150,
      render: (text: string) => (
        <Space>
          <TableOutlined />
          <span>{text}</span>
        </Space>
      ),
    },
    {
      title: '字段名',
      dataIndex: 'columnName',
      width: 150,
      render: (text: string) => (
        <Space>
          <ColumnHeightOutlined />
          <span>{text}</span>
        </Space>
      ),
    },
    {
      title: '问题类型',
      dataIndex: 'issueType',
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
      title: '严重程度',
      dataIndex: 'issueLevel',
      width: 100,
      valueEnum: {
        'high': { text: '高', status: 'Error' },
        'medium': { text: '中', status: 'Warning' },
        'low': { text: '低', status: 'Default' },
      },
      render: (_, record) => {
        const color = record.issueLevel === 'high' ? 'red' : record.issueLevel === 'medium' ? 'orange' : 'blue';
        const text = record.issueLevel === 'high' ? '高' : record.issueLevel === 'medium' ? '中' : '低';
        return <Tag color={color}>{text}</Tag>;
      },
    },
    {
      title: '问题描述',
      dataIndex: 'issueDesc',
      ellipsis: true,
    },
    {
      title: '问题数量',
      dataIndex: 'issueCount',
      width: 100,
      sorter: true,
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      valueEnum: {
        1: { text: '待处理', status: 'Error' },
        2: { text: '处理中', status: 'Warning' },
        3: { text: '已处理', status: 'Success' },
        0: { text: '已忽略', status: 'Default' },
      },
    },
    {
      title: '处理人',
      dataIndex: 'handler',
      width: 100,
      hideInSearch: true,
    },
    {
      title: '最后检查时间',
      dataIndex: 'lastCheckTime',
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
            setCurrentRecord(record);
            form.setFieldsValue({
              status: record.status,
              handler: record.handler,
              handleRemark: record.handleRemark,
            });
            setModalVisible(true);
          }}
        >
          处理
        </Button>,
      ],
    },
  ];

  return (
    <PageContainer>
      <ProTable<IssueListItem>
        headerTitle="质量问题列表"
        actionRef={actionRef}
        rowKey="key"
        search={{
          labelWidth: 120,
        }}
        request={async (params) => {
          const response = await queryIssues(params);
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
          <Tag key="ai" color="blue" icon={<RobotOutlined />}>
            AI智能诊断
          </Tag>,
        ]}
      />

      <Modal
        title="处理质量问题"
        open={modalVisible}
        onCancel={() => {
          setModalVisible(false);
          form.resetFields();
        }}
        onOk={() => {
          form.submit();
        }}
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleUpdateStatus}
        >
          <Form.Item
            name="status"
            label="处理状态"
            rules={[{ required: true, message: '请选择处理状态' }]}
          >
            <Select>
              <Select.Option value={1}>待处理</Select.Option>
              <Select.Option value={2}>处理中</Select.Option>
              <Select.Option value={3}>已处理</Select.Option>
              <Select.Option value={0}>已忽略</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item
            name="handler"
            label="处理人"
          >
            <Input placeholder="请输入处理人" />
          </Form.Item>
          <Form.Item
            name="handleRemark"
            label="处理备注"
          >
            <TextArea rows={4} placeholder="请输入处理备注" />
          </Form.Item>
        </Form>
      </Modal>
    </PageContainer>
  );
};

export default IssueList;

