<template>
  <div class="message-item flex" :class="{ self: isSelf, other: !isSelf }">
    <div class="message-item-avatar">
      <img v-if="message.avatarUrl" :src="message.avatarUrl" :alt="message.senderName" class="w-full h-full object-cover rounded-full" />
      <div v-else class="w-full h-full flex items-center justify-center bg-border/50 text-border/50 text-xs font-medium">
        {{ message.senderName[0] }}
      </div>
    </div>
    <div class="message-item-content flex-1">
      <div class="message-item-header">
        <div class="message-item-author text-sm font-medium">{{ message.senderName }}</div>
        <div class="message-item-timestamp text-xs text-border/40">{{ formatTimestamp(message.timestamp) }}</div>
      </div>
      <div class="message-item-text text-sm whitespace-pre-wrap break-all">{{ message.content }}</div>
      <div v-if="message.deliveryState && !isSelf" class="message-item-status text-xs" :class="{ sent: message.deliveryState === 'sent', received: message.deliveryState === 'received', failed: message.deliveryState === 'failed' }">
        {{ message.deliveryState === 'sent' ? '✓' : message.deliveryState === 'received' ? '✓✓' : '✗' }}
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
    formatTimestamp(timestamp) {
      if (!timestamp) return ''
      
      const date = new Date(timestamp)
      const now = new Date()
      
      // If today, show time only
      if (date.toDateString() === now.toDateString()) {
        return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
      }
      
      // If yesterday, show "Yesterday"
      const yesterday = new Date(now)
      yesterday.setDate(yesterday.getDate() - 1)
      if (date.toDateString() === yesterday.toDateString()) {
        return 'Yesterday'
      }
      
      // If within last 7 days, show day and time
      const diffTime = Math.abs(now - date)
      const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24))
      if (diffDays < 7) {
        return date.toLocaleDateString(undefined, { weekday: 'short' }) + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
      }
      
      // Otherwise, show date
      return date.toLocaleDateString()
    }
  }
}
</script>

<style scoped>
/* Styles are handled in App.vue scoped styles */
</style>