import React, { useState, useRef, useEffect } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import { Input, Button, Table, message, Card, Space, Typography, Spin, Alert, Select } from 'antd';
import { SendOutlined, DatabaseOutlined } from '@ant-design/icons';
import { queryDatabase, DbQueryRequest, getDatabaseList, DatabaseInfo } from './service';
import styles from './index.less';

const { Text } = Typography;
const { Option } = Select;

const DbQuery: React.FC = () => {
  const [inputValue, setInputValue] = useState('');
  const [loading, setLoading] = useState(false);
  const [sqlQuery, setSqlQuery] = useState<string>('');
  const [queryResult, setQueryResult] = useState<Array<Record<string, any>>>([]);
  const [total, setTotal] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [columns, setColumns] = useState<any[]>([]);
  const [errorMessage, setErrorMessage] = useState<string>('');
  const [databaseList, setDatabaseList] = useState<DatabaseInfo[]>([]);
  const [selectedDatabase, setSelectedDatabase] = useState<string | undefined>(undefined);
  const inputRef = useRef<any>(null);

  // 获取数据库列表
  useEffect(() => {
    const fetchDatabaseList = async () => {
      try {
        const response = await getDatabaseList();
        if (response.success && response.data) {
          setDatabaseList(response.data);
        }
      } catch (error) {
        console.error('获取数据库列表失败:', error);
      }
    };
    fetchDatabaseList();
  }, []);

  // 处理查询
  const handleQuery = async () => {
    if (!inputValue.trim()) {
      message.warning('请输入查询问题');
      return;
    }

    setLoading(true);
    setSqlQuery('');
    setQueryResult([]);
    setTotal(0);
    setCurrentPage(1);
    setErrorMessage(''); // 清除之前的错误信息

    try {
      // 如果选择了数据库，获取完整的数据库信息
      let dbInfo: Partial<DbQueryRequest> = {};
      if (selectedDatabase) {
        const selectedDb = databaseList.find(db => db.database_name === selectedDatabase);
        if (selectedDb) {
          dbInfo = {
            database_name: selectedDb.database_name,
            datasource_type: selectedDb.datasource_type,
            host: selectedDb.host,
            port: selectedDb.port,
          };
        }
      }

      const params: DbQueryRequest = {
        question: inputValue.trim(),
        page: 1,
        page_size: pageSize,
        ...dbInfo,
      };

      const response = await queryDatabase(params);
      
      if (response.success && response.data) {
        setSqlQuery(response.data.sql_query || '');
        setQueryResult(response.data.query_result || []);
        setTotal(response.data.total || 0);
        setErrorMessage(''); // 成功时清除错误信息
        
        // 根据查询结果动态生成表格列
        if (response.data.query_result && response.data.query_result.length > 0) {
          const firstRow = response.data.query_result[0];
          const newColumns = Object.keys(firstRow).map((key) => ({
            title: key,
            dataIndex: key,
            key: key,
            ellipsis: true,
            width: 150,
          }));
          setColumns(newColumns);
        } else {
          setColumns([]);
        }
      } else {
        // 显示错误信息在警示条中，不使用弹窗
        // 优先使用 response.message，如果没有则使用默认提示
        const errorMsg = response?.message || '查询失败';
        setErrorMessage(errorMsg);
        setSqlQuery('');
        setQueryResult([]);
        setTotal(0);
        setColumns([]);
      }
    } catch (error: any) {
      console.error('查询失败:', error);
      // 显示错误信息在警示条中，不使用弹窗
      // 尝试从多个可能的属性中提取错误信息
      let errorMsg = '查询失败，请重试';
      if (error?.response?.data?.message) {
        errorMsg = error.response.data.message;
      } else if (error?.data?.message) {
        errorMsg = error.data.message;
      } else if (error?.message) {
        errorMsg = error.message;
      } else if (typeof error === 'string') {
        errorMsg = error;
      }
      setErrorMessage(errorMsg);
      setSqlQuery('');
      setQueryResult([]);
      setTotal(0);
      setColumns([]);
    } finally {
      setLoading(false);
    }
  };

  // 处理分页变化
  const handleTableChange = async (page: number, size?: number) => {
    if (!inputValue.trim()) {
      return;
    }

    setLoading(true);
    setCurrentPage(page);
    if (size) {
      setPageSize(size);
    }

    try {
      // 如果选择了数据库，获取完整的数据库信息
      let dbInfo: Partial<DbQueryRequest> = {};
      if (selectedDatabase) {
        const selectedDb = databaseList.find(db => db.database_name === selectedDatabase);
        if (selectedDb) {
          dbInfo = {
            database_name: selectedDb.database_name,
            datasource_type: selectedDb.datasource_type,
            host: selectedDb.host,
            port: selectedDb.port,
          };
        }
      }

      const params: DbQueryRequest = {
        question: inputValue.trim(),
        page: page,
        page_size: size || pageSize,
        ...dbInfo,
      };

      const response = await queryDatabase(params);
      
      if (response.success && response.data) {
        setQueryResult(response.data.query_result || []);
        setTotal(response.data.total || 0);
        setErrorMessage(''); // 成功时清除错误信息
      } else {
        // 显示错误信息在警示条中，不使用弹窗
        // 优先使用 response.message，如果没有则使用默认提示
        const errorMsg = response?.message || '查询失败';
        setErrorMessage(errorMsg);
        setQueryResult([]);
        setTotal(0);
      }
    } catch (error: any) {
      console.error('查询失败:', error);
      // 显示错误信息在警示条中，不使用弹窗
      // 尝试从多个可能的属性中提取错误信息
      let errorMsg = '查询失败，请重试';
      if (error?.response?.data?.message) {
        errorMsg = error.response.data.message;
      } else if (error?.data?.message) {
        errorMsg = error.data.message;
      } else if (error?.message) {
        errorMsg = error.message;
      } else if (typeof error === 'string') {
        errorMsg = error;
      }
      setErrorMessage(errorMsg);
      setQueryResult([]);
      setTotal(0);
    } finally {
      setLoading(false);
    }
  };


  return (
    <PageContainer>
      <div className={styles.dbQueryContainer}>
        <Card className={styles.queryCard}>
          <Space direction="vertical" size="large" style={{ width: '100%' }}>
            {/* 输入区域 */}
            <div className={styles.inputSection}>
              <div className={styles.inputWrapper}>
                <Select
                  value={selectedDatabase}
                  onChange={(value: string | undefined) => setSelectedDatabase(value)}
                  placeholder="选择数据库"
                  allowClear
                  className={styles.databaseSelect}
                  size="large"
                  showSearch
                  filterOption={(input: string, option: any) =>
                    (option?.children as unknown as string)?.toLowerCase().includes(input.toLowerCase())
                  }
                >
                  {databaseList.map((db) => {
                    const displayText = db.alias_name
                      ? `${db.alias_name}(${db.database_name})`
                      : db.database_name;
                    return (
                      <Option key={db.database_name} value={db.database_name}>
                        {displayText}
                      </Option>
                    );
                  })}
                </Select>
                <Input
                  ref={inputRef}
                  value={inputValue}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                    setInputValue(e.target.value);
                    // 用户输入时清除错误信息
                    if (errorMessage) {
                      setErrorMessage('');
                    }
                  }}
                  onPressEnter={(e: React.KeyboardEvent<HTMLInputElement>) => {
                    e.preventDefault();
                    if (!loading && inputValue.trim()) {
                      handleQuery();
                    }
                  }}
                  placeholder="请输入您的查询问题，例如：查询用户表中最近10条记录"
                  className={styles.textInput}
                  disabled={loading}
                  size="large"
                  prefix={<DatabaseOutlined style={{ color: '#999' }} />}
                />
                <Button
                  type="primary"
                  icon={<SendOutlined />}
                  onClick={handleQuery}
                  loading={loading}
                  disabled={!inputValue.trim()}
                  className={styles.sendButton}
                  size="large"
                >
                  查询
                </Button>
              </div>
            </div>

            {/* 错误提示警示条 */}
            {errorMessage && (
              <Alert
                message="查询失败"
                description={errorMessage}
                type="error"
                showIcon
                closable
                onClose={() => setErrorMessage('')}
                style={{ marginTop: 16 }}
              />
            )}

            {/* SQL查询显示 */}
            {sqlQuery && (
              <Card size="small" className={styles.sqlCard}>
                <Space>
                  <DatabaseOutlined />
                  <Text strong>执行的SQL:</Text>
                </Space>
                <div className={styles.sqlContent}>
                  <pre>{sqlQuery}</pre>
                </div>
              </Card>
            )}

            {/* 查询结果 */}
            {loading && queryResult.length === 0 ? (
              <div className={styles.loadingContainer}>
                <Spin size="large" tip="正在查询数据..." />
              </div>
            ) : queryResult.length > 0 ? (
              <Card 
                size="small" 
                className={styles.resultCard}
                title={
                  <Space>
                    <DatabaseOutlined />
                    <span>查询结果 ({total} 条)</span>
                  </Space>
                }
              >
                <Table
                  columns={columns}
                  dataSource={queryResult.map((item, index) => ({
                    ...item,
                    key: index,
                  }))}
                  pagination={{
                    current: currentPage,
                    pageSize: pageSize,
                    total: total,
                    showSizeChanger: true,
                    showQuickJumper: true,
                    showTotal: (total: number) => `共 ${total} 条`,
                    pageSizeOptions: ['10', '20', '50', '100'],
                    onChange: handleTableChange,
                    onShowSizeChange: handleTableChange,
                  }}
                  scroll={{ x: 'max-content' }}
                  size="small"
                />
              </Card>
            ) : sqlQuery ? (
              <Card size="small" className={styles.emptyCard}>
                <Text type="secondary">查询完成，但未返回数据</Text>
              </Card>
            ) : null}
          </Space>
        </Card>
      </div>
    </PageContainer>
  );
};

export default DbQuery;

