export interface QualityData {
  // 基础统计数据
  databaseCount: number;
  tableCount: number;
  columnCount: number;
  
  // 质量指标
  databaseBusinessRate: number; // 数据库业务关联率
  tableCommentRate: number;     // 数据表注释完备率
  columnCommentRate: number;    // 数据字段注释完备率
  tableAccuracyRate: number;    // 数据表备注准确度
  columnAccuracyRate: number;   // 数据字段备注准确度
  
  // 图表数据
  databaseQualityDataList: ChartDataItem[];
  tableQualityDataList: ChartDataItem[];
  columnQualityDataList: ChartDataItem[];
  tableCommentAccuracyDataList: ChartDataItem[];  // 表注释准确度分布
  columnCommentAccuracyDataList: ChartDataItem[]; // 字段注释准确度分布
}

export interface ChartDataItem {
  type: string;
  value: number;
}

export interface QualityParams {
  pageSize?: number;
  currentPage?: number;
  filter?: { [key: string]: any[] };
  sorter?: { [key: string]: any };
} 