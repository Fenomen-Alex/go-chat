<template>
  <div
    v-if="isOpen"
    class="fixed inset-0 z-50 flex items-center justify-center"
  >
    <div
      class="absolute inset-0 bg-black/50 backdrop-blur-sm transition-opacity"
      @click="closeDialog"
    ></div>
    <div
      class="bg-surface w-full max-w-md rounded-xl border border-border shadow-xl mx-4 p-6 animate-slide-up flex-shrink-0"
    >
      <h3 class="text-lg font-semibold text-text mb-6">Connect to Peer</h3>
      
      <div class="space-y-4">
        <div>
          <label class="block text-xs font-semibold text-text-muted uppercase tracking-wider mb-2">
            Multiaddress
          </label>
          <input
            type="text"
            v-model="addr"
            placeholder="/ip4/127.0.0.1/tcp/57171/p2p/12D3KooW..."
            class="w-full bg-input border border-border rounded-lg px-3 py-2 text-sm text-text focus:border-primary focus:ring-1 focus:ring-primary outline-none"
          />
          <p class="text-[10px] text-text-muted mt-1">
            Example: /ip4/192.168.1.50/tcp/57171/p2p/12D3KooW...
          </p>
        </div>

        <div>
          <label class="block text-xs font-semibold text-text-muted uppercase tracking-wider mb-2">
            Nickname (optional)
          </label>
          <input
            type="text"
            v-model="nickname"
            placeholder="Friendly name for this peer"
            class="w-full bg-input border border-border rounded-lg px-3 py-2 text-sm text-text focus:border-primary focus:ring-1 focus:ring-primary outline-none"
          />
        </div>

        <div v-if="error" class="p-3 rounded-lg bg-danger/10 border border-danger/20 text-danger text-sm">
          {{ error }}
        </div>

        <div v-if="isConnecting" class="flex items-center space-x-2 text-sm text-text-muted">
          <div class="w-3 h-3 rounded-full border-2 border-current border-t-transparent animate-pulse"></div>
          <span>Connecting...</span>
        </div>
      </div>

      <div class="pt-4 flex justify-end space-x-2">
        <button
          @click="closeDialog"
          class="px-4 py-2 text-sm text-text-muted hover:text-text transition-colors"
        >
          Cancel
        </button>
        <button
          :disabled="!addr.trim() || isConnecting"
          @click="connect"
          class="px-4 py-2 text-sm font-semibold text-primary hover:text-primary-dark disabled:bg-surface disabled:text-text-muted transition-colors"
        >
          Connect
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { WailsService } from '@/bindings'

interface Props {
  isOpen: boolean
  connectFn: (addr: string, nickname?: string) => Promise<void>
  isConnecting: boolean
  error: string | null
}

const props = defineProps<Props>()

const emit = defineEmits(['close', 'connect'])

const addr = ref('')
const nickname = ref('')

function closeDialog() {
  addr.value = ''
  nickname.value = ''
  emit('close')
}

async function connect() {
  if (!addr.value.trim()) return
  await props.connectFn(addr.value.trim(), nickname.value.trim() || undefined)
  emit('connect')
}
</script>

<style scoped>
.animate-slide-up {
  animation: slideUp 0.3s ease-out forwards;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.animate-pulse {
  animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}
</style>
