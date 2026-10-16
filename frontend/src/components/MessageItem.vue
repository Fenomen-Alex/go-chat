<template>
  <div
    class="message-wrapper flex items-end space-x-2 animate-fade-in"
    :class="{ 'justify-end': isSelf, 'justify-start': !isSelf }"
  >
    <!-- Other user's message -->
    <div v-if="!isSelf" class="flex items-end space-x-2">
      <div
        class="w-7 h-7 rounded-full bg-primary-light flex items-center justify-center overflow-hidden flex-shrink-0"
      >
        <div
          class="w-full h-full flex items-center justify-center text-primary font-medium text-xs"
        >
          {{ message.senderName?.charAt(0).toUpperCase() || '?' }}
        </div>
      </div>
      <div class="flex flex-col">
        <div
          class="message-bubble bg-message-other rounded-lg rounded-tr-none px-3 py-2 text-sm text-text max-w-[80%] break-words whitespace-pre-wrap"
        >
          {{ message.content }}
        </div>
        <div class="flex items-center space-x-1 mt-1">
          <span class="text-[10px] text-text-muted">{{ formatTime(message.timestamp) }}</span>
          <span
            v-if="message.deliveryState === 'read'"
            class="text-[10px] text-success"
          >
            ✓✓
          </span>
          <span
            v-else-if="message.deliveryState === 'sent'"
            class="text-[10px] text-text-muted"
          >
            ✓
          </span>
        </div>
      </div>
    </div>

    <!-- Self message -->
    <div v-else class="flex items-end space-x-2">
      <div class="flex flex-col items-end">
        <div
          class="message-bubble bg-message-self rounded-lg rounded-tl-none px-3 py-2 text-sm text-white max-w-[80%] break-words whitespace-pre-wrap"
        >
          {{ message.content }}
        </div>
        <div class="flex items-center space-x-1 mt-1">
          <span class="text-[10px] text-text-muted">{{ formatTime(message.timestamp) }}</span>
          <span
            v-if="message.deliveryState === 'read'"
            class="text-[10px] text-success"
          >
            ✓✓
          </span>
          <span
            v-else-if="message.deliveryState === 'sent'"
            class="text-[10px] text-text-muted"
          >
            ✓
          </span>
          <span v-else class="text-[10px] text-warning animate-pulse">●</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { MessageView } from '@/types'

interface Props {
  message: MessageView
  isSelf: boolean
}

const props = defineProps<Props>()

function formatTime(timestamp: string): string {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  return date.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' })
}
</script>

<style scoped>
.message-bubble {
  word-wrap: break-word;
  overflow-wrap: break-word;
}

.bg-message-other {
  background-color: rgba(55, 65, 81, 0.6);
  border: 1px solid rgba(75, 85, 99, 0.3);
  color: #f3f4f6;
}

.bg-message-self {
  background-color: #4f46e5;
  color: #ffffff;
}

.text-success {
  color: #22c55e;
}

.text-warning {
  color: #f59e0b;
}
</style>
