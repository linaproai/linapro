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
