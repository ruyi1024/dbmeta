import React, { useEffect, useState } from 'react';
import {
  Card,
  Table,
  Tooltip,
  Button,
  Divider,
  message,
  Row,
  Col,
  Space,
  DatePicker,
  Select,
} from 'antd';
import {
  RetweetOutlined,
  ReloadOutlined,
  SearchOutlined,
  DatabaseOutlined,
  BarChartOutlined,
} from '@ant-design/icons';
import ReactECharts from 'echarts-for-react';

import {
  getEventCharts,
  getEventChartsFull,
  getEventList,
  getEventAllDescription,
} from '@/services/event/eventList';
import { getStorageItem, setStorageItem } from '@/utils/storage';
import type { ColumnsType } from 'antd/lib/table';
import moment from 'moment';
import EventInfoView from '@/pages/event/event/eventInfo';
import EventChart from '@/pages/event/event/eventChart';
import { useLatest, useMap } from 'ahooks';

const EVENT_TABLE_KEY = 'lepus.eventList.table_pageSize_01';

export default (): React.ReactNode => {
  const [eventUuid, setEventUuid] = useState<string>();
  const [modalVisible, setModalVisible] = useState<boolean>(false);

  const [dateKeyword, setDateKeyword] = useState<string[]>([]);
  const [eventKeyword, setEventKeyword] = useState<string>('');
  const [groupKeyword, setGroupKeyword] = useState<string>('');
  const [eventEntityKeyword, setEventEntityKeyword] = useState<string>('');
  const [eventKeyKeyword, setEventKeyKeyword] = useState<string>('');
  const [typeKeyword, setTypeKeyword] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(false);
  const [pageSize, setPageSize] = useState<number>(getStorageItem(EVENT_TABLE_KEY) || 25);
  const [current, setCurrent] = useState<number>(1);
  const [total, setTotal] = useState<number>();
  const [sorter, setSorter] = useState<any>();
  const [list, setList] = useState<API.EventListRes>();
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const [dates, setDates] = useState<any[]>([]);
  const [activeTabKey, setActiveTabKey] = useState<string>('1');
  const [chartList, setChartList] = useState<any[]>([]);
  const [chartFull, setChartFull] = useState<any[]>([]);
  const [chartFullLoading, setChartFullLoading] = useState<boolean>(false);

  const [filtersEventType, setFiltersEventType] = useState<any>([]);
  const [filtersEventGroup, setFiltersEventGroup] = useState<any>([]);
  const [filtersEventEntity, setFiltersEventEntity] = useState<any>([]);
  const [filtersEventKey, setFiltersEventKey] = useState<any>([]);

  const latestChartFullRef = useLatest(chartFull);

  const [
    eventInfoMap,
    {
      set: setEventInfoMap,
      reset: resetEventInfoMap,
      get: getEventInfoMap
    }
  ] = useMap<string, string>();

  useEffect(() => {
    try {
      // eslint-disable-next-line @typescript-eslint/no-use-before-define
      did({});
      // eslint-disable-next-line @typescript-eslint/no-use-before-define
      didFiltersEventType();
      // @ts-ignore
      // eslint-disable-next-line @typescript-eslint/no-use-before-define
      didCharts();
      //console.info(chartList)
      // @ts-ignore
      // eslint-disable-next-line @typescript-eslint/no-use-before-define
      didEventAllInfo();
    } catch (e) {
      message.error(`get event error. ${e}`);
    }
  }, []);

  const didFiltersEventType = () => {
    setGroupKeyword('');
    setFiltersEventGroup([]);
    setEventEntityKeyword('');
    setFiltersEventEntity([]);
    setEventKeyKeyword('');
    setFiltersEventKey([]);
    fetch('/api/v1/event/type/list')
      .then((response) => response.json())
      .then((json) => {
        setFiltersEventType(json.data);
      })
      .catch((error) => {
        console.log('Fetch type list failed', error);
      });
  };

  const didEventAllInfo = () => {
    getEventAllDescription().then((res) => {
      if (res.success) {
        resetEventInfoMap();
        res.data.forEach((item: any) => {
          setEventInfoMap(`${item.eventType}.${item.eventKey}`, item.description)
        })
      }
    });
  };

  const didFiltersEventGroup = (eventType: string) => {
    setGroupKeyword('');
    setEventEntityKeyword('');
    setFiltersEventEntity([]);
    setEventKeyKeyword('');
    setFiltersEventKey([]);
    fetch('/api/v1/event/group/list?event_type=' + eventType)
      .then((response) => response.json())
      .then((json) => {
        setFiltersEventGroup(json.data);
      })
      .catch((error) => {
        console.log('Fetch group list failed', error);
      });
  };

  const didFiltersEventEntity = (eventType: string, eventGroup: string) => {
    console.info(eventType);
    setEventEntityKeyword('');
    fetch('/api/v1/event/entity/list?event_type=' + eventType + '&event_group=' + eventGroup)
      .then((response) => response.json())
      .then((json) => {
        setFiltersEventEntity(json.data);
      })
      .catch((error) => {
        console.log('Fetch entity list failed', error);
      });
  };

  const didFiltersEventKey = (eventType: string, eventGroup: string) => {
    console.info(eventType);
    setEventKeyKeyword('');
    fetch('/api/v1/event/key/list?event_type=' + eventType + '&event_group=' + eventGroup)
      .then((response) => response.json())
      .then((json) => {
        setFiltersEventKey(json.data);
      })
      .catch((error) => {
        console.log('Fetch key list failed', error);
      });
  };

  const didCharts = async (params: any) => {
    getEventCharts(params).then((res: any) => {
      if (res.success) {
        setChartList(res.data);
      } else {
        message.error('get event charts full fail. ', res.errorMsg);
      }
    });
  };

  const didChartsFull = (params: API.DidParams) => {
    if (typeKeyword == undefined || typeKeyword == '') {
      message.warning('请选择事件类型');
      return;
    }
    if (groupKeyword == undefined || groupKeyword == '') {
      message.warning('请选择事件分组');
      return;
    }
    if (eventEntityKeyword == undefined || eventEntityKeyword == '') {
      message.warning('请选择事件实体');
      return;
    }

    const data = {
      // eslint-disable-next-line no-nested-ternary
      eventKeyword: params.reset
        ? ''
        : params.eventKeyword !== undefined
          ? params.eventKeyword
          : eventKeyword || '',
      // eslint-disable-next-line no-nested-ternary
      groupKeyword: params.reset
        ? ''
        : params.groupKeyword !== undefined
          ? params.groupKeyword
          : groupKeyword || '',
      // eslint-disable-next-line no-nested-ternary
      typeKeyword: params.reset
        ? ''
        : params.typeKeyword !== undefined
          ? params.typeKeyword
          : typeKeyword || '',
      // eslint-disable-next-line no-nested-ternary
      eventEntityKeyword: params.reset
        ? ''
        : params.eventEntityKeyword !== undefined
          ? params.eventEntityKeyword
          : eventEntityKeyword || '',
      eventKeyKeyword: params.reset
        ? ''
        : params.eventKeyKeyword !== undefined
          ? params.eventKeyKeyword
          : eventKeyKeyword || '',
      startTime: params.reset ? moment().subtract(60, 'minutes') : dateKeyword[0],
      endTime: params.reset ? moment() : dateKeyword[1],
      ...params,
    };

    setChartFullLoading(true);
    message.info('图表创建中，请稍后...');
    getEventChartsFull(data).then((res: any) => {
      if (res.success) {
        setChartFull(res.data);
        setChartFullLoading(false);
      } else {
        setChartFullLoading(false);
        message.error('get event charts full fail. ', res.errorMsg);
      }
    });
  };

  const did = async (params: API.DidParams) => {
    setLoading(true);
    const limit = params && params.limit ? params.limit : pageSize;
    const data = {
      offset: limit * (current >= 2 ? -1 : 0),
      // eslint-disable-next-line no-nested-ternary
      eventKeyword: params.reset
        ? ''
        : params.eventKeyword !== undefined && params.eventKeyword !== ''
          ? params.eventKeyword
          : eventKeyword || '',
      // eslint-disable-next-line no-nested-ternary
      groupKeyword: params.reset
        ? ''
        : params.groupKeyword !== undefined && params.groupKeyword !== ''
          ? params.groupKeyword
          : groupKeyword || '',
      // eslint-disable-next-line no-nested-ternary
      typeKeyword: params.reset
        ? ''
        : params.typeKeyword !== undefined && params.typeKeyword !== ''
          ? params.typeKeyword
          : typeKeyword || '',
      // eslint-disable-next-line no-nested-ternary
      eventEntityKeyword: params.reset
        ? ''
        : params.eventEntityKeyword !== undefined && params.eventEntityKeyword !== ''
          ? params.eventEntityKeyword
          : eventEntityKeyword || '',
      eventKeyKeyword: params.reset
        ? ''
        : params.eventKeyKeyword !== undefined && params.eventKeyKeyword !== ''
          ? params.eventKeyKeyword
          : eventKeyKeyword || '',
      startTime: params.reset ? moment().subtract(60, 'minutes') : dateKeyword[0],
      endTime: params.reset ? moment() : dateKeyword[1],
      ...params,
    };
    if (sorter && sorter.field) {
      data.sorterField = `${sorter.field}`;
      data.sorterOrder = `${sorter.order}`;
    }
    getEventList(data)
      .then((res: any) => {
        if (res && res.success) {
          setList(res.data);
          setTotal(res.total);
          setLoading(false);
          // 分页数据不足自动切换到第一页
          if (res.total !== 0 && res.total < current * pageSize - 1) {
            setCurrent(1);
          }
        } else {
          message.error(`失败！${res.errorMsg}`);
          setLoading(false);
        }
      })
      .catch((e) => {
        message.error('get event data error.', e);
        setLoading(false);
      });
    didCharts(data);
    //didChartsFull(data);
  };

  const handleStandardTableChange = (
    pagination: { pageSize: number; current: number },
    filterArg: any,
    sort: any,
  ) => {
    const params = {
      limit: pagination.pageSize,
      offset: (pagination.current > 1 ? pagination.current - 1 : 0) * pagination.pageSize,
      sorterField: '',
      sorterOrder: '',
    };
    if (sort && sort.field) {
      params.sorterField = `${sort.field}`;
      params.sorterOrder = `${sort.order}`;
    }
    setSorter(sort);
    setCurrent(pagination.current);
    setPageSize(pagination.pageSize);
    setStorageItem(EVENT_TABLE_KEY, pagination.pageSize);
    did(params);
  };

  const reset = () => {
    setCurrent(1);
    setEventKeyword('');
    setDates([]);
    setDateKeyword([]);
    setGroupKeyword('');
    setTypeKeyword('');
    setEventEntityKeyword('');
    did({ reset: true });
  };

  const disabledDate = (cur: any) => {
    // if (!dates || dates.length === 0) {
    //   return false;
    // }
    // const tooLate = dates[0] && cur.diff(dates[0], 'days') > 7;
    // // @ts-ignore
    // const tooEarly = dates[1] && dates[1].diff(cur, 'days') > 7;
    // const afterToday = cur && cur > moment().endOf('day');
    // return (tooEarly || tooLate) || afterToday ;
    return cur && cur > moment().endOf('day');
  };

  const columns: ColumnsType<any> | [] = [
    {
      title: '事件时间',
      dataIndex: 'event_time',
      sorter: true,
      render: (text: any) => moment(text).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '事件类型',
      dataIndex: 'event_type',
      sorter: false,
    },
    {
      title: '事件组',
      dataIndex: 'event_group',
      sorter: false,
    },
    {
      title: '事件实体',
      dataIndex: 'event_entity',
      sorter: false,
    },
    {
      title: '事件标签',
      dataIndex: 'event_tag',
      sorter: false,
    },
    {
      title: '事件指标',
      dataIndex: 'event_key',
      sorter: false,
      //render: (text: string) => <span><Tooltip title="xxxxxx" placement="topRight" ><QuestionCircleTwoTone /></Tooltip>&nbsp;{text}</span>,
    },
    {
      title: '事件数据',
      sorter: false,
      dataIndex: 'event_value',
    },
    {
      title: '数据单位',
      dataIndex: 'event_unit',
    },
    {
      title: '操作',
      dataIndex: 'event_detail',
      render: (text: any, record: any) => (
        <a
          onClick={() => {
            setEventUuid(record.event_uuid);
            setModalVisible(true);
          }}
        >
          事件详情
        </a>
      ),
    },
  ];

  const chartOption = (d: any) => {
    return {
      dataset: {
        source: d,
      },
      grid: {
        top: 10,
        bottom: 30,
        right: 35,
        left: 60,
      },
      toolbox: {
        show: false,
      },
      tooltip: {
        trigger: 'none',
        axisPointer: {
          type: 'cross',
        },
      },
      xAxis: {
        type: 'time',
        boundaryGap: false,
      },
      yAxis: {
        type: 'value',
        boundaryGap: [0, '100%'],
        axisLine: {
          show: false,
        },
        axisTick: {
          show: false,
        },
        splitNumber: 5,
        splitLine: {
          show: false,
        },
      },

      series: [
        {
          type: 'bar',
          // showBackground: true,
          symbol: 'none',
          areaStyle: {},
        },
      ],
    };
  };

  // @ts-ignore
  return (
    <div>
      <Row>
        <Col span={24} style={{ paddingTop: 2 }}>
          <Card
            size="small"
            tabProps={{ size: 'middle' }}
            tabList={[
              {
                key: '1',
                tab: (
                  <span>
                    <DatabaseOutlined />
                    事件数据
                  </span>
                ),
              },
              {
                key: '2',
                tab: (
                  <span>
                    <BarChartOutlined />
                    事件图表
                  </span>
                ),
              },
            ]}
            activeTabKey={activeTabKey}
            onTabChange={(key) => {
              setActiveTabKey(key);
            }}
          >
            {activeTabKey === '1' && (
              <>
                <Col span={24}>
                  <Card size="small" title={<strong>数据筛选</strong>} bodyStyle={{ padding: 10 }}>
                    <Space>
                      <span>事件时间:</span>
                      <DatePicker.RangePicker
                        style={{ width: 380 }}
                        onCalendarChange={(val: any) => {
                          setDates(val);
                        }}
                        defaultValue={[moment().subtract(10, 'minutes'), moment()]}
                        showTime={{
                          format: 'HH:mm:ss',
                          defaultValue: [moment().subtract(10, 'minutes'), moment()],
                        }}
                        format="YYYY-MM-DD HH:mm:ss"
                        disabledDate={disabledDate}
                        onChange={(value, dateString) => {
                          console.log('---date picker --> ', value, dateString);
                          setDateKeyword(dateString);
                        }}
                        ranges={{
                          '30分钟': [moment().subtract(30, 'minutes'), moment()],
                          '1小时': [moment().subtract(1, 'hours'), moment()],
                          今天: [moment('00:00:00', 'HH:mm:ss'), moment()],
                          最近三天: [moment().subtract(3, 'days'), moment()],
                          本周: [moment().startOf('week'), moment().endOf('week')],
                          本月: [moment().startOf('month'), moment().endOf('month')],
                        }}
                        placeholder={['开始时间', '结束时间']}
                      />
                    </Space>
                    <Divider dashed style={{ margin: '10px 0' }} />
                    <Row>
                      <Col flex="auto">
                        <Space>
                          <span>事件类型:</span>
                          <Select
                            allowClear
                            style={{ width: 200 }}
                            value={typeKeyword}
                            onChange={(val) => {
                              // @ts-ignore
                              setTypeKeyword(val);
                              didFiltersEventGroup(val);
                              did({
                                typeKeyword: val,
                                groupKeyword: '',
                                eventEntityKeyword: '',
                                eventKeyKeyword: '',
                              });
                            }}
                            placeholder="筛选事件类型"
                          >
                            {filtersEventType &&
                              filtersEventType.map((item: any) => {
                                return (
                                  <Select.Option
                                    key={`${item.event_type}`}
                                    value={`${item.event_type}`}
                                  >
                                    {`${item.event_type}`}
                                  </Select.Option>
                                );
                              })}
                          </Select>
                          <span>事件分组:</span>
                          <Select
                            allowClear
                            style={{ width: 200 }}
                            value={groupKeyword}
                            onChange={(val) => {
                              // @ts-ignore
                              setGroupKeyword(val);
                              didFiltersEventEntity(typeKeyword, val);
                              didFiltersEventKey(typeKeyword, val);
                              did({ groupKeyword: val, eventEntityKeyword: '' });
                            }}
                            placeholder="筛选事件组"
                          >
                            {filtersEventGroup &&
                              filtersEventGroup.map((item: any) => {
                                return (
                                  <Select.Option
                                    key={`${item.event_group}`}
                                    value={`${item.event_group}`}
                                  >
                                    {`${item.event_group}`}
                                  </Select.Option>
                                );
                              })}
                          </Select>
                          <span>事件实体:</span>
                          <Select
                            mode="multiple"
                            allowClear
                            style={{ width: 200 }}
                            value={(eventEntityKeyword && eventEntityKeyword.split(',')) || []}
                            onChange={(val) => {
                              // @ts-ignore
                              setEventEntityKeyword(val.join(','));
                              did({ eventEntityKeyword: val.join(',') });
                            }}
                            placeholder="筛选事件实体"
                          >
                            {filtersEventEntity &&
                              filtersEventEntity.map((item: any) => {
                                return (
                                  // <Select.Option
                                  //   key={`${item.ip}:${item.port}`}
                                  //   value={`${item.ip}:${item.port}`}
                                  // >
                                  //   {`${item.ip}:${item.port}`}
                                  // </Select.Option>
                                  <Select.Option
                                    key={`${item.event_entity}`}
                                    value={`${item.event_entity}`}
                                  >
                                    {`${item.event_entity}`}
                                  </Select.Option>
                                );
                              })}
                          </Select>
                          <span>事件指标:</span>

                          <Select
                            mode="multiple"
                            allowClear
                            style={{ width: 200 }}
                            value={(eventKeyKeyword && eventKeyKeyword.split(',')) || []}
                            onChange={(val) => {
                              // @ts-ignore
                              setEventKeyKeyword(val.join(','));
                              did({ eventKeyKeyword: val.join(',') });
                            }}
                            placeholder="筛选事件指标"
                          >
                            {filtersEventKey &&
                              filtersEventKey.map((item: any) => {
                                return (
                                  <Select.Option
                                    key={`${item.event_key}`}
                                    value={`${item.event_key}`}
                                  >
                                    {`${item.event_key}`}
                                  </Select.Option>
                                );
                              })}
                          </Select>
                        </Space>
                      </Col>
                      <Col flex="280px" style={{ textAlign: 'right' }}>
                        <Space>
                          <Tooltip placement="top" title="重置搜索内容">
                            <Button
                              type="link"
                              icon={<RetweetOutlined />}
                              onClick={() => {
                                reset();
                                did({});
                              }}
                            >
                              重置
                            </Button>
                          </Tooltip>
                          <Tooltip placement="top" title="重载并刷新表格数据">
                            <Button
                              type="link"
                              icon={<ReloadOutlined />}
                              onClick={() => {
                                did({});
                              }}
                            >
                              刷新
                            </Button>
                          </Tooltip>
                          <Tooltip placement="top" title="搜索选择内容">
                            <Button
                              type="link"
                              icon={<SearchOutlined />}
                              onClick={() => {
                                did({});
                              }}
                            >
                              查询
                            </Button>
                          </Tooltip>
                        </Space>
                      </Col>
                    </Row>
                  </Card>
                </Col>

                <Col span={24} style={{ paddingTop: 10 }}>
                  <Card>
                    <div>
                      <ReactECharts style={{ height: 150 }} option={chartOption(chartList)} />
                    </div>
                  </Card>
                </Col>

                <Table
                  rowKey={(record) => record.id}
                  columns={columns}
                  // @ts-ignore
                  dataSource={list}
                  size="small"
                  loading={loading}
                  // @ts-ignore
                  onChange={handleStandardTableChange}
                  pagination={{
                    pageSize,
                    current,
                    total,
                    showSizeChanger: true,
                    pageSizeOptions: ['25', '50', '100'],
                    showQuickJumper: true,
                    showTotal: (t, range) => `第 ${range[0]}-${range[1]}条， 共 ${t}条`,
                  }}
                  scroll={{ x: 1300 }}
                  sticky
                />
              </>
            )}
            {activeTabKey === '2' && (
              <>
                <Col span={24}>
                  <Card size="small" title={<strong>图表绘制</strong>} bodyStyle={{ padding: 10 }}>
                    <Space>
                      <span>事件时间:</span>
                      <DatePicker.RangePicker
                        style={{ width: 380 }}
                        onCalendarChange={(val: any) => {
                          setDates(val);
                        }}
                        defaultValue={[moment().subtract(10, 'minutes'), moment()]}
                        showTime={{
                          format: 'HH:mm:ss',
                          defaultValue: [moment().subtract(10, 'minutes'), moment()],
                        }}
                        format="YYYY-MM-DD HH:mm:ss"
                        disabledDate={disabledDate}
                        onChange={(value, dateString) => {
                          console.log('---date picker --> ', value, dateString);
                          setDateKeyword(dateString);
                        }}
                        ranges={{
                          '30分钟': [moment().subtract(30, 'minutes'), moment()],
                          '1小时': [moment().subtract(1, 'hours'), moment()],
                          今天: [moment('00:00:00', 'HH:mm:ss'), moment()],
                          最近三天: [moment().subtract(3, 'days'), moment()],
                          本周: [moment().startOf('week'), moment().endOf('week')],
                          本月: [moment().startOf('month'), moment().endOf('month')],
                        }}
                        placeholder={['开始时间', '结束时间']}
                      />
                    </Space>
                    <Divider dashed style={{ margin: '10px 0' }} />
                    <Row>
                      <Col flex="auto">
                        <Space>
                          <span>事件类型:</span>
                          <Select
                            allowClear
                            style={{ width: 200 }}
                            value={typeKeyword}
                            onChange={(val) => {
                              // @ts-ignore
                              setTypeKeyword(val);
                              didFiltersEventGroup(val);
                            }}
                            placeholder="筛选事件类型"
                          >
                            {filtersEventType &&
                              filtersEventType.map((item: any) => {
                                return (
                                  <Select.Option
                                    key={`${item.event_type}`}
                                    value={`${item.event_type}`}
                                  >
                                    {`${item.event_type}`}
                                  </Select.Option>
                                );
                              })}
                          </Select>
                          <span>事件分组:</span>
                          <Select
                            allowClear
                            style={{ width: 200 }}
                            value={groupKeyword}
                            onChange={(val) => {
                              // @ts-ignore
                              setGroupKeyword(val);
                              didFiltersEventEntity(typeKeyword, val);
                              didFiltersEventKey(typeKeyword, val);
                            }}
                            placeholder="筛选事件组"
                          >
                            {filtersEventGroup &&
                              filtersEventGroup.map((item: any) => {
                                return (
                                  <Select.Option
                                    key={`${item.event_group}`}
                                    value={`${item.event_group}`}
                                  >
                                    {`${item.event_group}`}
                                  </Select.Option>
                                );
                              })}
                          </Select>
                          <span>事件实体:</span>
                          <Select
                            mode="multiple"
                            allowClear
                            style={{ width: 200 }}
                            value={(eventEntityKeyword && eventEntityKeyword.split(',')) || []}
                            onChange={(val) => {
                              // @ts-ignore
                              setEventEntityKeyword(val.join(','));
                            }}
                            placeholder="筛选事件实体"
                          >
                            {filtersEventEntity &&
                              filtersEventEntity.map((item: any) => {
                                return (
                                  // <Select.Option
                                  //   key={`${item.ip}:${item.port}`}
                                  //   value={`${item.ip}:${item.port}`}
                                  // >
                                  //   {`${item.ip}:${item.port}`}
                                  // </Select.Option>
                                  <Select.Option
                                    key={`${item.event_entity}`}
                                    value={`${item.event_entity}`}
                                  >
                                    {`${item.event_entity}`}
                                  </Select.Option>
                                );
                              })}
                          </Select>
                          <span>事件指标:</span>

                          <Select
                            mode="multiple"
                            allowClear
                            style={{ width: 200 }}
                            value={(eventKeyKeyword && eventKeyKeyword.split(',')) || []}
                            onChange={(val) => {
                              // @ts-ignore
                              setEventKeyKeyword(val.join(','));
                            }}
                            placeholder="筛选事件指标"
                          >
                            {filtersEventKey &&
                              filtersEventKey.map((item: any) => {
                                return (
                                  <Select.Option
                                    key={`${item.event_key}`}
                                    value={`${item.event_key}`}
                                  >
                                    {`${item.event_key}`}
                                  </Select.Option>
                                );
                              })}
                          </Select>
                        </Space>
                      </Col>
                      <Col flex="280px" style={{ textAlign: 'right' }}>
                        <Space>
                          <Tooltip placement="top" title="重置搜索内容">
                            <Button
                              type="link"
                              icon={<RetweetOutlined />}
                              onClick={() => {
                                reset();
                                did({});
                              }}
                            >
                              重置条件
                            </Button>
                          </Tooltip>
                          <Tooltip placement="top" title="搜索选择内容">
                            <Button
                              type="link"
                              icon={<SearchOutlined />}
                              onClick={() => {
                                didChartsFull({});
                              }}
                            >
                              创建图表
                            </Button>
                          </Tooltip>
                        </Space>
                      </Col>
                    </Row>
                  </Card>
                </Col>
                <Col span={24} style={{ paddingTop: 10 }}>
                  {latestChartFullRef.current && (
                    <EventChart loading={chartFullLoading} chartData={latestChartFullRef.current} eventInfo={(key: string) => getEventInfoMap(`${typeKeyword}.${key}`)} />
                  )}
                </Col>
              </>
            )}
          </Card>
        </Col>
      </Row>

      <EventInfoView
        eventUuid={eventUuid}
        modalVisible={modalVisible}
        onCancel={() => setModalVisible(false)}
      />
    </div>
  );
};
