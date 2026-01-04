export interface IssueListItem {
  key: number;
  databaseName?: string;
  tableName: string;
  columnName: string;
  issueType: string;
  issueLevel: string;
  issueDesc: string;
  issueCount: number;
  status: number;
  handler?: string;
  handleRemark?: string;
  lastCheckTime: string;
}

