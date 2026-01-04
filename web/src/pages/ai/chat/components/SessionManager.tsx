import React, { useState, useEffect } from 'react';
import { List, Button, Input, Popconfirm, message, Space, Typography } from 'antd';
import { PlusOutlined, DeleteOutlined, EditOutlined, CheckOutlined, CloseOutlined } from '@ant-design/icons';
import { getSessions, createSession, deleteSession, updateSessionTitle, ChatSession } from '../services/chatQuery';
import styles from './SessionManager.less';

const { Text } = Typography;

interface SessionManagerProps {
  currentSessionId: string;
  onSessionChange: (sessionId: string) => void;
  onSessionCreated?: (session: ChatSession) => void;
}

const SessionManager: React.FC<SessionManagerProps> = ({
  currentSessionId,
  onSessionChange,
  onSessionCreated,
}) => {
  const [sessions, setSessions] = useState<ChatSession[]>([]);
  const [loading, setLoading] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editingTitle, setEditingTitle] = useState<string>('');

  // 加载会话列表
  const loadSessions = async () => {
    setLoading(true);
    try {
      const response = await getSessions();
      if (response.success) {
        setSessions(response.data || []);
        // 如果有会话且当前没有选中，选择第一个
        if (response.data && response.data.length > 0 && !currentSessionId) {
          onSessionChange(response.data[0].session_id);
        }
      }
    } catch (error) {
      console.error('加载会话列表失败:', error);
      message.error('加载会话列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadSessions();
  }, []);

  // 创建新会话
  const handleCreateSession = async () => {
    try {
      const response = await createSession();
      if (response.success && response.data) {
        setSessions(prev => [response.data, ...prev]);
        onSessionChange(response.data.session_id);
        if (onSessionCreated) {
          onSessionCreated(response.data);
        }
        message.success('创建会话成功');
      }
    } catch (error) {
      console.error('创建会话失败:', error);
      message.error('创建会话失败');
    }
  };

  // 删除会话
  const handleDeleteSession = async (sessionId: string) => {
    try {
      const response = await deleteSession(sessionId);
      if (response.success) {
        setSessions(prev => prev.filter(s => s.session_id !== sessionId));
        // 如果删除的是当前会话，切换到第一个会话
        if (sessionId === currentSessionId) {
          const remainingSessions = sessions.filter(s => s.session_id !== sessionId);
          if (remainingSessions.length > 0) {
            onSessionChange(remainingSessions[0].session_id);
          } else {
            // 如果没有剩余会话，创建新会话
            handleCreateSession();
          }
        }
        message.success('删除会话成功');
      }
    } catch (error) {
      console.error('删除会话失败:', error);
      message.error('删除会话失败');
    }
  };

  // 开始编辑标题
  const handleStartEdit = (session: ChatSession) => {
    setEditingId(session.session_id);
    setEditingTitle(session.title);
  };

  // 保存标题
  const handleSaveTitle = async (sessionId: string) => {
    if (!editingTitle.trim()) {
      message.warning('标题不能为空');
      return;
    }
    try {
      const response = await updateSessionTitle(sessionId, editingTitle.trim());
      if (response.success) {
        setSessions(prev =>
          prev.map(s =>
            s.session_id === sessionId ? { ...s, title: editingTitle.trim() } : s
          )
        );
        setEditingId(null);
        setEditingTitle('');
        message.success('更新标题成功');
      }
    } catch (error) {
      console.error('更新标题失败:', error);
      message.error('更新标题失败');
    }
  };

  // 取消编辑
  const handleCancelEdit = () => {
    setEditingId(null);
    setEditingTitle('');
  };

  return (
    <div className={styles.sessionManager}>
      <div className={styles.header}>
        <Text strong>会话列表</Text>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          size="small"
          onClick={handleCreateSession}
        >
          新建会话
        </Button>
      </div>
      <List
        loading={loading}
        dataSource={sessions}
        renderItem={(session) => (
          <List.Item
            className={`${styles.sessionItem} ${
              session.session_id === currentSessionId ? styles.active : ''
            }`}
            onClick={() => onSessionChange(session.session_id)}
          >
            <div className={styles.sessionContent}>
              {editingId === session.session_id ? (
                <Space style={{ width: '100%' }}>
                  <Input
                    value={editingTitle}
                    onChange={(e) => setEditingTitle(e.target.value)}
                    onPressEnter={() => handleSaveTitle(session.session_id)}
                    size="small"
                    style={{ flex: 1 }}
                    autoFocus
                  />
                  <Button
                    type="link"
                    icon={<CheckOutlined />}
                    size="small"
                    onClick={() => handleSaveTitle(session.session_id)}
                  />
                  <Button
                    type="link"
                    icon={<CloseOutlined />}
                    size="small"
                    onClick={handleCancelEdit}
                  />
                </Space>
              ) : (
                <>
                  <div className={styles.sessionTitle}>{session.title}</div>
                  <Space className={styles.sessionActions}>
                    <Button
                      type="text"
                      icon={<EditOutlined />}
                      size="small"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleStartEdit(session);
                      }}
                    />
                    <Popconfirm
                      title="确定要删除此会话吗？"
                      onConfirm={(e) => {
                        e?.stopPropagation();
                        handleDeleteSession(session.session_id);
                      }}
                      onClick={(e) => e.stopPropagation()}
                    >
                      <Button
                        type="text"
                        danger
                        icon={<DeleteOutlined />}
                        size="small"
                        onClick={(e) => e.stopPropagation()}
                      />
                    </Popconfirm>
                  </Space>
                </>
              )}
            </div>
          </List.Item>
        )}
      />
    </div>
  );
};

export default SessionManager;

