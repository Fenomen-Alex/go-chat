<template>
  <div class="min-h-screen flex flex-col bg-dark text-light">
    <header class="bg-dark/80 backdrop-blur-sm border-b border-border/50 px-4 py-3 flex items-center justify-between">
      <div class="flex items-center space-x-3">
        <div class="w-8 h-8 bg-primary/20 rounded-lg flex items-center justify-center">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2-2h-5l-5 5v-5z" />
          </svg>
        </div>
        <h1 class="text-lg font-semibold">go-chat</h1>
      </div>
      <div class="flex items-center space-x-2 text-sm">
        <span id="connection-status" class="px-2 py-1 bg-success/20 rounded-full text-success">● Online</span>
        <button id="settings-btn" class="px-3 py-1 bg-primary/20 rounded hover:bg-primary/30 transition-colors">
          Settings
        </button>
      </div>
    </header>

    <main class="flex-1 overflow-hidden flex">
      <!-- Sidebar -->
      <aside class="w-64 bg-dark/90 border-r border-border/50 flex flex-col overflow-hidden">
        <nav class="flex-1 overflow-y-auto p-4 space-y-2">
          <!-- Workspaces/Organizations -->
          <div class="space-y-1 mb-4">
            <h2 class="text-xs font-medium text-uppercase tracking-wider text-border/50">WORKSPACES</h2>
            <div v-for="org in organizations" :key="org.orgId" class="flex items-center space-x-2 px-2 py-1 rounded hover:bg-dark/50 cursor-pointer" @click="selectOrganization(org)">
              <div class="w-3 h-3 bg-info rounded-full flex-shrink-0"></div>
              <span class="text-sm truncate max-w-xs">{{ org.name }}</span>
            </div>
          </div>

          <!-- Channels -->
          <div class="space-y-1 mb-4">
            <h2 class="text-xs font-medium text-uppercase tracking-wider text-border/50">CHANNELS</h2>
            <div v-for="channel in channels" :key="channel.channelId" class="flex items-center space-x-2 px-2 py-1 rounded hover:bg-dark/50 cursor-pointer" @click="selectChannel(channel)">
              <template v-if="channel.type === 'private'">
                <div class="w-3 h-3 bg-secondary/20 rounded-full flex-shrink-0">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-secondary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7H3m8 4H3m-4 8h10a2 2 0 002-2v-2a2 2 0 00-2-2h-2M5 7a2 2 0 012-2h2a2 2 0 012 2v2.586a1 1 0 01-.293.707l-1.414 1.414A2 2 0 019.414 11H5a2 2 0 01-2-2V7z" />
                  </svg>
                </div>
              </template>
              <template v-else>
                <div class="w-3 h-3 bg-primary/20 rounded-full flex-shrink-0">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H9a2 2 0 00-2 2v16l2-2 2-2 2-2 2 2zm2-12a2 2 0 014 0v7a2 2 0 01-2 2h-.016a1.023 1.023 0 00-.49-.28l-.458-.458A6.002 6.002 0 006 16a6.002 6.002 0 00-4 5.831v.169a6.002 6.002 0 002.536 4.93l.581.581a1.023 1.023 0 00.49.28h.016a2 2 0 012 2V9z" />
                  </svg>
                </div>
              </template>
              <span class="text-sm truncate max-w-xs">{{ channel.name }}</span>
              <span v-if="channel.unread > 0" class="ml-auto w-3 h-3 bg-danger rounded-full flex-shrink-0 text-xs flex items-center justify-center text-white">{{ channel.unread }}</span>
            </div>
          </div>

          <!-- Direct Messages -->
          <div class="space-y-1">
            <h2 class="text-xs font-medium text-uppercase tracking-wider text-border/50">DIRECT MESSAGES</h2>
            <div v-for="dm in dms" :key="dm.channelId" class="flex items-center space-x-2 px-2 py-1 rounded hover:bg-dark/50 cursor-pointer" @click="selectDM(dm)">
              <div class="w-3 h-3 bg-info/20 rounded-full flex-shrink-0">
                <img :src="dm.avatar" alt="Avatar" class="w-full h-full object-cover rounded-full" />
              </div>
              <span class="text-sm truncate max-w-xs">{{ dm.name }}</span>
              <span v-if="dm.unread > 0" class="ml-auto w-3 h-3 bg-danger rounded-full flex-shrink-0 text-xs flex items-center justify-center text-white">{{ dm.unread }}</span>
            </div>
          </div>
        </nav>

        <!-- Bottom controls -->
        <div class="px-4 py-3 border-t border-border/50">
          <button id="new-channel-btn" class="w-full flex items-center space-x-2 px-3 py-2 bg-primary/20 rounded hover:bg-primary/30 transition-colors text-sm">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
            </svg>
            New Channel
          </button>
        </div>
      </aside>

      <!-- Main content area -->
      <section class="flex-1 flex flex-col overflow-hidden border-l border-border/50">
        <!-- Header -->
        <header class="bg-dark/80 backdrop-blur-sm border-b border-border/50 px-4 py-3 flex items-center justify-between">
          <div class="flex items-center space-x-3">
            <div class="w-10 h-10 bg-primary/20 rounded-lg flex items-center justify-center">
              <svg v-if="selectedChannel && selectedChannel.type === 'private'" xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-secondary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 21V8a2 2 0 00-2-2h-1a2 2 0 00-1.414.586l-2 2V8a2 2 0 00-2-2h-1V8a2 2 0 00-2-2h-1V8a2 2 0 00-2-2h-.586l-.293-.293A2 2 0 009.172 4H5a2 2 0 00-2 2v13a2 2 0 002 2h5l2 2 2-2h2a2 2 0 002-2v-2a2 2 0 012 2h1a2 2 0 012-2v-1a1 1 0 00-1-1H5a1 1 0 00-1 1v5a1 1 0 001 1h7a1 1 0 001-1V8a1 1 0 011-1h4a1 1 0 001 1v1a1 1 0 001 1h4" />
              </svg>
              <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c0 1.105-.895 2-2 2s-2-.895-2-2 .895-2 2-2 2 .895 2 2zm0 12c0 1.105-.895 2-2 2s-2-.895-2-2 .895-2 2-2 2 .895 2 2zm0-6c0 1.105-.895 2-2 2s-2-.895-2-2 .895-2 2-2 2 .895 2 2z" />
              </svg>
            </div>
            <h2 class="text-lg font-semibold truncate max-w-xl" id="current-channel-title">
              {{ selectedChannel ? selectedChannel.name : 'Select a channel' }}
            </h2>
          </div>
          <div class="flex items-center space-x-2 text-sm">
            <button id="channel-info-btn" class="px-3 py-1 bg-primary/20 rounded hover:bg-primary/30 transition-colors">
              Info
            </button>
          </div>
        </header>

        <!-- Chat messages -->
        <section class="flex-1 overflow-y-auto p-4 space-y-3" id="chat-messages">
          <div v-if="!selectedChannel" class="text-center py-12">
            <p class="text-border/50">Select a channel to start chatting</p>
          </div>
          <div v-else class="flex flex-col space-y-2">
            <message-item
              v-for="message in messages"
              :key="message.id"
              :message="message"
              :is-self="message.senderPeerId === myPeerId"
            />
          </div>
        </section>

        <!-- Message input -->
        <form @submit.prevent="sendMessage" class="bg-dark/80 backdrop-blur-sm border-t border-border/50 px-4 py-3 flex items-center space-x-2">
          <div class="flex-1 relative">
            <textarea
              id="message-input"
              rows="1"
              maxlength="2000"
              placeholder="Type a message or /help..."
              class="w-full bg-input-bg/50 border border-border/50 rounded px-3 py-2 text-sm resize-none focus:outline-none focus:border-primary/50 focus:bg-input-bg/70"
              v-model="messageText"
              @keydown.shift.enter="onShiftEnter"
            ></textarea>
            <div class="absolute inset-y-0 right-2 flex items-center pointer-events-none text-border/50">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7H3m8 4H3m-4 8h10a2 2 0 002-2v-2a2 2 0 00-2-2h-2M5 7a2 2 0 012-2h2a2 2 0 012 2v2.586a1 1 0 01-.293.707l-1.414 1.414A2 2 0 019.414 11H5a2 2 0 01-2-2V7z" />
              </svg>
            </div>
          </div>
          <button type="submit" class="px-4 py-2 bg-primary rounded hover:bg-primary/80 transition-colors text-sm font-medium flex items-center space-x-1">
            Send
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 3l9 9 9-9" />
            </svg>
          </button>
        </form>
      </section>
    </main>
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
      // State
      organizations: [],
      channels: [],
      dms: [],
      messages: [],
      selectedChannel: null,
      selectedOrganization: null,
      messageText: '',
      myPeerId: '',
      isConnected: false,
      
      // UI state
      showSettings: false,
      showChannelInfo: false,
    }
  },
  async created() {
    // Initialize connection to backend
    try {
      const identity = await window.gochatgui.Identity()
      this.myPeerId = identity.peerId
      
      // Load initial data
      await this.loadData()
      
      // Set up event listeners for updates
      this.setupEventListeners()
    } catch (error) {
      console.error('Failed to initialize:', error)
    }
  },
  methods: {
    async loadData() {
      try {
        // Load organizations
        const orgsResponse = await window.gochatgui.Organizations()
        this.organizations = orgsResponse
        
        // Load channels
        const channelsResponse = await window.gochatgui.Channels()
        this.channels = channelsResponse
        
        // Load DMs
        const dmsResponse = await window.gochatgui.DMChannels()
        this.dms = dmsResponse
        
        // Load connection status
        const diagResponse = await window.gochatgui.ConnectedPeers()
        this.isConnected = diagResponse.isConnected
        
        // Update connection status UI
        this.updateConnectionStatus()
        
        // If we have a selected channel, load its messages
        if (this.selectedChannel) {
          await this.loadMessages()
        }
      } catch (error) {
        console.error('Failed to load data:', error)
      }
    },
    
    async loadMessages() {
      if (!this.selectedChannel) return
      
      try {
        const messagesResponse = await window.gochatgui.Messages({
          channelId: this.selectedChannel.channelId,
          limit: 50,
          offset: 0
        })
        this.messages = messagesResponse
        
        // Scroll to bottom
        this.$nextTick(() => {
          const messagesContainer = document.getElementById('chat-messages')
          if (messagesContainer) {
            messagesContainer.scrollTop = messagesContainer.scrollHeight
          }
        })
      } catch (error) {
        console.error('Failed to load messages:', error)
      }
    },
    
    setupEventListeners() {
      // Listen for refresh events from backend
      window.gochatgui.on('refresh', () => {
        this.loadData()
      })
      
      // Listen for new messages
      window.gochatgui.on('new-message', (data) => {
        // Add new message to current view if it's for the selected channel
        if (this.selectedChannel && data.channelId === this.selectedChannel.channelId) {
          this.messages.push(data)
          
          // Scroll to bottom
          this.$nextTick(() => {
            const messagesContainer = document.getElementById('chat-messages')
            if (messagesContainer) {
              messagesContainer.scrollTop = messagesContainer.scrollHeight
            }
          })
        }
      })
      
      // Listen for connection status changes
      window.gochatgui.on('connection-changed', (data) => {
        this.isConnected = data.connected
        this.updateConnectionStatus()
      })
    },
    
    updateConnectionStatus() {
      const statusEl = document.getElementById('connection-status')
      if (statusEl) {
        if (this.isConnected) {
          statusEl.textContent = '● Online'
          statusEl.className = 'px-2 py-1 bg-success/20 rounded-full text-success'
        } else {
          statusEl.textContent = '● Offline'
          statusEl.className = 'px-2 py-1 bg-danger/20 rounded-full text-danger'
        }
      }
    },
    
    selectOrganization(org) {
      this.selectedOrganization = org
      // TODO: Filter channels by organization
    },
    
    selectChannel(channel) {
      this.selectedChannel = channel
      this.selectedOrganization = null
      this.loadMessages()
    },
    
    selectDM(dm) {
      this.selectedChannel = dm
      this.selectedOrganization = null
      this.loadMessages()
    },
    
    async sendMessage() {
      if (!this.selectedChannel || !this.messageText.trim()) return
      
      try {
        await window.gochatgui.SendMessage(
          this.selectedChannel.channelId,
          this.messageText,
          'text'
        )
        
        // Clear input
        this.messageText = ''
        
        // Focus textarea
        this.$nextTick(() => {
          const textarea = document.getElementById('message-input')
          if (textarea) textarea.focus()
        })
      } catch (error) {
        console.error('Failed to send message:', error)
        // TODO: Show error notification
      }
    },
    
    onShiftEnter(e) {
      // Allow shift+enter for newlines
      if (e.shiftKey) {
        // Allow default behavior (newline)
        return
      }
      // Prevent enter from submitting if we don't want to send
      e.preventDefault()
      this.sendMessage()
    }
  }
}
</script>

<style scoped>
/* Message item component styles */
.message-item {
  max-width: 80%;
}

.message-item.self {
  align-self: flex-end;
  background-color: rgba(88, 101, 242, 0.2);
}

.message-item.other {
  align-self: flex-start;
  background-color: rgba(255, 255, 255, 0.05);
}

.message-item-content {
  display: flex;
  margin-bottom: 4px;
}

.message-item-avatar {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  margin-right: 8px;
  flex-shrink: 0;
}

.message-item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2px;
}

.message-item-author {
  font-weight: 600;
}

.message-item-timestamp {
  font-size: 0.75rem;
  opacity: 0.7;
}

.message-item-text {
  margin-bottom: 4px;
  word-wrap: break-word;
}

.message-item-status {
  font-size: 0.75rem;
  opacity: 0.6;
  margin-top: 2px;
}

.message-item-status.sent {
  color: #57F287;
}

.message-item-status.received {
  color: #57F287;
}

.message-item-status.sending {
  color: #FFA657;
}

.message-item-status.failed {
  color: #ED4245;
}
</style>