import React, { useEffect, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import { Row, Col, Card, Space } from 'antd';
import { 
  PieChartTwoTone
} from '@ant-design/icons';
import styles from './index.less';
import PieChart from '@/components/Chart/PieChart';
import { StatisticCard } from '@ant-design/pro-components';

const { Divider } = StatisticCard;

export default (): React.ReactNode => {
  const [dashboardData, setDashboardData] = useState<any>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [datasourcePieData, setDatasourcePieData] = useState<any>([{ type: 'noData', value: 1 }]);
  const [databasePieData, setDatabasePieData] = useState<any>([{ type: 'noData', value: 1 }]);
  const [tablePieData, setTablePieData] = useState<any>([{ type: 'noData', value: 1 }]);
  const [columnPieData, setColumnPieData] = useState<any>([{ type: 'noData', value: 1 }]);

  useEffect(() => {
    try {
      fetch(`/api/v1/meta/dashboard/info`)
        .then((response) => response.json())
        .then((json) => {
          console.info(json.data);
          setDashboardData(json.data);
          setDatasourcePieData(json.data.datasourcePieDataList);
          setDatabasePieData(json.data.databasePieDataList);
          setTablePieData(json.data.tablePieDataList);
          setColumnPieData(json.data.columnPieDataList);
          setLoading(false);
        })
        .catch((error) => {
          console.log('fetch dashboard data failed', error);
          setLoading(false);
        });
    } catch (e) {
      setLoading(false);
    }
  }, []);

  return (
    <>
      {/* 统计卡片区域 */}
      <div className={styles.statisticsSection}>
        <StatisticCard.Group className={styles.statisticGroup}>
          <StatisticCard
            statistic={{
              title: '数据源类型',
              value: dashboardData?.datasourceTypeCount || 0,
              status: 'default',
            }}
            loading={loading}
            className={styles.statisticCard}
          />
          <Divider />
          <StatisticCard
            statistic={{
              title: '机房数量',
              value: dashboardData?.datasourceIdcCount || 0,
              status: 'success',
            }}
            loading={loading}
            className={styles.statisticCard}
          />
          <Divider />
          <StatisticCard
            statistic={{
              title: '环境数量',
              value: dashboardData?.datasourceEnvCount || 0,
              status: 'processing',
            }}
            loading={loading}
            className={styles.statisticCard}
          />
          <Divider />
          <StatisticCard
            statistic={{
              title: '实例总数',
              value: dashboardData?.datasourceCount || 0,
              status: 'warning',
            }}
            loading={loading}
            className={styles.statisticCard}
          />
          <Divider />
          <StatisticCard
            statistic={{
              title: '数据库数',
              value: dashboardData?.databaseCount || 0,
              status: 'success',
            }}
            loading={loading}
            className={styles.statisticCard}
          />
          <Divider />
          <StatisticCard
            statistic={{
              title: '数据表数',
              value: dashboardData?.tableCount || 0,
              status: 'processing',
            }}
            loading={loading}
            className={styles.statisticCard}
          />
          <Divider />
          <StatisticCard
            statistic={{
              title: '字段总数',
              value: dashboardData?.columnCount || 0,
              status: 'warning',
            }}
            loading={loading}
            className={styles.statisticCard}
          />
        </StatisticCard.Group>
      </div>

      {/* 图表区域 */}
      <div className={styles.chartsSection}>
        <Row gutter={[24, 24]}>
          <Col xs={24} sm={24} md={12} lg={12} xl={12}>
            <Card 
              title={
                <Space>
                  <PieChartTwoTone twoToneColor={['#1890ff', '#91d5ff']} />
                  <span>数据源实例分布</span>
                </Space>
              } 
              bordered={false}
              className={styles.chartCard}
              headStyle={{ borderBottom: '1px solid #f0f0f0' }}
            >
              <PieChart data={datasourcePieData} loading={loading} height={330} />
            </Card>
          </Col>
          <Col xs={24} sm={24} md={12} lg={12} xl={12}>
            <Card 
              title={
                <Space>
                  <PieChartTwoTone twoToneColor={['#52c41a', '#b7eb8f']} />
                  <span>数据库分布</span>
                </Space>
              } 
              bordered={false}
              className={styles.chartCard}
              headStyle={{ borderBottom: '1px solid #f0f0f0' }}
            >
              <PieChart data={databasePieData} loading={loading} height={330} />
            </Card>
          </Col>
        </Row>

        <Row gutter={[24, 24]} style={{ marginTop: 24 }}>
          <Col xs={24} sm={24} md={12} lg={12} xl={12}>
            <Card 
              title={
                <Space>
                  <PieChartTwoTone twoToneColor={['#1890ff', '#91d5ff']} />
                  <span>数据表分布</span>
                </Space>
              } 
              bordered={false}
              className={styles.chartCard}
              headStyle={{ borderBottom: '1px solid #f0f0f0' }}
            >
              <PieChart data={tablePieData} loading={loading} height={330} />
            </Card>
          </Col>
          <Col xs={24} sm={24} md={12} lg={12} xl={12}>
            <Card 
              title={
                <Space>
                  <PieChartTwoTone twoToneColor={['#faad14', '#ffe58f']} />
                  <span>数据字段分布</span>
                </Space>
              } 
              bordered={false}
              className={styles.chartCard}
              headStyle={{ borderBottom: '1px solid #f0f0f0' }}
            >
              <PieChart data={columnPieData} loading={loading} height={330} />
            </Card>
          </Col>
        </Row>
      </div>
    </>
  );
};
