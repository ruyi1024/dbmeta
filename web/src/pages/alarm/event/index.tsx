import React, { useEffect, useRef, useState } from 'react';
import { Suspense } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import { message, Popconfirm, Space, Table } from 'antd';
import ProTable, { ProColumns, ActionType } from '@ant-design/pro-table';
import { TableListItem } from './data.d';

import { queryEvent, queryAlarmAnalysis } from './service';
import Tag from 'antd/es/tag';
import { batchUpdateAlarmStatus } from "@/services/alarm/alarm";
import EventInfoView from "@/pages/monitor/event/eventInfo";


const TableList: React.FC<any> = () => {
  useEffect(() => {
    try {
      console.info("init page.")
    } catch (e) {
      message.error(`get event error. ${e}`);
    }
  }, []);

  const [eventUuid, setEventUuid] = useState<string>();
  const [modalVisible, setModalVisible] = useState<boolean>(false);

  const actionRef = useRef<ActionType>();

  const [alarmAnalysisData, setAlarmAnalysisData] = useState({});
  const didQuery = async () => {
    try {
      didQuery
      const analysisData = await queryAlarmAnalysis();
      setAlarmAnalysisData(analysisData);
      return
    } catch (e) {
      return { success: false, msg: e }
    }
  }

  // useEffect(() => {
  //   didQuery();
  // });

  const columns: ProColumns<TableListItem>[] = [
    {
      title: '告警标题',
      dataIndex: 'alarm_title',
      hideInForm: true,
      sorter: false,
      search: false,
      width: 280,
      copyable: true,
      ellipsis: true,
      tip: '标题过长会自动收缩',
      render: (text, value) => {
        // const nodes = <span><Tag color="volcano">未处理</Tag>{value.alarm_level}+{value.alarm_title}</span>
        // return nodes
        switch (value.status) {
          case 0:
            return <span><Tag color="volcano">未处理</Tag>[{value.alarm_level}] {value.alarm_title}</span>
          case 1:
            return <span><Tag color="blue">处理中</Tag>[{value.alarm_level}] {value.alarm_title}</span>
          case 2:
            return <span><Tag color="green">已完成</Tag>[{value.alarm_level}] {value.alarm_title}</span>
          default:
            return <span><Tag>未知</Tag>[{value.alarm_level}] {value.alarm_title}</span>
        }
      }
    },
    /*
    {
      title: '事件处理',
      dataIndex: 'status',
      hideInForm: true,
      sorter: false,
      // @ts-ignore
      render: (_, value) => {
        switch (value.status) {
          case 0:
            return <Tag color="volcano">未处理</Tag>
          case 1:
            return <Tag color="blue">处理中</Tag>
          case 2:
            return <Tag color="green">已完成</Tag>
          default:
            return <Tag>未知</Tag>
        }
      }
    },
     */
    {
      title: '事件类型',
      dataIndex: 'event_type',
      hideInForm: true,
      sorter: false,
      width: 100,
    },
    {
      title: '事件组',
      dataIndex: 'event_group',
      hideInForm: false,
      sorter: false,
      width: 85,
    },
    {
      title: '事件实体',
      dataIndex: 'event_entity',
      hideInForm: false,
      sorter: false,
      width: 160,
      ellipsis: true,
    },
    {
      title: '事件标签',
      dataIndex: 'event_tag',
      hideInForm: false,
      sorter: false,
      width: 130,
    },
    {
      title: '触发规则',
      dataIndex: 'event_key',
      hideInForm: false,
      sorter: false,
      width: 140,
      ellipsis: true,
      render: (text, value) => {
        const nodes = <span>{value.event_key} [{value.event_value}{value.event_unit}{value.alarm_rule}{value.alarm_value}{value.event_unit}]</span>
        return nodes
      }
    },

    {
      title: '邮件',
      dataIndex: 'send_mail',
      filters: false,
      onFilter: false,
      valueEnum: {
        '0': { text: '', status: 'Default' },
        '1': { text: '', status: 'Success' },
        '2': { text: '', status: 'Error' },
      },
      sorter: false,
      search: false,
      width: 55,
    },
    {
      title: '短信',
      dataIndex: 'send_sms',
      filters: false,
      onFilter: false,
      valueEnum: {
        '0': { text: '', status: 'Default' },
        '1': { text: '', status: 'Success' },
        '2': { text: '', status: 'Error' },
      },
      sorter: false,
      search: false,
      width: 55,
    },
    {
      title: '电话',
      dataIndex: 'send_phone',
      filters: false,
      onFilter: false,
      valueEnum: {
        '0': { text: '', status: 'Default' },
        '1': { text: '', status: 'Success' },
        '2': { text: '', status: 'Error' },
      },
      sorter: false,
      search: false,
      width: 55,
    },
    {
      title: '微信',
      dataIndex: 'send_wechat',
      filters: false,
      onFilter: false,
      valueEnum: {
        '0': { text: '', status: 'Default' },
        '1': { text: '', status: 'Success' },
        '2': { text: '', status: 'Error' },
      },
      sorter: false,
      search: false,
      width: 55,
    },
    {
      title: 'API',
      dataIndex: 'send_webhook',
      filters: false,
      onFilter: false,
      valueEnum: {
        '0': { text: '', status: 'Default' },
        '1': { text: '', status: 'Success' },
        '2': { text: '', status: 'Error' },
      },
      sorter: false,
      search: false,
      width: 55,
    },

    {
      title: '告警时间',
      dataIndex: 'gmt_created',
      valueType: 'dateTime',
      hideInForm: true,
      search: false,
      width: 180,
    },
    {
      title: '操作',
      dataIndex: 'event_detail',
      width: 100,
      render: (text: any, record: any) => (
        <a
          onClick={() => {
            setEventUuid(record.event_uuid);
            setModalVisible(true);
          }}
        >
          告警事件分析
        </a>
      ),
    },
  ];


  const updateStatus = (ids: number[], updStatus: number) => {
    console.log(ids, updStatus);
    batchUpdateAlarmStatus({ ids, status: updStatus }).then(res => {
      if (res.success) {
        // @ts-ignore
        actionRef.current.reload();
      }
    })
  }

  return (
    <PageContainer>

      {/*<IntroduceRow loading={loading} analysisData={alarmAnalysisData || {}} />*/}


      <ProTable<TableListItem>
        headerTitle="数据列表"
        cardBordered
        actionRef={actionRef}
        rowKey="id"
        search={{
          labelWidth: 120,
        }}
        request={(params, sorter) => queryEvent({ ...params, sorter })}
        columns={columns}
        pagination={{
          pageSize: 50,
        }}
        rowSelection={{
          selections: [Table.SELECTION_ALL, Table.SELECTION_INVERT],
          // defaultSelectedRowKeys: [1],
          renderCell: (checked, record, index, originNode) => {
            if (record.status !== 2) {
              return originNode;
            }
            return false;
          }
        }}
        tableAlertRender={({ selectedRows, onCleanSelected }) => (
          <Space size={12}>
            <span>
              {
                selectedRows &&
                <>
                  已选 {selectedRows.filter(item => item && (item.status == 0 || item.status == 1)).length} 项 <a style={{ marginLeft: 8 }} onClick={onCleanSelected}>取消选择</a>
                </>
              }
            </span>
            <span>
              <Space>
                {selectedRows && selectedRows.filter(item => item && item.status == 0).length > 0 &&
                  <>
                    <strong>
                      {selectedRows && `未处理  ${selectedRows.filter(item => item && item.status == 0).length} 项`}
                    </strong>
                    <span>操作：</span>
                    <Popconfirm
                      title="是否标记为处理中 ？"
                      onConfirm={() => updateStatus(selectedRows.filter(item => item && item.status == 0).map(item => item.id), 1)}
                      okText="是"
                      cancelText="否"
                    >
                      <a>处理中</a>
                    </Popconfirm>
                    <Popconfirm
                      title="是否标记为已完成 ？"
                      onConfirm={() => updateStatus(selectedRows.filter(item => item && item.status == 1).map(item => item.id), 2)}
                      okText="是"
                      cancelText="否"
                    >
                      <a>已完成</a>
                    </Popconfirm>
                  </>
                }
              </Space>
            </span>
            <span>
              <Space>
                {
                  selectedRows && selectedRows.filter(item => item && item.status == 1).length > 0 &&
                  <>
                    <strong>
                      {selectedRows && `处理中 ${selectedRows.filter(item => item && item.status == 1).length} 项`}
                    </strong>
                    <span>操作：</span>
                    <Popconfirm
                      title="是否标记为已完成 ？"
                      onConfirm={() => updateStatus(selectedRows.filter(item => item && item.status == 1).map(item => item.id), 2)}
                      okText="是"
                      cancelText="否"
                    >
                      <a>已完成</a>
                    </Popconfirm>
                  </>
                }
              </Space>
            </span>
          </Space>
        )}
        sticky
        tableAlertOptionRender={() => {
          return "";
        }}
      />

      <EventInfoView
        eventUuid={eventUuid}
        modalVisible={modalVisible}
        onCancel={() => setModalVisible(false)}
      />

    </PageContainer>
  );
};

export default TableList;
