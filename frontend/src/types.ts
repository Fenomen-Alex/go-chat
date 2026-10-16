export interface IdentityView {
  display_name: string
  peer_id: string
  public_key: string
  avatar_color: string
  fingerprint: string
  created_at: string
  is_default: boolean
}

export interface PeerView {
  peer_id: string
  display_name: string
  status: string
  trusted: boolean
  blocked: boolean
  favorite: boolean
  avatar_color: string
  fingerprint: string
  first_seen: string
  last_seen: string
}

export interface ChannelView {
  channel_id: string
  name: string
  channel_type: string
  org_id: string
  description: string
  unread_count: number
  archived: boolean
}

export interface MessageView {
  message_id: string
  channel_id: string
  sender_peer_id: string
  sender_name: string
  content: string
  content_type: string
  timestamp: string
  is_self: boolean
  delivery_state: string
}

export interface DiagnosticsView {
  peer_id: string
  listening_addrs: string[]
  connected_peers: PeerView[]
  connected_count: number
  is_connected: boolean
}

export interface OrganizationView {
  org_id: string
  name: string
  description: string
  icon: string
  owner_peer_id: string
  visibility: string
  max_members: number
  history_ret: number
  attachment_limit: number
  archived: boolean
  created_at: string
  updated_at: string
}

export interface InviteView {
  invite_id: string
  sender_peer_id: string
  target_peer_id: string
  org_id: string
  channel_id: string
  invite_type: string
  message: string
  one_time: boolean
  max_uses: number
  use_count: number
  expires_at: string
  created_at: string
}

export interface ConnectionView {
  id: string
  addr: string
  nickname: string
  connected: boolean
  last_connected: string
}

export interface ChatView {
  channelId: string
  displayName: string
  channelType: string
  unreadCount: number
  avatarUrl: string
  lastMessage: string
  lastMessageTime: string
  status: string
}
