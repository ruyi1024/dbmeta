import React, { useState } from 'react';
import { PageContainer } from '@ant-design/pro-components';
import { Tabs } from 'antd';
import { 
  DashboardOutlined,
  BarChartOutlined
} from '@ant-design/icons';

// 导入各个子页面组件
import Dashboard from './dashboard/index';
import EventList from './event/index';

const EventPage: React.FC = () => {
  const [activeTab, setActiveTab] = useState('dashboard');

  const tabItems = [
    {
      key: 'dashboard',
      label: (
        <span>
          <DashboardOutlined /> 事件概览
        </span>
      ),
      children: (
        <div>
          <Dashboard />
        </div>
      ),
    },
    {
      key: 'event',
      label: (
        <span>
          <BarChartOutlined /> 事件数据查询与分析
        </span>
      ),
      children: (
        <div>
          <EventList />
        </div>
      ),
    },
  ];

  return (
    <PageContainer title="事件平台">
      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        items={tabItems}
        style={{ backgroundColor: '#fff', padding: '16px' }}
      />
    </PageContainer>
  );
};

export default EventPage;
