import React, { useState, useRef, useEffect } from 'react';
import { flushSync } from 'react-dom';
import { PageContainer } from '@ant-design/pro-layout';
import { message, Segmented } from 'antd';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { chatQuery, createSession, ChatQueryRequest } from './services/chatQuery';
import styles from './index.less';

const antMessage = message;

interface Message {
  type: 'user' | 'ai';
  content: string;
  think?: string; // 思考/推理内容
  timestamp: number;
  isSystemMessage?: boolean;
  options?: string[];
  showFeedback?: boolean;
  feedbackSubmitted?: boolean;
  sqlQuery?: string;
  queryResult?: Array<Record<string, any>>;
}

interface ChatSession {
  id: string;
  title: string;
  lastMessage: string;
  timestamp: number;
}


const AIChat: React.FC = () => {
  const [messages, setMessages] = useState<Message[]>([
    {
      type: 'ai',
      content: '您好！我是AIDBA，很高兴为您服务。在这里我可以帮助您解答关于数据库管理问题、执行数据查询、性能问题分析。请问有什么我可以帮助您的吗？',
      timestamp: Date.now(),
    }
  ]);
  const [inputValue, setInputValue] = useState('');
  const [loading, setLoading] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const messageListRef = useRef<HTMLDivElement>(null);
  const [shouldAutoScroll, setShouldAutoScroll] = useState(false);
  const [isInitialLoad, setIsInitialLoad] = useState(true);
  const [currentSessionId, setCurrentSessionId] = useState(`session_${Date.now()}`);
  
  // 模式切换状态
  const [mode, setMode] = useState<'chat' | 'agent'>('chat');
  
  // Agent模式相关状态
  const [agentSessionId, setAgentSessionId] = useState<string>('');
  
  // 添加状态来存储历史会话的消息
  const [sessionMessages, setSessionMessages] = useState<Record<string, Message[]>>({});
  
  // 添加反馈统计状态
  const [feedbackStats, setFeedbackStats] = useState<{
    helpRate: number;
    totalFeedback: number;
    helpfulCount: number;
    hotQuestions: Array<{ question: string; count: number }>;
  }>({
    helpRate: 0,
    totalFeedback: 0,
    helpfulCount: 0,
    hotQuestions: []
  });
  
  // 添加推荐规则状态
  const [recommendedRules, setRecommendedRules] = useState<Array<{ rule_name: string; priority: number }>>([]);
  
  // 组件挂载时加载反馈统计和推荐规则
  useEffect(() => {
    fetchFeedbackStats();
    fetchRecommendedRules();
  }, []);

  // 获取推荐规则列表
  const fetchRecommendedRules = async () => {
    try {
      const response = await fetch('/api/v1/ai/chat/rules/recommended', {
        method: 'GET',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include'
      });

      if (response.ok) {
        const data = await response.json();
        if (data.success && data.data) {
          setRecommendedRules(data.data);
        }
      }
    } catch (error) {
      console.error('获取推荐规则失败:', error);
    }
  };

  // 处理推荐规则的点击（以agent模式提交）
  const handleAgentRecommendation = async (ruleName: string) => {
    if (!ruleName.trim()) {
      return;
    }

    // 切换到agent模式
    if (mode !== 'agent') {
      setMode('agent');
    }

    const userMessage: Message = {
      type: 'user',
      content: ruleName,
          timestamp: Date.now(),
    };

    const updatedMessages = [...messages, userMessage];
    setMessages(updatedMessages);
    
    // 保存当前会话的消息到历史记录
    setSessionMessages(prev => ({
      ...prev,
      [currentSessionId]: updatedMessages
    }));
    
    setInputValue('');
    setLoading(true);
    
    // 发送消息时，如果用户在底部附近，则保持自动滚动
    if (messageListRef.current) {
      const { scrollTop, scrollHeight, clientHeight } = messageListRef.current;
      const isNearBottom = scrollHeight - scrollTop <= clientHeight + 200;
      if (isNearBottom) {
        setShouldAutoScroll(true);
      }
    }

    try {
      // 确保使用agent模式
      let sessionId = agentSessionId;
      if (!sessionId) {
        // 如果没有会话ID，先创建会话
        try {
          const result = await createSession();
          if (result.success && result.data) {
            sessionId = result.data.session_id;
            setAgentSessionId(sessionId);
          } else {
            throw new Error('无法创建Agent会话');
          }
        } catch (error) {
          throw new Error('创建Agent会话失败，请稍后重试');
        }
      }

      const queryParams: ChatQueryRequest = {
        session_id: sessionId,
        question: ruleName,
        reset_context: true, // 点击agent推荐时，重置上下文，开始新流程
      };

      const result = await chatQuery(queryParams);

      if (!result.success) {
        throw new Error(result.message || 'Agent查询失败');
      }

      // 安全获取响应数据
      const responseData = result.data;
      if (!responseData) {
        throw new Error('响应数据格式错误：缺少data字段');
      }
      
      // 格式化Agent响应
      // 清理多轮对话上下文标记（隐藏技术标记）
      let content = responseData.answer || '抱歉，我暂时无法回答这个问题。';
      content = content.replace(/<!--MULTI_ROUND_CONTEXT:.*?-->/g, '');
      
      // 注意：不要在这里添加SQL，因为前端会单独显示SQL和查询结果
      // SQL和查询结果会通过 sqlQuery 和 queryResult 字段单独显示

      const aiMessage: Message = {
        type: 'ai',
        content: content,
        timestamp: Date.now(),
        showFeedback: false, // Agent模式下不显示反馈
        sqlQuery: responseData.sql_query,
        queryResult: responseData.query_result,
        options: responseData.options, // 多轮对话的选择选项
      };

      const finalMessages = [...updatedMessages, aiMessage];
      setMessages(finalMessages);
      
      // 保存到历史记录
      setSessionMessages(prev => ({
        ...prev,
        [currentSessionId]: finalMessages
      }));

      // 只有在用户发送消息后才启用自动滚动
      setShouldAutoScroll(true);
    } catch (error) {
      console.error('Agent推荐查询失败:', error);
      const errorMessage: Message = {
          type: 'ai',
        content: `抱歉，服务暂时不可用。${error instanceof Error ? error.message : '请稍后重试，或者您可以尝试重新描述一下问题。'}`,
          timestamp: Date.now(),
        showFeedback: false, // Agent模式下不显示反馈
      };
      const finalMessages = [...updatedMessages, errorMessage];
      setMessages(finalMessages);
      setSessionMessages(prev => ({
        ...prev,
        [currentSessionId]: finalMessages
      }));
    } finally {
      setLoading(false);
    }
  };

  // 切换到Agent模式时自动创建会话
  useEffect(() => {
    if (mode === 'agent' && !agentSessionId) {
      createAgentSession();
    }
  }, [mode]);

  // 创建Agent会话
  const createAgentSession = async () => {
    try {
      const result = await createSession();
      if (result.success && result.data) {
        setAgentSessionId(result.data.session_id);
        console.log('Agent会话创建成功:', result.data.session_id);
      } else {
        antMessage.error('创建Agent会话失败');
      }
    } catch (error) {
      console.error('创建Agent会话失败:', error);
      antMessage.error('创建Agent会话失败，请稍后重试');
    }
  };

  // 监听 sessionMessages 变化，更新历史会话列表
  useEffect(() => {
    const sessionsList = generateChatSessions(sessionMessages);
    setChatSessions(sessionsList);
    console.log('历史会话列表已更新，共', sessionsList.length, '个会话');
  }, [sessionMessages]);

  // 历史会话数据 - 从 sessionMessages 动态生成
  const [chatSessions, setChatSessions] = useState<ChatSession[]>([]);
  
  // 从 sessionMessages 生成历史会话列表
  const generateChatSessions = (sessions: Record<string, Message[]>) => {
    const sessionsList: ChatSession[] = [];
    
    Object.entries(sessions).forEach(([sessionId, messages]) => {
      if (messages && messages.length > 0) {
        // 找到最后一条用户消息作为标题和最后消息
        const lastUserMessage = messages.filter(msg => msg.type === 'user').pop();
        if (lastUserMessage) {
          const session: ChatSession = {
            id: sessionId,
            title: lastUserMessage.content.substring(0, 20) + (lastUserMessage.content.length > 20 ? '...' : ''),
            lastMessage: lastUserMessage.content,
            timestamp: lastUserMessage.timestamp,
          };
          sessionsList.push(session);
        }
      }
    });
    
    // 按时间戳倒序排列，最新的在前面
    return sessionsList.sort((a, b) => b.timestamp - a.timestamp);
  };

  // 聊天模式的快捷问题建议（一般性咨询）
  const chatQuickQuestions = [
    '如何优化SQL查询性能？',
    '数据库连接池如何配置？',
    '如何监控数据库状态？',
    '备份策略有哪些建议？',
    '索引的最佳实践是什么？',
    '如何处理数据库死锁？',
    '数据库安全审计怎么做？',
    '如何防范SQL注入攻击？',
    '数据库权限管理最佳实践？',
    '如何检测数据库异常访问？'
  ];

  // Agent模式的快捷问题建议（具体查询操作）
  const agentQuickQuestions = [
    '查询所有数据库的状态',
    '显示当前数据库连接数',
    '查看最近执行的慢查询',
    '统计各表的记录数量',
    '查询数据库性能指标',
    '查看表结构信息',
    '显示数据库空间使用情况',
    '查询最近的错误日志',
    '统计各数据库的大小',
    '查看活跃会话信息'
  ];

  // 根据当前模式选择建议问题
  const quickQuestions = mode === 'agent' ? agentQuickQuestions : chatQuickQuestions;

  const scrollToBottom = () => {
    if (shouldAutoScroll && messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  };

  // 监听消息变化，实时滚动到底部
  useEffect(() => {
    if (isInitialLoad) {
      setIsInitialLoad(false);
      // 初始加载时不滚动，保持页面顶部
      return;
    }
    // 只在有新消息且用户没有手动滚动时自动滚动
    if (shouldAutoScroll && messages.length > 0) {
      // 使用 requestAnimationFrame 确保 DOM 更新后再滚动，更及时
      requestAnimationFrame(() => {
    scrollToBottom();
      });
    }
  }, [messages, shouldAutoScroll]); // 监听消息内容和滚动标志的变化

  // 获取反馈统计数据
  const fetchFeedbackStats = async () => {
    try {
      const response = await fetch('/api/v1/ai/feedback/stats', {
        method: 'GET',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include'
      });

      if (response.ok) {
        const data = await response.json();
        setFeedbackStats(data);
      } else {
        // 如果API不存在，使用模拟数据
        const mockData = {
          helpRate: 85.6,
          totalFeedback: 1250,
          helpfulCount: 1070,
          hotQuestions: [
            { question: '如何配置防火墙规则？', count: 156 },
            { question: '服务器CPU使用率过高怎么办？', count: 142 },
            { question: '数据库连接超时怎么解决？', count: 128 },
            { question: '如何备份重要数据？', count: 115 },
            { question: '网络延迟高怎么优化？', count: 98 },
            { question: '系统日志如何查看？', count: 87 },
            { question: '如何设置监控告警？', count: 76 },
            { question: '服务器内存不足怎么办？', count: 65 },
            { question: '如何配置负载均衡？', count: 54 },
            { question: 'SSL证书过期怎么处理？', count: 43 }
          ]
        };
        setFeedbackStats(mockData);
      }
    } catch (error) {
      console.error('获取反馈统计数据失败:', error);
      // 使用模拟数据作为后备
      const mockData = {
        helpRate: 85.6,
        totalFeedback: 1250,
        helpfulCount: 1070,
        hotQuestions: [
          { question: '如何配置防火墙规则？', count: 156 },
          { question: '服务器CPU使用率过高怎么办？', count: 142 },
          { question: '数据库连接超时怎么解决？', count: 128 },
          { question: '如何备份重要数据？', count: 115 },
          { question: '网络延迟高怎么优化？', count: 98 },
          { question: '系统日志如何查看？', count: 87 },
          { question: '如何设置监控告警？', count: 76 },
          { question: '服务器内存不足怎么办？', count: 65 },
          { question: '如何配置负载均衡？', count: 54 },
          { question: 'SSL证书过期怎么处理？', count: 43 }
        ]
      };
      setFeedbackStats(mockData);
    }
  };


  // 监听消息列表滚动，判断用户是否主动滚动
  const handleScroll = () => {
    if (messageListRef.current) {
      const { scrollTop, scrollHeight, clientHeight } = messageListRef.current;
      const isAtBottom = scrollHeight - scrollTop <= clientHeight + 100; // 增加容差
      setShouldAutoScroll(isAtBottom);
    }
  };

  // 在Agent模式下，新问题应该总是重置上下文，除非是点击选择按钮（多轮对话继续）
  // 在Chat模式下，不需要重置上下文
  const handleSend = async (question?: string, resetContext?: boolean) => {
    // 如果没有明确指定 resetContext，根据模式决定：
    // - Agent模式：默认为 true（新问题总是重置上下文）
    // - Chat模式：默认为 false（保持对话上下文）
    if (resetContext === undefined) {
      resetContext = mode === 'agent';
    }
    const questionText = question || inputValue;
    if (!questionText.trim()) {
      antMessage.warning('请输入问题');
      return;
    }

    const userMessage: Message = {
      type: 'user',
      content: questionText,
      timestamp: Date.now(),
    };

    const updatedMessages = [...messages, userMessage];
    setMessages(updatedMessages);
    
    // 保存当前会话的消息到历史记录
      setSessionMessages(prev => ({
        ...prev,
        [currentSessionId]: updatedMessages
      }));
    
    setInputValue('');
    setLoading(true);
    
    // 发送消息时，如果用户在底部附近，则保持自动滚动
    if (messageListRef.current) {
      const { scrollTop, scrollHeight, clientHeight } = messageListRef.current;
      const isNearBottom = scrollHeight - scrollTop <= clientHeight + 200;
      if (isNearBottom) {
        setShouldAutoScroll(true);
      }
    }

    try {
      if (mode === 'chat') {
        // 聊天模式：调用 /api/v1/ai/chat（流式输出）
        // 创建AI消息占位符
        const aiMessageIndex = updatedMessages.length;
        const initialAiMessage: Message = {
          type: 'ai',
          content: '',
          think: '',
          timestamp: Date.now(),
          showFeedback: mode === 'chat', // 仅在Chat模式下显示反馈
        };
        const messagesWithPlaceholder = [...updatedMessages, initialAiMessage];
        setMessages(messagesWithPlaceholder);
        setSessionMessages(prev => ({
          ...prev,
          [currentSessionId]: messagesWithPlaceholder
        }));

        try {
      const response = await fetch('/api/v1/ai/chat', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({ question: questionText }),
      });

          if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
          }

          // 检查是否是流式响应
          const contentType = response.headers.get('content-type');
          console.log('响应Content-Type:', contentType);
          if (contentType && contentType.includes('text/event-stream')) {
            // 处理SSE流式响应
            const reader = response.body?.getReader();
            const decoder = new TextDecoder();
            let buffer = '';

            if (!reader) {
              throw new Error('无法读取响应流');
            }

            let hasData = false;
            const startTime = Date.now();
            console.log(`[流式开始] 开始读取流式响应，时间: ${new Date().toISOString()}`);
            
            // 使用递归函数持续读取，避免阻塞
            const readStream = async (): Promise<void> => {
              try {
                const { done, value } = await reader.read();
                
                if (done) {
                  console.log(`[流式结束] 流式响应完成，总耗时: ${Date.now() - startTime}ms`);
                  // 处理剩余的buffer数据
                  if (buffer.trim()) {
                    const events = buffer.split('\n\n');
                    for (const event of events) {
                      if (!event.trim()) continue;
                      processSSEEvent(event);
                    }
                  }
                  if (!hasData && buffer.trim() === '') {
                    // 如果没有任何数据，显示错误消息
                    setMessages(prev => {
                      const newMessages = [...prev];
                      if (newMessages[aiMessageIndex]) {
                        newMessages[aiMessageIndex] = {
                          ...newMessages[aiMessageIndex],
                          content: '抱歉，AI服务没有返回任何数据。请检查AI模型配置是否正确。',
                        };
                      }
                      return newMessages;
                    });
                  }
                  return;
                }

                if (value && value.length > 0) {
                  hasData = true;
                  const receiveTime = Date.now();
                  const elapsed = receiveTime - startTime;
                  const decoded = decoder.decode(value, { stream: true });
                  buffer += decoded;
                  
                  // 调试：实时打印接收到的数据块
                  console.log(`[流式数据] 收到 ${value.length} 字节，耗时: ${elapsed}ms，时间: ${new Date().toISOString()}`);
                  console.log(`[流式数据] 内容预览:`, decoded.substring(0, 100));
                  
                  // 立即处理所有完整的事件，不等待
                  const events = buffer.split('\n\n');
                  buffer = events.pop() || ''; // 保留最后一个不完整的事件
                  
                  // 处理每个完整的事件，立即更新UI
                  for (const event of events) {
                    if (!event.trim()) continue;
                    const processTime = Date.now();
                    console.log(`[SSE事件] 处理事件，耗时: ${processTime - startTime}ms，时间: ${new Date().toISOString()}`, event.substring(0, 200));
                    processSSEEvent(event);
                  }
                }
                
                // 继续读取下一个数据块（不等待，立即递归）
                readStream();
              } catch (error) {
                console.error('[流式错误]', error);
                throw error;
              }
            };
            
            // 开始读取流
            await readStream();
            
            // 处理SSE事件的辅助函数
            function processSSEEvent(event: string) {
              let eventType = '';
              const dataLines: string[] = [];
              
              // 解析SSE事件（支持多行 data）
              const lines = event.split('\n');
              for (const line of lines) {
                const trimmedLine = line.trim();
                // 跳过空行和注释
                if (!trimmedLine || trimmedLine.startsWith(':')) {
                  continue;
                }
                // 支持 event: 和 event:message 两种格式（冒号后可能有空格也可能没有）
                if (trimmedLine.startsWith('event:')) {
                  eventType = trimmedLine.substring(6).trim();
                } else if (trimmedLine.startsWith('data:')) {
                  // 支持多行 data（SSE 标准支持）
                  dataLines.push(trimmedLine.substring(5).trim());
                }
              }
              
              // 合并多行 data（用 \n 连接）
              const eventData = dataLines.join('\n');

              if (eventData) {
                try {
                  const parsed = JSON.parse(eventData);
                  if (eventType === 'think' && parsed.content) {
                    // 使用 flushSync 强制立即更新UI，确保实时显示
                    flushSync(() => {
                      setMessages(prev => {
                        const newMessages = [...prev];
                        if (newMessages[aiMessageIndex]) {
                          const currentThink = newMessages[aiMessageIndex].think || '';
                          newMessages[aiMessageIndex] = {
                            ...newMessages[aiMessageIndex],
                            think: currentThink + parsed.content,
                          };
                        }
                        return newMessages;
                      });
                    });
                    // sessionMessages 更新不需要立即刷新，可以异步
        setSessionMessages(prev => ({
          ...prev,
                      [currentSessionId]: prev[currentSessionId]?.map((msg, idx) =>
                        idx === aiMessageIndex
                          ? { ...msg, think: (msg.think || '') + parsed.content }
                          : msg
                      ) || []
                    }));
                    // 自动滚动到底部
                    setShouldAutoScroll(true);
                  } else if (eventType === 'message' && parsed.content) {
                    // 清理多轮对话上下文标记
                    let cleanContent = parsed.content.replace(/<!--MULTI_ROUND_CONTEXT:.*?-->/g, '');
                    // 使用 flushSync 强制立即更新UI，确保实时显示
                    flushSync(() => {
                      setMessages(prev => {
                        const newMessages = [...prev];
                        if (newMessages[aiMessageIndex]) {
                          // 清理现有内容中的标记，然后追加新内容
                          const currentContent = (newMessages[aiMessageIndex].content || '').replace(/<!--MULTI_ROUND_CONTEXT:.*?-->/g, '');
                          newMessages[aiMessageIndex] = {
                            ...newMessages[aiMessageIndex],
                            content: currentContent + cleanContent,
                          };
                        }
                        return newMessages;
                      });
                    });
                    // sessionMessages 更新不需要立即刷新，可以异步
        setSessionMessages(prev => ({
          ...prev,
                      [currentSessionId]: prev[currentSessionId]?.map((msg, idx) =>
                        idx === aiMessageIndex
                          ? { ...msg, content: msg.content + parsed.content }
                          : msg
                      ) || []
                    }));
                    // 自动滚动到底部
                    setShouldAutoScroll(true);
                  } else if (eventType === 'error' || parsed.error) {
                    throw new Error(parsed.error || parsed.message || '流式响应错误');
                  } else if (eventType === 'done') {
                    // 流式输出完成
                  }
                } catch (e) {
                  // 忽略JSON解析错误，可能是其他格式的数据
                  if (e instanceof Error && e.message !== '流式响应错误') {
                    console.warn('解析SSE数据失败:', e, eventData);
                  } else {
                    throw e;
                  }
                }
              }
            }
          } else {
            // 非流式响应（兼容旧版本）
            const data = await response.json();
            setMessages(prev => {
              const newMessages = [...prev];
              if (newMessages[aiMessageIndex]) {
                newMessages[aiMessageIndex] = {
                  ...newMessages[aiMessageIndex],
                  content: (data.answer || data.data?.answer || '抱歉，我暂时无法回答这个问题。').replace(/<!--MULTI_ROUND_CONTEXT:.*?-->/g, ''),
                };
              }
              return newMessages;
            });
      setSessionMessages(prev => ({
        ...prev,
              [currentSessionId]: prev[currentSessionId]?.map((msg, idx) =>
                idx === aiMessageIndex
                  ? { ...msg, content: data.answer || data.data?.answer || '抱歉，我暂时无法回答这个问题。' }
                  : msg
              ) || []
      }));
    }
          // 只有在用户发送消息后才启用自动滚动
          setShouldAutoScroll(true);
        } catch (error) {
          // 如果流式请求失败，更新错误消息
          setMessages(prev => {
            const newMessages = [...prev];
            if (newMessages[aiMessageIndex]) {
              newMessages[aiMessageIndex] = {
                ...newMessages[aiMessageIndex],
                content: `抱歉，服务暂时不可用。${error instanceof Error ? error.message : '请稍后重试。'}`,
              };
            }
            return newMessages;
          });
         setSessionMessages(prev => ({
           ...prev,
            [currentSessionId]: prev[currentSessionId]?.map((msg, idx) =>
              idx === aiMessageIndex
                ? { ...msg, content: `抱歉，服务暂时不可用。${error instanceof Error ? error.message : '请稍后重试。'}` }
                : msg
            ) || []
          }));
          setLoading(false);
       return;
        }
    } else {
        // Agent模式：调用 /api/v1/ai/chat/query
        // Agent模式：调用 /api/v1/ai/chat/query
        let sessionId = agentSessionId;
        if (!sessionId) {
          // 如果没有会话ID，先创建会话
          try {
            const result = await createSession();
            if (result.success && result.data) {
              sessionId = result.data.session_id;
              setAgentSessionId(sessionId);
            } else {
              throw new Error('无法创建Agent会话');
            }
          } catch (error) {
            throw new Error('创建Agent会话失败，请稍后重试');
          }
        }

        const queryParams: ChatQueryRequest = {
          session_id: sessionId,
          question: questionText,
          reset_context: resetContext, // 根据参数决定是否重置上下文
        };

        const result = await chatQuery(queryParams);

        if (!result.success) {
          throw new Error(result.message || 'Agent查询失败');
        }

        // 安全获取响应数据
        const responseData = result.data;
        if (!responseData) {
          throw new Error('响应数据格式错误：缺少data字段');
        }
        
        // 格式化Agent响应
        // 清理多轮对话上下文标记（隐藏技术标记）
        let content = responseData.answer || '抱歉，我暂时无法回答这个问题。';
        content = content.replace(/<!--MULTI_ROUND_CONTEXT:.*?-->/g, '');
        
        // 如果有SQL查询，添加到内容中
        // if (responseData.sql_query) {
        //   content = `**生成的SQL:**\n\`\`\`sql\n${responseData.sql_query}\n\`\`\`\n\n${content}`;
        // }

             const aiMessage: Message = {
         type: 'ai',
          content: content,
         timestamp: Date.now(),
          showFeedback: false, // Agent模式下不显示反馈
          sqlQuery: responseData.sql_query,
          queryResult: responseData.query_result,
          options: responseData.options, // 多轮对话的选择选项
       };

        const finalMessages = [...updatedMessages, aiMessage];
        setMessages(finalMessages);
      
        // 保存包含AI回复的完整消息到历史记录
        setSessionMessages(prev => ({
          ...prev,
          [currentSessionId]: finalMessages
        }));
      }
    } catch (error) {
      console.error('AI请求失败:', error);
      const aiMessage: Message = {
         type: 'ai',
        content: `抱歉，服务暂时不可用。${error instanceof Error ? error.message : '请稍后重试，或者您可以尝试重新描述一下问题。'}`,
         timestamp: Date.now(),
        showFeedback: mode === 'chat', // 仅在Chat模式下显示反馈
       };
      const finalMessages = [...updatedMessages, aiMessage];
      setMessages(finalMessages);
      
      // 保存包含错误消息的完整消息到历史记录
        setSessionMessages(prev => ({
          ...prev,
        [currentSessionId]: finalMessages
      }));
    } finally {
      setLoading(false);
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleQuickQuestion = (question: string) => {
    handleSend(question);
  };

  // 处理选项点击
  const handleOptionClick = (option: string) => {
    setInputValue(option);
    handleSend(option, false); // 点击选择按钮时，不重置上下文，继续多轮对话
  };

  // 已移除所有引导流程相关函数

  const handleClearChat = () => {
      setMessages([
        {
          type: 'ai',
        content: '您好！我是AIDBA，很高兴为您服务。在这里我可以帮助您解答关于数据库管理、SQL查询、性能优化、安全审计等相关问题。请问有什么我可以帮助您的吗？',
          timestamp: Date.now(),
        }
      ]);
      setCurrentSessionId(`session_${Date.now()}`);
    setIsInitialLoad(true); // 重置初始加载状态，不自动滚动
    setShouldAutoScroll(false);
    
    // 保存当前会话到历史记录（如果当前会话有内容）
    if (messages.length > 1) {
      setSessionMessages(prev => ({
        ...prev,
        [currentSessionId]: messages
      }));
    }
  };

  const handleNewChat = () => {
    // 保存当前会话到历史记录
    if (messages.length > 1) {
      setSessionMessages(prev => ({
        ...prev,
        [currentSessionId]: messages
      }));
    }
    
    // 创建新的会话ID
    const newSessionId = `session_${Date.now()}`;
    setCurrentSessionId(newSessionId);
    
    // 显示默认欢迎消息
      setMessages([
        {
          type: 'ai',
        content: '您好！我是AIDBA，很高兴为您服务。在这里我可以帮助您解答关于数据库管理问题、执行数据查询、性能问题分析。请问有什么我可以帮助您的吗？',
          timestamp: Date.now(),
        }
      ]);
    
    setIsInitialLoad(true); // 重置初始加载状态，不自动滚动
    setShouldAutoScroll(false);
  };

  const handleSessionClick = (session: ChatSession) => {
    setCurrentSessionId(session.id);
    
    // 加载对应会话的历史消息
    const sessionHistory = sessionMessages[session.id];
    if (sessionHistory && sessionHistory.length > 0) {
      // 如果有历史消息，加载它们
      setMessages(sessionHistory);
      console.log(`加载会话 ${session.id} 的历史消息，共 ${sessionHistory.length} 条`);
    } else {
      // 如果没有历史消息，显示默认欢迎消息
      const defaultMessage: Message = {
        type: 'ai',
        content: '您好！我是AIDBA，很高兴为您服务。在这里我可以帮助您解答关于数据库管理问题、执行数据查询、性能问题分析。请问有什么我可以帮助您的吗？',
        timestamp: Date.now(),
      };
      setMessages([defaultMessage]);
      console.log(`会话 ${session.id} 没有历史消息，显示默认消息`);
    }
    
    setIsInitialLoad(true); // 切换会话时重置初始加载状态，不自动滚动
    setShouldAutoScroll(false);
    antMessage.info(`切换到会话: ${session.title}`);
  };

  const scrollToBottomManually = () => {
    setShouldAutoScroll(true);
    if (messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  };

  const formatTime = (timestamp: number) => {
    const date = new Date(timestamp);
    const now = new Date();
    const isToday = date.toDateString() === now.toDateString();
    const isYesterday = new Date(now.getTime() - 24 * 60 * 60 * 1000).toDateString() === date.toDateString();
    
    if (isToday) {
      return `今天 ${date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })}`;
    } else if (isYesterday) {
      return `昨天 ${date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })}`;
    } else {
      const year = date.getFullYear();
      const month = String(date.getMonth() + 1).padStart(2, '0');
      const day = String(date.getDate()).padStart(2, '0');
      const time = date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
      return `${year}-${month}-${day} ${time}`;
    }
  };

  // 处理反馈提交
  const handleFeedback = async (messageIndex: number, isHelpful: boolean) => {
    const currentMessage = messages[messageIndex];
    if (!currentMessage || currentMessage.feedbackSubmitted) {
      return;
    }

    try {
      // 获取用户信息（未登录时使用guest）
      const user = 'guest'; // 这里可以根据实际登录状态获取用户信息
      
      // 获取对应的用户问题（向前查找最近的用户消息）
      let userQuestion = '';
      for (let i = messageIndex - 1; i >= 0; i--) {
        const message = messages[i];
        if (message && message.type === 'user') {
          userQuestion = message.content;
          break;
        }
      }
      
      // 如果找不到用户问题，使用默认值
      if (!userQuestion.trim()) {
        userQuestion = '用户咨询问题';
      }

      // 调试信息
      console.log('提交反馈数据:', {
        user: user,
        question: userQuestion,
        answer: currentMessage.content,
        isHelpful: isHelpful ? 1 : 0,
        timestamp: Math.floor(Date.now() / 1000) // 转换为秒级时间戳
      });

      // 提交反馈到后端
      const response = await fetch('/api/v1/ai/feedback', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({
          user: user,
          question: userQuestion,
          answer: currentMessage.content,
          isHelpful: isHelpful ? 1 : 0,
          timestamp: Math.floor(Date.now() / 1000) // 转换为秒级时间戳
        }),
      });

      if (response.ok) {
        // 更新消息状态，标记反馈已提交
        const updatedMessages = [...messages];
        updatedMessages[messageIndex] = {
          ...updatedMessages[messageIndex],
          feedbackSubmitted: true
        };
        setMessages(updatedMessages);
        
        // 保存到会话历史
        if (currentSessionId !== 'default') {
          setSessionMessages(prev => ({
            ...prev,
            [currentSessionId]: updatedMessages
          }));
        }
        
        antMessage.success('感谢您的反馈！');
      } else {
        antMessage.error('反馈提交失败，请稍后重试');
      }
    } catch (error) {
      console.error('提交反馈失败:', error);
      antMessage.error('反馈提交失败，请稍后重试');
    }
  };

  return (
    <PageContainer>
      <div className={styles.aiChatLayout}>
        {/* 左侧侧边栏 */}
        <div className={styles.sidebar}>

          {/* 新建会话按钮 */}
          <div className={styles.newChatSection}>
            <button className={`${styles.newChatButton} ${loading ? styles.disabled : ''}`} onClick={handleNewChat} disabled={loading}>
              <span className={styles.newChatIcon}>💬</span>
              <span>新建会话</span>
              <span className={styles.shortcut}>Ctrl K</span>
            </button>
          </div>


          {/* 历史会话 */}
          <div className={styles.historySection}>
            <div className={styles.historyHeader}>
              <span className={styles.historyIcon}>🕒</span>
              <span>历史会话</span>
            </div>
                         <div className={styles.sessionList}>
               {chatSessions.length > 0 ? (
                 <>
                   {chatSessions.map((session) => (
                     <div 
                       key={session.id} 
                       className={`${styles.sessionItem} ${currentSessionId === session.id ? styles.activeSession : ''} ${loading ? styles.disabled : ''}`}
                       onClick={() => {
                         if (!loading) {
                           handleSessionClick(session);
                         }
                       }}
                     >
                       <div className={styles.sessionTitle}>{session.title}</div>
                     </div>
                   ))}
                   <div className={styles.viewAllSessions}>查看全部</div>
                 </>
               ) : (
                 <div className={styles.noSessions}>
                   <div style={{ 
                     textAlign: 'center', 
                     color: '#64748b', 
                     fontSize: '12px', 
                     padding: '20px 10px',
                     fontStyle: 'italic'
                   }}>
                     暂无历史会话
                   </div>
                 </div>
               )}
             </div>
          </div>
        </div>

        {/* 中间聊天区域 */}
        <div className={styles.chatContainer}>
          {/* 聊天头部 */}
          <div className={styles.chatHeader}>
            <div className={styles.headerInfo}>
              <div 
                className={styles.aiAvatar}
                style={{
                  width: 40,
                  height: 40,
                  borderRadius: '50%',
                  background: 'linear-gradient(135deg, #1a365d, #2c5282)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  color: '#fff',
                  fontSize: '18px',
                  border: '3px solid #fff',
                  boxShadow: '0 4px 12px rgba(26, 54, 93, 0.15)'
                }}
              >
                🤖
              </div>
              <div className={styles.headerText}>
                <h5 style={{ margin: 0, color: '#1a365d', fontSize: '16px', fontWeight: 600 }}>
                  AIDBA
                </h5>
                <div style={{ fontSize: '12px', color: '#64748b' }}>
                  <span className={styles.onlineIndicator}></span>
                  在线 · 通常几秒内回复
                </div>
              </div>
            </div>
            <div className={styles.headerActions}>
              <button 
                className={`${styles.clearButton} ${loading ? styles.disabled : ''}`}
                onClick={handleClearChat}
                disabled={loading}
                style={{
                  padding: '6px 12px',
                  border: '1px solid #e2e8f0',
                  borderRadius: '8px',
                  background: 'transparent',
                  color: '#64748b',
                  cursor: loading ? 'not-allowed' : 'pointer',
                  fontSize: '12px'
                }}
              >
                🗑️ 清空对话
              </button>
            </div>
          </div>

          <div style={{ height: '1px', background: 'rgba(26, 54, 93, 0.08)' }}></div>

          {/* 消息列表 */}
          <div 
            className={styles.messageList}
            ref={messageListRef}
            onScroll={handleScroll}
          >
            {messages.map((message, index) => (
              <div key={index} className={`${styles.messageWrapper} ${styles[message.type]}`}>
                <div className={styles.messageItem}>
                  {message.type === 'ai' && (
                    <div 
                      className={styles.aiAvatarSmall}
                      style={{
                        width: 28,
                        height: 28,
                        borderRadius: '50%',
                        background: message.isSystemMessage 
                          ? 'linear-gradient(135deg, #059669, #10b981)'
                            : 'linear-gradient(135deg, #1a365d, #2c5282)',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        color: '#fff',
                        fontSize: '12px',
                        marginRight: 12,
                        border: '2px solid #fff',
                        boxShadow: '0 2px 8px rgba(26, 54, 93, 0.15)'
                      }}
                    >
                      {message.isSystemMessage ? '📊' : '🤖'}
                    </div>
                  )}
                  {message.type === 'user' ? (
                    <>
                      <div className={styles.userMessageContent}>
                        <div className={styles.userMessageText}>{message.content}</div>
                        <div className={styles.userMessageTime}>
                          {formatTime(message.timestamp)}
                        </div>
                      </div>
                      <div 
                        className={styles.userAvatarSmall}
                        style={{
                          width: 32,
                          height: 32,
                          borderRadius: '50%',
                          background: 'linear-gradient(135deg, #667eea, #764ba2)',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          color: '#fff',
                          fontSize: '14px',
                          marginLeft: 12,
                          border: '2px solid #fff',
                          boxShadow: '0 2px 8px rgba(102, 126, 234, 0.2)',
                          fontWeight: 600
                        }}
                      >
                        👤
                      </div>
                    </>
                  ) : (
                  <div className={styles.messageBubble}>
                    <div className={styles.messageContent}>
                        {/* 显示思考内容（仅在Chat模式下显示） */}
                        {message.think && mode === 'chat' && (
                          <div className={styles.thinkContent}>
                            <div className={styles.thinkLabel}>💭 思考过程：</div>
                            <div className={styles.thinkText}>{message.think}</div>
                          </div>
                        )}
                        {/* 使用ReactMarkdown渲染消息内容 */}
                        <ReactMarkdown
                        remarkPlugins={[remarkGfm]}
                        components={{
                          code({ node, inline, className, children, ...props }: any) {
                            const match = /language-(\w+)/.exec(className || '');
                            const language = match ? match[1] : '';
                            return !inline && language ? (
                              <pre
                                style={{
                                  background: '#f5f5f5',
                                  padding: '12px',
                                  borderRadius: '6px',
                                  overflow: 'auto',
                                  fontSize: '13px',
                                  lineHeight: '1.5',
                                  margin: '8px 0',
                                  border: '1px solid #e0e0e0'
                                }}
                              >
                                <code
                                  className={className}
                                  style={{
                                    fontFamily: 'Monaco, Consolas, "Courier New", monospace',
                                    color: '#333'
                                  }}
                                  {...props}
                                >
                                  {String(children).replace(/\n$/, '')}
                                </code>
                              </pre>
                            ) : (
                              <code
                                className={className}
                                style={{
                                  background: '#f5f5f5',
                                  padding: '2px 6px',
                                  borderRadius: '3px',
                                  fontSize: '0.9em',
                                  fontFamily: 'Monaco, Consolas, "Courier New", monospace',
                                  color: '#d73a49'
                                }}
                                {...props}
                              >
                                {children}
                              </code>
                            );
                          },
                          table({ node, ...props }: any) {
                            return (
                              <div style={{ overflowX: 'auto', margin: '12px 0' }}>
                                <table
                                  style={{
                                    width: '100%',
                                    borderCollapse: 'collapse',
                                    border: '1px solid #e0e0e0',
                                    borderRadius: '4px',
                                    overflow: 'hidden'
                                  }}
                                  {...props}
                                />
                              </div>
                            );
                          },
                          th({ node, ...props }: any) {
                            return (
                              <th
                                style={{
                                  padding: '10px 12px',
                                  background: '#fafafa',
                                  borderBottom: '2px solid #e0e0e0',
                                  textAlign: 'left',
                                  fontWeight: 600,
                                  color: '#333'
                                }}
                                {...props}
                              />
                            );
                          },
                          td({ node, ...props }: any) {
                            return (
                              <td
                                style={{
                                  padding: '10px 12px',
                                  borderBottom: '1px solid #f0f0f0',
                                  color: '#666'
                                }}
                                {...props}
                              />
                            );
                          },
                          blockquote({ node, ...props }: any) {
                            return (
                              <blockquote
                                style={{
                                  margin: '12px 0',
                                  padding: '8px 16px',
                                  borderLeft: '4px solid #1a365d',
                                  background: '#f8f9fa',
                                  color: '#666',
                                  fontStyle: 'italic'
                                }}
                                {...props}
                              />
                            );
                          },
                          h1({ node, ...props }: any) {
                            return (
                              <h1
                                style={{
                                  fontSize: '24px',
                                  fontWeight: 600,
                                  margin: '16px 0 12px 0',
                                  color: '#1a365d',
                                  borderBottom: '2px solid #e0e0e0',
                                  paddingBottom: '8px'
                                }}
                                {...props}
                              />
                            );
                          },
                          h2({ node, ...props }: any) {
                            return (
                              <h2
                                style={{
                                  fontSize: '20px',
                                  fontWeight: 600,
                                  margin: '14px 0 10px 0',
                                  color: '#1a365d'
                                }}
                                {...props}
                              />
                            );
                          },
                          h3({ node, ...props }: any) {
                            return (
                              <h3
                                style={{
                                  fontSize: '18px',
                                  fontWeight: 600,
                                  margin: '12px 0 8px 0',
                                  color: '#2c5282'
                                }}
                                {...props}
                              />
                            );
                          },
                          ul({ node, ...props }: any) {
                            return (
                              <ul
                                style={{
                                  margin: '8px 0',
                                  paddingLeft: '24px',
                                  lineHeight: '1.8'
                                }}
                                {...props}
                              />
                            );
                          },
                          ol({ node, ...props }: any) {
                            return (
                              <ol
                                style={{
                                  margin: '8px 0',
                                  paddingLeft: '24px',
                                  lineHeight: '1.8'
                                }}
                                {...props}
                              />
                            );
                          },
                          li({ node, ...props }: any) {
                            return (
                              <li
                                style={{
                                  margin: '4px 0',
                                  color: '#4a5568'
                                }}
                                {...props}
                              />
                            );
                          },
                          p({ node, ...props }: any) {
                            return (
                              <p
                                style={{
                                  margin: '8px 0',
                                  lineHeight: '1.8',
                                  color: '#2d3748'
                                }}
                                {...props}
                              />
                            );
                          },
                          strong({ node, ...props }: any) {
                            return (
                              <strong
                                style={{
                                  fontWeight: 600,
                                  color: '#1a365d'
                                }}
                                {...props}
                              />
                            );
                          },
                          hr({ node, ...props }: any) {
                            return (
                              <hr
                                style={{
                                  margin: '16px 0',
                                  border: 'none',
                                  borderTop: '1px solid #e0e0e0'
                                }}
                                {...props}
                              />
                            );
                          }
                        }}
                      >
                      {message.content}
                      </ReactMarkdown>
                      
                      {/* 显示SQL查询（如果有） */}
                      {message.sqlQuery && (
                        <div style={{ 
                          marginTop: '12px', 
                          padding: '12px', 
                          background: '#f5f5f5', 
                          borderRadius: '8px',
                          border: '1px solid #e0e0e0'
                        }}>
                          <div style={{ 
                            fontSize: '12px', 
                            color: '#666', 
                            marginBottom: '8px',
                            fontWeight: 600
                          }}>
                            📊 执行的SQL:
                          </div>
                          <pre style={{ 
                            margin: 0, 
                            padding: '8px', 
                            background: '#fff', 
                            borderRadius: '4px',
                            overflow: 'auto',
                            fontSize: '13px',
                            fontFamily: 'monospace',
                            color: '#333'
                          }}>
                            <code>{message.sqlQuery}</code>
                          </pre>
                        </div>
                      )}
                      
                      {/* 显示查询结果（如果有） */}
                      {message.queryResult && message.queryResult.length > 0 && (
                        <div style={{ 
                          marginTop: '12px',
                          padding: '12px',
                          background: '#f5f5f5',
                          borderRadius: '8px',
                          border: '1px solid #e0e0e0',
                          maxHeight: '400px',
                          overflow: 'auto'
                        }}>
                          <div style={{ 
                            fontSize: '12px', 
                            color: '#666', 
                            marginBottom: '8px',
                            fontWeight: 600
                          }}>
                            📋 查询结果 ({message.queryResult.length} 条):
                          </div>
                          <table style={{ 
                            width: '100%', 
                            borderCollapse: 'collapse',
                            background: '#fff',
                            borderRadius: '4px',
                            fontSize: '12px'
                          }}>
                            <thead>
                              <tr style={{ background: '#fafafa', borderBottom: '2px solid #e0e0e0' }}>
                                {message.queryResult[0] && Object.keys(message.queryResult[0]).map((key) => (
                                  <th key={key} style={{ 
                                    padding: '8px', 
                                    textAlign: 'left',
                                    fontWeight: 600,
                                    color: '#333',
                                    borderRight: '1px solid #e0e0e0'
                                  }}>
                                    {key}
                                  </th>
                                ))}
                              </tr>
                            </thead>
                            <tbody>
                              {message.queryResult.map((row, rowIndex) => (
                                <tr key={rowIndex} style={{ 
                                  borderBottom: '1px solid #f0f0f0'
                                }}>
                                  {message.queryResult && message.queryResult[0] && Object.keys(message.queryResult[0]).map((key) => (
                                    <td key={key} style={{ 
                                      padding: '8px',
                                      borderRight: '1px solid #f0f0f0',
                                      color: '#666'
                                    }}>
                                      {row[key] !== null && row[key] !== undefined 
                                        ? String(row[key]) 
                                        : 'NULL'}
                                    </td>
                                  ))}
                                </tr>
                              ))}
                            </tbody>
                          </table>
                        </div>
                      )}
                      
                      {message.options && (
                        <div className={styles.optionButtons}>
                          {message.options.map((option, optionIndex) => (
                            <button
                              key={optionIndex}
                              className={`${styles.optionButton} ${loading ? styles.disabled : ''}`}
                              disabled={loading}
                              onClick={() => {
                                if (!loading) {
                                   handleOptionClick(option);
                                 }
                              }}
                            >
                              {option}
                            </button>
                          ))}
                        </div>
                      )}
                    </div>
                                         <div className={styles.messageTime}>
                       {formatTime(message.timestamp)}
                     </div>
                     
                    {/* 反馈选项（仅在Chat模式下显示） */}
                    {message.type === 'ai' && message.showFeedback && !message.feedbackSubmitted && mode === 'chat' && (
                       <div className={styles.feedbackSection}>
                         <div className={styles.feedbackText}>该回答是否对你有帮助？</div>
                         <div className={styles.feedbackButtons}>
                           <button
                             className={`${styles.feedbackButton} ${styles.helpful}`}
                             onClick={() => handleFeedback(index, true)}
                             disabled={loading}
                           >
                             <span className={styles.feedbackIcon}>👍</span>
                             有帮助
                           </button>
                           <button
                             className={`${styles.feedbackButton} ${styles.notHelpful}`}
                             onClick={() => handleFeedback(index, false)}
                             disabled={loading}
                           >
                             <span className={styles.feedbackIcon}>👎</span>
                             无帮助
                           </button>
                         </div>
                       </div>
                     )}
                     
                    {/* 反馈已提交的提示（仅在Chat模式下显示） */}
                    {message.type === 'ai' && message.feedbackSubmitted && mode === 'chat' && (
                       <div className={styles.feedbackSubmitted}>
                         <span className={styles.feedbackSubmittedIcon}>✅</span>
                         感谢您的反馈
                       </div>
                     )}
                    </div>
                  )}
                </div>
              </div>
            ))}
            
            {loading && (
              <div className={`${styles.messageWrapper} ${styles.ai}`}>
                <div className={styles.messageItem}>
                  <div 
                    className={styles.aiAvatarSmall}
                    style={{
                      width: 28,
                      height: 28,
                      borderRadius: '50%',
                      background: 'linear-gradient(135deg, #1a365d, #2c5282)',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      color: '#fff',
                      fontSize: '12px',
                      marginRight: 12,
                      border: '2px solid #fff',
                      boxShadow: '0 2px 8px rgba(26, 54, 93, 0.15)'
                    }}
                  >
                    🤖
                  </div>
                  <div className={styles.typingIndicator}>
                    <span></span>
                    <span></span>
                    <span></span>
                  </div>
                </div>
              </div>
            )}
            
            <div ref={messagesEndRef} />
          </div>

          {/* 快捷问题 */}
          {messages.length <= 1 && (
            <div className={styles.quickQuestions}>
              <div className={styles.quickQuestionsTitle}>
                😊 AI助手建议您可以问：
              </div>
              <div className={styles.questionTags}>
                {quickQuestions.map((question, index) => (
                  <div 
                    key={index}
                    className={`${styles.questionTag} ${loading ? styles.disabled : ''}`}
                    onClick={() => {
                      if (!loading) {
                        handleQuickQuestion(question);
                      }
                    }}
                  >
                    {question}
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* 输入区域 */}
          <div className={styles.inputArea}>
            {/* 模式切换按钮 */}
            <div style={{ marginBottom: '12px', display: 'flex', justifyContent: 'center' }}>
              <Segmented
                options={[
                  { label: '💬 聊天', value: 'chat' },
                  { label: '🤖 Agent', value: 'agent' }
                ]}
                value={mode}
                onChange={(value: string | number) => {
                  const newMode = value as 'chat' | 'agent';
                  setMode(newMode);
                  if (newMode === 'agent' && !agentSessionId) {
                    createAgentSession();
                  }
                }}
                style={{
                  background: '#f0f0f0',
                  borderRadius: '8px',
                  padding: '4px'
                }}
              />
            </div>


            <div className={styles.inputWrapper}>
              <textarea
                value={inputValue}
                onChange={(e) => setInputValue(e.target.value)}
                onKeyPress={handleKeyPress}
                placeholder={mode === 'agent' ? '请输入您的问题，Agent可以帮您查询数据库...' : '请输入您的问题...'}
                className={styles.textArea}
                disabled={loading}
                rows={1}
                style={{
                  minHeight: '40px',
                  maxHeight: '120px',
                  resize: 'none',
                  overflow: 'auto'
                }}
              />
              <button
                onClick={() => handleSend()}
                disabled={!inputValue.trim() || loading}
                className={styles.sendButton}
                style={{
                  width: 48,
                  height: 48,
                  borderRadius: '50%',
                  border: 'none',
                  background: loading || !inputValue.trim() 
                    ? 'linear-gradient(135deg, #e2e8f0, #cbd5e0)' 
                    : mode === 'agent'
                      ? 'linear-gradient(135deg, #7c3aed, #a855f7)'
                      : 'linear-gradient(135deg, #1a365d, #2c5282)',
                  color: '#fff',
                  cursor: loading || !inputValue.trim() ? 'not-allowed' : 'pointer',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: '16px',
                  boxShadow: '0 4px 16px rgba(26, 54, 93, 0.25)',
                  transition: 'all 0.3s'
                }}
              >
                {loading ? '⏳' : '📤'}
              </button>
            </div>
            {!shouldAutoScroll && (
              <div className={styles.scrollHint} onClick={scrollToBottomManually}>
                有新消息 ↓
              </div>
            )}
          </div>
        </div>

        {/* 右侧统计面板 */}
        <div className={styles.statsPanel}>
          {/* 帮助达成率 */}
          <div className={styles.helpRateSection}>
            <div className={styles.sectionTitle}>
              <span className={styles.sectionIcon}>📊</span>
              问题帮助达成率
            </div>
            <div className={styles.helpRateCard}>
              <div className={styles.helpRateNumber}>
                {feedbackStats.helpRate.toFixed(1)}%
              </div>
              <div className={styles.helpRateLabel}>解决率</div>
              <div className={styles.helpRateDetails}>
                <span>有帮助: {feedbackStats.helpfulCount}</span>
                <span>总计: {feedbackStats.totalFeedback}</span>
              </div>
            </div>
          </div>

          {/* 热门问题 */}
          <div className={styles.hotQuestionsSection}>
            <div className={styles.sectionTitle}>
              <span className={styles.sectionIcon}>🤖</span>
              智能助手推荐
            </div>
            <div className={styles.hotQuestionsList}>
              {recommendedRules.map((item, index) => (
                <div 
                  key={index} 
                  className={styles.hotQuestionItem}
                  onClick={() => {
                    if (!loading) {
                      handleAgentRecommendation(item.rule_name);
                    }
                  }}
                >
                  <div className={styles.questionRank}>
                    <span className={styles.rankNumber}>{index + 1}</span>
                  </div>
                  <div className={styles.questionContent}>
                    <div className={styles.questionText}>{item.rule_name}</div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </PageContainer>
  );
};

export default AIChat; 