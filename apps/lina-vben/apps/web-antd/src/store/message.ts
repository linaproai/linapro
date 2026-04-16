import type { UserMessage } from '#/api/system/message/model';

import { ref } from 'vue';

import { defineStore } from 'pinia';

import {
  messageClear,
  messageDelete,
  messageList,
  messageRead,
  messageReadAll,
  messageUnreadCount,
} from '#/api/system/message';

const DEFAULT_POLL_INTERVAL = 60_000; // 60 seconds

export const useMessageStore = defineStore('message', () => {
  const unreadCount = ref(0);
  const messages = ref<UserMessage[]>([]);
  const messagesTotal = ref(0);
  let pollTimer: ReturnType<typeof setInterval> | null = null;

  /** Fetch unread count from server, and refresh message list if count changed */
  async function fetchUnreadCount() {
    try {
      const newCount = await messageUnreadCount();
      if (newCount !== unreadCount.value) {
        unreadCount.value = newCount;
        // Refresh message list when unread count changes
        await fetchMessages();
      }
    } catch {
      // Silently ignore polling errors
    }
  }

  /** Start polling for unread count
   * @param interval - polling interval in milliseconds, defaults to 60000
   */
  function startPolling(interval: number = DEFAULT_POLL_INTERVAL) {
    stopPolling();
    fetchUnreadCount();
    pollTimer = setInterval(fetchUnreadCount, interval);
  }

  /** Stop polling */
  function stopPolling() {
    if (pollTimer) {
      clearInterval(pollTimer);
      pollTimer = null;
    }
  }

  /** Fetch message list */
  async function fetchMessages(pageNum = 1, pageSize = 20) {
    const res = await messageList({ pageNum, pageSize });
    messages.value = res.items;
    messagesTotal.value = res.total;
  }

  /** Mark single message as read */
  async function markRead(id: number) {
    await messageRead(id);
    const msg = messages.value.find((m) => m.id === id);
    if (msg) {
      msg.isRead = 1;
    }
    await fetchUnreadCount();
  }

  /** Mark all messages as read */
  async function markAllRead() {
    await messageReadAll();
    messages.value.forEach((m) => (m.isRead = 1));
    unreadCount.value = 0;
  }

  /** Remove single message */
  async function removeMessage(id: number) {
    await messageDelete(id);
    messages.value = messages.value.filter((m) => m.id !== id);
    messagesTotal.value = Math.max(0, messagesTotal.value - 1);
    await fetchUnreadCount();
  }

  /** Clear all messages */
  async function clearAll() {
    await messageClear();
    messages.value = [];
    messagesTotal.value = 0;
    unreadCount.value = 0;
  }

  function $reset() {
    stopPolling();
    unreadCount.value = 0;
    messages.value = [];
    messagesTotal.value = 0;
  }

  return {
    $reset,
    clearAll,
    fetchMessages,
    fetchUnreadCount,
    markAllRead,
    markRead,
    messages,
    messagesTotal,
    removeMessage,
    startPolling,
    stopPolling,
    unreadCount,
  };
});
