import React, { useEffect, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import { Row, Col, Card, Alert, message, Tooltip, Table, Space } from 'antd';
import { InfoCircleOutlined, SmileTwoTone, PieChartTwoTone, LineChartOutlined, ProfileTwoTone, SoundTwoTone } from '@ant-design/icons';
import styles from './index.less';
import { ChartCard, MiniArea, MiniBar } from './components/Charts';
import Trend from './components/Trend';
import { Gauge } from '@ant-design/charts';
import PieChart from '@/components/Chart/PieChart';
import LineChart from '@/components/Chart/LineChart';
import BarChart from '@/components/Chart/BarChart';
import { StatisticCard } from '@ant-design/pro-components';
import moment from "moment";

const { Divider } = StatisticCard;



const demoPie = [
  {
    type: '分类一',
    value: 27,
  },
  {
    type: '分类二',
    value: 25,
  },
  {
    type: '分类三',
    value: 18,
  },
  {
    type: '分类四',
    value: 15,
  },
  {
    type: '分类五',
    value: 10,
  },
  {
    type: '其他',
    value: 5,
  },
];

export default (): React.ReactNode => {
  const [dashboardData, setDashboardData] = useState<any>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [queryStatusPieDataList, setQueryStatusPieDataList] = useState<any>([{ type: 'noData', value: 1 }]);
  const [queryTypePieDataList, setQueryTypePieDataList] = useState<any>([{ type: 'noData', value: 1 }]);
  const [sensitiveDsTypePieDataList, setSensitiveDsTypePieDataList] = useState<any>([{ type: 'noData', value: 1 }]);
  const [sensitiveTypePieDataList, setSensitiveTypePieDataList] = useState<any>([{ type: 'noData', value: 1 }]);

  // const columns_event = [
  //   {
  //     title: '事件时间',
  //     dataIndex: 'event_time',
  //   },
  //   {
  //     title: '事件类型',
  //     dataIndex: 'event_type',
  //   },
  //   {
  //     title: '事件组',
  //     dataIndex: 'event_group',
  //   },
  //   {
  //     title: '事件实体',
  //     dataIndex: 'event_entity',
  //   },
  //   {
  //     title: '事件指标',
  //     dataIndex: 'event_key',
  //   },
  //   {
  //     title: '事件值',
  //     dataIndex: 'event_value',
  //     render: (_: any, record: any) => <>{record.event_value}</>,
  //   },
  // ];

  // const columns_alarm = [
  //   {
  //     title: '告警时间',
  //     dataIndex: 'event_time',
  //   },
  //   {
  //     title: '告警信息',
  //     dataIndex: 'alarm_title',
  //   },
  //   {
  //     title: '告警级别',
  //     dataIndex: 'alarm_level',
  //   },
  //   {
  //     title: '事件类型',
  //     dataIndex: 'event_type',
  //   },
  //   {
  //     title: '事件实体',
  //     dataIndex: 'event_entity',
  //   },
  // ];

  useEffect(() => {
    try {
      fetch(`/api/v1/safe/dashboard/info`)
        .then((response) => response.json())
        .then((json) => {
          console.info(json.data);
          return (
            setDashboardData(json.data),
            setQueryStatusPieDataList(json.data.queryStatusPieDataList),
            setQueryTypePieDataList(json.data.queryTypePieDataList),
            setSensitiveDsTypePieDataList(json.data.sensitiveDsTypePieDataList),
            setSensitiveTypePieDataList(json.data.sensitiveTypePieDataList),
            setLoading(false)
          );
        })
        .catch((error) => {
          console.log('fetch dashboard data failed', error);
        });
    } catch (e) {
      message.error(`get data error. ${e}`)
    }
  }, []);

  const columns_intercept = [
    {
      title: '拦截原因',
      dataIndex: 'result',
    },
    {
      title: '执行类型',
      dataIndex: 'sql_type',
    },
    {
      title: '类型',
      dataIndex: 'datasource_type',
    },
    {
      title: '数据库',
      dataIndex: 'database',
    },
    {
      title: '用户',
      dataIndex: 'username',
    },
    {
      title: '日期',
      dataIndex: 'gmt_created',
    },
  ];

  const columns_sensitive = [
    {
      title: '敏感类型',
      dataIndex: 'rule_name',
    },
    {
      title: '源类型',
      dataIndex: 'datasource_type',
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
      title: '数据字段',
      dataIndex: 'column_name',
    },
    {
      title: '日期',
      dataIndex: 'gmt_created',
    },
  ];


  return (
    <PageContainer>
      <Row gutter={[16, 24]} style={{ marginTop: '10px' }}>
        <Col span={24}>
          <StatisticCard.Group>
            <StatisticCard
              statistic={{
                title: '今日查询',
                value: dashboardData && dashboardData.todayQueryCount,
                status: 'default',
              }}
            />
            <Divider />
            <StatisticCard
              statistic={{
                title: '查询总数',
                value: dashboardData && dashboardData.totalQueryCount,
                status: 'success',
              }}
            />
            <Divider />
            <StatisticCard
              statistic={{
                title: '高危拦截',
                value: dashboardData && dashboardData.totalInterceptCount,
                status: 'processing',
              }}
            />
            <Divider />
            <StatisticCard
              statistic={{
                title: '敏感数据库',
                value: dashboardData && dashboardData.sensitiveDatabaseCount,
                status: 'warning',
              }}
            />
            <StatisticCard
              statistic={{
                title: '敏感数据表',
                value: dashboardData && dashboardData.sensitiveTableCount,
                status: 'warning',
              }}
            />
            <StatisticCard
              statistic={{
                title: '敏感数据字段',
                value: dashboardData && dashboardData.sensitiveColumnCount,
                status: 'warning',
              }}
            />
            <Divider />
            <StatisticCard
              statistic={{
                title: '敏感数据保护次数',
                value: dashboardData && dashboardData.sensitiveQueryCount,
                status: 'processing',
              }}
            />
          </StatisticCard.Group>
        </Col>
      </Row>

      <Row gutter={[16, 24]} style={{ marginTop: '15px' }}>
        <Col span={12}>
          <Card title={<span><PieChartTwoTone />&nbsp;SQL执行状态分布</span>} bordered={false}>
            <PieChart data={queryStatusPieDataList} loading={loading} height={330} />
          </Card>
        </Col>
        <Col span={12}>
          <Card title={<span><PieChartTwoTone />&nbsp;SQL执行类型分布</span>} bordered={false}>
            <PieChart data={queryTypePieDataList} loading={loading} height={330} />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 24]} style={{ marginTop: '15px' }}>
        <Col span={12}>
          <Card title={<span><LineChartOutlined />&nbsp;近15日SQL查询趋势</span>} bordered={false}>
            {dashboardData.queryDayLineDataList && (
              <LineChart data={dashboardData.query15DayLineDataList} unit="" />
            )}
          </Card>
        </Col>
        <Col span={12}>
          <Card title={<span><LineChartOutlined />&nbsp;近15日SQL拦截趋势</span>} bordered={false}>
            {dashboardData.query15DayInterceptLineDataList && (
              <LineChart data={dashboardData.query15DayInterceptLineDataList} unit="" />
            )}
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 24]} style={{ marginTop: '15px' }}>
        <Col span={24}>
          <Card title={<span><LineChartOutlined />&nbsp;年度SQL查询和风险拦截统计</span>} bordered={false}>
            {dashboardData.queryMonthBarDataList && (
              <BarChart data={dashboardData.queryMonthBarDataList} unit="" />
            )}
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 24]} style={{ marginTop: '15px' }}>
        <Col span={12}>
          <Card title={<span><PieChartTwoTone />&nbsp;敏感数据数据源分布</span>} bordered={false}>
            <PieChart data={sensitiveDsTypePieDataList} loading={loading} height={330} />
          </Card>
        </Col>
        <Col span={12}>
          <Card title={<span><PieChartTwoTone />&nbsp;敏感数据类型分布</span>} bordered={false}>
            <PieChart data={sensitiveTypePieDataList} loading={loading} height={330} />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 24]} style={{ marginTop: '15px' }}>
        <Col span={12}>
          <Card title={<span><ProfileTwoTone />&nbsp;SQL执行最新拦截记录</span>} bordered={false} style={{ paddingBottom: '8px' }}>
            <Table
              columns={columns_intercept}
              loading={loading}
              dataSource={dashboardData.queryNewInterceptDataList}
              size="small"
              pagination={false}
            />
          </Card>
        </Col>
        <Col span={12}>
          <Card title={<span><ProfileTwoTone />&nbsp;敏感信息最新探测记录</span>} bordered={false} style={{ paddingBottom: '8px' }}>
            <Table
              columns={columns_sensitive}
              loading={loading}
              dataSource={dashboardData.queryNewSensitiveDataList}
              size="small"
              pagination={false}
            />
          </Card>
        </Col>
      </Row>

    </PageContainer>
  );
};
