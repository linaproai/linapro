export interface UserMessage {
  id: number;
  userId: number;
  title: string;
  type: number;
  sourceType: string;
  sourceId: number;
  isRead: number;
  readAt: string | null;
  createdAt: string;
}

export interface UserMessageDetail {
  id: number;
  title: string;
  type: number;
  sourceType: string;
  sourceId: number;
  content: string;
  createdByName: string;
  createdAt: string;
}
