import { request } from 'umi';

/** 聊天查询请求 */
export interface ChatQueryRequest {
  session_id: string;
  question: string;
  datasource_id?: number;
  database_name?: string;
  table_name?: string;
  reset_context?: boolean; // 是否重置上下文（新问题时设置为true，不使用之前的多轮对话缓存）
}

/** 聊天查询响应 */
export interface ChatQueryResponse {
  answer: string;
  sql_query?: string;
  query_result?: Array<Record<string, any>>;
  timestamp: number;
  options?: string[]; // 多轮对话的选择选项（当问题类型为select时）
}

/** 会话信息 */
export interface ChatSession {
  id: number;
  session_id: string;
  user_name: string;
  title: string;
  gmt_created: string;
  gmt_updated: string;
}

/** 消息信息 */
export interface ChatMessage {
  id: number;
  session_id: string;
  role: 'user' | 'assistant';
  content: string;
  sql_query: string;
  query_result: Array<Record<string, any>>;
  gmt_created: string;
}

/** 问题流程项 */
export interface QuestionFlowItem {
  key: string; // 参数键名
  question: string; // 提示问题
  type?: 'text' | 'select' | 'number' | 'email'; // 输入类型
  options?: string[]; // 选项列表（当type为select时使用，静态选项）
  options_sql?: string; // 获取选项的SQL（当type为select时，如果设置了此字段，会执行SQL获取选项列表）
  required?: boolean; // 是否必填
  validation?: string; // 验证规则
  description?: string; // 参数描述
}

/** 语义规则 */
export interface SemanticSqlRule {
  id: number;
  rule_name: string;
  semantic_pattern: string;
  sql_template: string;
  query_type: string;
  description: string;
  enabled: number;
  priority: number;
  use_local_db?: number; // 0: 使用远程数据源, 1: 使用本地MySQL
  multi_round_enabled?: number; // 0: 单轮对话, 1: 多轮对话
  question_flow?: QuestionFlowItem[]; // 问题流程配置
  parameter_mapping?: Record<string, string>; // 参数映射配置
  gmt_created: string;
  gmt_updated: string;
}

/** 发送查询请求 */
export async function chatQuery(params: ChatQueryRequest) {
  return request<{
    success: boolean;
    data: ChatQueryResponse;
  }>('/api/v1/ai/chat/query', {
    method: 'POST',
    data: params,
  });
}

/** 获取会话列表 */
export async function getSessions() {
  return request<{
    success: boolean;
    data: ChatSession[];
  }>('/api/v1/ai/chat/sessions', {
    method: 'GET',
  });
}

/** 创建新会话 */
export async function createSession() {
  return request<{
    success: boolean;
    data: ChatSession;
  }>('/api/v1/ai/chat/sessions', {
    method: 'POST',
  });
}

/** 删除会话 */
export async function deleteSession(sessionId: string) {
  return request<{
    success: boolean;
    message: string;
  }>(`/api/v1/ai/chat/sessions/${sessionId}`, {
    method: 'DELETE',
  });
}

/** 获取会话消息历史 */
export async function getSessionMessages(sessionId: string) {
  return request<{
    success: boolean;
    data: ChatMessage[];
  }>(`/api/v1/ai/chat/sessions/${sessionId}/messages`, {
    method: 'GET',
  });
}

/** 更新会话标题 */
export async function updateSessionTitle(sessionId: string, title: string) {
  return request<{
    success: boolean;
    message: string;
  }>(`/api/v1/ai/chat/sessions/${sessionId}/title`, {
    method: 'PUT',
    data: { title },
  });
}

/** 获取语义规则列表 */
export async function getRules() {
  return request<{
    success: boolean;
    data: SemanticSqlRule[];
  }>('/api/v1/ai/chat/rules', {
    method: 'GET',
  });
}

/** 创建规则 */
export async function createRule(params: Partial<SemanticSqlRule>) {
  return request<{
    success: boolean;
    data: SemanticSqlRule;
  }>('/api/v1/ai/chat/rules', {
    method: 'POST',
    data: params,
  });
}

/** 更新规则 */
export async function updateRule(id: number, params: Partial<SemanticSqlRule>) {
  return request<{
    success: boolean;
    message: string;
  }>(`/api/v1/ai/chat/rules/${id}`, {
    method: 'PUT',
    data: params,
  });
}

/** 删除规则 */
export async function deleteRule(id: number) {
  return request<{
    success: boolean;
    message: string;
  }>(`/api/v1/ai/chat/rules/${id}`, {
    method: 'DELETE',
  });
}

