import React, { useState, useEffect } from 'react';
import { Modal, message } from 'antd';
import ProForm, { ProFormText, ProFormSelect, ProFormTextArea } from '@ant-design/pro-form';
import type { TaskListItem } from '../data.d';
import { request } from 'umi';

export interface CreateFormProps {
  modalVisible: boolean;
  onCancel: () => void;
  onSubmit: (values: TaskListItem) => Promise<void>;
}

interface DatabaseInfo {
  id: number;
  database_name: string;
  alias_name?: string;
  datasource_type?: string;
  host?: string;
  port?: string;
}

const CreateForm: React.FC<CreateFormProps> = (props) => {
  const { modalVisible, onCancel, onSubmit } = props;
  const [databaseList, setDatabaseList] = useState<DatabaseInfo[]>([]);
  const [selectedDatabaseId, setSelectedDatabaseId] = useState<number | undefined>(undefined);

  // 获取数据库列表
  useEffect(() => {
    if (modalVisible) {
      const fetchDatabaseList = async () => {
        try {
          const response = await request('/api/v1/meta/database/list', {
            method: 'GET',
            params: {
              is_deleted: 0,
            },
          });
          if (response.success && response.data) {
            setDatabaseList(response.data);
          }
        } catch (error) {
          console.error('获取数据库列表失败:', error);
          message.error('获取数据库列表失败');
        }
      };
      fetchDatabaseList();
    }
  }, [modalVisible]);

  // 处理数据库选择变化
  const handleDatabaseChange = (value: string) => {
    const db = databaseList.find((d) => d.database_name === value);
    if (db) {
      setSelectedDatabaseId(db.id);
    }
  };

  return (
    <Modal
      destroyOnClose
      title="新建评估任务"
      open={modalVisible}
      onCancel={() => {
        setSelectedDatabaseId(undefined);
        onCancel();
      }}
      footer={null}
      width={600}
    >
      <ProForm
        onFinish={async (value) => {
          // 根据选择的数据库名称，找到对应的数据库ID，设置datasourceId
          const selectedDb = databaseList.find((d) => d.database_name === value.databaseName);
          if (selectedDb) {
            value.datasourceId = selectedDb.id;
          }
          await onSubmit(value as TaskListItem);
        }}
        initialValues={{
          taskType: '全量',
          status: 'pending',
        }}
      >
        <ProFormText
          name="taskName"
          label="任务名称"
          rules={[{ required: true, message: '请输入任务名称' }]}
          placeholder="请输入任务名称"
        />
        <ProFormSelect
          name="taskType"
          label="任务类型"
          rules={[{ required: true, message: '请选择任务类型' }]}
          options={[
            { label: '全量评估', value: '全量' },
            { label: '增量评估', value: '增量' },
            { label: '定时评估', value: '定时' },
          ]}
        />
        <ProFormSelect
          name="databaseName"
          label="数据库"
          rules={[{ required: true, message: '请选择数据库' }]}
          placeholder="请选择数据库"
          showSearch
          options={databaseList.map((db) => {
            const displayText = db.alias_name
              ? `${db.alias_name}(${db.database_name})`
              : db.database_name;
            return {
              label: displayText,
              value: db.database_name,
            };
          })}
          fieldProps={{
            onChange: handleDatabaseChange,
          }}
        />
        <ProFormTextArea
          name="tableFilter"
          label="表过滤条件(JSON)"
          placeholder='例如: {"include": ["table1", "table2"]}'
          fieldProps={{
            rows: 3,
          }}
        />
        <ProFormTextArea
          name="scheduleConfig"
          label="调度配置(JSON)"
          placeholder='例如: {"cron": "0 0 2 * * ?"}'
          fieldProps={{
            rows: 3,
          }}
        />
      </ProForm>
    </Modal>
  );
};

export default CreateForm;

