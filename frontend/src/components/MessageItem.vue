<template>
  <div class="message-wrapper flex" :class="{ 'justify-end': isSelf, 'justify-start': !isSelf }">
    <!-- Other user's message -->
    <div v-if="!isSelf" class="flex items-end space-x-2 animate-fade-in">
      <div class="w-7 h-7 rounded-full bg-primary/10 flex items-center justify-center overflow-hidden flex-shrink-0">
        <div class="w-full h-full flex items-center justify-center text-primary font-medium text-xs">{{ message?.senderName?.[0] || 'U' }}</div>
      </div>
      <div class="flex flex-col">
        <div class="message-bubble bg-message-other rounded-lg rounded-tr-none px-3 py-2 text-sm text-chat-text max-w-[80%] break-words whitespace-pre-wrap" :class="{ 'rounded-tr-lg': !isSelf }">
          {{ message.content }}
        </div>
        <div class="flex items-center space-x-1 mt-0.5">
          <span class="text-[10px] text-chat-text-muted">{{ formatTime(message.timestamp) }}</span>
          <span v-if="message.deliveryState === 'read'" class="text-[10px] text-success">✓✓</span>
          <span v-else-if="message.deliveryState === 'sent'" class="text-[10px] text-chat-text-muted">✓</span>
        </div>
      </div>
    </div>

    <!-- Self message -->
    <div v-else class="flex items-end space-x-2 animate-fade-in">
      <div class="flex flex-col">
        <div class="message-bubble bg-message-self rounded-lg rounded-tl-none px-3 py-2 text-sm text-white max-w-[80%] break-words whitespace-pre-wrap" :class="{ 'rounded-tl-lg': isSelf }">
          {{ message.content }}
        </div>
        <div class="flex items-center space-x-1 mt-0.5 justify-end">
          <span class="text-[10px] text-chat-text-muted">{{ formatTime(message.timestamp) }}</span>
          <span v-if="message.deliveryState === 'read'" class="text-[10px] text-success">✓✓</span>
          <span v-else-if="message.deliveryState === 'sent'" class="text-[10px] text-chat-text-muted">✓</span>
          <span v-else class="text-[10px] text-warning animate-pulse">●</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'MessageItem',
  props: {
    message: {
      type: Object,
      required: true
    },
    isSelf: {
      type: Boolean,
      default: false
    }
  },
  methods: {
    formatTime(timestamp) {
      if (!timestamp) return ''
      const date = new Date(timestamp)
      return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    }
  }
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
  background-color: #7c3aed;
  color: #ffffff;
}

.text-success {
  color: #22c55e;
}

.text-warning {
  color: #f59e0b;
}
</style>
