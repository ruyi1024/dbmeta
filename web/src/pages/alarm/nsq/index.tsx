import { PlusOutlined, FormOutlined, DeleteOutlined } from '@ant-design/icons';
import { Button, Divider, Input, Form, Card, Row, Col } from 'antd';
import React, { useState, useRef } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import ProTable, { ProColumns, ActionType } from '@ant-design/pro-table';

import { queryLevel, updateLevel, addLevel, removeLevel } from './service';
import { useAccess } from 'umi';



const nsqFrame: React.FC<{}> = () => {


  return (
    <PageContainer>
      <iframe
        title="Embedded Page"
        width="100%"
        height="600"
        src="http://47.102.103.146:4171/counter"
        allowFullScreen
      />
    </PageContainer>
  );
};

export default nsqFrame;
