<template>
  <section
    class="flex-1 flex flex-col bg-app-bg min-w-0 min-h-0"
  >
    <!-- Empty state -->
    <div
      v-if="!conversation && !loading"
      class="flex-1 flex flex-col items-center justify-center text-center px-8 min-h-0"
    >
      <div
        class="w-24 h-24 rounded-full bg-primary-light flex items-center justify-center mb-6"
      >
        <svg
          class="w-12 h-12 text-primary"
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
      </div>
      <h3 class="text-lg font-semibold text-text mb-2">Select a conversation</h3>
      <p class="text-text-muted text-sm">
        Choose a chat from the sidebar to start messaging
      </p>
    </div>

    <!-- Loading state -->
    <div
      v-else-if="loading && !conversation"
      class="flex-1 flex items-center justify-center min-h-0"
    >
      <div class="text-text-muted text-sm">Loading...</div>
    </div>

    <!-- Active conversation -->
    <div v-else class="flex flex-col h-full min-h-0">
      <ConversationHeader
        :conversation="conversation"
        :is-online="isOnline"
        @back="goBack"
        @toggle-details="toggleDetails"
      />

      <MessageList
        :messages="messages"
        :my-peer-id="myPeerId"
        :loading-messages="loadingMessages"
      />

      <MessageComposer
        :channel-id="conversation?.channelId || null"
        :send-message-fn="sendMessage"
        :is-sending="isSending"
        @send="onMessageSent"
      />
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import type { ChatView, MessageView } from '@/types'
import ConversationHeader from './ConversationHeader.vue'
import MessageList from './MessageList.vue'
import MessageComposer from './MessageComposer.vue'

interface Props {
  conversation: ChatView | null
  messages: MessageView[]
  myPeerId: string
  loading: boolean
  loadingMessages: boolean
  isOnline: boolean
  isSending: boolean
  sendMessageFn: (content: string) => Promise<void>
}

const props = defineProps<Props>()

const emit = defineEmits(['back', 'toggle-details', 'message-sent'])

function goBack() {
  emit('back')
}

function toggleDetails() {
  emit('toggle-details')
}

async function sendMessage(content: string) {
  await props.sendMessageFn(content)
}

function onMessageSent() {
  emit('message-sent')
}

const isOnline = computed(() => props.isOnline)
</script>

<style scoped></style>
