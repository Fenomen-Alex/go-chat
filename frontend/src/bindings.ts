type WailsService = {
  AcceptInvite: (arg: { invite_id: string }) => Promise<{ channel_id: string }>
  Channels: () => Promise<any[]>
  Close: () => Promise<void>
  Connect: (arg: { addr: string }) => Promise<void>
  ConnectedPeers: () => Promise<DiagnosticsView>
  CreateChannel: (arg: { name: string; channel_type: string; description: string }) => Promise<void>
  CreateOrg: (arg: { name: string; description: string }) => Promise<void>
  DMChannels: () => Promise<any[]>
  DisconnectAll: () => Promise<void>
  Identity: () => Promise<IdentityView>
  InviteToChannel: (arg: { channel_id: string; peer_id: string }) => Promise<void>
  ListConnections: () => Promise<any[]>
  ListPendingInvites: () => Promise<any[]>
  Messages: (input: { channel_id: string; limit: number; offset: number }) => Promise<MessageView[]>
  OpenDM: (peerID: string) => Promise<{ channel_id: string }>
  Organizations: () => Promise<OrganizationView[]>
  Peers: () => Promise<PeerView[]>
  SaveConnection: (input: { addr: string; nickname?: string }) => Promise<void>
  SendMessage: (channelID: string, content: string, contentType: string) => Promise<void>
  SetDisplayName: (input: { name: string }) => Promise<void>
}

type WailsRuntime = {
  EventsOn: (event: string, callback: (data: any) => void) => void
  EventsOff: (event: string, ...additionalEventNames: string[]) => void
  EventsEmit: (event: string, data?: any) => void
  LogInfo: (message: string) => void
  LogError: (message: string) => void
}

declare global {
  interface Window {
    go?: {
      gui?: {
        Service: WailsService
      }
      runtime?: WailsRuntime
    }
    gochatgui?: WailsService & {
      on: (event: string, callback: (data: any) => void) => void
    }
  }
}

import type { DiagnosticsView, IdentityView, MessageView, OrganizationView, PeerView } from './types'

export function loadWailsBindings(): Promise<WailsService & { on: (event: string, callback: (data: any) => void) => void }> {
  return new Promise((resolve, reject) => {
    let attempts = 0
    const maxAttempts = 50

    function tryLoad() {
      attempts++
      const go = window.go
      if (go?.gui?.Service) {
        const svc = go.gui.Service
        const bindings: WailsService & { on: (event: string, callback: (data: any) => void) => void } = {
          AcceptInvite: (arg) => svc.AcceptInvite(arg),
          Channels: () => svc.Channels(),
          Close: () => svc.Close(),
          Connect: (arg) => svc.Connect(arg),
          ConnectedPeers: () => svc.ConnectedPeers(),
          CreateChannel: (arg) => svc.CreateChannel(arg),
          CreateOrg: (arg) => svc.CreateOrg(arg),
          DMChannels: () => svc.DMChannels(),
          DisconnectAll: () => svc.DisconnectAll(),
          Identity: () => svc.Identity(),
          InviteToChannel: (arg) => svc.InviteToChannel(arg),
          ListConnections: () => svc.ListConnections(),
          ListPendingInvites: () => svc.ListPendingInvites(),
          Messages: (input) => svc.Messages(input),
          OpenDM: (peerID) => svc.OpenDM(peerID),
          Organizations: () => svc.Organizations(),
          Peers: () => svc.Peers(),
          SaveConnection: (input) => svc.SaveConnection(input),
          SendMessage: (channelID, content, contentType) => svc.SendMessage(channelID, content, contentType),
          SetDisplayName: (input) => svc.SetDisplayName(input),
          on: (event, callback) => {
            if (window.go?.runtime?.EventsOn) {
              window.go.runtime.EventsOn(event, callback)
            }
          },
        }
        window.gochatgui = bindings
        resolve(bindings)
      } else if (attempts < maxAttempts) {
        setTimeout(tryLoad, 100)
      } else {
        reject(new Error('Wails bindings not available after ' + maxAttempts + ' attempts'))
      }
    }

    tryLoad()
  })
}
