<template>
  <div class="overflow-y-auto pr-1 min-h-0">
    <!-- Empty state -->
    <div
      v-if="chats.length === 0 && !loading"
      class="px-4 py-12 text-center"
    >
      <svg
        class="w-16 h-16 mx-auto text-text-muted mb-4"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
        />
      </svg>
      <p class="text-sm text-text-muted">No conversations yet</p>
      <p class="text-xs text-text-muted mt-1">
        Connect to a peer to start messaging
      </p>
    </div>

    <!-- Chat list -->
    <div v-for="chat in filteredChats" :key="chat.channelId"
      @click="selectChat(chat)"
      class="flex items-center space-x-3 px-4 py-3 cursor-pointer group hover:bg-hover rounded-lg transition-colors duration-150 mb-0.5"
      :class="{ 'bg-active border-r-2 border-primary': selectedChat?.channelId === chat.channelId }"
    >
      <!-- Avatar -->
      <div
        class="w-10 h-10 rounded-full bg-primary-light flex items-center justify-center overflow-hidden flex-shrink-0"
      >
        <img
          v-if="chat.avatarUrl"
          :src="chat.avatarUrl"
          class="w-full h-full object-cover"
        />
        <div
          v-else
          class="w-full h-full flex items-center justify-center text-primary font-semibold text-sm"
        >
          {{ chat.displayName?.charAt(0).toUpperCase() || '?' }}
        </div>
      </div>

      <!-- Info -->
      <div class="flex-1 min-w-0">
        <div class="flex items-center space-x-2">
          <h4 class="text-sm font-semibold text-text truncate">
            {{ chat.displayName }}
          </h4>
          <span
            v-if="chat.unreadCount > 0"
            class="w-2.5 h-2.5 rounded-full bg-danger flex-shrink-0"
          ></span>
        </div>
        <p
          class="text-xs text-text-muted truncate line-clamp-1"
          :class="{ 'font-semibold text-text': chat.unreadCount > 0 }"
        >
          {{ chat.lastMessage || 'No messages yet' }}
        </p>
      </div>

      <!-- Time -->
      <div class="text-[10px] text-text-muted flex-shrink-0 whitespace-nowrap">
        {{ chat.lastMessageTime || '' }}
      </div>

      <!-- Unread badge -->
      <span
        v-if="chat.unreadCount > 0"
        class="w-5 h-5 rounded-full bg-primary flex items-center justify-center text-[10px] text-white font-semibold flex-shrink-0"
      >
        {{ chat.unreadCount > 9 ? '9+' : chat.unreadCount }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ChatView } from '@/types'

interface Props {
  chats: ChatView[]
  searchQuery: string
  selectedChat: ChatView | null
  loading: boolean
}

const props = defineProps<Props>()

const emit = defineEmits(['select'])

const filteredChats = computed(() => {
  const query = props.searchQuery.toLowerCase().trim()
  if (!query) return props.chats
  return props.chats.filter((c) =>
    c.displayName.toLowerCase().includes(query) ||
    (c.lastMessage && c.lastMessage.toLowerCase().includes(query))
  )
})

function selectChat(chat: ChatView) {
  emit('select', chat)
}
</script>

<style scoped></style>
