<template>
  <div
    ref="scrollContainer"
    class="flex-1 overflow-y-auto p-4 space-y-4 min-h-0"
  >
    <!-- Date separator groups -->
    <div
      v-for="(group, index) in groupedMessages"
      :key="index"
      class="flex flex-col space-y-2 animate-fade-in"
    >
      <div v-if="group.date" class="flex items-center space-x-2 my-1">
        <span class="h-[1px] bg-border flex-1"></span>
        <span
          class="text-[10px] text-text-muted px-2 bg-app-bg uppercase tracking-wider font-semibold"
        >
          {{ group.date }}
        </span>
        <span class="h-[1px] bg-border flex-1"></span>
      </div>

      <MessageItem
        v-for="msg in group.messages"
        :key="msg.message_id"
        :message="msg"
        :is-self="msg.sender_peer_id === myPeerId"
      />
    </div>

    <!-- Empty state -->
    <div
      v-if="messages.length === 0 && !loadingMessages"
      class="flex flex-col items-center justify-center min-h-[200px] text-center py-20 opacity-60"
    >
      <p class="text-text-muted text-sm">No messages yet</p>
      <p class="text-[11px] text-text-muted mt-1">Say hello!</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUpdated, nextTick } from 'vue'
import type { MessageView } from '@/types'
import MessageItem from './MessageItem.vue'

interface Props {
  messages: MessageView[]
  myPeerId: string
  loadingMessages: boolean
}

const props = defineProps<Props>()

const scrollContainer = ref<HTMLDivElement | null>(null)

const groupedMessages = computed(() => {
  const groups: Array<{ date: string | null; messages: MessageView[] }> = []
  let currentGroup: { date: string | null; messages: MessageView[] } | null = null
  let currentDate: string | null = null

  for (const msg of props.messages) {
    const msgDate = formatDateOnly(msg.timestamp)
    if (msgDate !== currentDate) {
      if (currentGroup) groups.push(currentGroup)
      currentDate = msgDate
      currentGroup = { date: currentDate, messages: [] }
    }
    currentGroup.messages.push(msg)
  }
  if (currentGroup) groups.push(currentGroup)
  return groups
})

function scrollToBottom() {
  nextTick(() => {
    if (scrollContainer.value) {
      scrollContainer.value.scrollTop = scrollContainer.value.scrollHeight
    }
  })
}

onMounted(() => {
  scrollToBottom()
})

onUpdated(() => {
  scrollToBottom()
})

function formatTime(timestamp: string): string {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  return date.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' })
}

function formatDateOnly(timestamp: string): string {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  const now = new Date()
  if (date.toDateString() === now.toDateString()) return 'Today'
  const yesterday = new Date(now)
  yesterday.setDate(yesterday.getDate() - 1)
  if (date.toDateString() === yesterday.toDateString()) return 'Yesterday'
  return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
}
</script>

<style scoped>
.animate-fade-in {
  animation: fadeIn 0.2s ease-out forwards;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
