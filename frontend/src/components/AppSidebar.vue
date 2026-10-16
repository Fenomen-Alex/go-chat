<template>
  <aside
    class="w-72 bg-surface border-r border-border flex flex-col shrink-0 z-10 flex-shrink-0 min-h-0"
  >
    <div class="flex-1 overflow-y-auto min-h-0">
      <!-- Profile Section -->
      <div class="px-4 py-4 border-b border-border">
        <div class="flex items-center space-x-3 cursor-pointer group" @click="openProfile">
          <div
            class="w-10 h-10 rounded-full bg-primary-light flex items-center justify-center overflow-hidden group-hover:scale-105 transition-transform duration-200 flex-shrink-0"
          >
            <img
              v-if="identity.avatarUrl"
              :src="identity.avatarUrl"
              class="w-full h-full object-cover"
            />
            <div
              v-else
              class="w-full h-full flex items-center justify-center text-primary font-semibold text-lg"
            >
              {{ identity.displayName?.charAt(0).toUpperCase() || '?' }}
            </div>
          </div>
          <div class="flex-1 min-w-0">
            <h3 class="text-sm font-semibold text-text truncate">
              {{ identity.displayName }}
            </h3>
            <p class="text-xs text-text-muted truncate">
              {{ identity.peerId ? identity.peerId.slice(-8) : '' }}
            </p>
          </div>
        </div>
      </div>

      <!-- Workspace Nav -->
      <WorkspaceNav :organization="organization" />

      <!-- Search -->
      <div class="px-4 py-3 border-b border-border">
        <div class="relative">
          <svg
            class="w-4 h-4 absolute left-3 top-3 text-text-muted"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
            />
          </svg>
          <input
            type="text"
            placeholder="Search conversations"
            v-model="searchQuery"
            class="w-full bg-input pl-10 pr-4 py-2 rounded-lg text-sm text-text border border-border focus:border-primary focus:ring-1 focus:ring-primary outline-none transition-colors"
          />
        </div>
      </div>

      <!-- Channel List -->
      <ChannelList
        :chats="chats"
        :search-query="searchQuery"
        :selected-chat="selectedChat"
        :loading="loading"
        @select="onSelectChat"
      />
    </div>

    <!-- New Chat / Connect Button -->
    <div class="px-4 py-3 border-t border-border shrink-0">
      <button
        @click="startNewChat"
        class="w-full flex items-center justify-center space-x-2 px-3 py-2 bg-hover rounded-lg text-sm font-medium text-primary transition-colors hover:bg-active"
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
            d="M12 6v6m0 0v6m0-6h6m-6 0H6"
          />
        </svg>
        <span>New chat</span>
      </button>
    </div>
  </aside>
</template>

<script setup lang="ts">
import type { ChatView, IdentityView, OrganizationView } from '@/types'
import { ref } from 'vue'
import WorkspaceNav from './WorkspaceNav.vue'
import ChannelList from './ChannelList.vue'

interface Props {
  identity: IdentityView
  organization: OrganizationView | null
  chats: ChatView[]
  selectedChat: ChatView | null
  loading: boolean
}

const props = defineProps<Props>()

const emit = defineEmits(['select', 'profile', 'new-chat'])

const searchQuery = ref('')

function onSelectChat(chat: ChatView) {
  emit('select', chat)
}

function openProfile() {
  emit('profile')
}

function startNewChat() {
  emit('new-chat')
}
</script>

<style scoped></style>
