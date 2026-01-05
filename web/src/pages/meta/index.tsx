import React, {useState } from 'react';
import { PageContainer } from '@ant-design/pro-components';
import { Tabs } from 'antd';
import { 
  DashboardOutlined,
  CloudServerOutlined,
  DatabaseOutlined,
  TableOutlined,
  UnorderedListOutlined,
  CheckCircleOutlined
} from '@ant-design/icons';

import styles from './index.less';


// 导入各个子页面组件
import Dashboard from './dashboard';
import InstanceList from './instance';
import DatabaseList from './database';
import TableList from './table';
import ColumnList from './column';
import QualityDashboard from './quality/index';


const MetaDashboard: React.FC = () => {
  const [activeTab, setActiveTab] = useState('overview');



  const tabItems = [
    {
      key: 'overview',
      label: (
        <span>
          <DashboardOutlined /> 概览
        </span>
      ),
      children: (
        <div>
          <Dashboard />
        </div>
      ),
    },
    {
      key: 'instance',
      label: (
        <span>
          <CloudServerOutlined /> 实例查询
        </span>
      ),
      children: (
        <div>
          <InstanceList />
        </div>
      ),
    },
    {
      key: 'database',
      label: (
        <span>
          <DatabaseOutlined /> 数据库查询
        </span>
      ),
      children: (
        <div>
          <DatabaseList />
        </div>
      ),
    },
    {
      key: 'table',
      label: (
        <span>
          <TableOutlined /> 数据表查询
        </span>
      ),
      children: (
        <div>
          <TableList />
        </div>
      ),
    },
    {
      key: 'column',
      label: (
        <span>
          <UnorderedListOutlined /> 数据列查询
        </span>
      ),
      children: (
        <div>
          <ColumnList />
        </div>
      ),
    },
    {
      key: 'quality',
      label: (
        <span>
          <CheckCircleOutlined /> 数据字典质量
        </span>
      ),
      children: (
        <div>
          <QualityDashboard />
        </div>
      ),
    },
  ];

  return (
    <PageContainer title="数据字典信息查询平台" className={styles.metaDashboard}>
      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        items={tabItems}
        style={{ backgroundColor: '#fff', padding: '16px' }}
      />
    </PageContainer>
  );
};

export default MetaDashboard;
