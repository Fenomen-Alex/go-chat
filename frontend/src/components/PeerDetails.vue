<template>
  <aside
    v-if="isOpen"
    class="w-80 bg-surface border-l border-border flex flex-col shrink-0 z-20 flex-shrink-0 min-h-0 animate-slide-up"
  >
    <!-- Details header -->
    <div class="h-16 bg-surface border-b border-border flex items-center justify-between px-4 shrink-0 flex-shrink-0">
      <div class="flex items-center space-x-3">
        <div
          class="w-8 h-8 rounded-full bg-primary-light flex items-center justify-center overflow-hidden flex-shrink-0"
        >
          <div
            class="w-full h-full flex items-center justify-center text-primary font-semibold text-sm"
          >
            {{ selectedPeer?.display_name?.charAt(0).toUpperCase() || selectedChannel?.name?.charAt(0).toUpperCase() || '?' }}
          </div>
        </div>
        <div class="flex-1 min-w-0">
          <h4 class="text-sm font-semibold text-text truncate">
            {{ selectedPeer?.display_name || selectedChannel?.name || 'Details' }}
          </h4>
          <p class="text-xs text-text-muted">
            {{ selectedPeer ? selectedPeer.peer_id : selectedChannel?.channel_id || '' }}
          </p>
        </div>
      </div>
      <button
        class="w-8 h-8 rounded-full hover:bg-hover flex items-center justify-center text-text-muted transition-colors"
        @click="closeDetails"
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
            d="M6 18L18 6M6 6l12 12"
          />
        </svg>
      </button>
    </div>

    <div class="flex-1 overflow-y-auto p-4 space-y-6 min-h-0">
      <!-- Peer info -->
      <div v-if="selectedPeer" class="space-y-4">
        <h5 class="text-xs font-semibold text-text-muted uppercase tracking-wider">Peer</h5>
        <div class="space-y-3">
          <div class="flex items-start space-x-3">
            <div class="w-8 h-8 rounded-full bg-primary-light flex items-center justify-center flex-shrink-0">
            <span class="text-primary font-semibold text-xs">
              {{ selectedPeer?.display_name?.charAt(0).toUpperCase() || selectedChannel?.name?.charAt(0).toUpperCase() || '?' }}
            </span>
            </div>
            <div class="flex-1 min-w-0">
              <p class="text-sm text-text">{{ selectedPeer.display_name }}</p>
              <p class="text-xs text-text-muted truncate">
                {{ selectedPeer.peer_id }}
              </p>
            </div>
          </div>
          <div class="flex items-center space-x-2">
            <span
              class="w-2 h-2 rounded-full"
              :class="{ 'bg-success': selectedPeer.status === 'online', 'bg-text-muted': selectedPeer.status !== 'online' }"
            ></span>
            <span
              class="text-xs font-medium"
              :class="{ 'text-success': selectedPeer.status === 'online', 'text-text-muted': selectedPeer.status !== 'online' }"
            >
              {{ selectedPeer.status }}
            </span>
          </div>
          <div class="flex items-center space-x-2">
            <span class="text-xs text-text-muted">Trusted:</span>
            <span
              class="w-4 h-2 rounded-full"
              :class="{ 'bg-success': selectedPeer.trusted, 'bg-text-muted': !selectedPeer.trusted }"
            ></span>
          </div>
          <div class="flex items-center space-x-2">
            <span class="text-xs text-text-muted">Blocked:</span>
            <span
              class="w-4 h-2 rounded-full"
              :class="{ 'bg-danger': selectedPeer.blocked, 'bg-text-muted': !selectedPeer.blocked }"
            ></span>
          </div>
          <div class="flex items-center space-x-2">
            <span class="text-xs text-text-muted">Favorite:</span>
            <span
              class="w-4 h-2 rounded-full"
              :class="{ 'bg-warning': selectedPeer.favorite, 'bg-text-muted': !selectedPeer.favorite }"
            ></span>
          </div>
        </div>
      </div>

      <!-- Channel info -->
      <div v-else-if="selectedChannel" class="space-y-4">
        <h5 class="text-xs font-semibold text-text-muted uppercase tracking-wider">Channel</h5>
        <div class="space-y-3">
          <div class="flex items-start space-x-3">
            <div class="w-8 h-8 rounded-full bg-primary-light flex items-center justify-center flex-shrink-0">
            <span class="text-primary font-semibold text-xs">
              {{ selectedChannel?.name?.charAt(0).toUpperCase() || '?' }}
            </span>
            </div>
            <div class="flex-1 min-w-0">
              <p class="text-sm text-text">{{ selectedChannel.name }}</p>
              <p class="text-xs text-text-muted truncate">
                {{ selectedChannel.channel_id }}
              </p>
              <p v-if="selectedChannel.description" class="text-xs text-text-muted">
                {{ selectedChannel.description }}
              </p>
            </div>
          </div>
          <div class="flex items-center space-x-2">
            <span class="text-xs text-text-muted">Type:</span>
            <span class="text-xs font-medium text-primary">
              {{ selectedChannel.channel_type }}
            </span>
          </div>
          <div v-if="selectedChannel.unread_count > 0" class="flex items-center space-x-2">
            <span class="text-xs text-text-muted">Unread:</span>
            <span class="text-xs font-semibold text-text">
              {{ selectedChannel.unread_count }}
            </span>
          </div>
        </div>
      </div>

      <!-- Actions -->
      <div class="space-y-2">
        <button
          v-if="selectedPeer?.peer_id"
          class="w-full flex items-center space-x-2 px-3 py-2 bg-hover rounded-lg text-sm font-medium text-primary transition-colors hover:bg-active"
          @click="copyPeerId"
        >
          <svg
            class="w-4 h-4"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3"
            />
          </svg>
          Copy Peer ID
        </button>
        <button
          class="w-full flex items-center space-x-2 px-3 py-2 bg-hover rounded-lg text-sm font-medium text-text-muted transition-colors hover:bg-active"
          @click="closeDetails"
        >
          <svg
            class="w-4 h-4"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
          Close
        </button>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { PeerView, ChannelView } from '@/types'

interface Props {
  selectedPeer: PeerView | null
  selectedChannel: ChannelView | null
  isOpen: boolean
}

const props = defineProps<Props>()

const emit = defineEmits(['close', 'copy-peer-id'])

function closeDetails() {
  emit('close')
}

function copyPeerId() {
  if (props.selectedPeer?.peer_id) {
    navigator.clipboard.writeText(props.selectedPeer.peer_id)
    emit('copy-peer-id')
  }
}

const hasContent = computed(() => !!props.selectedPeer || !!props.selectedChannel)
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
</style>
