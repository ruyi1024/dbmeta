import React, { useEffect, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import { Row, Col, Card, Progress, message } from 'antd';
import { PieChartTwoTone, CheckCircleOutlined } from '@ant-design/icons';
import styles from './index.less';
import PieChart from '@/components/Chart/PieChart';
import { StatisticCard } from '@ant-design/pro-components';
import { queryQualityData } from './service';

const { Divider } = StatisticCard;

export default (): React.ReactNode => {
  const [qualityData, setQualityData] = useState<any>({});
  const [loading, setLoading] = useState<boolean>(true);
  const [databaseQualityData, setDatabaseQualityData] = useState<any>([{ type: 'noData', value: 1 }]);
  const [tableQualityData, setTableQualityData] = useState<any>([{ type: 'noData', value: 1 }]);
  const [columnQualityData, setColumnQualityData] = useState<any>([{ type: 'noData', value: 1 }]);
  const [tableCommentAccuracyData, setTableCommentAccuracyData] = useState<any>([{ type: 'noData', value: 1 }]);
  const [columnCommentAccuracyData, setColumnCommentAccuracyData] = useState<any>([{ type: 'noData', value: 1 }]);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const response = await queryQualityData();
        if (response.success) {
          const data = response.data;
          setQualityData(data);
          setDatabaseQualityData(data.databaseQualityDataList || [{ type: 'noData', value: 1 }]);
          setTableQualityData(data.tableQualityDataList || [{ type: 'noData', value: 1 }]);
          setColumnQualityData(data.columnQualityDataList || [{ type: 'noData', value: 1 }]);
          setTableCommentAccuracyData(data.tableCommentAccuracyDataList || [{ type: 'noData', value: 1 }]);
          setColumnCommentAccuracyData(data.columnCommentAccuracyDataList || [{ type: 'noData', value: 1 }]);
        } else {
          message.error('获取数据失败');
        }
      } catch (error) {
        console.error('获取元数据质量数据失败:', error);
        message.error('获取数据失败，请检查网络连接');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  const getQualityColor = (rate: number) => {
    if (rate >= 80) return '#52c41a';
    if (rate >= 60) return '#faad14';
    return '#ff4d4f';
  };

  return (
    <PageContainer title="数据字典质量大盘">
      <Row gutter={[16, 24]} style={{ marginTop: '10px' }}>
        <Col span={24}>
          <StatisticCard.Group className={styles.statisticGroup}>
            <StatisticCard
              statistic={{
                title: <span style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                  <span style={{ fontSize: '16px' }}>🗄️</span>
                  数据库总数
                </span>,
                value: qualityData?.databaseCount || 0,
                status: 'default',
              }}
              loading={loading}
            />
            <Divider />
            <StatisticCard
              statistic={{
                title: <span style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                  <span style={{ fontSize: '16px' }}>📋</span>
                  数据表总数
                </span>,
                value: qualityData?.tableCount || 0,
                status: 'success',
              }}
              loading={loading}
            />
            <Divider />
            <StatisticCard
              statistic={{
                title: <span style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                  <span style={{ fontSize: '16px' }}>📝</span>
                  数据字段总数
                </span>,
                value: qualityData?.columnCount || 0,
                status: 'processing',
              }}
              loading={loading}
            />
          </StatisticCard.Group>
        </Col>
      </Row>

      <Row gutter={[16, 24]} style={{ marginTop: '15px' }}>
        <Col span={8}>
          <Card 
            title={<span><CheckCircleOutlined style={{ color: getQualityColor(qualityData?.databaseBusinessRate || 0) }} />&nbsp;数据库业务关联率</span>}
            bordered={false}
            className={styles.qualityCard}
          >
            <div style={{ textAlign: 'center', padding: '20px 0' }}>
              <Progress
                type="circle"
                percent={qualityData?.databaseBusinessRate || 0}
                strokeColor={getQualityColor(qualityData?.databaseBusinessRate || 0)}
                format={(percent) => `${percent}%`}
                size={120}
              />
              <div style={{ marginTop: '16px', fontSize: '14px', color: '#666' }}>
                已关联业务数据库占比
              </div>
            </div>
          </Card>
        </Col>
        <Col span={8}>
          <Card 
            title={<span><CheckCircleOutlined style={{ color: getQualityColor(qualityData?.tableCommentRate || 0) }} />&nbsp;数据表注释完备率</span>}
            bordered={false}
            className={styles.qualityCard}
          >
            <div style={{ textAlign: 'center', padding: '20px 0' }}>
              <Progress
                type="circle"
                percent={qualityData?.tableCommentRate || 0}
                strokeColor={getQualityColor(qualityData?.tableCommentRate || 0)}
                format={(percent) => `${percent}%`}
                size={120}
              />
              <div style={{ marginTop: '16px', fontSize: '14px', color: '#666' }}>
                有注释的数据表占比
              </div>
            </div>
          </Card>
        </Col>
        <Col span={8}>
          <Card 
            title={<span><CheckCircleOutlined style={{ color: getQualityColor(qualityData?.columnCommentRate || 0) }} />&nbsp;数据字段注释完备率</span>}
            bordered={false}
            className={styles.qualityCard}
          >
            <div style={{ textAlign: 'center', padding: '20px 0' }}>
              <Progress
                type="circle"
                percent={qualityData?.columnCommentRate || 0}
                strokeColor={getQualityColor(qualityData?.columnCommentRate || 0)}
                format={(percent) => `${percent}%`}
                size={120}
              />
              <div style={{ marginTop: '16px', fontSize: '14px', color: '#666' }}>
                有注释的数据字段占比
              </div>
            </div>
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 24]} style={{ marginTop: '15px' }}>
        <Col span={12}>
          <Card 
            title={<span><CheckCircleOutlined style={{ color: getQualityColor(qualityData?.tableAccuracyRate || 0) }} />&nbsp;数据表备注准确度</span>}
            bordered={false}
            className={styles.qualityCard}
          >
            <div style={{ textAlign: 'center', padding: '20px 0' }}>
              <Progress
                type="circle"
                percent={qualityData?.tableAccuracyRate || 0}
                strokeColor={getQualityColor(qualityData?.tableAccuracyRate || 0)}
                format={(percent) => `${percent}%`}
                size={120}
              />
              <div style={{ marginTop: '16px', fontSize: '14px', color: '#666' }}>
                数据表备注准确度评估
              </div>
            </div>
          </Card>
        </Col>
        <Col span={12}>
          <Card 
            title={<span><CheckCircleOutlined style={{ color: getQualityColor(qualityData?.columnAccuracyRate || 0) }} />&nbsp;数据字段备注准确度</span>}
            bordered={false}
            className={styles.qualityCard}
          >
            <div style={{ textAlign: 'center', padding: '20px 0' }}>
              <Progress
                type="circle"
                percent={qualityData?.columnAccuracyRate || 0}
                strokeColor={getQualityColor(qualityData?.columnAccuracyRate || 0)}
                format={(percent) => `${percent}%`}
                size={120}
              />
              <div style={{ marginTop: '16px', fontSize: '14px', color: '#666' }}>
                数据字段备注准确度评估
              </div>
            </div>
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 24]} style={{ marginTop: '15px' }}>
        <Col span={12}>
          <Card 
            title={<span><PieChartTwoTone />&nbsp;数据库业务关联情况</span>} 
            bordered={false}
            className={styles.chartCard}
          >
            <PieChart data={databaseQualityData} loading={loading} height={330} />
          </Card>
        </Col>
        <Col span={12}>
          <Card 
            title={<span><PieChartTwoTone />&nbsp;数据表注释完备情况</span>} 
            bordered={false}
            className={styles.chartCard}
          >
            <PieChart data={tableQualityData} loading={loading} height={330} />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 24]} style={{ marginTop: '15px' }}>
        <Col span={12}>
          <Card 
            title={<span><PieChartTwoTone />&nbsp;数据字段注释完备情况</span>} 
            bordered={false}
            className={styles.chartCard}
          >
            <PieChart data={columnQualityData} loading={loading} height={330} />
          </Card>
        </Col>
        <Col span={12}>
          <Card 
            title={<span><PieChartTwoTone />&nbsp;表注释准确度分布</span>} 
            bordered={false}
            className={styles.chartCard}
          >
            <PieChart data={tableCommentAccuracyData} loading={loading} height={330} />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 24]} style={{ marginTop: '15px' }}>
        <Col span={12}>
          <Card 
            title={<span><PieChartTwoTone />&nbsp;字段注释准确度分布</span>} 
            bordered={false}
            className={styles.chartCard}
          >
            <PieChart data={columnCommentAccuracyData} loading={loading} height={330} />
          </Card>
        </Col>
      </Row>
    </PageContainer>
  );
}; 