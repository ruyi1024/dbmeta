import { BorderLeftOutlined, HeartOutlined, RightSquareOutlined, TableOutlined, UnorderedListOutlined } from '@ant-design/icons';
import { PageContainer, ProColumns, ProTable } from '@ant-design/pro-components';
import { Card, Tabs, Form, Select, Button, Input, message, Table, Alert, Space, List, Row, Col, AutoComplete, Drawer } from 'antd';

import React, { useState, useEffect, useRef } from 'react';
import moment from 'moment';
import styles from './index.less';



import WaterMarkContent from '@/components/WaterMarkContent'




const Index: React.FC = () => {


  const [currentUserinfo, setCurrentUserinfo] = useState({ "chineseName": "", "username": "" });
  const [currentDate, setCurrentDate] = useState<string>("");



  return (
    <PageContainer>
      <WaterMarkContent text={currentUserinfo.chineseName + "-" + currentDate}>
      <Row style={{ marginTop: '10px' }}>
      <Col span={12}>
      <iframe
        title="Embedded Page"
        width="100%"
        height="600"
        src="http://47.116.68.40/chatbot/0eUg8JcxR4MuFEQL"
        allow="microphone"
        allowFullScreen
      /></Col>
      </Row>
      
      </WaterMarkContent>
    </PageContainer>
  );
};

export default Index;

