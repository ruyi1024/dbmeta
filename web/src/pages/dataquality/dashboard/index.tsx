import React, { useEffect, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import { Row, Col, Card, Progress, Table, Tag, Space, Tooltip } from 'antd';
import { 
  PieChartTwoTone, 
  CheckCircleOutlined, 
  TableOutlined,
  ColumnHeightOutlined,
  ExclamationCircleOutlined,
  RobotOutlined,
  ThunderboltOutlined,
  BulbOutlined
} from '@ant-design/icons';
import styles from './index.less';
import PieChart from '@/components/Chart/PieChart';
import { StatisticCard } from '@ant-design/pro-components';
import type { ColumnsType } from 'antd/es/table';

const { Divider } = StatisticCard;

// 模拟数据（API失败时使用）
const mockData = {
  totalTables: 1250,
  totalColumns: 15680,
  totalIssues: 342,
  fieldCompleteness: 87.5,
  fieldAccuracy: 92.3,
  tableCompleteness: 89.2,
  dataConsistency: 85.6,
  dataUniqueness: 94.1,
  dataTimeliness: 88.7,
  completenessData: [
    { type: '完整', value: 13700 },
    { type: '缺失', value: 1980 },
  ],
  accuracyData: [
    { type: '准确', value: 14470 },
    { type: '格式错误', value: 850 },
    { type: '范围错误', value: 360 },
  ],
  consistencyData: [
    { type: '一致', value: 13420 },
    { type: '不一致', value: 2260 },
  ],
  uniquenessData: [
    { type: '唯一', value: 14750 },
    { type: '重复', value: 930 },
  ],
  issueList: [],
  aiAnalysis: {
    overallScore: 88.2,
    overallLevel: '良好',
    analysisTime: '2024-01-15 10:30:00',
    recommendations: [
      {
        type: 'high',
        title: '字段完整性待提升',
        desc: '检测到1980个字段存在数据缺失，建议优先处理user_info表的email字段，空值率超过20%',
        priority: '高',
      },
      {
        type: 'medium',
        title: '数据格式规范性问题',
        desc: '发现850个字段存在格式错误，主要集中在phone、email等联系信息字段，建议统一格式规范',
        priority: '中',
      },
      {
        type: 'low',
        title: '数据一致性优化',
        desc: '部分关联表数据存在不一致情况，建议检查外键约束和数据同步机制',
        priority: '低',
      },
    ],
    insights: [
      '整体数据质量评分为88.2分，处于良好水平',
      '数据唯一性表现最佳，达到94.1%',
      '字段准确性需要重点关注，存在格式和范围错误',
      '建议优先处理高优先级问题，预计可提升整体质量5-8分',
    ],
    trendAnalysis: '近30天数据质量呈上升趋势，较上月提升2.3%',
  },
};

export default (): React.ReactNode => {
  const [loading, setLoading] = useState<boolean>(true);
  const [dashboardData, setDashboardData] = useState<any>(mockData);

  useEffect(() => {
    // 从API获取数据
    const fetchData = async () => {
      try {
        const response = await fetch('/api/v1/dataquality/dashboard/info');
        const result = await response.json();
        if (result.code === 200) {
          // 处理API返回的数据结构
          const data = result.data;
          setDashboardData({
            totalTables: data.totalTables || 0,
            totalColumns: data.totalColumns || 0,
            totalIssues: data.totalIssues || 0,
            fieldCompleteness: data.fieldCompleteness || 0,
            fieldAccuracy: data.fieldAccuracy || 0,
            tableCompleteness: data.tableCompleteness || 0,
            dataConsistency: data.dataConsistency || 0,
            dataUniqueness: data.dataUniqueness || 0,
            dataTimeliness: data.dataTimeliness || 0,
            completenessData: data.completenessData || [],
            accuracyData: data.accuracyData || [],
            consistencyData: data.consistencyData || [],
            uniquenessData: data.uniquenessData || [],
            issueList: data.issueList || [],
            aiAnalysis: data.aiAnalysis || mockData.aiAnalysis,
          });
        } else {
          // 如果API失败，使用模拟数据
          setDashboardData(mockData);
        }
      } catch (error) {
        console.error('获取数据质量信息失败:', error);
        // API失败时使用模拟数据
        setDashboardData(mockData);
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, []);

  const getQualityColor = (rate: number) => {
    if (rate >= 90) return '#52c41a';
    if (rate >= 80) return '#1890ff';
    if (rate >= 70) return '#faad14';
    return '#ff4d4f';
  };

  const getQualityStatus = (rate: number) => {
    if (rate >= 90) return 'success';
    if (rate >= 80) return 'processing';
    if (rate >= 70) return 'warning';
    return 'error';
  };

  const getIssueLevelColor = (level: string) => {
    switch (level) {
      case 'high':
        return 'red';
      case 'medium':
        return 'orange';
      case 'low':
        return 'blue';
      default:
        return 'default';
    }
  };

  const getIssueLevelText = (level: string) => {
    switch (level) {
      case 'high':
        return '高';
      case 'medium':
        return '中';
      case 'low':
        return '低';
      default:
        return '未知';
    }
  };

  const columns: ColumnsType<any> = [
    {
      title: '表名',
      dataIndex: 'tableName',
      key: 'tableName',
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
      key: 'columnName',
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
      key: 'issueType',
      width: 120,
    },
    {
      title: '严重程度',
      dataIndex: 'issueLevel',
      key: 'issueLevel',
      width: 100,
      render: (level: string) => (
        <Tag color={getIssueLevelColor(level)}>{getIssueLevelText(level)}</Tag>
      ),
    },
    {
      title: '问题描述',
      dataIndex: 'issueDesc',
      key: 'issueDesc',
      ellipsis: {
        showTitle: false,
      },
      render: (desc: string) => (
        <Tooltip placement="topLeft" title={desc}>
          {desc}
        </Tooltip>
      ),
    },
    {
      title: '问题数量',
      dataIndex: 'issueCount',
      key: 'issueCount',
      width: 100,
      sorter: (a: any, b: any) => a.issueCount - b.issueCount,
    },
    {
      title: '最后检查时间',
      dataIndex: 'lastCheckTime',
      key: 'lastCheckTime',
      width: 180,
    },
    {
      title: 'AI分析',
      key: 'aiAnalysis',
      width: 120,
      render: (_: any, record: any) => (
        <Tooltip 
          title={
            <div>
              <div style={{ marginBottom: 8 }}>
                <strong>AI评估：</strong>该问题属于{record.issueLevel === 'high' ? '高' : record.issueLevel === 'medium' ? '中' : '低'}优先级
              </div>
              <div>
                <strong>建议：</strong>建议优先处理，预计修复后可提升数据质量{record.issueLevel === 'high' ? '3-5' : record.issueLevel === 'medium' ? '1-3' : '0.5-1'}分
              </div>
            </div>
          }
        >
          <Tag color="blue" icon={<RobotOutlined />}>
            AI分析
          </Tag>
        </Tooltip>
      ),
    },
  ];

  return (
    <PageContainer 
      title={
        <Space>
          <RobotOutlined style={{ color: '#1890ff', fontSize: '20px' }} />
          <span>AI数据质量分析</span>
          <Tag color="blue" icon={<ThunderboltOutlined />} style={{ marginLeft: 8 }}>
            AI智能评估
          </Tag>
        </Space>
      }
      className={styles.dataQualityDashboard}
    >
      {/* AI分析概览卡片 */}
      <div className={styles.aiAnalysisSection}>
        <Card 
          className={styles.aiAnalysisCard}
          bordered={false}
        >
          <Row gutter={24}>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <div className={styles.aiScoreCard}>
                <div className={styles.aiScoreHeader}>
                  <RobotOutlined className={styles.aiIcon} />
                  <span className={styles.aiLabel}>AI综合评分</span>
                </div>
                <div className={styles.aiScoreValue}>
                  {dashboardData?.aiAnalysis?.overallScore || 0}
                  <span className={styles.aiScoreUnit}>分</span>
                </div>
                <div className={styles.aiScoreLevel}>
                  <Tag color={getQualityColor(dashboardData?.aiAnalysis?.overallScore || 0)}>
                    {dashboardData?.aiAnalysis?.overallLevel || '良好'}
                  </Tag>
                </div>
                <div className={styles.aiAnalysisTime}>
                  <span>分析时间：{dashboardData?.aiAnalysis?.analysisTime || '--'}</span>
                </div>
              </div>
            </Col>
            <Col xs={24} sm={24} md={16} lg={16} xl={16}>
              <div className={styles.aiInsightsCard}>
                <div className={styles.aiInsightsHeader}>
                  <BulbOutlined className={styles.aiInsightsIcon} />
                  <span>AI智能洞察</span>
                </div>
                <div className={styles.aiInsightsList}>
                  {(dashboardData?.aiAnalysis?.insights || []).map((insight: string, index: number) => (
                    <div key={index} className={styles.aiInsightItem}>
                      <CheckCircleOutlined className={styles.aiInsightIcon} />
                      <span>{insight}</span>
                    </div>
                  ))}
                </div>
                <div className={styles.aiTrend}>
                  <ThunderboltOutlined style={{ color: '#52c41a', marginRight: 8 }} />
                  <span>{dashboardData?.aiAnalysis?.trendAnalysis || ''}</span>
                </div>
              </div>
            </Col>
          </Row>
        </Card>
      </div>

      {/* AI优化建议区域 */}
      <div className={styles.aiRecommendationsSection}>
        <Card 
          title={
            <Space>
              <RobotOutlined style={{ color: '#1890ff' }} />
              <span>AI优化建议</span>
            </Space>
          }
          bordered={false}
          className={styles.aiRecommendationsCard}
        >
          <Row gutter={[16, 16]}>
            {(dashboardData?.aiAnalysis?.recommendations || []).map((rec: any, index: number) => (
              <Col xs={24} sm={24} md={12} lg={8} xl={8} key={index}>
                <Card
                  className={styles.recommendationCard}
                  style={{
                    borderLeft: `4px solid ${
                      rec.type === 'high' ? '#ff4d4f' : rec.type === 'medium' ? '#faad14' : '#1890ff'
                    }`,
                  }}
                >
                  <div className={styles.recommendationHeader}>
                    <Tag 
                      color={
                        rec.type === 'high' ? 'red' : rec.type === 'medium' ? 'orange' : 'blue'
                      }
                    >
                      {rec.priority}优先级
                    </Tag>
                  </div>
                  <div className={styles.recommendationTitle}>{rec.title}</div>
                  <div className={styles.recommendationDesc}>{rec.desc}</div>
                </Card>
              </Col>
            ))}
          </Row>
        </Card>
      </div>

      {/* 统计卡片区域 */}
      <div className={styles.statisticsSection}>
        <StatisticCard.Group className={styles.statisticGroup}>
          <StatisticCard
            statistic={{
              title: '数据表总数',
              value: dashboardData?.totalTables || 0,
              status: 'default',
            }}
            loading={loading}
            className={styles.statisticCard}
          />
          <Divider />
          <StatisticCard
            statistic={{
              title: '数据字段总数',
              value: dashboardData?.totalColumns || 0,
              status: 'success',
            }}
            loading={loading}
            className={styles.statisticCard}
          />
          <Divider />
          <StatisticCard
            statistic={{
              title: '质量问题总数',
              value: dashboardData?.totalIssues || 0,
              status: 'warning',
            }}
            loading={loading}
            className={styles.statisticCard}
          />
          <Divider />
          <StatisticCard
            statistic={{
              title: '整体质量评分',
              value: (
                Math.round(
                  ((dashboardData?.fieldCompleteness || 0) +
                   (dashboardData?.fieldAccuracy || 0) +
                   (dashboardData?.tableCompleteness || 0) +
                   (dashboardData?.dataConsistency || 0) +
                   (dashboardData?.dataUniqueness || 0) +
                   (dashboardData?.dataTimeliness || 0)) / 6
                )
              ),
              suffix: '/ 100',
              status: getQualityStatus(
                Math.round(
                  ((dashboardData?.fieldCompleteness || 0) +
                   (dashboardData?.fieldAccuracy || 0) +
                   (dashboardData?.tableCompleteness || 0) +
                   (dashboardData?.dataConsistency || 0) +
                   (dashboardData?.dataUniqueness || 0) +
                   (dashboardData?.dataTimeliness || 0)) / 6
                )
              ),
            }}
            loading={loading}
            className={styles.statisticCard}
          />
        </StatisticCard.Group>
      </div>

      {/* 质量指标区域 */}
      <div className={styles.qualityMetricsSection}>
        <Row gutter={[24, 24]}>
          <Col xs={24} sm={12} md={12} lg={8} xl={8}>
            <Card 
              title={
                <Space>
                  <CheckCircleOutlined style={{ color: getQualityColor(dashboardData?.fieldCompleteness || 0) }} />
                  <span>字段完整性</span>
                  <Tag color="blue" icon={<RobotOutlined />} style={{ fontSize: '11px' }}>
                    AI评估
                  </Tag>
                </Space>
              }
              bordered={false}
              className={styles.metricCard}
            >
              <div className={styles.metricContent}>
                <div className={styles.metricValue}>
                  {dashboardData?.fieldCompleteness || 0}%
                </div>
                <Progress
                  percent={dashboardData?.fieldCompleteness || 0}
                  strokeColor={getQualityColor(dashboardData?.fieldCompleteness || 0)}
                  showInfo={false}
                />
                <div className={styles.metricDesc}>
                  字段数据完整率，反映数据缺失情况
                </div>
              </div>
            </Card>
          </Col>
          <Col xs={24} sm={12} md={12} lg={8} xl={8}>
            <Card 
              title={
                <Space>
                  <CheckCircleOutlined style={{ color: getQualityColor(dashboardData?.fieldAccuracy || 0) }} />
                  <span>字段准确性</span>
                  <Tag color="blue" icon={<RobotOutlined />} style={{ fontSize: '11px' }}>
                    AI评估
                  </Tag>
                </Space>
              }
              bordered={false}
              className={styles.metricCard}
            >
              <div className={styles.metricContent}>
                <div className={styles.metricValue}>
                  {dashboardData?.fieldAccuracy || 0}%
                </div>
                <Progress
                  percent={dashboardData?.fieldAccuracy || 0}
                  strokeColor={getQualityColor(dashboardData?.fieldAccuracy || 0)}
                  showInfo={false}
                />
                <div className={styles.metricDesc}>
                  字段数据准确率，反映格式和范围正确性
                </div>
              </div>
            </Card>
          </Col>
          <Col xs={24} sm={12} md={12} lg={8} xl={8}>
            <Card 
              title={
                <Space>
                  <CheckCircleOutlined style={{ color: getQualityColor(dashboardData?.tableCompleteness || 0) }} />
                  <span>表完整性</span>
                  <Tag color="blue" icon={<RobotOutlined />} style={{ fontSize: '11px' }}>
                    AI评估
                  </Tag>
                </Space>
              }
              bordered={false}
              className={styles.metricCard}
            >
              <div className={styles.metricContent}>
                <div className={styles.metricValue}>
                  {dashboardData?.tableCompleteness || 0}%
                </div>
                <Progress
                  percent={dashboardData?.tableCompleteness || 0}
                  strokeColor={getQualityColor(dashboardData?.tableCompleteness || 0)}
                  showInfo={false}
                />
                <div className={styles.metricDesc}>
                  表结构完整性，反映表结构规范性
                </div>
              </div>
            </Card>
          </Col>
          <Col xs={24} sm={12} md={12} lg={8} xl={8}>
            <Card 
              title={
                <Space>
                  <CheckCircleOutlined style={{ color: getQualityColor(dashboardData?.dataConsistency || 0) }} />
                  <span>数据一致性</span>
                  <Tag color="blue" icon={<RobotOutlined />} style={{ fontSize: '11px' }}>
                    AI评估
                  </Tag>
                </Space>
              }
              bordered={false}
              className={styles.metricCard}
            >
              <div className={styles.metricContent}>
                <div className={styles.metricValue}>
                  {dashboardData?.dataConsistency || 0}%
                </div>
                <Progress
                  percent={dashboardData?.dataConsistency || 0}
                  strokeColor={getQualityColor(dashboardData?.dataConsistency || 0)}
                  showInfo={false}
                />
                <div className={styles.metricDesc}>
                  跨表数据一致性，反映关联数据正确性
                </div>
              </div>
            </Card>
          </Col>
          <Col xs={24} sm={12} md={12} lg={8} xl={8}>
            <Card 
              title={
                <Space>
                  <CheckCircleOutlined style={{ color: getQualityColor(dashboardData?.dataUniqueness || 0) }} />
                  <span>数据唯一性</span>
                  <Tag color="blue" icon={<RobotOutlined />} style={{ fontSize: '11px' }}>
                    AI评估
                  </Tag>
                </Space>
              }
              bordered={false}
              className={styles.metricCard}
            >
              <div className={styles.metricContent}>
                <div className={styles.metricValue}>
                  {dashboardData?.dataUniqueness || 0}%
                </div>
                <Progress
                  percent={dashboardData?.dataUniqueness || 0}
                  strokeColor={getQualityColor(dashboardData?.dataUniqueness || 0)}
                  showInfo={false}
                />
                <div className={styles.metricDesc}>
                  主键和业务唯一性，反映重复数据情况
                </div>
              </div>
            </Card>
          </Col>
          <Col xs={24} sm={12} md={12} lg={8} xl={8}>
            <Card 
              title={
                <Space>
                  <CheckCircleOutlined style={{ color: getQualityColor(dashboardData?.dataTimeliness || 0) }} />
                  <span>数据及时性</span>
                  <Tag color="blue" icon={<RobotOutlined />} style={{ fontSize: '11px' }}>
                    AI评估
                  </Tag>
                </Space>
              }
              bordered={false}
              className={styles.metricCard}
            >
              <div className={styles.metricContent}>
                <div className={styles.metricValue}>
                  {dashboardData?.dataTimeliness || 0}%
                </div>
                <Progress
                  percent={dashboardData?.dataTimeliness || 0}
                  strokeColor={getQualityColor(dashboardData?.dataTimeliness || 0)}
                  showInfo={false}
                />
                <div className={styles.metricDesc}>
                  数据更新及时性，反映数据新鲜度
                </div>
              </div>
            </Card>
          </Col>
        </Row>
      </div>

      {/* 图表区域 */}
      <div className={styles.chartsSection}>
        <Row gutter={[24, 24]}>
          <Col xs={24} sm={24} md={12} lg={12} xl={12}>
            <Card 
              title={
                <Space>
                  <PieChartTwoTone twoToneColor={['#52c41a', '#ff4d4f']} />
                  <span>字段完整性分布</span>
                </Space>
              } 
              bordered={false}
              className={styles.chartCard}
            >
              <PieChart data={dashboardData?.completenessData || []} loading={loading} height={300} />
            </Card>
          </Col>
          <Col xs={24} sm={24} md={12} lg={12} xl={12}>
            <Card 
              title={
                <Space>
                  <PieChartTwoTone twoToneColor={['#1890ff', '#faad14', '#ff4d4f']} />
                  <span>字段准确性分布</span>
                </Space>
              } 
              bordered={false}
              className={styles.chartCard}
            >
              <PieChart data={dashboardData?.accuracyData || []} loading={loading} height={300} />
            </Card>
          </Col>
        </Row>
        <Row gutter={[24, 24]} style={{ marginTop: 24 }}>
          <Col xs={24} sm={24} md={12} lg={12} xl={12}>
            <Card 
              title={
                <Space>
                  <PieChartTwoTone twoToneColor={['#52c41a', '#ff7875']} />
                  <span>数据一致性分布</span>
                </Space>
              } 
              bordered={false}
              className={styles.chartCard}
            >
              <PieChart data={dashboardData?.consistencyData || []} loading={loading} height={300} />
            </Card>
          </Col>
          <Col xs={24} sm={24} md={12} lg={12} xl={12}>
            <Card 
              title={
                <Space>
                  <PieChartTwoTone twoToneColor={['#52c41a', '#ff7875']} />
                  <span>数据唯一性分布</span>
                </Space>
              } 
              bordered={false}
              className={styles.chartCard}
            >
              <PieChart data={dashboardData?.uniquenessData || []} loading={loading} height={300} />
            </Card>
          </Col>
        </Row>
      </div>

      {/* 问题列表区域 */}
      <div className={styles.issuesSection}>
        <Card 
          title={
            <Space>
              <ExclamationCircleOutlined style={{ color: '#ff4d4f' }} />
              <span>质量问题列表</span>
              <Tag color="blue" icon={<RobotOutlined />} style={{ marginLeft: 8 }}>
                AI智能诊断
              </Tag>
            </Space>
          }
          bordered={false}
          className={styles.issuesCard}
        >
          <Table
            columns={columns}
            dataSource={dashboardData?.issueList || []}
            loading={loading}
            pagination={{
              pageSize: 10,
              showSizeChanger: true,
              showTotal: (total: number) => `共 ${total} 条问题`,
            }}
            rowKey="key"
          />
        </Card>
      </div>
    </PageContainer>
  );
};

