import { PlusOutlined, FormOutlined, DeleteOutlined } from '@ant-design/icons';
import { Button, Divider, message, Popconfirm, Select } from 'antd';
import React, { useState, useRef, useEffect } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import ProTable, { ProColumns, ActionType } from '@ant-design/pro-table';
import CreateForm from './components/CreateForm';
import UpdateForm from './components/UpdateForm';
import { TableListItem } from './data.d';
import { queryRule, updateRule, addRule, removeRule } from './service';
import { useAccess } from 'umi';

/**
 * 添加节点
 * @param fields
 */
const handleAdd = async (fields: TableListItem) => {
  const hide = message.loading('正在添加');
  try {
    await addRule({ ...fields });
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
    await updateRule({
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
    await removeRule({
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

const formInitValue = { "title": "", "event_type": "", "event_group": "", "event_key": "", "event_entity": "", "alarm_rule": "", "alarm_value": "", "alarm_times": "", "alarm_sleep": "", "level_id": "", "channel_id": "", "enable": "" }

const TableList: React.FC<{}> = () => {
  const [createModalVisible, handleModalVisible] = useState<boolean>(false);
  const [updateModalVisible, handleUpdateModalVisible] = useState<boolean>(false);
  const [formValues, setFormValues] = useState(formInitValue);
  const actionRef = useRef<ActionType>();
  const access = useAccess();

  const [levelList, setLevelList] = useState<any[]>([{ "id": 0, "level_name": "" }]);
  const [levelEnum, setLevelEnum] = useState<{}>({})
  const [channelList, setChannelList] = useState<any[]>([{ "id": 0, "name": "" }]);
  const [channelEnum, setChannelEnum] = useState<{}>({})

  useEffect(() => {

    fetch('/api/v1/alarm/level')
      .then((response) => response.json())
      .then((json) => {
        setLevelList(json.data);
        const valueDict: { [key: number]: string } = {}
        json.data.forEach((record: { id: string | number; level_name: string; }) => { valueDict[record.id] = record.level_name });
        setLevelEnum(valueDict)
      })
      .catch((error) => {
        console.log('Fetch level list failed', error);
      });

    fetch('/api/v1/alarm/channel')
      .then((response) => response.json())
      .then((json) => {
        setChannelList(json.data);
        const valueDict: { [key: string]: string } = {}
        json.data.forEach((record: { id: string | number; name: string; }) => { valueDict[record.id] = record.name });
        setChannelEnum(valueDict)
      })
      .catch((error) => {
        console.log('Fetch cluster list failed', error);
      });

  }, []);

  console.info(formValues);
  console.info(formValues.level_id);
  const columns: ProColumns<TableListItem>[] = [
    {
      title: '规则名称',
      dataIndex: 'title',
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
      initialValue: formValues.title,
    },
    {
      title: '事件类型',
      dataIndex: 'event_type',
      hideInForm: false,
      formItemProps: {
        rules: [
          {
            required: true,
            message: '此项为必填项',
          },
        ],
      },
      initialValue: formValues.event_type,
    },
    {
      title: '事件指标',
      dataIndex: 'event_key',
      hideInForm: false,
      formItemProps: {
        rules: [
          {
            required: true,
            message: '此项为必填项',
          },
        ],
      },
      initialValue: formValues.event_key,
    },
    {
      title: '事件组',
      dataIndex: 'event_group',
      hideInForm: false,
      initialValue: formValues.event_group,
    },
    {
      title: '事件实体',
      dataIndex: 'event_entity',
      hideInForm: false,
      initialValue: formValues.event_entity,
    },
    {
      title: '告警规则',
      dataIndex: 'alarm_rule',
      hideInForm: false,
      search: false,
      formItemProps: {
        rules: [
          {
            required: true,
            message: '此项为必填项',
          },
        ],
      },
      valueEnum: {
        '=': { text: '等于' },
        '!=': { text: '不等于' },
        '>': { text: '大于' },
        '>=': { text: '大于等于' },
        '<': { text: '小于' },
        '<=': { text: '小于等于' },
      },
      initialValue: formValues.alarm_rule,
    },
    {
      title: '告警值',
      dataIndex: 'alarm_value',
      hideInForm: false,
      search: false,
      formItemProps: {
        rules: [
          {
            required: true,
            message: '此项为必填项',
          },
        ],
      },
      initialValue: formValues.alarm_value,
    },
    {
      title: '告警级别',
      dataIndex: 'level_id',
      //initialValue: parseInt(formValues.level_id), //类型问题导致表单默认值异常，需要转换成整数
      initialValue: formValues.level_id,
      formItemProps: {
        rules: [
          {
            required: true,
            message: '此项为必填项',
          },
        ],
      },
      renderFormItem: (_, { type, defaultRender, ...rest }, form) => {
        return <Select showSearch>
          {levelList && levelList.map(item => <Option key={item.level_name} value={item.id}>{item.level_name}</Option>)}
        </Select>
      },
      valueEnum: levelEnum,
    },
    {
      title: '通知渠道',
      dataIndex: 'channel_id',
      initialValue: formValues.channel_id,
      hideInForm: false,
      search: false,
      formItemProps: {
        rules: [
          {
            required: true,
            message: '此项为必填项',
          },
        ],
      },
      renderFormItem: (_, { type, defaultRender, ...rest }, form) => {
        return <Select showSearch>
          {channelList && channelList.map(item => <Option key={item.name} value={item.id}>{item.name}</Option>)}
        </Select>
      },
      valueEnum: channelEnum,
    },
    {
      title: '限流次数',
      dataIndex: 'alarm_times',
      hideInForm: false,
      search: false,
      formItemProps: {
        rules: [
          {
            required: true,
            message: '此项为必填项',
          },
        ],
      },
      valueEnum: {
        1: { text: '1次' },
        2: { text: '2次' },
        3: { text: '3次' },
        4: { text: '4次' },
        5: { text: '5次' },
        6: { text: '6次' },
        7: { text: '7次' },
        8: { text: '8次' },
      },
      initialValue: formValues.alarm_times,
      renderFormItem: (_, { type, defaultRender, ...rest }, form) => {
        return (
          <Select>
            <OptionComponent value={1}>1次</OptionComponent>
            <OptionComponent value={2}>2次</OptionComponent>
            <OptionComponent value={3}>3次</OptionComponent>
            <OptionComponent value={4}>4次</OptionComponent>
            <OptionComponent value={5}>5次</OptionComponent>
            <OptionComponent value={6}>6次</OptionComponent>
            <OptionComponent value={7}>7次</OptionComponent>
            <OptionComponent value={8}>8次</OptionComponent>
          </Select>
        );
      },
    },
    {
      title: '风控时间',
      dataIndex: 'alarm_sleep',
      hideInForm: false,
      search: false,
      formItemProps: {
        rules: [
          {
            required: true,
            message: '此项为必填项',
          },
        ],
      },
      valueEnum: {
        300: { text: '5分钟' },
        900: { text: '15分钟' },
        1800: { text: '30分钟' },
        3600: { text: '1小时' },
        10800: { text: '3小时' },
        21600: { text: '6小时' },
        43200: { text: '12小时' },
        86400: { text: '24小时' },
      },
      initialValue: formValues.alarm_sleep,
      renderFormItem: (_, { type, defaultRender, ...rest }, form) => {
        return (
          <Select>
            <OptionComponent value={300}>5分钟</OptionComponent>
            <OptionComponent value={900}>15分钟</OptionComponent>
            <OptionComponent value={1800}>30分钟</OptionComponent>
            <OptionComponent value={3600}>1小时</OptionComponent>
            <OptionComponent value={10800}>3小时</OptionComponent>
            <OptionComponent value={21600}>6小时</OptionComponent>
            <OptionComponent value={43200}>12小时</OptionComponent>
            <OptionComponent value={86400}>24小时</OptionComponent>
          </Select>
        );
      },
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
    // {
    //   title: '创建时间',
    //   dataIndex: 'gmt_created',
    //   sorter: true,
    //   valueType: 'dateTime',
    //   hideInForm: true,
    //   search:false,
    // },
    // {
    //   title: '修改时间',
    //   dataIndex: 'gmt_updated',
    //   sorter: true,
    //   valueType: 'dateTime',
    //   hideInForm: true,
    //   search:false,
    // },
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
            title={`确认要删除数据【${record.title}】,删除后不可恢复，是否继续？`}
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
        request={(params, sorter, filter) => queryRule({ ...params, sorter, filter })}
        columns={columns}
        pagination={{
          pageSize: 20,
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
