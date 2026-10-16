<template>
  <div
    class="bg-surface/80 backdrop-blur-md border-t border-border px-4 py-3 shrink-0 z-10 flex-shrink-0 min-h-[80px]"
  >
    <div class="flex items-end space-x-2 max-w-4xl mx-auto">
      <button
        class="w-9 h-9 rounded-lg hover:bg-hover flex items-center justify-center text-text-muted transition-colors"
        title="Emoji"
        @click="openEmojiPicker"
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
            d="M2 10h20M4.931 4.931a1 1 0 000 1.414l13.23 13.23a1 1 0 001.414 0l1.414-1.414a1 1 0 000-1.414l-13.23-13.23a1 1 0 00-1.414 0zM21 12a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
      </button>
      <button
        class="w-9 h-9 rounded-lg hover:bg-hover flex items-center justify-center text-text-muted transition-colors"
        title="Attach file"
        @click="openFilePicker"
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
            d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
          />
        </svg>
      </button>
      <div class="flex-1 relative min-w-0">
        <textarea
          id="message-input"
          v-model="messageText"
          placeholder="Type a message..."
          rows="1"
          class="w-full bg-input border border-border rounded-lg px-3 py-2 text-sm text-text resize-none focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary min-h-[38px] max-h-[120px] leading-tight transition-colors"
          @keydown.enter.prevent="sendMessage"
        ></textarea>
      </div>
      <button
        :disabled="!canSend"
        class="w-9 h-9 rounded-lg bg-primary flex items-center justify-center text-white hover:bg-primary-hover disabled:bg-surface disabled:text-text-muted transition-all duration-200"
        title="Send"
        @click="sendMessage"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z"
          />
        </svg>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

interface Props {
  channelId: string | null
  sendMessageFn: (content: string) => Promise<void>
  isSending: boolean
}

const props = defineProps<Props>()

const messageText = ref('')

const canSend = computed(() => {
  return !!props.channelId && messageText.value.trim().length > 0 && !props.isSending
})

async function sendMessage() {
  if (!props.channelId || !messageText.value.trim() || props.isSending) return
  const content = messageText.value.trim()
  await props.sendMessageFn(content)
  messageText.value = ''
}

function openEmojiPicker() {
  // Placeholder for emoji picker
}

function openFilePicker() {
  // Placeholder for file attachment
}
</script>

<style scoped></style>
