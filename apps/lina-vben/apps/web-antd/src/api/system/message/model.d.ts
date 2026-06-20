export interface UserMessage {
  id: number;
  userId: number;
  title: string;
  categoryCode: string;
  typeLabel: string;
  typeColor: string;
  sourceType: string;
  sourceId: string;
  isRead: number;
  readAt: number | null;
  createdAt: number | null;
}

export interface UserMessageDetail {
  id: number;
  title: string;
  categoryCode: string;
  typeLabel: string;
  typeColor: string;
  sourceType: string;
  sourceId: string;
  content: string;
  createdByName: string;
  createdAt: number | null;
}
