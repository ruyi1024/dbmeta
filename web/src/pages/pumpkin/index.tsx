import React, { useState } from 'react';
import { PageContainer } from '@ant-design/pro-components';
import { Tabs } from 'antd';
import { 
  DashboardOutlined,
  DatabaseOutlined,
  TableOutlined
} from '@ant-design/icons';

// 导入各个子页面组件
import Overview from './overview/index';
import DatabaseCapacity from './database/index';
import TableCapacity from './table/index';

const PumpkinPage: React.FC = () => {
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
          <Overview />
        </div>
      ),
    },
    {
      key: 'database',
      label: (
        <span>
          <DatabaseOutlined /> 数据库容量分析
        </span>
      ),
      children: (
        <div>
          <DatabaseCapacity />
        </div>
      ),
    },
    {
      key: 'table',
      label: (
        <span>
          <TableOutlined /> 数据表容量分析
        </span>
      ),
      children: (
        <div>
          <TableCapacity />
        </div>
      ),
    },
  ];

  return (
    <PageContainer title="数据容量数据平台">
      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        items={tabItems}
        style={{ backgroundColor: '#fff', padding: '16px' }}
      />
    </PageContainer>
  );
};

export default PumpkinPage;
