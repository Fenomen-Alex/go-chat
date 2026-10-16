<template>
  <div class="app min-h-screen w-full flex flex-col bg-app-bg text-text overflow-hidden">
    <!-- Top App Bar -->
    <header class="h-16 bg-surface border-b border-border flex items-center justify-between px-4 shrink-0 z-20 flex-shrink-0">
      <div class="flex items-center space-x-3 cursor-pointer" @click="openProfile">
        <div class="w-10 h-10 rounded-full bg-primary-light flex items-center justify-center overflow-hidden flex-shrink-0">
          <img v-if="myIdentity?.avatarUrl" :src="myIdentity.avatarUrl" class="w-full h-full object-cover" />
          <div v-else class="w-full h-full flex items-center justify-center text-primary font-semibold text-lg">{{ myIdentity?.displayName[0] || 'U' }}</div>
        </div>
        <div class="min-w-0 flex flex-col">
          <h1 class="text-sm font-semibold text-text truncate">go-chat</h1>
          <p class="text-xs text-text-muted">{{ isConnected ? 'Online' : 'Offline' }}</p>
        </div>
      </div>

      <div class="flex items-center space-x-2">
        <button class="w-8 h-8 rounded-full hover:bg-hover flex items-center justify-center text-text-muted transition-colors" title="New chat" @click="startNewChat">
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
          </svg>
        </button>
        <button class="w-8 h-8 rounded-full hover:bg-hover flex items-center justify-center text-text-muted transition-colors" title="Settings" @click="openSettings">
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
        </button>
      </div>
    </header>

    <main class="flex-1 overflow-hidden flex relative min-h-0">
      <!-- Left Sidebar: Chats -->
      <aside class="w-80 bg-surface border-r border-border flex flex-col shrink-0 z-10 flex-shrink-0 min-h-0">
        <div class="flex-1 overflow-y-auto min-h-0">
          <!-- Profile Section -->
          <div class="px-4 py-4 border-b border-border">
            <div class="flex items-center space-x-3 cursor-pointer group" @click="openProfile">
              <div class="w-10 h-10 rounded-full bg-primary-light flex items-center justify-center overflow-hidden group-hover:scale-105 transition-transform duration-200 flex-shrink-0">
                <img v-if="myIdentity?.avatarUrl" :src="myIdentity.avatarUrl" class="w-full h-full object-cover" />
                <div v-else class="w-full h-full flex items-center justify-center text-primary font-semibold text-lg">{{ myIdentity?.displayName[0] || 'U' }}</div>
              </div>
              <div class="flex-1 min-w-0">
                <h3 class="text-sm font-semibold text-text truncate">{{ myIdentity?.displayName || 'User' }}</h3>
                <p class="text-xs text-text-muted truncate">{{ myIdentity?.peerId ? myIdentity.peerId.slice(-8) : '' }}</p>
              </div>
            </div>
          </div>

          <!-- Search -->
          <div class="px-4 py-3 border-b border-border">
            <div class="relative">
              <svg class="w-4 h-4 absolute left-3 top-3 text-text-muted" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
              <input type="text" placeholder="Search" v-model="searchQuery" class="w-full bg-input pl-10 pr-4 py-2 rounded-lg text-sm text-text border border-border focus:border-primary focus:ring-1 focus:ring-primary outline-none transition-colors" />
            </div>
          </div>

          <!-- Chats List -->
          <div class="overflow-y-auto pr-1 min-h-0">
            <div v-if="filteredChats.length === 0 && !loading" class="px-4 py-12 text-center">
              <svg class="w-16 h-16 mx-auto text-text-muted mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
              </svg>
              <p class="text-sm text-text-muted">No conversations yet</p>
              <p class="text-xs text-text-muted mt-1">Connect to a peer to start messaging</p>
            </div>

            <div v-for="chat in filteredChats" :key="chat.channelId"
                 @click="selectChat(chat)"
                 class="flex items-center space-x-3 px-4 py-3 cursor-pointer group hover:bg-hover rounded-lg transition-colors duration-150 mb-0.5"
                 :class="{ 'bg-active border-r-2 border-primary': selectedChat?.channelId === chat.channelId }">
              <!-- Avatar -->
              <div class="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center overflow-hidden flex-shrink-0">
                <img v-if="chat?.avatarUrl" :src="chat.avatarUrl" class="w-full h-full object-cover" />
                <div v-else class="w-full h-full flex items-center justify-center text-primary font-semibold text-sm">{{ chat?.displayName[0] || 'U' }}</div>
              </div>

              <!-- Info -->
              <div class="flex-1 min-w-0">
                <div class="flex items-center space-x-2">
                  <h4 class="text-sm font-semibold text-text truncate">{{ chat?.displayName || 'Unknown' }}</h4>
                  <span v-if="chat?.unreadCount > 0" class="w-2.5 h-2.5 rounded-full bg-danger flex-shrink-0"></span>
                </div>
                <p class="text-xs text-text-muted truncate line-clamp-1" :class="{ 'font-semibold text-text': chat?.unreadCount > 0 }">{{ chat?.lastMessage || 'No messages yet' }}</p>
              </div>

              <!-- Time -->
              <div class="text-[10px] text-text-muted flex-shrink-0 whitespace-nowrap">{{ chat.lastMessageTime || '' }}</div>
            </div>
          </div>
        </div>

        <!-- New Chat Button -->
        <div class="px-4 py-3 border-t border-border shrink-0">
          <button @click="startNewChat" class="w-full flex items-center justify-center space-x-2 px-3 py-2 bg-hover rounded-lg text-sm font-medium text-primary transition-colors hover:bg-active">
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
            </svg>
            <span>New chat</span>
          </button>
        </div>
      </aside>

      <!-- Center: Chat View -->
      <section class="flex-1 flex flex-col bg-app-bg min-w-0 min-h-0">
        <!-- Empty State -->
        <div v-if="!selectedChat && !loading" class="flex-1 flex flex-col items-center justify-center text-center px-8 min-h-0">
          <div class="w-24 h-24 rounded-full bg-primary/5 flex items-center justify-center mb-6">
            <svg class="w-12 h-12 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
            </svg>
          </div>
          <h3 class="text-lg font-semibold text-text mb-2">Select a conversation</h3>
          <p class="text-text-muted text-sm">Choose a chat from the sidebar to start messaging</p>
        </div>

        <div v-else-if="loading && !selectedChat" class="flex-1 flex items-center justify-center min-h-0">
          <div class="text-text-muted text-sm">Loading...</div>
        </div>

        <div v-else class="flex flex-col h-full min-h-0">
          <!-- Chat Header -->
          <header class="h-16 bg-surface/80 backdrop-blur-md border-b border-border flex items-center justify-between px-4 shrink-0 z-10 flex-shrink-0">
            <div class="flex items-center space-x-3 cursor-pointer" @click="goBack">
              <svg class="w-5 h-5 text-text-muted" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
              </svg>
              <div class="w-10 h-10 rounded-full bg-primary-light flex items-center justify-center overflow-hidden flex-shrink-0">
                <img v-if="selectedChat?.avatarUrl" :src="selectedChat.avatarUrl" class="w-full h-full object-cover" />
                <div v-else class="w-full h-full flex items-center justify-center text-primary font-semibold text-lg">{{ selectedChat?.displayName[0] || 'U' }}</div>
              </div>
              <div class="flex-1 min-w-0">
                <h4 class="text-sm font-semibold text-text truncate">{{ selectedChat?.displayName || 'Unknown' }}</h4>
                <p class="text-xs text-text-muted" :class="{ 'text-success': isChatOnline, 'text-text-muted': !isChatOnline }">
                  {{ isChatOnline ? 'Online' : 'Offline' }}
                </p>
              </div>
            </div>
            <div class="flex items-center space-x-2 shrink-0">
              <button class="w-8 h-8 rounded-full hover:bg-hover flex items-center justify-center text-text-muted transition-colors" title="Voice call">
                <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14.572V3.571L14.448 7.04A1 1 0 0113 7.914v6.086a1 1 0 01-1.447.894L5 10.571V3.571z" />
                </svg>
              </button>
              <button class="w-8 h-8 rounded-full hover:bg-hover flex items-center justify-center text-text-muted transition-colors" title="Video call">
                <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14.572V3.571L14.448 7.04A1 1 0 0113 7.914v6.086a1 1 0 01-1.447.894L5 10.571V3.571z" />
                </svg>
              </button>
              <button class="w-8 h-8 rounded-full hover:bg-hover flex items-center justify-center text-text-muted transition-colors" title="Chat info">
                <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </button>
            </div>
          </header>

          <!-- Messages -->
          <div class="flex-1 overflow-y-auto p-4 space-y-4 chat-scroll-area min-h-0" id="chat-scroll-area">
            <div v-for="(group, index) in groupedMessages" :key="index" class="flex flex-col space-y-2 animate-fade-in">
              <div v-if="group.date" class="flex items-center space-x-2 my-1">
                <span class="h-[1px] bg-border flex-1"></span>
                <span class="text-[10px] text-text-muted px-2 bg-app-bg uppercase tracking-wider font-semibold">{{ group.date }}</span>
                <span class="h-[1px] bg-border flex-1"></span>
              </div>

              <message-item
                v-for="msg in group.messages"
                :key="msg.messageId"
                :message="msg"
                :is-self="msg.sender_peer_id === myPeerId"
              />
            </div>

            <div v-if="messages.length === 0 && !loadingMessages" class="flex flex-col items-center justify-center min-h-[200px] text-center py-20 opacity-60">
              <p class="text-text-muted text-sm">No messages yet</p>
              <p class="text-[11px] text-text-muted mt-1">Say hello!</p>
            </div>
          </div>

          <!-- Input Area -->
          <div class="bg-surface/80 backdrop-blur-md border-t border-border px-4 py-3 shrink-0 z-10 flex-shrink-0 min-h-[80px]">
            <div class="flex items-end space-x-2 max-w-4xl mx-auto">
              <button class="w-9 h-9 rounded-lg hover:bg-hover flex items-center justify-center text-text-muted transition-colors" title="Emoji">
                <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2 10h20M4.931 4.931a1 1 0 000 1.414l13.23 13.23a1 1 0 001.414 0l1.414-1.414a1 1 0 000-1.414l-13.23-13.23a1 1 0 00-1.414 0zM21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </button>
              <button class="w-9 h-9 rounded-lg hover:bg-hover flex items-center justify-center text-text-muted transition-colors" title="Attach file">
                <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
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
                :disabled="!messageText.trim() || !selectedChat || isSending"
                class="w-9 h-9 rounded-lg bg-primary flex items-center justify-center text-white hover:bg-primary-hover disabled:bg-surface disabled:text-text-muted transition-all duration-200" title="Send"
                @click="sendMessage"
              >
                <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
                </svg>
              </button>
            </div>
          </div>
        </div>
      </section>
    </main>

    <!-- Settings Modal -->
    <div v-if="showSettings" class="fixed inset-0 z-50 flex items-center justify-center">
      <div class="absolute inset-0 bg-black/50 backdrop-blur-sm transition-opacity" @click="showSettings = false"></div>
      <div class="bg-surface w-full max-w-md rounded-xl border border-border shadow-xl mx-4 p-6 animate-slide-up flex-shrink-0">
        <h3 class="text-lg font-semibold text-text mb-6">Settings</h3>
        <div class="space-y-5">
          <div>
            <label class="block text-xs font-semibold text-text-muted uppercase tracking-wider mb-2">Display name</label>
            <input type="text" v-model="newDisplayName" class="w-full bg-input border border-border rounded-lg px-3 py-2 text-sm text-text focus:border-primary focus:ring-1 focus:ring-primary outline-none" />
          </div>
          <div class="pt-2 flex justify-end space-x-2">
            <button @click="showSettings = false" class="px-4 py-2 text-sm text-text-muted hover:text-text transition-colors">Cancel</button>
            <button @click="saveDisplayName" class="px-4 py-2 text-sm font-semibold text-primary hover:text-primary-dark transition-colors">Save</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import MessageItem from './components/MessageItem.vue'

export default {
  name: 'App',
  components: {
    MessageItem
  },
  data() {
    return {
      myIdentity: {
        displayName: 'User',
        peerId: '',
        avatarUrl: ''
      },
      myPeerId: '',
      chats: [],
      selectedChat: null,
      messages: [],
      messageText: '',
      isConnected: false,
      isSending: false,
      loading: true,
      loadingMessages: false,
      searchQuery: '',
      showSettings: false,
      newDisplayName: ''
    }
  },
  computed: {
    filteredChats() {
      const query = this.searchQuery.toLowerCase().trim()
      if (!query) return this.chats
      return this.chats.filter(c =>
        c.displayName.toLowerCase().includes(query) ||
        (c.lastMessage && c.lastMessage.toLowerCase().includes(query))
      )
    },
    groupedMessages() {
      const groups = []
      let currentGroup = null
      let currentDate = null

      for (const msg of this.messages) {
        const msgDate = this.formatDateOnly(msg.timestamp)
        if (msgDate !== currentDate) {
          if (currentGroup) groups.push(currentGroup)
          currentDate = msgDate
          currentGroup = { date: currentDate, messages: [] }
        }
        currentGroup.messages.push(msg)
      }
      if (currentGroup) groups.push(currentGroup)
      return groups
    },
    isChatOnline() {
      if (!this.selectedChat) return false
      return this.selectedChat.status === 'online'
    }
  },
  async created() {
    try {
      const identity = await window.gochatgui.Identity()
      this.myIdentity = {
        displayName: identity.display_name || 'User',
        peerId: identity.peer_id || '',
        avatarUrl: ''
      }
      this.myPeerId = identity.peer_id || ''
      this.newDisplayName = this.myIdentity.displayName

      await this.loadChats()
      this.setupEventListeners()
    } catch (error) {
      console.error('Failed to initialize:', error)
    } finally {
      this.loading = false
    }
  },
  methods: {
    async loadChats() {
      try {
        const [channelsResponse, dmsResponse, peersResponse] = await Promise.all([
          window.gochatgui.Channels(),
          window.gochatgui.DMChannels(),
          window.gochatgui.Peers()
        ])

        const peerMap = {}
        for (const p of (peersResponse || [])) {
          peerMap[p.peer_id] = p
        }

        const chats = []
        for (const ch of (channelsResponse || [])) {
          const peer = peerMap[ch.channel_id] || {}
          chats.push({
            channelId: ch.channel_id,
            displayName: ch.name || 'Channel',
            channelType: ch.channel_type,
            unreadCount: ch.unread_count || 0,
            avatarUrl: '',
            lastMessage: '',
            lastMessageTime: '',
            status: peer.status || 'offline'
          })
        }
        for (const dm of (dmsResponse || [])) {
          const peer = peerMap[dm.channel_id] || {}
          chats.push({
            channelId: dm.channel_id,
            displayName: dm.name || 'Direct Message',
            channelType: dm.channel_type,
            unreadCount: dm.unread_count || 0,
            avatarUrl: '',
            lastMessage: '',
            lastMessageTime: '',
            status: peer.status || 'offline'
          })
        }

        chats.sort((a, b) => {
          if (a.unreadCount > 0 && b.unreadCount === 0) return -1
          if (a.unreadCount === 0 && b.unreadCount > 0) return 1
          return 0
        })

        this.chats = chats
        await this.updateConnectionStatus()
      } catch (error) {
        console.error('Failed to load chats:', error)
      } finally {
        this.loading = false
      }
    },
    async loadMessages(channelId) {
      this.loadingMessages = true
      try {
        const response = await window.gochatgui.Messages({
          channel_id: channelId,
          limit: 100,
          offset: 0
        })
        this.messages = (response || []).reverse()
        this.scrollToBottom()
      } catch (error) {
        console.error('Failed to load messages:', error)
        this.messages = []
      } finally {
        this.loadingMessages = false
      }
    },
    setupEventListeners() {
      window.gochatgui.on('new-message', (data) => {
        if (this.selectedChat && data.channel_id === this.selectedChat.channelId) {
          this.messages.push(data)
          this.scrollToBottom()
          this.updateChatListItem(data)
        } else {
          const chat = this.chats.find(c => c.channelId === data.channel_id)
          if (chat) {
            chat.unreadCount = (chat.unreadCount || 0) + 1
            chat.lastMessage = data.content
            chat.lastMessageTime = this.formatTime(data.timestamp)
          }
        }
      })

      window.gochatgui.on('connection-changed', (data) => {
        this.isConnected = data.connected
      })
    },
    async updateConnectionStatus() {
      try {
        const diag = await window.gochatgui.ConnectedPeers()
        this.isConnected = diag.is_connected
      } catch (e) {
        console.error('Failed to get connection status:', e)
      }
    },
    selectChat(chat) {
      this.selectedChat = chat
      this.messages = []
      this.loadMessages(chat.channelId)
      chat.unreadCount = 0
      this.scrollToBottom()
    },
    goBack() {
      this.selectedChat = null
      this.messages = []
    },
    async sendMessage() {
      if (!this.selectedChat || !this.messageText.trim() || this.isSending) return

      const content = this.messageText.trim()
      this.isSending = true
      this.messageText = ''

      try {
        await window.gochatgui.SendMessage(
          this.selectedChat.channelId,
          content,
          'text'
        )
      } catch (error) {
        console.error('Failed to send message:', error)
      } finally {
        this.isSending = false
        const textarea = document.getElementById('message-input')
        if (textarea) textarea.focus()
      }
    },
    updateChatListItem(data) {
      const chat = this.chats.find(c => c.channelId === data.channel_id)
      if (chat) {
        chat.lastMessage = data.content
        chat.lastMessageTime = this.formatTime(data.timestamp)
        const index = this.chats.indexOf(chat)
        if (index > 0) {
          this.chats.splice(index, 1)
          this.chats.unshift(chat)
        }
      }
    },
    startNewChat() {
      // Placeholder
    },
    openProfile() {
      // Placeholder
    },
    openSettings() {
      this.showSettings = true
    },
    saveDisplayName() {
      if (!this.newDisplayName.trim()) return
      window.gochatgui.SetDisplayName({ name: this.newDisplayName.trim() })
        .then(() => {
          this.myIdentity.displayName = this.newDisplayName.trim()
          this.showSettings = false
        })
        .catch(err => console.error('Failed to save display name:', err))
    },
    scrollToBottom() {
      this.$nextTick(() => {
        const container = document.getElementById('chat-scroll-area')
        if (container) {
          container.scrollTop = container.scrollHeight
        }
      })
    },
    formatTime(timestamp) {
      if (!timestamp) return ''
      const date = new Date(timestamp)
      return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    },
    formatDateOnly(timestamp) {
      if (!timestamp) return ''
      const date = new Date(timestamp)
      const now = new Date()
      if (date.toDateString() === now.toDateString()) return 'Today'
      const yesterday = new Date(now)
      yesterday.setDate(yesterday.getDate() - 1)
      if (date.toDateString() === yesterday.toDateString()) return 'Yesterday'
      return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
    }
  },
  updated() {
    if (this.selectedChat) {
      this.scrollToBottom()
    }
  }
}
</script>

<style scoped>
.app {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
}
</style>
