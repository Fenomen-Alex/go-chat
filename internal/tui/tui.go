package tui

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"go-chat/internal/app"
	"go-chat/internal/crypto"
	"go-chat/internal/storage"
	"go-chat/internal/tunnel"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MessageItem struct {
	Sender        string
	SenderPeerID  string
	Content       string
	Timestamp     string
	DeliveryState string
}

type Model struct {
	app            *app.App
	ready          bool
	width          int
	height         int
	channelList    []*storage.Channel
	selectedChan   int

	dmList       []*storage.Channel
	selectedDM   int

	chatView       viewport.Model
	messages       []MessageItem

	input          textinput.Model
	inputMode      bool

	statusText     string
	statusLog      []string

	peerList       []*storage.Peer
	logEntries     []string

	showHelp       bool
	showPeers      bool
	showLogs       bool

	dmFocused      bool

	pendingConnect string
	needsName      bool
	namePromptErr  string

	loading    bool
	loadingMsg string
	unread     map[string]int
	lastMsgCnt map[string]int
}

func NewModel(a *app.App) *Model {
	ti := textinput.New()
	ti.Placeholder = "Type a message or /help..."
	ti.Focus()
	ti.CharLimit = 2000
	ti.Width = 60

	a.Logger.SetConsoleOutput(false)

	m := &Model{
		app:        a,
		input:      ti,
		inputMode:  true,
		needsName:  a.IsDefaultName(),
		statusText: fmt.Sprintf("PeerID: %s | /myaddr to see shareable address", a.PeerID()),
		unread:     make(map[string]int),
		lastMsgCnt: make(map[string]int),
	}

	return m
}

type refreshMsg struct{}
type statusMsg string
type firstLaunchMsg struct{}
type loadingTickMsg time.Time

func (m *Model) Init() tea.Cmd {
	m.loadChannels()
	m.loadDMs()
	m.loadPeers()
	m.loadLogs()
	return tea.Batch(textinput.Blink, m.waitForEvent(), m.firstLaunchCmd())
}

func (m *Model) firstLaunchCmd() tea.Cmd {
	return func() tea.Msg {
		return firstLaunchMsg{}
	}
}

func (m *Model) waitForEvent() tea.Cmd {
	return func() tea.Msg {
		<-m.app.RefreshCh
		return refreshMsg{}
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case refreshMsg:
		m.updateUnread()
		m.loadChannels()
		m.loadDMs()
		m.loadPeers()
		m.loadLogs()
		return m, m.waitForEvent()

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		leftPanelWidth := 68 // 2 panels x (width(32) + rounded border(2))
		inputHeight := 3
		statusHeight := 1
		chatHeight := m.height - inputHeight - statusHeight - 4

		chatW := m.width - leftPanelWidth - 8
		if chatW < 20 {
			chatW = 20
		}

		if m.chatView.Height == 0 {
			m.chatView = viewport.New(chatW, chatHeight)
		} else {
			m.chatView.Width = chatW
			m.chatView.Height = chatHeight
		}

		m.input.Width = m.width - 56
		if m.needsName {
			m.input.Width = 42
		}

	case loadingTickMsg:
		if m.loading {
			elapsed := time.Since(time.Time(msg)).Truncate(time.Second)
			dots := elapsed.String()
			if elapsed > 15*time.Second {
				m.loadingMsg = "Still connecting... " + dots
			} else {
				m.loadingMsg = "Connecting... " + dots
			}
			return m, m.loadingTick()
		}
		return m, nil

	case statusMsg:
		m.loading = false
		s := string(msg)
		m.addStatus(s)
		if strings.HasPrefix(s, "Connected to ") {
			m.loadChannels()
			m.loadDMs()
			m.loadPeers()
		}

	case firstLaunchMsg:
		m.addStatus("Tab to navigate | ? for help | /channel create <name>")

	case tea.KeyMsg:
		if m.needsName {
			switch msg.String() {
			case "ctrl+c", "ctrl+q":
				return m, tea.Quit
			case "enter":
				return m, m.handleInput()
			default:
				m.namePromptErr = ""
				var cmd tea.Cmd
				m.input, cmd = m.input.Update(msg)
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}

		hotkeyHandled := false
		switch msg.String() {
		case "ctrl+c", "ctrl+q":
			return m, tea.Quit

		case "tab":
			hotkeyHandled = true
			if m.inputMode {
				m.inputMode = false
				m.dmFocused = false
				m.loadPeers()
			} else if !m.dmFocused {
				m.dmFocused = true
			} else {
				m.inputMode = true
				m.dmFocused = false
			}

		case "up":
			hotkeyHandled = true
			if m.inputMode {
				m.chatView.LineUp(1)
			} else if m.dmFocused {
				if m.selectedDM > 0 {
					m.selectedDM--
					m.loadDMMessages()
				}
			} else if m.selectedChan > 0 {
				m.selectedChan--
				m.loadMessages()
			}

		case "pgup":
			hotkeyHandled = true
			m.chatView.HalfViewUp()
		case "pgdown":
			hotkeyHandled = true
			m.chatView.HalfViewDown()

		case "down":
			hotkeyHandled = true
			if m.inputMode {
				m.chatView.LineDown(1)
			} else if m.dmFocused {
				if m.selectedDM < len(m.dmList)-1 {
					m.selectedDM++
					m.loadDMMessages()
				}
			} else if m.selectedChan < len(m.channelList)-1 {
				m.selectedChan++
				m.loadMessages()
			}

		case "enter":
			if m.inputMode {
				return m, m.handleInput()
			}
			hotkeyHandled = true
			m.inputMode = true
			m.loadPeers()

		case "?":
			hotkeyHandled = true
			m.showHelp = !m.showHelp
			if m.showHelp {
				m.showPeers = false
				m.showLogs = false
			}

		case "P":
			hotkeyHandled = true
			m.showPeers = !m.showPeers
			if m.showPeers {
				m.loadPeers()
				m.showHelp = false
				m.showLogs = false
			}

		case "L":
			hotkeyHandled = true
			m.showLogs = !m.showLogs
			if m.showLogs {
				m.loadLogs()
				m.showHelp = false
				m.showPeers = false
			}
		}

		if !hotkeyHandled && m.inputMode {
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			cmds = append(cmds, cmd)
		} else if !m.inputMode {
			m.chatView, _ = m.chatView.Update(msg)
		}

		return m, tea.Batch(cmds...)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) handleInput() tea.Cmd {
	text := strings.TrimSpace(m.input.Value())
	m.input.SetValue("")

	if text == "" {
		return nil
	}

	if m.needsName {
		if strings.HasPrefix(text, "/") {
			return m.handleCommand(text)
		}
		name := strings.TrimSpace(text)
		if m.app.IsReservedDisplayName(name) {
			m.namePromptErr = "Invalid or reserved name. Please choose another."
			return nil
		}
		m.needsName = false
		m.app.SetDisplayName(name)
		m.input.Placeholder = "Type a message or /help..."
		m.addStatus(fmt.Sprintf("Display name set to '%s'", name))
		if m.pendingConnect != "" {
			addr := m.pendingConnect
			m.pendingConnect = ""
			m.addStatus(fmt.Sprintf("Connecting to %s...", addr))
			m.loading = true
			m.loadingMsg = "Connecting..."
			appPtr := m.app
			return tea.Batch(func() tea.Msg {
				if err := appPtr.Connect(addr); err != nil {
					return statusMsg(fmt.Sprintf("Connect error: %v", err))
				}
				appPtr.SaveConnection(addr)
				return statusMsg(fmt.Sprintf("Connected to %s", addr))
			}, m.loadingTick())
		}
		return nil
	}

	if m.showHelp || m.showPeers || m.showLogs {
		m.showHelp = false
		m.showPeers = false
		m.showLogs = false
	}

	if strings.HasPrefix(text, "/") {
		return m.handleCommand(text)
	}

	if m.dmFocused {
		if len(m.dmList) == 0 {
			m.addStatus("No DM selected. Use /dm <peer_id> to start a conversation.")
			return nil
		}
		channelID := m.dmList[m.selectedDM].ChannelID
		if err := m.app.SendMessage(channelID, text, "text"); err != nil {
			m.addStatus(fmt.Sprintf("Error: %v", err))
			return nil
		}
		msg := MessageItem{
			Sender:        m.app.Identity().DisplayName,
			SenderPeerID:  m.app.PeerID(),
			Content:       text,
			Timestamp:     "now",
			DeliveryState: "sent",
		}
		m.messages = append(m.messages, msg)
		m.chatView.SetContent(m.renderMessages())
		m.chatView.GotoBottom()
		return nil
	}

	if len(m.channelList) == 0 {
		m.addStatus("No channel selected. Use Tab to navigate and select a channel.")
		return nil
	}

	channelID := m.channelList[m.selectedChan].ChannelID
	if err := m.app.SendMessage(channelID, text, "text"); err != nil {
		m.addStatus(fmt.Sprintf("Error: %v", err))
		return nil
	}

	msg := MessageItem{
		Sender:        m.app.Identity().DisplayName,
		SenderPeerID:  m.app.PeerID(),
		Content:       text,
		Timestamp:     "now",
		DeliveryState: "sent",
	}
	m.messages = append(m.messages, msg)
	m.chatView.SetContent(m.renderMessages())

	m.chatView.GotoBottom()

	return nil
}

func (m *Model) handleCommand(text string) tea.Cmd {
	parts := strings.Fields(text)
	cmd := strings.ToLower(parts[0])

	switch cmd {
	case "/help":
		m.showHelp = !m.showHelp

	case "/connect":
		if len(parts) < 2 {
			m.addStatus("Usage: /connect <multiaddr> | /connect <index>  (saved connections)")
			return nil
		}
		arg := parts[1]
		var idx int
		if _, err := fmt.Sscanf(arg, "%d", &idx); err == nil {
			conns, err := m.app.ListConnections()
			if err != nil {
				m.addStatus(fmt.Sprintf("Error: %v", err))
				return nil
			}
			if idx < 1 || idx > len(conns) {
				m.addStatus("Invalid connection index. Use /connections to list.")
				return nil
			}
			arg = conns[idx-1].Address
		}
		if m.app.IsDefaultName() {
			m.pendingConnect = arg
			m.addStatus(fmt.Sprintf("Set your display name to connect to %s", arg))
			m.input.SetValue("")
			m.input.Placeholder = "Enter your display name..."
			m.inputMode = true
			return nil
		}
		m.addStatus(fmt.Sprintf("Connecting to %s...", arg))
		m.loading = true
		m.loadingMsg = "Connecting..."
		appPtr := m.app
		connAddr := arg
		return tea.Batch(func() tea.Msg {
			if err := appPtr.Connect(connAddr); err != nil {
				return statusMsg(fmt.Sprintf("Connect error: %v", err))
			}
			appPtr.SaveConnection(connAddr)
			return statusMsg(fmt.Sprintf("Connected to %s", connAddr))
		}, m.loadingTick())

	case "/disconnect":
		m.app.DisconnectAll()
		m.addStatus("Disconnected")

	case "/peers":
		m.loadPeers()
		m.showPeers = !m.showPeers

	case "/channel":
		if len(parts) < 2 {
			m.addStatus("Usage: /channel create <name> | /channel private <name> [desc] | /channel list")
			return nil
		}
		switch parts[1] {
		case "create":
			if len(parts) < 3 {
				m.addStatus("Usage: /channel create <name>")
				return nil
			}
			name := strings.Join(parts[2:], " ")
			ch, err := m.app.CreateChannel(name, "text")
			if err != nil {
				m.addStatus(fmt.Sprintf("Error: %v", err))
				return nil
			}
			m.loadChannels()
			for i, c := range m.channelList {
				if c.ChannelID == ch.ChannelID {
					m.selectedChan = i
					break
				}
			}
			m.loadMessages()
			m.addStatus(fmt.Sprintf("Channel '%s' created", name))
		case "private":
			if len(parts) < 3 {
				m.addStatus("Usage: /channel private <name> [description]")
				return nil
			}
			name := parts[2]
			desc := ""
			if len(parts) > 3 {
				desc = strings.Join(parts[3:], " ")
			}
			ch, err := m.app.CreatePrivateChannel(name, desc)
			if err != nil {
				m.addStatus(fmt.Sprintf("Error: %v", err))
				return nil
			}
			m.loadChannels()
			for i, c := range m.channelList {
				if c.ChannelID == ch.ChannelID {
					m.selectedChan = i
					break
				}
			}
			m.loadMessages()
			m.addStatus(fmt.Sprintf("Private channel '%s' created. Use /invite <channel_id> <peer_id> to add members.", name))
		case "list":
			m.loadChannels()
		default:
			m.addStatus("Unknown channel command. Use: create, private, list")
		}

	case "/invite":
		if len(parts) < 2 {
			m.addStatus("Usage: /invite <channel_id> <peer_id> or /invite accept <invite_id>")
			return nil
		}
		if parts[1] == "accept" {
			if len(parts) < 3 {
				m.addStatus("Usage: /invite accept <invite_id>")
				return nil
			}
			inviteID := parts[2]
			channelID, err := m.app.AcceptInvite(inviteID)
			if err != nil {
				m.addStatus(fmt.Sprintf("Error: %v", err))
				return nil
			}
			m.loadChannels()
			for i, c := range m.channelList {
				if c.ChannelID == channelID {
					m.selectedChan = i
					break
				}
			}
			m.loadMessages()
			m.addStatus(fmt.Sprintf("Joined channel %s", channelID))
			return nil
		}
		if len(parts) < 3 {
			m.addStatus("Usage: /invite <channel_id> <peer_id>")
			return nil
		}
		channelID := parts[1]
		peerID := parts[2]
		inviteID, err := m.app.InviteToChannel(channelID, peerID)
		if err != nil {
			m.addStatus(fmt.Sprintf("Error: %v", err))
			return nil
		}
		m.addStatus(fmt.Sprintf("Invite sent (%s) to %s for channel %s", inviteID, peerID, channelID))

	case "/invites":
		invites, err := m.app.ListPendingInvites()
		if err != nil {
			m.addStatus(fmt.Sprintf("Error: %v", err))
			return nil
		}
		if len(invites) == 0 {
			m.addStatus("No pending invites.")
			return nil
		}
		for _, inv := range invites {
			m.addStatus(fmt.Sprintf("Invite %s from %s for channel %s",
				inv.InviteID, inv.SenderPeerID, inv.ChannelID))
		}
		m.addStatus("Use /invite accept <invite_id> to accept")

	case "/dm":
		if len(parts) < 2 {
			m.addStatus("Usage: /dm <peer_id>")
			return nil
		}
		peerID := parts[1]
		dmID, err := m.app.OpenDM(peerID)
		if err != nil {
			m.addStatus(fmt.Sprintf("Error: %v", err))
			return nil
		}
		m.addStatus(fmt.Sprintf("DM with %s", peerID))
		m.loadDMs()
		for i, ch := range m.dmList {
			if ch.ChannelID == dmID {
				m.selectedDM = i
				break
			}
		}
		m.dmFocused = true
		m.inputMode = false
		m.loadDMMessages()

	case "/myaddr":
		peerID := m.app.PeerID()
		allAddrs := m.app.AllAddrs()
		var lines []string
		lines = append(lines, fmt.Sprintf("=== Peer ID: %s ===", peerID))
		for _, addr := range allAddrs {
			lines = append(lines, fmt.Sprintf("  %s", addr))
		}
		lines = append(lines, "---")
		if len(allAddrs) > 0 {
			lines = append(lines, fmt.Sprintf("Give this to peers: /connect %s", allAddrs[0]))
		}
		lines = append(lines, "LAN: auto-discovered via mDNS | Internet: use /connect")
		m.addStatus(lines[0])
		for _, line := range lines[1:] {
			m.messages = append(m.messages, MessageItem{
				Sender:    "● system",
				Content:   line,
				Timestamp: "now",
			})
		}
		m.chatView.SetContent(m.renderMessages())
		m.chatView.GotoBottom()

	case "/relay":
		if len(parts) < 2 {
			m.addStatus("Usage: /relay <multiaddr>  or  /relay connect <addr>")
			m.addStatus("Set relay_peers in config.yaml to auto-connect on startup")
			return nil
		}
		if parts[1] == "connect" && len(parts) >= 3 {
			if err := m.app.Connect(parts[2]); err != nil {
				m.addStatus(fmt.Sprintf("Relay connect error: %v", err))
				return nil
			}
			m.addStatus(fmt.Sprintf("Connected to relay: %s", parts[2]))
		} else {
			if err := m.app.Connect(parts[1]); err != nil {
				m.addStatus(fmt.Sprintf("Relay connect error: %v", err))
				return nil
			}
			m.addStatus(fmt.Sprintf("Connected to relay: %s", parts[1]))
		}

	case "/profile":
		id := m.app.Identity()
		fp := crypto.Fingerprint(id.PublicKey)
		m.addStatus(fmt.Sprintf("Profile: %s | PeerID: %s | Fingerprint: %s",
			id.DisplayName, id.PeerID, fp))

	case "/tunnel":
		if len(parts) < 2 {
			m.addStatus("Usage: /tunnel <server-addr>  (e.g. /tunnel 1.2.3.4:1234)")
			return nil
		}
		serverAddr := parts[1]

		localPort := 0
		for _, addr := range m.app.AllAddrs() {
			if p := extractPort(addr); p > 0 {
				localPort = p
				break
			}
		}

		if localPort == 0 {
			m.addStatus("Could not determine local port.")
			return nil
		}

		m.loading = true
		m.loadingMsg = "Setting up tunnel..."

		return func() tea.Msg {
			publicPort, err := tunnel.RunClient(serverAddr, localPort)
			if err != nil {
				return statusMsg(fmt.Sprintf("Tunnel error: %v", err))
			}
			host, _, _ := net.SplitHostPort(serverAddr)
			return statusMsg(fmt.Sprintf("Tunnel active! Share: /connect /ip4/%s/tcp/%d/p2p/%s", host, publicPort, m.app.PeerID()))
		}

	case "/publicip":
		m.loading = true
		m.loadingMsg = "Looking up public IP..."
		return func() tea.Msg {
			ip, err := fetchPublicIP()
			if err != nil {
				return statusMsg(fmt.Sprintf("Public IP lookup failed: %v", err) + "\nTry: curl ifconfig.me  (in another terminal)")
			}
			port := 0
			for _, addr := range m.app.AllAddrs() {
				if p := extractPort(addr); p > 0 {
					port = p
					break
				}
			}
			if port > 0 {
				return statusMsg(fmt.Sprintf("If UPnP works or port %d is forwarded: /connect /ip4/%s/tcp/%d/p2p/%s", port, ip, port, m.app.PeerID()))
			}
			return statusMsg(fmt.Sprintf("Public IP: %s", ip))
		}

	case "/connections":
		conns, err := m.app.ListConnections()
		if err != nil {
			m.addStatus(fmt.Sprintf("Error: %v", err))
			return nil
		}
		if len(conns) == 0 {
			m.addStatus("No saved connections. Use /connect to connect to a peer.")
			return nil
		}
		m.showHelp = false
		m.showPeers = false
		lines := []string{"Saved connections:"}
		for i, c := range conns {
			addr := c.Address
			if utf8.RuneCountInString(addr) > 50 {
				addr = string([]rune(addr)[:50]) + "..."
			}
			nick := c.Nickname
			if nick == "" {
				nick = "-"
			}
			lines = append(lines, fmt.Sprintf("  %d. %s  (%s)", i+1, addr, nick))
		}
		lines = append(lines, "Reconnect: /connect <index>")
		for _, line := range lines {
			m.messages = append(m.messages, MessageItem{
				Sender:    "● system",
				Content:   line,
				Timestamp: "now",
			})
		}
		m.chatView.SetContent(m.renderMessages())
		m.chatView.GotoBottom()

	case "/name":
		if len(parts) < 2 {
			m.addStatus(fmt.Sprintf("Current name: %s", m.app.Identity().DisplayName))
			m.addStatus("Usage: /name <new name>")
			return nil
		}
		name := strings.Join(parts[1:], " ")
		if m.app.IsReservedDisplayName(name) {
			m.addStatus("Invalid or reserved name")
			return nil
		}
		m.app.SetDisplayName(name)
		m.addStatus(fmt.Sprintf("Display name changed to '%s'", name))

	case "/quit":
		return tea.Quit

	default:
		m.addStatus(fmt.Sprintf("Unknown command: %s. Type /help for help.", cmd))
	}

	return nil
}

func (m *Model) loadingTick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return loadingTickMsg(t)
	})
}

func (m *Model) addStatus(msg string) {
	m.statusText = msg
	m.statusLog = append(m.statusLog, msg)
	if len(m.statusLog) > 100 {
		m.statusLog = m.statusLog[len(m.statusLog)-100:]
	}
}

func (m *Model) updateUnread() {
	if m.dmFocused {
		if len(m.dmList) == 0 {
			return
		}
		curID := m.dmList[m.selectedDM].ChannelID
		for _, ch := range m.dmList {
			if ch.ChannelID == curID {
				continue
			}
			cnt, err := m.app.CountChannelMessages(ch.ChannelID)
			if err != nil {
				continue
			}
			prev := m.lastMsgCnt[ch.ChannelID]
			if cnt > prev {
				m.unread[ch.ChannelID] += cnt - prev
			}
			m.lastMsgCnt[ch.ChannelID] = cnt
		}
		return
	}
	if len(m.channelList) == 0 {
		return
	}
	curID := m.channelList[m.selectedChan].ChannelID
	for _, ch := range m.channelList {
		if ch.ChannelID == curID {
			continue
		}
		cnt, err := m.app.CountChannelMessages(ch.ChannelID)
		if err != nil {
			continue
		}
		prev := m.lastMsgCnt[ch.ChannelID]
		if cnt > prev {
			m.unread[ch.ChannelID] += cnt - prev
		}
		m.lastMsgCnt[ch.ChannelID] = cnt
	}
}

func (m *Model) loadLogs() {
	entries := m.app.Logger.UIMessages()
	m.logEntries = nil
	for _, e := range entries {
		m.logEntries = append(m.logEntries, fmt.Sprintf("[%s] %s", e.Level, e.Message))
	}
}

func (m *Model) loadChannels() {
	var currentID string
	if len(m.channelList) > 0 && m.selectedChan < len(m.channelList) {
		currentID = m.channelList[m.selectedChan].ChannelID
	}

	channels, err := m.app.ListChannels()
	if err != nil {
		m.addStatus(fmt.Sprintf("Error loading channels: %v", err))
		return
	}
	m.channelList = channels

	if currentID != "" {
		found := false
		for i, ch := range m.channelList {
			if ch.ChannelID == currentID {
				m.selectedChan = i
				found = true
				break
			}
		}
		if !found {
			m.selectedChan = 0
		}
	} else if m.selectedChan >= len(m.channelList) {
		m.selectedChan = 0
	}
	if m.selectedChan < 0 {
		m.selectedChan = 0
	}

	if !m.dmFocused {
		m.loadMessages()
	}
}

func (m *Model) loadDMs() {
	var currentID string
	if len(m.dmList) > 0 && m.selectedDM < len(m.dmList) {
		currentID = m.dmList[m.selectedDM].ChannelID
	}

	dms, err := m.app.ListDMChannels()
	if err != nil {
		m.addStatus(fmt.Sprintf("Error loading DMs: %v", err))
		return
	}
	m.dmList = dms

	if currentID != "" {
		found := false
		for i, ch := range m.dmList {
			if ch.ChannelID == currentID {
				m.selectedDM = i
				found = true
				break
			}
		}
		if !found {
			m.selectedDM = 0
		}
	} else if m.selectedDM >= len(m.dmList) {
		m.selectedDM = 0
	}
	if m.selectedDM < 0 {
		m.selectedDM = 0
	}

	if m.dmFocused {
		m.loadDMMessages()
	}
}

func (m *Model) loadMessages() {
	if len(m.channelList) == 0 {
		m.messages = nil
		m.chatView.SetContent("")
		return
	}
	channelID := m.channelList[m.selectedChan].ChannelID
	msgs, err := m.app.ListMessages(channelID, 500, 0)
	if err != nil {
		m.addStatus(fmt.Sprintf("Error loading messages: %v", err))
		return
	}

	m.unread[channelID] = 0
	m.lastMsgCnt[channelID] = len(msgs)
	m.messages = nil
	for i := len(msgs) - 1; i >= 0; i-- {
		msg := msgs[i]
		sender := m.app.GetPeerDisplayName(msg.SenderPeerID)
		m.messages = append(m.messages, MessageItem{
			Sender:        sender,
			SenderPeerID:  msg.SenderPeerID,
			Content:       msg.Content,
			Timestamp:     msg.CreatedAt.Format("15:04"),
			DeliveryState: msg.DeliveryState,
		})
	}
	m.chatView.SetContent(m.renderMessages())
	m.chatView.GotoBottom()
}

func (m *Model) loadDMMessages() {
	if len(m.dmList) == 0 {
		m.messages = nil
		m.chatView.SetContent("")
		return
	}
	channelID := m.dmList[m.selectedDM].ChannelID
	msgs, err := m.app.ListMessages(channelID, 500, 0)
	if err != nil {
		m.addStatus(fmt.Sprintf("Error loading DM messages: %v", err))
		return
	}

	m.unread[channelID] = 0
	m.lastMsgCnt[channelID] = len(msgs)
	m.messages = nil
	for i := len(msgs) - 1; i >= 0; i-- {
		msg := msgs[i]
		sender := m.app.GetPeerDisplayName(msg.SenderPeerID)
		m.messages = append(m.messages, MessageItem{
			Sender:        sender,
			SenderPeerID:  msg.SenderPeerID,
			Content:       msg.Content,
			Timestamp:     msg.CreatedAt.Format("15:04"),
			DeliveryState: msg.DeliveryState,
		})
	}
	m.chatView.SetContent(m.renderMessages())
	m.chatView.GotoBottom()
}

func (m *Model) loadPeers() {
	peers, err := m.app.ListPeers()
	if err != nil {
		m.addStatus(fmt.Sprintf("Error loading peers: %v", err))
		return
	}
	m.peerList = peers
}

func (m *Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}
	if m.needsName {
		return m.renderNameModal()
	}

	channelPanel := m.renderChannelPanel()
	dmPanel := m.renderDMPanel()
	chatPanel := m.renderChatPanel()
	statusBar := m.renderStatusBar()

	leftSide := lipgloss.JoinVertical(lipgloss.Top, channelPanel, dmPanel)

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, leftSide, chatPanel)

	input := InputStyle.Render(m.input.View())
	if !m.inputMode {
		input = DimmedInputStyle.Render(m.input.View())
	}

	body := lipgloss.JoinVertical(lipgloss.Left, topRow, input, statusBar)

	return AppStyle.Render(body)
}

func (m *Model) renderNameModal() string {
	modalW := 50

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Render("Welcome to go-chat!")

	prompt := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#B9BBBE")).
		Render("Enter your display name to get started:")

	input := m.input.View()

	parts := []string{title, "", prompt, "", input}
	if m.namePromptErr != "" {
		parts = append(parts, "", ErrorStyle.Render(m.namePromptErr))
	}
	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	box := lipgloss.NewStyle().
		Width(modalW).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#57F287")).
		Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m *Model) renderChannelPanel() string {
	style := ChannelPanelStyle
	if !m.inputMode && !m.dmFocused {
		style = ChannelPanelFocusedStyle
	}

	if len(m.channelList) == 0 {
		return style.Render(DimmedStyle.Render("No channels\n\n/channel create <name>"))
	}

	var items []string
	for i, ch := range m.channelList {
		prefix := "# "
		if ch.ChannelType == "private" {
			prefix = PrivateChannelIcon + " "
		}
		name := ch.Name
		if utf8.RuneCountInString(name) > 22 {
			name = string([]rune(name)[:22])
		}
		unreadStr := ""
		if cnt := m.unread[ch.ChannelID]; cnt > 0 && i != m.selectedChan {
			unreadStr = fmt.Sprintf(" (%d)", cnt)
		}
		disp := "  " + prefix + name + unreadStr
		if i == m.selectedChan {
			items = append(items, SelectedChannelStyle.Render(disp))
		} else {
			items = append(items, ChannelItemStyle.Render(disp))
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, items...)
	countStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#808080")).Padding(0, 1)
	return style.Render(TitleStyle.Render("Channels") + "\n" + content + "\n" + countStyle.Render(fmt.Sprintf("%d channels", len(m.channelList))))
}

func (m *Model) renderDMPanel() string {
	style := DMPanelStyle
	if !m.inputMode && m.dmFocused {
		style = DMPanelFocusedStyle
	}

	if len(m.dmList) == 0 {
		return style.Render(DimmedStyle.Render("No DMs\n\n/dm <peer_id>"))
	}

	var items []string
	for i, ch := range m.dmList {
		name := ch.Name
		if utf8.RuneCountInString(name) > 24 {
			name = string([]rune(name)[:24])
		}
		unreadStr := ""
		if cnt := m.unread[ch.ChannelID]; cnt > 0 {
			unreadStr = fmt.Sprintf(" (%d)", cnt)
		}
		disp := "  @ " + name + unreadStr
		if i == m.selectedDM {
			items = append(items, SelectedDMStyle.Render(disp))
		} else {
			items = append(items, DMItemStyle.Render(disp))
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, items...)
	countStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#808080")).Padding(0, 1)
	return style.Render(DMTitleStyle.Render("Direct Messages") + "\n" + content + "\n" + countStyle.Render(fmt.Sprintf("%d conversations", len(m.dmList))))
}

func (m *Model) renderChatPanel() string {
	messages := m.renderMessages()

	header := ""
	if m.dmFocused {
		if len(m.dmList) > 0 && m.selectedDM < len(m.dmList) {
			ch := m.dmList[m.selectedDM]
			header = DMHeaderStyle.Render(" @ "+ch.Name+" ") + "\n"
		}
	} else if len(m.channelList) > 0 && m.selectedChan < len(m.channelList) {
		ch := m.channelList[m.selectedChan]
		prefix := " # "
		if ch.ChannelType == "private" {
			prefix = " " + PrivateChannelIcon + " "
		}
		header = ChannelHeaderStyle.Render(prefix+ch.Name+" ") + "\n"
	}

	chatContent := header + "\n" + messages
	m.chatView.SetContent(chatContent)

	panel := ChatPanelStyle.Render(m.chatView.View())

	if m.showHelp || m.showPeers || m.showLogs {
		pw := m.chatView.Width - 4
		ph := m.chatView.Height - 4
		if pw < 50 {
			pw = 50
		}
		if ph < 20 {
			ph = 20
		}
		overlayContent := m.helpView()
		if m.showPeers {
			overlayContent = m.peersView()
		}
		if m.showLogs {
			overlayContent = m.logsView()
		}
		overlayBorderColor := lipgloss.Color("#57F287")
		if m.showLogs {
			overlayBorderColor = lipgloss.Color("#FEE75C")
		}
		overlay := lipgloss.NewStyle().
			Width(pw).
			Height(ph).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(overlayBorderColor).
			Render(overlayContent)
		panel = lipgloss.Place(m.chatView.Width+2, m.chatView.Height+2,
			lipgloss.Center, lipgloss.Center, overlay)
	}

	return panel
}

func (m *Model) renderMessages() string {
	if m.showHelp || m.showPeers || m.showLogs {
		return ""
	}
	if len(m.messages) == 0 {
		return DimmedStyle.Render("  No messages yet. Start typing!")
	}

	chatWidth := m.chatView.Width - 4
	if chatWidth < 30 {
		chatWidth = 40
	}

	var items []string
	for _, msg := range m.messages {
		senderStyle := lipgloss.NewStyle().Bold(true).Foreground(senderColor(msg.SenderPeerID))
		if msg.SenderPeerID == m.app.PeerID() {
			senderStyle = SelfSenderStyle
		}

		deliveryMark := ""
		if msg.SenderPeerID == m.app.PeerID() {
			switch msg.DeliveryState {
			case "sent":
				deliveryMark = "✓"
			case "received":
				deliveryMark = "✓✓"
			default:
				deliveryMark = "○"
			}
		}

		wrapWidth := chatWidth - 20
		if wrapWidth < 20 {
			wrapWidth = 20
		}
		wrapped := lipgloss.NewStyle().Width(wrapWidth).Render(msg.Content)
		contentLines := strings.Split(wrapped, "\n")
		for j, cl := range contentLines {
			if j == 0 {
				line := fmt.Sprintf("%s %s %s%s",
					TimeStyle.Render(msg.Timestamp),
					senderStyle.Render(msg.Sender),
					deliveryMark,
					cl,
				)
				items = append(items, line)
			} else {
				items = append(items, fmt.Sprintf("%s %s  %s",
					TimeStyle.Render(""),
					strings.Repeat(" ", lipgloss.Width(senderStyle.Render(msg.Sender))),
					cl,
				))
			}
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, items...)
}

func (m *Model) renderStatusBar() string {
	modeBadge := ModeBadgeInput.Render(" INPUT ")
	if !m.inputMode {
		modeBadge = ModeBadgeNav.Render(" NAV ")
	}

	ctx := ""
	if m.dmFocused {
		if len(m.dmList) > 0 && m.selectedDM < len(m.dmList) {
			ctx = "@ " + m.dmList[m.selectedDM].Name
		}
	} else if len(m.channelList) > 0 && m.selectedChan < len(m.channelList) {
		ch := m.channelList[m.selectedChan]
		prefix := "#"
		if ch.ChannelType == "private" {
			prefix = PrivateChannelIcon
		}
		ctx = prefix + ch.Name
	}

	statusText := m.statusText
	if ctx != "" {
		statusText = ctx + " │ " + m.statusText
	}
	if m.loading {
		statusText = "⟳ " + m.loadingMsg
	}

	left := StatusStyle.Render(statusText)
	onlinePeers := 0
	for _, p := range m.peerList {
		if p.Status == "online" {
			onlinePeers++
		}
	}
	right := StatusStyle.Render(fmt.Sprintf("%s  Peers: %d", modeBadge, onlinePeers))

	barWidth := m.width - 4
	leftLen := lipgloss.Width(left)
	rightLen := lipgloss.Width(right)
	gap := barWidth - leftLen - rightLen
	if gap < 1 {
		gap = 1
	}

	spacer := strings.Repeat(" ", gap)
	return lipgloss.JoinHorizontal(lipgloss.Left, left, spacer, right)
}

func (m *Model) helpView() string {
	return HelpStyle.Render(`Commands:
  /help               Show this help
  /myaddr             Show your local addresses
  /publicip           Look up your public IP
  /connect <addr>     Connect to a peer directly
  /connect <index>    Reconnect to a saved connection
  /connections        List saved connections
  /relay <addr>       Connect via a relay peer
  /tunnel <addr>      Create TCP tunnel
  /disconnect         Disconnect all peers
  /peers              List known peers
  /channel create     Create a public channel
  /channel private    Create a private channel (invite only)
  /invite <ch> <peer> Invite peer to private channel
  /invite accept <id> Accept a channel invite
  /invites            List pending invites
  /dm <peer>          Open direct message
  /name [name]        Show or set your display name
  /profile            Show your profile
  /quit               Quit

Keys:
  Tab        Cycle: input -> channels -> DMs
  Arrows     Navigate channels/DMs (nav mode)
  Enter      Send message / confirm
  ?          Toggle help
  P          Toggle peers
  L          Toggle logs
  Ctrl+C     Quit

Internet:
  /publicip           Show your public IP
  /tunnel <addr>      TCP tunnel via a public server
  /relay <addr>       libp2p relay (requires public server)`)
}

func (m *Model) peersView() string {
	if len(m.peerList) == 0 {
		return DimmedStyle.Render("  No peers connected.")
	}
	var items []string
	for _, p := range m.peerList {
		id := p.PeerID
		if utf8.RuneCountInString(id) > 16 {
			id = string([]rune(id)[:16])
		}
		items = append(items, fmt.Sprintf("  %s (%s) [%s]", p.DisplayName, p.Status, id))
	}
	return strings.Join(items, "\n")
}

func (m *Model) logsView() string {
	var lines []string
	statusStart := 0
	if len(m.statusLog) > 3 {
		statusStart = len(m.statusLog) - 3
	}
	for i := statusStart; i < len(m.statusLog); i++ {
		entry := m.statusLog[i]
		if utf8.RuneCountInString(entry) > 60 {
			entry = string([]rune(entry)[:60])
		}
		lines = append(lines, DimmedStyle.Render("∙ "+entry))
	}

	maxLogLines := 12 - (len(m.statusLog) - statusStart)
	if maxLogLines < 0 {
		maxLogLines = 0
	}
	logStart := 0
	if len(m.logEntries) > maxLogLines {
		logStart = len(m.logEntries) - maxLogLines
	}
	for i := logStart; i < len(m.logEntries); i++ {
		entry := m.logEntries[i]
		if utf8.RuneCountInString(entry) > 55 {
			entry = string([]rune(entry)[:55])
		}
		lines = append(lines, DimmedStyle.Render(entry))
	}

	if len(lines) == 0 {
		return DimmedStyle.Render("  No log entries.")
	}
	return strings.Join(lines, "\n")
}

func extractPort(addr string) int {
	if !strings.HasPrefix(addr, "/") {
		return 0
	}
	parts := strings.Split(addr, "/")
	for i, part := range parts {
		if (part == "tcp" || part == "udp" || part == "quic" || part == "quic-v1") && i+1 < len(parts) {
			var port int
			if _, err := fmt.Sscanf(parts[i+1], "%d", &port); err == nil {
				return port
			}
		}
	}
	return 0
}

func fetchPublicIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(ip)), nil
}
