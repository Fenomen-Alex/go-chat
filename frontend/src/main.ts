import { createApp } from 'vue'
import App from './App.vue'
import './index.css'

const app = createApp(App)

// Global error handler so blank screens show errors
app.config.errorHandler = (err, vm, info) => {
  console.error('Vue error:', err)
  console.error('Info:', info)
  const el = document.getElementById('app')
  if (el) {
    el.innerHTML = `<div style="padding:20px;color:#f1f5f9;background:#0f172a;font-family:sans-serif">
      <h3 style="color:#ef4444">Vue Error</h3>
      <pre style="color:#94a3b8;white-space:pre-wrap">${String(err).replace(/</g,'&lt;')}</pre>
      <p>Info: ${info}</p>
    </div>`
  }
}

app.config.warnHandler = (msg, vm, trace) => {
  console.warn('Vue warning:', msg, trace)
}

// Load Wails bindings then mount Vue
async function init() {
  try {
    // Dynamically load Wails runtime and bindings
    await import('/wailsjs/runtime/runtime.js')
    await import('/wailsjs/go/gui/Service.js')

    // Wait for DOM ready
    if (document.readyState === 'loading') {
      await new Promise((resolve) => {
        document.addEventListener('DOMContentLoaded', resolve, { once: true })
      })
    }

    // Set up window.gochatgui alias from Wails runtime bindings
    const go = (window as any).go
    if (go?.gui?.Service) {
      const svc = go.gui.Service
      ;(window as any).gochatgui = {
        AcceptInvite: (arg: any) => svc.AcceptInvite(arg),
        Channels: () => svc.Channels(),
        Close: () => svc.Close(),
        Connect: (arg: any) => svc.Connect(arg),
        ConnectedPeers: () => svc.ConnectedPeers(),
        CreateChannel: (arg: any) => svc.CreateChannel(arg),
        CreateOrg: (arg: any) => svc.CreateOrg(arg),
        DMChannels: () => svc.DMChannels(),
        DisconnectAll: () => svc.DisconnectAll(),
        Identity: () => svc.Identity(),
        InviteToChannel: (arg: any) => svc.InviteToChannel(arg),
        ListConnections: () => svc.ListConnections(),
        ListPendingInvites: () => svc.ListPendingInvites(),
        Messages: (input: any) => svc.Messages(input),
        OpenDM: (peerID: string) => svc.OpenDM(peerID),
        Organizations: () => svc.Organizations(),
        Peers: () => svc.Peers(),
        SaveConnection: (input: any) => svc.SaveConnection(input),
        SendMessage: (channelID: string, content: string, contentType: string) =>
          svc.SendMessage(channelID, content, contentType),
        SetDisplayName: (input: any) => svc.SetDisplayName(input),
        on: (event: string, callback: (data: any) => void) => {
          if (window.go?.runtime?.EventsOn) {
            window.go.runtime.EventsOn(event, callback)
          }
        },
      }
      app.mount('#app')
    } else {
      const el = document.getElementById('app')
      if (el) {
        el.innerHTML = `<div style="padding:20px;color:#f1f5f9;background:#0f172a;font-family:sans-serif">
          <h3 style="color:#ef4444">Wails bindings not loaded</h3>
          <p style="color:#94a3b8">window.gochatgui could not be initialized from Wails runtime.</p>
        </div>`
      }
    }
  } catch (err) {
    console.error('Failed to initialize:', err)
    const el = document.getElementById('app')
    if (el) {
      el.innerHTML = `<div style="padding:20px;color:#f1f5f9;background:#0f172a;font-family:sans-serif">
        <h3 style="color:#ef4444">Initialization Error</h3>
        <pre style="color:#94a3b8;white-space:pre-wrap">${String(err)}</pre>
      </div>`
    }
  }
}

init()
