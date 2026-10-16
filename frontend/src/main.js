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

// Load Wails runtime and bindings dynamically
async function loadWailsBindings() {
  try {
    await import('/wailsjs/runtime/runtime.js')
    await import('/wailsjs/go/gui/Service.js')
  } catch (err) {
    console.error('Failed to load Wails bindings:', err)
  }
}

async function init() {
  await loadWailsBindings()

  // Wait for DOM to be ready before mounting Vue
  if (document.readyState === 'loading') {
    await new Promise((resolve) => {
      document.addEventListener('DOMContentLoaded', resolve, { once: true })
    })
  }

  // Set up window.gochatgui alias from Wails runtime bindings
  const setupBindings = () => {
    if (window.go?.gui?.Service) {
      const svc = window.go.gui.Service
      window.gochatgui = {
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
      return true
    }
    return false
  }

  const bindingsReady = setupBindings()

  if (!bindingsReady) {
    const el = document.getElementById('app')
    if (el) {
      el.innerHTML = `<div style="padding:20px;color:#f1f5f9;background:#0f172a;font-family:sans-serif">
        <h3 style="color:#ef4444">Wails bindings not loaded</h3>
        <p style="color:#94a3b8">window.gochatgui could not be initialized from Wails runtime.</p>
        <p style="color:#94a3b8;font-size:12px">Make sure wailsjs runtime and Service scripts are served correctly.</p>
      </div>`
    }
  } else {
    app.mount('#app')
  }
}

init()
