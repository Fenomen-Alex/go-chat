<template>
  <header
    class="h-16 bg-surface/80 backdrop-blur-md border-b border-border flex items-center justify-between px-4 shrink-0 z-10 flex-shrink-0"
  >
    <div class="flex items-center space-x-3 cursor-pointer" @click="goBack">
      <svg
        class="w-5 h-5 text-text-muted"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M15 19l-7-7 7-7"
        />
      </svg>
      <div class="w-10 h-10 rounded-full bg-primary-light flex items-center justify-center overflow-hidden flex-shrink-0">
        <img
          v-if="conversation.avatarUrl"
          :src="conversation.avatarUrl"
          class="w-full h-full object-cover"
        />
        <div
          v-else
          class="w-full h-full flex items-center justify-center text-primary font-semibold text-lg"
        >
          {{ conversation?.displayName?.charAt(0).toUpperCase() || '?' }}
        </div>
      </div>
      <div class="flex-1 min-w-0">
        <h4 class="text-sm font-semibold text-text truncate">
          {{ conversation.displayName }}
        </h4>
        <p
          class="text-xs text-text-muted"
          :class="{ 'text-success': isOnline, 'text-text-muted': !isOnline }"
        >
          {{ isOnline ? 'Online' : 'Offline' }}
        </p>
      </div>
    </div>

    <div class="flex items-center space-x-2 shrink-0">
      <button
        class="w-8 h-8 rounded-full hover:bg-hover flex items-center justify-center text-text-muted transition-colors"
        title="Voice call"
      >
        <svg
          class="w-5 h-5"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14.572V3.571L14.448 7.04A1 1 0 0113 7.914v6.086a1 1 0 01-1.447.894L5 10.571V3.571z"
          />
        </svg>
      </button>
      <button
        class="w-8 h-8 rounded-full hover:bg-hover flex items-center justify-center text-text-muted transition-colors"
        title="Video call"
      >
        <svg
          class="w-5 h-5"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14.572V3.571L14.448 7.04A1 1 0 0113 7.914v6.086a1 1 0 01-1.447.894L5 10.571V3.571z"
          />
        </svg>
      </button>
      <button
        class="w-8 h-8 rounded-full hover:bg-hover flex items-center justify-center text-text-muted transition-colors"
        title="Conversation info"
        @click="toggleDetails"
      >
        <svg
          class="w-5 h-5"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
      </button>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ChatView } from '@/types'

interface Props {
  conversation: ChatView
  isOnline: boolean
}

const props = defineProps<Props>()

const emit = defineEmits(['back', 'toggle-details'])

function goBack() {
  emit('back')
}

function toggleDetails() {
  emit('toggle-details')
}

const isOnline = computed(() => props.isOnline)
</script>

<style scoped></style>
