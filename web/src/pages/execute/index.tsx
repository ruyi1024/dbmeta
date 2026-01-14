import { BorderLeftOutlined, HeartOutlined, RightSquareOutlined, TableOutlined, UnorderedListOutlined, ThunderboltOutlined, DeleteOutlined, PlusOutlined, CloseOutlined } from '@ant-design/icons';
import { PageContainer, ProColumns, ProTable } from '@ant-design/pro-components';
import { Card, Tabs, Form, Select, Button, Input, message, Table, Alert, Space, List, Row, Col, AutoComplete, Drawer, Modal } from 'antd';

import React, { useState, useEffect, useRef } from 'react';
import moment from 'moment';
import styles from './index.less';

// 引入ACE编辑器
import AceEditor from 'react-ace'
// 引入对应的mode
import 'ace-builds/src-noconflict/mode-mysql'
// 引入对应的theme
import 'ace-builds/src-noconflict/theme-github'
// 如果要有代码提示，下面这句话必须引入!!!
import 'ace-builds/src-noconflict/ext-language_tools'
// js中实现SQL格式化
import { format } from 'sql-formatter'
import 'ace-builds/src-noconflict/mode-sql'
import { Ace } from 'ace-builds';

//导出excel
import * as Exceljs from 'exceljs';
import { Workbook } from 'exceljs';
import { saveAs } from 'file-saver';

import WaterMarkContent from '@/components/WaterMarkContent'

// favoriteColumns 将在组件内部定义，以便访问组件状态

// 可调整列宽的列头组件
const ResizableHeaderCell = (props: any) => {
  const { onResize, width, ...restProps } = props;

  if (!width) {
    return <th {...restProps} />;
  }

  return (
    <th {...restProps} style={{ position: 'relative', width, minWidth: width }}>
      {restProps.children}
      {onResize && (
        <span
          className="react-resizable-handle"
          onClick={(e) => {
            e.stopPropagation();
          }}
          onMouseDown={(e) => {
            e.preventDefault();
            const startX = e.clientX;
            const startWidth = width;
            const minWidth = 50;

            const handleMouseMove = (e: MouseEvent) => {
              const diff = e.clientX - startX;
              const newWidth = Math.max(minWidth, startWidth + diff);
              onResize(e as any, { size: { width: newWidth } });
            };

            const handleMouseUp = () => {
              document.removeEventListener('mousemove', handleMouseMove);
              document.removeEventListener('mouseup', handleMouseUp);
            };

            document.addEventListener('mousemove', handleMouseMove);
            document.addEventListener('mouseup', handleMouseUp);
          }}
          style={{
            position: 'absolute',
            right: 0,
            top: 0,
            bottom: 0,
            width: '5px',
            cursor: 'col-resize',
            userSelect: 'none',
            touchAction: 'none',
            zIndex: 1,
            backgroundColor: 'transparent',
          }}
        />
      )}
    </th>
  );
};


const Index: React.FC = () => {

  const [form] = Form.useForm();

  const [formValues, setFormValues] = useState({
    datasource: "",
    database: "",
    table: "",
    sql: "",
  });

  const [typeList, setTypeList] = useState<any[]>([{ id: 0, cluster_name: '' }]);
  const [datasourceList, setDatasourceList] = useState([]);
  const [databaseList, setDatabaseList] = useState([]);
  const [tableList, setTableList] = useState([]);
  const [favoriteList, setFavoriteList] = useState([]);
  const [openFavorite, setOpenFavorite] = useState(false);
  const [openAiGenerate, setOpenAiGenerate] = useState(false);
  const [aiQuestion, setAiQuestion] = useState('');
  const [aiGenerating, setAiGenerating] = useState(false);

  const [type, setType] = useState<string>('');
  const [datasource, setDatasource] = useState<string>('');
  const [database, setDatabase] = useState<string>('');
  const [table, setTable] = useState<string>('');

  // 标签页相关状态
  interface TabItem {
    key: string;
    label: string;
    sqlContent: string;
    selectedSql: string; // 选中的SQL文本
    loading: boolean;
    tableDataTotal: number;
    tableDataList: any;
    tableDataColumn: any;
    tableDataSuccess: boolean;
    tableDataMsg: string;
    queryTimes: number;
    columnWidths: { [key: string]: number };
  }

  const [tabs, setTabs] = useState<TabItem[]>([
    {
      key: '1',
      label: '查询 1',
      sqlContent: '',
      selectedSql: '', // 选中的SQL文本
      loading: false,
      tableDataTotal: 0,
      tableDataList: null,
      tableDataColumn: null,
      tableDataSuccess: false,
      tableDataMsg: '',
      queryTimes: 0,
      columnWidths: {},
    },
  ]);
  const [activeTabKey, setActiveTabKey] = useState<string>('1');
  const [tabCounter, setTabCounter] = useState<number>(2);

  // 获取当前活动标签页
  const getCurrentTab = () => tabs.find(tab => tab.key === activeTabKey) || tabs[0];
  
  // 兼容旧代码的getter（从当前标签页获取）
  const sqlContent = getCurrentTab()?.sqlContent || '';
  const loading = getCurrentTab()?.loading || false;
  const tableDataTotal = getCurrentTab()?.tableDataTotal || 0;
  const tableDataList = getCurrentTab()?.tableDataList;
  const tableDataColumn = getCurrentTab()?.tableDataColumn;
  const tableDataSuccess = getCurrentTab()?.tableDataSuccess || false;
  const tableDataMsg = getCurrentTab()?.tableDataMsg || '';
  const queryTimes = getCurrentTab()?.queryTimes || 0;
  const columnWidths = getCurrentTab()?.columnWidths || {};

  const [currentUserinfo, setCurrentUserinfo] = useState({ "chineseName": "", "username": "" });
  const [currentDate, setCurrentDate] = useState<string>("");

  const editorRef = React.createRef()

  useEffect(() => {
    const currentDate = moment().format("YYYYMMDD");
    setCurrentDate(currentDate)
    //获取登录用户信息
    fetch('/api/v1/currentUser')
      .then((response) => response.json())
      .then((json) => {
        setCurrentUserinfo(json.data);
      })
      .catch((error) => {
        console.log('Fetch current userinfo failed', error);
      });

    //获取数据源类型  
    fetch('/api/v1/query/datasource_type')
      .then((response) => response.json())
      .then((json) => {
        setTypeList(json.data);
        const valueDict: { [key: number]: string } = {};
        json.data.forEach((record: { id: string | number; name: string; }) => {
          valueDict[record.id] = record.name;
        });
      })
      .catch((error) => {
        console.log('Fetch type list failed', error);
      });

  }, []);


  //获取数据源
  const didQueryDatasource = (val: string) => {
    setDatabaseList([]);
    setTableList([]);
    setDatabase("");
    setTable("");
    // 清空所有标签页的SQL内容
    setTabs(prevTabs => prevTabs.map(tab => ({ ...tab, sqlContent: '' })));
    form.setFieldsValue({ "datasource": "", "database": "", "table": "", "sql": "" });
    const formValue = form.getFieldsValue();
    const type = formValue.type;
    setType(val);
    fetch('/api/v1/query/datasource?type=' + type)
      .then((response) => response.json())
      .then((json) => setDatasourceList(json.data))
      .catch((error) => {
        console.log('fetch datasource list failed', error);
      });
  };

  //获取数据库
  const didQueryDatabase = (val: string) => {
    setDatabaseList([]);
    setTableList([]);
    setDatabase("");
    setTable("");
    // 清空所有标签页的SQL内容
    setTabs(prevTabs => prevTabs.map(tab => ({ ...tab, sqlContent: '' })));
    form.setFieldsValue({ "database": "", "table": "", "sql": "" });
    setDatasource(val);
    fetch('/api/v1/query/database?datasource=' + val + '&type=' + type)
      .then((response) => response.json())
      .then((json) => setDatabaseList(json.data))
      .catch((error) => {
        console.log('fetch database list failed', error);
      });
  };

  //获取数据表
  const didQueryTable = (val: string) => {
    setDatabase(val);
    // 只清空当前标签页的SQL内容
    updateTab(activeTabKey, { sqlContent: '' });
    form.setFieldsValue({ "table": "", "sql": "" });
    fetch('/api/v1/query/table?datasource=' + datasource + '&database=' + val + '&type=' + type)
      .then((response) => response.json())
      .then((json) => (json.data == null ? [] : json.data))
      .then((data) => (setTableList(data)))
      .catch((error) => {
        console.log('fetch table list failed', error);
      });
  };


  //点击表名事件
  const onClickTable = (val: string) => {
    didSetTable(val);
  };

  //点击表名后填充SQL内容
  const didSetTable = (val: string) => {
    setTable(val);
    let sql = ''
    if (type == 'MySQL' || type == 'TiDB' || type == 'Doris' || type == "MariaDB" || type == "GreatSQL" || type == "OceanBase" || type == 'ClickHouse' || type == 'PostgreSQL') {
      sql = "select * from " + val + " limit 100"
    }
    if (type == 'Oracle') {
      sql = "select * from " + database + '.' + val + " where rownum<=100"
    }
    if (type == 'SQLServer') {
      sql = "select top 100 * from " + val
    }
    if (type == 'MongoDB') {
      sql = "select.from('" + val + "')" + ".where('_id','!=','').limit(100)"
    }
    updateTab(activeTabKey, { sqlContent: sql });
    form.setFieldsValue({
      sql: sql
    });
  };

  //自动提示
  const complete = (editor: Ace.Editor, tableDataList: any[]) => {
    const completers = tableDataList.map(item => ({
      name: item.table_name,
      value: item.table_name,
      score: 100,
      meta: '',
    }));
    console.log(completers)
    editor.completers.push({
      getCompletions(editor, session, pos, prefix, callback) {
        callback(null, completers);
      },
    });

  }

  // 更新标签页数据
  const updateTab = (key: string, updates: Partial<TabItem>) => {
    setTabs(prevTabs => 
      prevTabs.map(tab => 
        tab.key === key ? { ...tab, ...updates } : tab
      )
    );
  }

  // 新增标签页
  const addTab = () => {
    const newKey = tabCounter.toString();
    const newTab: TabItem = {
      key: newKey,
      label: `查询 ${tabCounter}`,
      sqlContent: '',
      selectedSql: '',
      loading: false,
      tableDataTotal: 0,
      tableDataList: null,
      tableDataColumn: null,
      tableDataSuccess: false,
      tableDataMsg: '',
      queryTimes: 0,
      columnWidths: {},
    };
    setTabs(prevTabs => [...prevTabs, newTab]);
    setActiveTabKey(newKey);
    setTabCounter(prev => prev + 1);
    form.setFieldsValue({ sql: '' });
  }

  // 删除标签页
  const removeTab = (targetKey: string) => {
    if (tabs.length === 1) {
      message.warning('至少需要保留一个标签页');
      return;
    }
    
    const newTabs = tabs.filter(tab => tab.key !== targetKey);
    setTabs(newTabs);
    
    // 如果删除的是当前活动标签，切换到其他标签
    if (targetKey === activeTabKey) {
      const index = tabs.findIndex(tab => tab.key === targetKey);
      const newActiveKey = index > 0 ? tabs[index - 1].key : newTabs[0].key;
      setActiveTabKey(newActiveKey);
      const activeTab = newTabs.find(tab => tab.key === newActiveKey);
      if (activeTab) {
        form.setFieldsValue({ sql: activeTab.sqlContent });
      }
    }
  }

  // 切换标签页
  const handleTabChange = (key: string) => {
    setActiveTabKey(key);
    const tab = tabs.find(t => t.key === key);
    if (tab) {
      form.setFieldsValue({ sql: tab.sqlContent });
    }
  }

  //编辑器内容改变
  const onChangeContent = (value: string) => {
    form.setFieldsValue({ "sql": value });
    updateTab(activeTabKey, { sqlContent: value });
  }

  // 选择内容改变（当用户选择SQL文本时）
  const onSelectionChange = (selection: any) => {
    if (!editorRef.current) {
      return;
    }
    const editor = (editorRef.current as any).editor;
    if (!editor) {
      return;
    }
    
    const selectedText = editor.getSelectedText();
    if (selectedText && selectedText.trim()) {
      // 有选中文本，保存选中的SQL
      updateTab(activeTabKey, { selectedSql: selectedText.trim() });
    } else {
      // 没有选中文本，清空选中的SQL
      updateTab(activeTabKey, { selectedSql: '' });
    }
  }

  // 获取要执行的SQL（如果有选中，使用选中的；否则使用全部）
  const getSqlToExecute = (): string => {
    const currentTab = getCurrentTab();
    if (currentTab?.selectedSql && currentTab.selectedSql.trim()) {
      return currentTab.selectedSql;
    }
    return currentTab?.sqlContent || '';
  }

  //智能生成SQL
  const handleAiGenerate = () => {
    if (type == "" || datasource == "" || database == "") {
      message.warning("请先选择数据源类型、数据源和数据库");
      return;
    }
    setOpenAiGenerate(true);
    setAiQuestion('');
  }

  //关闭智能生成SQL弹窗
  const closeAiGenerate = () => {
    setOpenAiGenerate(false);
    setAiQuestion('');
    setAiGenerating(false);
  }

  //执行智能生成SQL
  const doAiGenerate = async () => {
    if (!aiQuestion.trim()) {
      message.warning("请输入要生成的SQL描述");
      return;
    }

    setAiGenerating(true);
    try {
      // 从datasource中提取host和port
      const [host, port] = datasource.split(':');
      
      const params = {
        question: aiQuestion.trim(),
        datasource_type: type,
        database_name: database,
        host: host,
        port: port,
        page: 1,
        page_size: 1, // 只需要生成SQL，不需要查询结果
      };

      const headers = new Headers();
      headers.append('Content-Type', 'application/json');
      
      const response = await fetch('/api/v1/ai/dbquery', {
        method: 'POST',
        headers: headers,
        body: JSON.stringify(params),
      });

      const json = await response.json();
      
      if (json.success && json.data && json.data.sql_query) {
        const generatedSQL = json.data.sql_query;
        updateTab(activeTabKey, { sqlContent: generatedSQL });
        form.setFieldsValue({ sql: generatedSQL });
        message.success("SQL生成成功");
        closeAiGenerate();
      } else {
        message.error(json.message || "SQL生成失败");
      }
    } catch (error) {
      console.error('生成SQL失败:', error);
      message.error("生成SQL失败，请稍后重试");
    } finally {
      setAiGenerating(false);
    }
  }

  //格式化SQL
  const beautifySql = () => {
    if (type == "Redis") {
      message.warning("Redis数据源不支持该功能");
      return;
    }
    const currentTab = getCurrentTab();
    if (type == "" || database == "" || !currentTab?.sqlContent) {
      message.warning("数据源/数据库/SQL不完整，无法格式化SQL");
      return;
    }
    const formatted = format(currentTab.sqlContent);
    updateTab(activeTabKey, { sqlContent: formatted });
    form.setFieldsValue({ sql: formatted });
  }

  //收藏SQL
  const favoriteSql = () => {
    if (type == "" || datasource == "" || sqlContent == "") {
      message.warning("数据源/SQL不完整，无法收藏SQL");
      return;
    }
    const headers = new Headers();
    const params = { "datasource_type": type, "datasource": datasource, "database_name": database, "content": sqlContent };
    headers.append('Content-Type', 'application/json');
    fetch('/api/v1/favorite/list', {
      method: 'post',
      headers: headers,
      body: JSON.stringify(params),
    })
      .then((response) => response.json())
      .then((json) => {
        if (json.success == true) {
          message.success("加入收藏夹成功.")
        } else {
          message.success("加入收藏夹失败.")
        }
      })
      .catch((error) => {
        console.log('fetch data failed', error);
      });
  }

  //打开收藏夹
  const showDrawer = () => {
    if (type == "" || datasource == "") {
      message.warning("选择数据源后才能打开收藏夹");
      return;
    }
    // 构建查询参数，确保包含数据库名条件
    let queryParams = 'datasource=' + datasource + '&datasource_type=' + type;
    if (database && database !== '') {
      queryParams += '&database_name=' + database;
    }
    fetch('/api/v1/favorite/list?' + queryParams)
      .then((response) => response.json())
      .then((json) => setFavoriteList(json.data == null ? [] : json.data))
      .catch((error) => {
        console.log('fetch favorite list failed', error);
      });
    setOpenFavorite(true);

  }
  //关闭收藏夹
  const closeDrawer = () => {
    setFavoriteList([]);
    setOpenFavorite(false);
  }

  //刷新收藏夹列表
  const refreshFavoriteList = () => {
    if (type == "" || datasource == "") {
      return;
    }
    // 构建查询参数，确保包含数据库名条件
    let queryParams = 'datasource=' + datasource + '&datasource_type=' + type;
    if (database && database !== '') {
      queryParams += '&database_name=' + database;
    }
    fetch('/api/v1/favorite/list?' + queryParams)
      .then((response) => response.json())
      .then((json) => setFavoriteList(json.data == null ? [] : json.data))
      .catch((error) => {
        console.log('fetch favorite list failed', error);
      });
  }

  //删除收藏的SQL
  const deleteFavorite = (id: number) => {
    Modal.confirm({
      title: '确认删除',
      content: '确定要删除这条收藏的SQL吗？',
      onOk: () => {
        const headers = new Headers();
        headers.append('Content-Type', 'application/json');
        fetch('/api/v1/favorite/list', {
          method: 'DELETE',
          headers: headers,
          body: JSON.stringify({ id: id }),
        })
          .then((response) => response.json())
          .then((json) => {
            if (json.success) {
              message.success('删除成功');
              // 刷新收藏夹列表
              refreshFavoriteList();
            } else {
              message.error('删除失败：' + (json.msg || '未知错误'));
            }
          })
          .catch((error) => {
            console.log('delete favorite failed', error);
            message.error('删除失败');
          });
      },
    });
  }

  //收藏夹表格列定义
  const favoriteColumns: ProColumns[] = [
    {
      title: '收藏时间',
      dataIndex: 'gmt_created',
      width: 260,
    },
    {
      title: '收藏内容',
      dataIndex: 'content',
      copyable: true,
      tip: '点击复制图标可以复制完整SQL',
      ellipsis: true,
    },
    {
      title: '操作',
      valueType: 'option',
      width: 80,
      render: (_, record) => [
        <Button
          key="delete"
          type="link"
          danger
          size="small"
          icon={<DeleteOutlined />}
          onClick={() => deleteFavorite(record.id)}
        >
          删除
        </Button>,
      ],
    },
  ]


  //表单提交查询执行请求
  const asyncFetch = (values: {}) => {
    console.info(values);
    updateTab(activeTabKey, { loading: true });
    // 如果有选中的SQL，使用选中的；否则使用表单中的SQL
    const sqlToExecute = getSqlToExecute() || values["sql"] || '';
    const params = { ...values, "sql": sqlToExecute, "query_type": "execute" };
    const headers = new Headers();
    headers.append('Content-Type', 'application/json');
    fetch('/api/v1/query/doQuery', {
      method: 'post',
      headers: headers,
      body: JSON.stringify(params),
    })
      .then((response) => response.json())
      .then((json) => {
        console.info(json.data);
        const currentTab = getCurrentTab();
        // 初始化列宽（如果还没有设置）
        const currentWidths = currentTab?.columnWidths || {};
        if (json.columns && Array.isArray(json.columns)) {
          const initialWidths: { [key: string]: number } = {};
          json.columns.forEach((col: any) => {
            if (col.dataIndex && !currentWidths[col.dataIndex]) {
              initialWidths[col.dataIndex] = col.width || 150;
            }
          });
          if (Object.keys(initialWidths).length > 0) {
            updateTab(activeTabKey, {
              loading: false,
              tableDataSuccess: json.success,
              tableDataMsg: json.msg,
              tableDataList: json.data,
              tableDataColumn: json.columns,
              tableDataTotal: json.total,
              queryTimes: json.times,
              columnWidths: { ...currentWidths, ...initialWidths },
            });
          } else {
            updateTab(activeTabKey, {
              loading: false,
              tableDataSuccess: json.success,
              tableDataMsg: json.msg,
              tableDataList: json.data,
              tableDataColumn: json.columns,
              tableDataTotal: json.total,
              queryTimes: json.times,
            });
          }
        } else {
          updateTab(activeTabKey, {
            loading: false,
            tableDataSuccess: json.success,
            tableDataMsg: json.msg,
            tableDataList: json.data,
            tableDataColumn: json.columns,
            tableDataTotal: json.total,
            queryTimes: json.times,
          });
        }
      })
      .catch((error) => {
        console.log('fetch data failed', error);
      });
  };


  const onFinish = (fieldValue: []) => {
    const values = {
      datasource_type: fieldValue["type"],
      datasource: fieldValue["datasource"],
      database: fieldValue["database"],
      table: fieldValue["table"],
      sql: fieldValue["sql"],
    };
    setFormValues(values);
    asyncFetch(values);
  };

  const onFinishFailed = (errorInfo: any) => {
    console.info(errorInfo);
    message.error('执行查询未完成.');
  };

  //点击按钮提交
  const queryPost = (query_type: any) => {
    if (query_type != 'doExplain' && (table == "" || table == null)) {
      message.error('请先点击左侧表名称选择表.');
      return;
    }
    const currentTab = getCurrentTab();
    updateTab(activeTabKey, { loading: true });
    // 如果有选中的SQL，使用选中的；否则使用全部SQL
    const sqlToExecute = getSqlToExecute() || currentTab?.sqlContent || '';
    const params = { "datasource_type": type, "datasource": datasource, "database": database, "table": table, "sql": sqlToExecute, "query_type": query_type };
    const headers = new Headers();
    headers.append('Content-Type', 'application/json');
    fetch('/api/v1/query/doQuery', {
      method: 'post',
      headers: headers,
      body: JSON.stringify(params),
    })
      .then((response) => response.json())
      .then((json) => {
        const currentTab = getCurrentTab();
        const currentWidths = currentTab?.columnWidths || {};
        if (json.columns && Array.isArray(json.columns)) {
          const initialWidths: { [key: string]: number } = {};
          json.columns.forEach((col: any) => {
            if (col.dataIndex && !currentWidths[col.dataIndex]) {
              initialWidths[col.dataIndex] = col.width || 150;
            }
          });
          if (Object.keys(initialWidths).length > 0) {
            updateTab(activeTabKey, {
              loading: false,
              tableDataSuccess: json.success,
              tableDataMsg: json.msg,
              tableDataList: json.data,
              tableDataColumn: json.columns,
              tableDataTotal: json.total,
              queryTimes: json.times,
              columnWidths: { ...currentWidths, ...initialWidths },
            });
          } else {
            updateTab(activeTabKey, {
              loading: false,
              tableDataSuccess: json.success,
              tableDataMsg: json.msg,
              tableDataList: json.data,
              tableDataColumn: json.columns,
              tableDataTotal: json.total,
              queryTimes: json.times,
            });
          }
        } else {
          updateTab(activeTabKey, {
            loading: false,
            tableDataSuccess: json.success,
            tableDataMsg: json.msg,
            tableDataList: json.data,
            tableDataColumn: json.columns,
            tableDataTotal: json.total,
            queryTimes: json.times,
          });
        }
      })
      .catch((error) => {
        console.log('fetch data failed', error);
        message.error('执行查询失败');
      });
  }

  //导出excel模块
  const generateHeaders = (columns: any) => {
    return columns.map((col: { title: any; dataIndex: any; width: number; }) => {
      const obj: ITableHeaer = {
        header: col.title,
        key: col.dataIndex,
        width: col.width / 5 || 20,
      }
      return obj;
    }
    )
  }
  const saveWorkBook = (workbook: Workbook, fileName: string) => {
    workbook.xlsx.writeBuffer().then((data: BlobPart) => {
      const blob = new Blob([data], { type: '' });
      saveAs(blob, fileName);
    })
  }
  const exportExcel = () => {
    //创建工作簿
    const workbook = new Exceljs.Workbook();
    //添加sheet
    const worksheet = workbook.addWorksheet("数据结果");
    //设置sheet默认行高
    worksheet.properties.defaultRowHeight = 20;
    //设置列
    worksheet.columns = generateHeaders(tableDataColumn);
    //添加行
    let rows = worksheet.addRows(tableDataList);
    //设置字体和对齐方式
    rows?.forEach(row => {
      row.font = {
        size: 11,
        name: '宋体',
      }
      row.alignment = { vertical: 'middle', 'horizontal': 'left', wrapText: false };
    })
    //设置首行样式
    let headerRow = worksheet.getRow(1);
    headerRow.eachCell((cell, _colNum) => {
      //设置背景
      cell.fill = {
        type: 'pattern',
        pattern: 'solid',
        fgColor: { argb: '0099CC' }
      }
      //设置字体
      cell.font = {
        bold: true,
        italic: false,
        size: 11,
        name: '宋体',
        color: { argb: 'FFFFFF' }
      }
      //设置对齐
      cell.alignment = { vertical: 'middle', 'horizontal': 'center', wrapText: false };
    })

    //生成文件名
    const date = new Date();
    const year = date.getFullYear().toString();
    const month = (date.getMonth() + 1).toString();
    const day = date.getDate().toString();
    const hour = date.getHours().toString();
    const minute = date.getMinutes().toString();
    const second = date.getSeconds().toString();
    const exportFileName = type + "-" + year + month + day + hour + minute + second + '.xlsx'
    //导出文件
    saveWorkBook(workbook, exportFileName)
    //记录日志
    writeLog("exportExcel");

  }

  //前端调用记录日志方法
  const writeLog = (doType: string) => {
    const params = { "datasource_type": type, "datasource": datasource, "database": database, "sql": sqlContent, "query_type": doType };
    const headers = new Headers();
    headers.append('Content-Type', 'application/json');
    fetch('/api/v1/query/writeLog', {
      method: 'post',
      headers: headers,
      body: JSON.stringify(params),
    })
      .then((response) => response.json())
      .then((json) => {
        if (json.success == true) {
          return true;
        }
        return false;
      })
      .catch((error) => {
        return false;
      });
  }

  //阻止copy
  const handleCopy = (event: React.ClipboardEvent<HTMLDivElement>) => {
    message.warning("数据复制已被记录，请注意数据安全");
    writeLog("copyData");
    //event.preventDefault();
  };



  return (
    (<PageContainer title="数据查询平台">
      <WaterMarkContent text={currentUserinfo.chineseName + "-" + currentDate}>
        <Row style={{ marginTop: '10px' }}><Col span={24}><Card>
          <Form
            style={{ marginTop: 0 }}
            form={form}
            onFinish={onFinish}
            onFinishFailed={onFinishFailed}
            initialValues={{}}
            name={'sqlForm'}
            layout="inline"
          >

            <Form.Item
              name={'type'}
              label="数据源类型"
              rules={[{ required: true, message: '请选择数据源类型' }]}
            >
              {/* <Radio.Group defaultValue="" onChange={(val) => {didQueryDatasource(val); }} >
            {typeList && typeList.map(item => <Radio.Button value={item.name}>{item.name}</Radio.Button>)}
          </Radio.Group> */}
              <Select
                showSearch style={{ width: 240 }}
                placeholder="请选择数据源类型"
                onChange={(val) => { didQueryDatasource(val); }}
              >
                {typeList && typeList.map(item => <Option key={item.name} value={item.name}>{item.name}</Option>)}
              </Select>
            </Form.Item>

            <Form.Item
              name={'datasource'}
              label="数据源"
              rules={[{ required: true, message: '请选择数据源' }]}
            >
              <Select
                showSearch style={{ width: 320 }}
                placeholder="请选择数据源"
                value={datasource}
                onChange={(val) => {
                  didQueryDatabase(val);
                }}
              >
                {datasourceList && datasourceList.map(item => <Option key={item.host + ":" + item.port} value={item.host + ":" + item.port}>{item.name}[{item.status == 1 ? "可用" : "不可用"}] </Option>)}
              </Select>
            </Form.Item>

            {type != "Redis" &&
              <Form.Item
                name={'database'}
                label="数据库"
                rules={[{ required: true, message: '请选择数据库' }]}
              >
                <Select showSearch style={{ width: 240 }} placeholder="请选择数据库" value={database}
                  onChange={(val) => {
                    didQueryTable(val);
                  }}
                >
                  {databaseList && databaseList.map(item => <Option key={item.database_name} value={item.database_name}>{item.database_name}</Option>)}
                </Select>
              </Form.Item>
            }
          </Form>
        </Card></Col></Row>

        <Row>
          {type != "Redis" &&
            <Col span={4}>
              <Card size='small' title="数据表" extra={<a href='javascript:void(0)' onClick={event => didQueryTable(database)}>刷新</a>} style={{ width: '100%', height: '750px', overflow: 'auto' }}>
                <List
                  size="small"
                  dataSource={tableList}
                  renderItem={tableList != null && (item => <List.Item><a href='javascript:void(0)' onClick={event => onClickTable(item.table_name)}><TableOutlined /> {item.table_name}</a></List.Item>)}
                />
              </Card>
            </Col>
          }

          <Col span={20}>
            <Card>
              {database && database.length > 0 &&
                <Alert message={"当前查询引擎：" + type + "，数据库: " + database} type="info" showIcon closable />
              }
              {type == "Redis" &&
                <Space direction='vertical'>
                  <Alert message="请选择查询数据源，再输入命令，当前支持的命令有：RANDOMKEY、EXISTS、TYPE、TTL、GET、HLEN、HKEYS、HGET、HGETALL、LLEN、LINDEX、LRANGE、SCARD、SMEMBERS、SISMEMBER、ZCARD、ZCOUNT、ZRANGE" type="info" showIcon closable />
                </Space>
              }
              
              {/* SQL标签页 */}
              <Tabs
                type="editable-card"
                activeKey={activeTabKey}
                onChange={handleTabChange}
                onEdit={(targetKey, action) => {
                  if (action === 'add') {
                    addTab();
                  } else {
                    removeTab(targetKey as string);
                  }
                }}
                hideAdd={false}
                style={{ marginTop: 8 }}
              >
                {tabs.map(tab => (
                  <Tabs.TabPane
                    key={tab.key}
                    tab={tab.label}
                    closable={tabs.length > 1}
                  >
                    <Form
                      style={{ marginTop: 8 }}
                      form={form}
                      onFinish={onFinish}
                      onFinishFailed={onFinishFailed}
                      initialValues={{}}
                      name={'sqlForm'}
                      layout="horizontal"
                    >

                <Form.Item
                  name={'sql'}
                  rules={[{ required: true, message: '请输入SQL查询命令' }]}
                >
                  {/* <TextArea
                autoSize={{minRows: 4, maxRows: 8}}
                defaultValue={sqlContent}
                value={sqlContent}
              /> */}
                  <AceEditor
                    ref={editorRef}
                    placeholder="请输入执行的SQL命令（可选择部分SQL执行，未选择时执行全部）"
                    mode="mysql"
                    theme="textmate"
                    name="blah2"
                    fontSize={14}
                    showPrintMargin={true}
                    showGutter={true}
                    highlightActiveLine={true}
                    style={{ width: '100%', height: '200px', border: '1px solid #ccc' }}
                    value={sqlContent}
                    editorProps={{
                      $blockScrolling: false,
                    }}
                    onChange={(value) => onChangeContent(value)} //获取输入框的内容
                    onSelectionChange={onSelectionChange} // 监听选择变化
                    onLoad={editor => complete(editor, tableList)}

                    // 设置编辑器格式化和代码提示 
                    setOptions={{
                      useWorker: false,
                      enableBasicAutocompletion: true,
                      enableLiveAutocompletion: true,
                      // 自动提词此项必须设置为true
                      enableSnippets: true,
                      showLineNumbers: true,
                      tabSize: 1,
                    }}
                  />
                  <div style={{ marginTop: 8, marginBottom: 8 }}>
                    <Space>
                      <Button htmlType='button' type='dashed' icon={<ThunderboltOutlined />} size='small' onClick={() => handleAiGenerate()}>智能生成SQL</Button>
                      <Button htmlType='button' type='dashed' icon={<BorderLeftOutlined />} size='small' onClick={() => beautifySql()}>格式化SQL语句</Button>
                      <Button htmlType='button' type='dashed' icon={<HeartOutlined />} size='small' onClick={() => favoriteSql()}>加入收藏夹</Button>
                      <Button htmlType='button' type='dashed' icon={<UnorderedListOutlined />} size='small' onClick={() => showDrawer()}>打开收藏夹</Button>
                      {getCurrentTab()?.selectedSql && (
                        <span style={{ color: '#1890ff', fontSize: '12px' }}>
                          已选择SQL，将只执行选中的部分
                        </span>
                      )}
                    </Space>
                  </div>
                </Form.Item>

                <Form.Item wrapperCol={{ offset: 0, span: 16 }}>
                  <Space>
                    <Button type="primary" htmlType="submit" icon={<RightSquareOutlined />}>执行语句</Button>

                    {(type == "MySQL" || type == "TiDB" || type == "Doris" || type == "MariaDB" || type == "GreatSQL" || type == "OceanBase") &&
                      <>
                        <Button type="default" htmlType="button" onClick={() => queryPost("doExplain")}>
                          查看执行计划
                        </Button>
                        <Button type="default" htmlType="button" onClick={() => queryPost("showIndex")}>
                          查看表索引
                        </Button>
                        <Button type="default" htmlType="button" onClick={() => queryPost("showColumn")}>
                          查看表结构
                        </Button>
                        <Button type="default" htmlType="button" onClick={() => queryPost("showCreate")}>
                          查看建表语句
                        </Button>
                        <Button type="default" htmlType="button" onClick={() => queryPost("showTableSize")}>
                          查看表容量
                        </Button>
                      </>
                    }

                    {(type == "Oracle") &&
                      <>
                        <Button type="default" htmlType="button" onClick={() => queryPost("doExplain")}>
                          查看执行计划
                        </Button>
                        <Button type="default" htmlType="button" onClick={() => queryPost("showIndex")}>
                          查看表索引
                        </Button>
                        <Button type="default" htmlType="button" onClick={() => queryPost("showColumn")}>
                          查看表结构
                        </Button>
                        <Button type="default" htmlType="button" onClick={() => queryPost("showCreate")}>
                          查看建表语句
                        </Button>
                        <Button type="default" htmlType="button" onClick={() => queryPost("showTableSize")}>
                          查看表容量
                        </Button>
                      </>
                    }
                    {(type == "PostgreSQL") &&
                      <>
                        <Button type="default" htmlType="button" onClick={() => queryPost("doExplain")}>
                          查看执行计划
                        </Button>
                        <Button type="default" htmlType="button" onClick={() => queryPost("showIndex")}>
                          查看表索引
                        </Button>
                        <Button type="default" htmlType="button" onClick={() => queryPost("showColumn")}>
                          查看表结构
                        </Button>
                        {/* <Button type="default" htmlType="button"  onClick={()=>queryPost("showCreate")}>
                查看建表语句
              </Button> */}
                        <Button type="default" htmlType="button" onClick={() => queryPost("showTableSize")}>
                          查看表容量
                        </Button>
                      </>
                    }
                    {(type == "ClickHouse") &&
                      <>
                        <Button type="default" htmlType="button" onClick={() => queryPost("showColumn")}>
                          查看表结构
                        </Button>
                        <Button type="default" htmlType="button" onClick={() => queryPost("showCreate")}>
                          查看建表语句
                        </Button>
                        <Button type="default" htmlType="button" onClick={() => queryPost("showTableSize")}>
                          查看表容量
                        </Button>
                      </>
                    }
                    </Space>
                  </Form.Item>
                    </Form>
                  </Tabs.TabPane>
                ))}
              </Tabs>
            </Card>

            <Card>
              {tableDataSuccess == false && tableDataMsg != "" &&
                <Alert type="error" message={"执行失败：" + tableDataMsg} banner />
              }
              {tableDataSuccess == true && tableDataMsg != "" &&
                <Alert type="success" message={"执行成功，耗时：" + queryTimes + "毫秒," + tableDataMsg} banner />
              }
              {tableDataSuccess == true && tableDataTotal >= 0 &&
                <div style={{ whiteSpace: 'pre-wrap', marginTop: '10px' }} onCopy={(e) => handleCopy(e)}>
                  <div style={{ width: '100%', float: 'right', marginBottom: '10px' }}>{"查询到" + tableDataTotal + "条数据"} <Button icon={<RightSquareOutlined />} onClick={exportExcel}>查询结果导出Excel</Button></div>
                  <Table
                    bordered
                    loading={loading}
                    scroll={{ 
                      scrollToFirstRowOnChange: true, 
                      x: 'max-content',
                      y: 'calc(100vh - 250px)' // 设置表格最大高度，表头会自动固定（原300px减少45px，相当于高度增加15%）
                    }}
                    className={styles.tableStyle}
                    dataSource={tableDataList}
                    columns={tableDataColumn ? tableDataColumn.map((col: any) => {
                      const dataIndex = col.dataIndex || col.key;
                      const currentWidth = columnWidths[dataIndex] || col.width || 150;
                      return {
                        ...col,
                        width: currentWidth,
                        onHeaderCell: (column: any) => ({
                          width: currentWidth,
                          onResize: (e: MouseEvent, { size }: { size: { width: number } }) => {
                            const currentTab = getCurrentTab();
                            const newWidths = { ...(currentTab?.columnWidths || {}), [dataIndex]: size.width };
                            updateTab(activeTabKey, { columnWidths: newWidths });
                          },
                        }),
                      };
                    }) : []}
                    size={'small'}
                    sticky={{ 
                      offsetHeader: 64 // 距离顶部64px（菜单高度），表头固定在此位置
                    }}
                    components={{
                      header: {
                        cell: ResizableHeaderCell,
                      },
                    }}
                  />
                </div>
              }
            </Card>
            <Alert type="info" message="支持MySQL/MariaDB/GreatSQL/TiDB/Doris/OceanBase/ClickHouse/Oracle/PostgreSQL/SQLServer/MongoDB/Redis数据查询导出，如有需要请联系管理员申请权限。" banner closable />
          </Col>
        </Row>

        <Drawer
          title="SQL收藏夹"
          placement='right'
          width={1200}
          onClose={closeDrawer}
          open={openFavorite}
          extra={
            <Space>
              <Button onClick={closeDrawer}>关闭</Button>
            </Space>
          }
        >
          <ProTable rowKey="id" search={false} dataSource={favoriteList} columns={favoriteColumns} size="middle" />

        </Drawer>

        <Modal
          title="智能生成SQL"
          open={openAiGenerate}
          onCancel={closeAiGenerate}
          footer={[
            <Button key="cancel" onClick={closeAiGenerate}>
              取消
            </Button>,
            <Button key="generate" type="primary" loading={aiGenerating} onClick={doAiGenerate}>
              生成
            </Button>,
          ]}
          width={600}
        >
          <div style={{ marginBottom: 16 }}>
            <Alert
              message="提示"
              description="请输入您想要生成的SQL描述，例如：查询用户表中年龄大于18的所有记录"
              type="info"
              showIcon
              style={{ marginBottom: 16 }}
            />
            <Input.TextArea
              rows={6}
              placeholder="请输入SQL描述，例如：查询用户表中年龄大于18的所有记录"
              value={aiQuestion}
              onChange={(e) => setAiQuestion(e.target.value)}
              onPressEnter={(e) => {
                if (e.ctrlKey || e.metaKey) {
                  doAiGenerate();
                }
              }}
            />
            <div style={{ marginTop: 8, color: '#999', fontSize: 12 }}>
              提示：按 Ctrl+Enter 或 Cmd+Enter 快速生成
            </div>
          </div>
        </Modal>

      </WaterMarkContent>
    </PageContainer>)
  );
};

export default Index;

