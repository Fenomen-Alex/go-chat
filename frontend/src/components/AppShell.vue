<template>
  <div
    class="app min-h-screen w-full flex flex-col bg-app-bg text-text overflow-hidden"
  >
    <!-- Top App Bar -->
    <header
      class="h-16 bg-surface border-b border-border flex items-center justify-between px-4 shrink-0 z-20 flex-shrink-0"
    >
      <div class="flex items-center space-x-3 cursor-pointer" @click="openProfile">
        <div
          class="w-10 h-10 rounded-full bg-primary-light flex items-center justify-center overflow-hidden flex-shrink-0"
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
        <div class="min-w-0 flex flex-col">
          <h1 class="text-sm font-semibold text-text truncate">go-chat</h1>
          <p class="text-xs text-text-muted">
            {{ isConnected ? 'Online' : 'Offline' }}
          </p>
        </div>
      </div>

      <div class="flex items-center space-x-2">
        <button
          class="w-8 h-8 rounded-full hover:bg-hover flex items-center justify-center text-text-muted transition-colors"
          title="Connect to peer"
          @click="openConnectionDialog"
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
              d="M13 10V3L4 14h7v7l9-11h-7z"
            />
          </svg>
        </button>
        <button
          class="w-8 h-8 rounded-full hover:bg-hover flex items-center justify-center text-text-muted transition-colors"
          title="Settings"
          @click="openSettings"
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
              d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
              />
          </svg>
        </button>
      </div>
    </header>

    <!-- Main shell: 3-column grid -->
    <main
      class="flex-1 overflow-hidden grid grid-cols-[250px_minmax(0,1fr)_270px] min-h-0"
    >
      <!-- Left Sidebar -->
      <AppSidebar
        :identity="identity"
        :organization="organization"
        :chats="chats"
        :selected-chat="selectedChat"
        :loading="loading"
        @select="onSelectChat"
        @profile="openProfile"
        @new-chat="openConnectionDialog"
      />

      <!-- Center Conversation -->
      <ConversationView
        :conversation="selectedChat"
        :messages="messages"
        :my-peer-id="myPeerId"
        :loading="loading"
        :loading-messages="loadingMessages"
        :is-online="isOnline"
        :is-sending="isSending"
        :send-message-fn="sendMessage"
        @back="goBack"
        @toggle-details="toggleDetails"
        @message-sent="onMessageSent"
      />

      <!-- Right Details Sidebar -->
      <PeerDetails
        :selected-peer="selectedPeer"
        :selected-channel="selectedChannel"
        :is-open="showDetails"
        @close="toggleDetails"
        @copy-peer-id="copyPeerId"
      />
    </main>

    <!-- Modals -->
    <ConnectionDialog
      :is-open="showConnectionDialog"
      :is-connecting="isConnecting"
      :error="connectionError"
      :connect-fn="handleConnect"
      @close="closeConnectionDialog"
    />

    <SettingsView
      :is-open="showSettings"
      :current-name="identity.displayName"
      :save-fn="saveDisplayName"
      :is-saving="isSaving"
      @close="closeSettings"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive } from 'vue'
import type { ChatView, IdentityView, MessageView, OrganizationView, PeerView } from '@/types'
import { loadWailsBindings, type WailsService } from '@/bindings'
import AppSidebar from './AppSidebar.vue'
import ConversationView from './ConversationView.vue'
import PeerDetails from './PeerDetails.vue'
import ConnectionDialog from './ConnectionDialog.vue'
import SettingsView from './SettingsView.vue'

// State
const identity = reactive<IdentityView>({
  displayName: 'User',
  peerId: '',
  publicKey: '',
  avatarColor: '',
  fingerprint: '',
  createdAt: '',
  isDefault: false,
})

const myPeerId = ref('')
const chats = ref<ChatView[]>([])
const selectedChat = ref<ChatView | null>(null)
const messages = ref<MessageView[]>([])
const organization = ref<OrganizationView | null>(null)
const loading = ref(true)
const loadingMessages = ref(false)
const isConnected = ref(false)
const isOnline = ref(false)
const isSending = ref(false)
const isSaving = ref(false)
const showDetails = ref(false)
const showConnectionDialog = ref(false)
const showSettings = ref(false)
const isConnecting = ref(false)
const connectionError = ref<string | null>(null)

let bindings: WailsService | null = null

// Computed
const selectedPeer = computed<PeerView | null>(() => {
  if (!selectedChat.value || selectedChat.value.channelType !== 'dm') return null
  return {
    peer_id: selectedChat.value.channelId,
    display_name: selectedChat.value.displayName,
    status: selectedChat.value.status,
    trusted: true,
    blocked: false,
    favorite: false,
    avatar_color: '',
    fingerprint: '',
    first_seen: '',
    last_seen: '',
  }
})

const selectedChannel = computed<ChannelView | null>(() => {
  if (!selectedChat.value || selectedChat.value.channelType === 'dm') return null
  return {
    org_id: '',
    name: selectedChat.value.displayName,
    description: '',
    channel_id: selectedChat.value.channelId,
    channel_type: selectedChat.value.channelType,
    unread_count: selectedChat.value.unreadCount,
    archived: false,
  }
})

// Initialize
async function init() {
  try {
    bindings = await loadWailsBindings()
    
    // Load identity
    const ident = await bindings.Identity()
    identity.displayName = ident.display_name
    identity.peerId = ident.peer_id
    identity.publicKey = ident.public_key
    identity.avatarColor = ident.avatar_color
    identity.fingerprint = ident.fingerprint
    identity.createdAt = ident.created_at
    identity.isDefault = ident.is_default
    myPeerId.value = ident.peer_id

    // Load data
    await loadChats()
    setupEventListeners()
  } catch (err) {
    console.error('Failed to initialize:', err)
  } finally {
    loading.value = false
  }
}

async function loadChats() {
  try {
    const [channelsResponse, dmsResponse, peersResponse] = await Promise.all([
      bindings!.Channels(),
      bindings!.DMChannels(),
      bindings!.Peers(),
    ])

    const peerMap = new Map<string, any>()
    for (const p of (peersResponse || [])) {
      peerMap.set(p.peer_id, p)
    }

    const chatList: ChatView[] = []

    for (const ch of (channelsResponse || [])) {
      const peer = peerMap.get(ch.channel_id) || {}
      chatList.push({
        channelId: ch.channel_id,
        displayName: ch.name || 'Channel',
        channelType: ch.channel_type,
        unreadCount: ch.unread_count || 0,
        avatarUrl: '',
        lastMessage: '',
        lastMessageTime: '',
        status: peer.status || 'offline',
      })
    }

    for (const dm of (dmsResponse || [])) {
      const peer = peerMap.get(dm.channel_id) || {}
      chatList.push({
        channelId: dm.channel_id,
        displayName: dm.name || 'Direct Message',
        channelType: dm.channel_type,
        unreadCount: dm.unread_count || 0,
        avatarUrl: '',
        lastMessage: '',
        lastMessageTime: '',
        status: peer.status || 'offline',
      })
    }

    chatList.sort((a, b) => {
      if (a.unreadCount > 0 && b.unreadCount === 0) return -1
      if (a.unreadCount === 0 && b.unreadCount > 0) return 1
      return 0
    })

    chats.value = chatList

    const diag = await bindings.ConnectedPeers()
    isConnected.value = diag.is_connected
  } catch (err) {
    console.error('Failed to load chats:', err)
  } finally {
    loading.value = false
  }
}

async function loadMessages(channelId: string) {
  loadingMessages.value = true
  try {
    const response = await bindings!.Messages({
      channel_id: channelId,
      limit: 100,
      offset: 0,
    })
    messages.value = (response || []).reverse()
  } catch (err) {
    console.error('Failed to load messages:', err)
    messages.value = []
  } finally {
    loadingMessages.value = false
  }
}

function setupEventListeners() {
  bindings!.on('new-message', (data: any) => {
    if (selectedChat.value && data.channel_id === selectedChat.value.channelId) {
      messages.value.push(data)
      scrollToBottom()
      updateChatListItem(data)
    } else {
      const chat = chats.value.find((c) => c.channelId === data.channel_id)
      if (chat) {
        chat.unreadCount = (chat.unreadCount || 0) + 1
        chat.lastMessage = data.content
        chat.lastMessageTime = formatTime(data.timestamp)
      }
    }
  })

  bindings!.on('connection-changed', (data: any) => {
    isConnected.value = data.connected
  })
}

// Actions
function onSelectChat(chat: ChatView) {
  selectedChat.value = chat
  messages.value = []
  loadMessages(chat.channelId)
  chat.unreadCount = 0
  showDetails.value = true
}

function goBack() {
  selectedChat.value = null
  messages.value = []
  showDetails.value = false
}

function toggleDetails() {
  showDetails.value = !showDetails.value
}

async function sendMessage(content: string) {
  if (!selectedChat.value || isSending.value) return
  isSending.value = true
  try {
    await bindings!.SendMessage(selectedChat.value.channelId, content, 'text')
  } catch (err) {
    console.error('Failed to send message:', err)
  } finally {
    isSending.value = false
  }
}

function onMessageSent() {
  // Message was sent, UI will update via event listener
}

function openProfile() {
  showSettings.value = true
}

function openConnectionDialog() {
  showConnectionDialog.value = true
}

function openSettings() {
  showSettings.value = true
}

function closeConnectionDialog() {
  showConnectionDialog.value = false
  connectionError.value = null
}

async function handleConnect(addr: string, nickname?: string) {
  isConnecting.value = true
  connectionError.value = null
  try {
    await bindings!.Connect({ addr })
    showConnectionDialog.value = false
    await loadChats()
  } catch (err: any) {
    connectionError.value = err.message || 'Failed to connect'
  } finally {
    isConnecting.value = false
  }
}

function closeSettings() {
  showSettings.value = false
}

async function saveDisplayName() {
  if (!identity.displayName.trim()) return
  isSaving.value = true
  try {
    await bindings!.SetDisplayName({ name: identity.displayName.trim() })
    showSettings.value = false
  } catch (err: any) {
    console.error('Failed to save display name:', err)
  } finally {
    isSaving.value = false
  }
}

function copyPeerId() {
  if (selectedChat.value?.channelId) {
    navigator.clipboard.writeText(selectedChat.value.channelId)
  }
}

function formatTime(timestamp: string): string {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  return date.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' })
}

function updateChatListItem(data: any) {
  const chat = chats.value.find((c) => c.channelId === data.channel_id)
  if (chat) {
    chat.lastMessage = data.content
    chat.lastMessageTime = formatTime(data.timestamp)
    const index = chats.value.indexOf(chat)
    if (index > 0) {
      chats.value.splice(index, 1)
      chats.value.unshift(chat)
    }
  }
}

// Expose for parent (none - AppShell is root)
init()
</script>

<style scoped>
.app {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
}
</style>
