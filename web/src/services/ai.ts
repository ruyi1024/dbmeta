import { request } from 'umi';

export interface ChatResponse {
  answer: string;
  timestamp: number;
}

export async function chatWithAI(params: { question: string }) {
  return request<ChatResponse>('/api/v1/ai/chat', {
    method: 'POST',
    data: params,
  });
}

export async function getChatHistory() {
  return request<ChatResponse[]>('/api/v1/ai/chat/history', {
    method: 'GET',
  });
}

export async function clearChatHistory() {
  return request('/api/v1/ai/chat/history', {
    method: 'DELETE',
  });
} 