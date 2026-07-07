package tui

import (
	"fmt"
	"strings"

	"go-chat/internal/app"
	"go-chat/internal/crypto"
	"go-chat/internal/storage"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MessageItem struct {
	Sender    string
	Content   string
	Timestamp string
}

type Model struct {
	app            *app.App
	ready          bool
	width          int
	height         int

	orgList        []*storage.Organization
	selectedOrg    int

	channelList    []*storage.Channel
	selectedChan   int

	chatView       viewport.Model
	messages       []MessageItem

	input          textinput.Model
	inputMode      bool

	statusText     string
	statusLog      []string

	peerList       []*storage.Peer

	showHelp       bool
	showPeers      bool

	err            error
}

func NewModel(a *app.App) *Model {
	ti := textinput.New()
	ti.Placeholder = "Type a message or /help..."
	ti.Focus()
	ti.CharLimit = 2000
	ti.Width = 60

	return &Model{
		app:       a,
		input:     ti,
		inputMode: true,
		statusText: fmt.Sprintf("PeerID: %s | /myaddr to see shareable address", a.PeerID()),
	}
}

func (m *Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		headerHeight := 1
		inputHeight := 3
		statusHeight := 1
		chatHeight := m.height - headerHeight - inputHeight - statusHeight - 4

		if m.chatView.Height == 0 {
			m.chatView = viewport.New(m.width-50, chatHeight)
		} else {
			m.chatView.Width = m.width - 50
			m.chatView.Height = chatHeight
		}

		m.input.Width = m.width - 54

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+q":
			return m, tea.Quit

		case "tab":
			m.inputMode = !m.inputMode

		case "up":
			if !m.inputMode && m.selectedChan > 0 {
				m.selectedChan--
				m.loadMessages()
			} else if m.inputMode {
				m.chatView.LineUp(1)
			}

		case "down":
			if !m.inputMode && m.selectedChan < len(m.channelList)-1 {
				m.selectedChan++
				m.loadMessages()
			} else if m.inputMode {
				m.chatView.LineDown(1)
			}

		case "left":
			if !m.inputMode && m.selectedOrg > 0 {
				m.selectedOrg--
				m.loadChannels()
			}

		case "right":
			if !m.inputMode && m.selectedOrg < len(m.orgList)-1 {
				m.selectedOrg++
				m.loadChannels()
			}

		case "enter":
			if m.inputMode {
				return m, m.handleInput()
			}

		case "?":
			m.showHelp = !m.showHelp

		case "P":
			m.showPeers = !m.showPeers
			if m.showPeers {
				m.loadPeers()
			}
		}
	}

	if m.inputMode {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.chatView, _ = m.chatView.Update(msg)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) handleInput() tea.Cmd {
	text := strings.TrimSpace(m.input.Value())
	m.input.SetValue("")

	if text == "" {
		return nil
	}

	if strings.HasPrefix(text, "/") {
		return m.handleCommand(text)
	}

	if len(m.channelList) == 0 {
		m.addStatus("No channel selected. Join or create a channel first.")
		return nil
	}

	channelID := m.channelList[m.selectedChan].ChannelID
	if err := m.app.SendMessage(channelID, text, "text"); err != nil {
		m.addStatus(fmt.Sprintf("Error: %v", err))
		return nil
	}

	msg := MessageItem{
		Sender:    m.app.Identity().DisplayName,
		Content:   text,
		Timestamp: "now",
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
			m.addStatus("Usage: /connect <multiaddr>")
			return nil
		}
		if err := m.app.Connect(parts[1]); err != nil {
			m.addStatus(fmt.Sprintf("Connect error: %v", err))
			return nil
		}
		m.addStatus(fmt.Sprintf("Connected to %s", parts[1]))

	case "/disconnect":
		m.app.DisconnectAll()
		m.addStatus("Disconnected")

	case "/peers":
		m.loadPeers()
		m.showPeers = !m.showPeers

	case "/org":
		if len(parts) < 2 {
			m.addStatus("Usage: /org create <name> or /org list")
			return nil
		}
		switch parts[1] {
		case "create":
			if len(parts) < 3 {
				m.addStatus("Usage: /org create <name>")
				return nil
			}
			name := strings.Join(parts[2:], " ")
			org, err := m.app.CreateOrg(name, "")
			if err != nil {
				m.addStatus(fmt.Sprintf("Error: %v", err))
				return nil
			}
			m.orgList = append(m.orgList, org)
			m.loadChannels()
			m.addStatus(fmt.Sprintf("Organization '%s' created", name))
		case "list":
			m.loadOrgs()
		default:
			m.addStatus("Unknown org command")
		}

	case "/channel":
		if len(parts) < 2 {
			m.addStatus("Usage: /channel create <name> or /channel list")
			return nil
		}
		switch parts[1] {
		case "create":
			if len(parts) < 3 {
				m.addStatus("Usage: /channel create <name>")
				return nil
			}
			name := strings.Join(parts[2:], " ")
			orgID := ""
			if len(m.orgList) > 0 {
				orgID = m.orgList[m.selectedOrg].OrgID
			}
			ch, err := m.app.CreateChannel(orgID, name, "text", "")
			if err != nil {
				m.addStatus(fmt.Sprintf("Error: %v", err))
				return nil
			}
			m.channelList = append(m.channelList, ch)
			m.addStatus(fmt.Sprintf("Channel '%s' created", name))
		case "list":
			m.loadChannels()
		default:
			m.addStatus("Unknown channel command")
		}

	case "/dm":
		if len(parts) < 2 {
			m.addStatus("Usage: /dm <peer_id>")
			return nil
		}
		peerID := parts[1]
		if err := m.app.OpenDM(peerID); err != nil {
			m.addStatus(fmt.Sprintf("Error: %v", err))
			return nil
		}
		m.addStatus(fmt.Sprintf("DM with %s", peerID))

	case "/myaddr":
		peerID := m.app.PeerID()
		allAddrs := m.app.AllAddrs()
		m.addStatus(fmt.Sprintf("=== Peer ID: %s ===", peerID))
		for _, addr := range allAddrs {
			m.addStatus(fmt.Sprintf("  %s", addr))
		}
		m.addStatus("---")
		if len(allAddrs) > 0 {
			m.addStatus(fmt.Sprintf("Share: /connect %s", allAddrs[0]))
		}
		m.addStatus("Peer can also find you via mDNS (LAN) or DHT (internet)")

	case "/profile":
		id := m.app.Identity()
		fp := crypto.Fingerprint(id.PublicKey)
		m.addStatus(fmt.Sprintf("Profile: %s | PeerID: %s | Fingerprint: %s",
			id.DisplayName, id.PeerID, fp))

	case "/quit":
		return tea.Quit

	default:
		m.addStatus(fmt.Sprintf("Unknown command: %s. Type /help for help.", cmd))
	}

	return nil
}

func (m *Model) addStatus(msg string) {
	m.statusText = msg
	m.statusLog = append(m.statusLog, msg)
	if len(m.statusLog) > 100 {
		m.statusLog = m.statusLog[len(m.statusLog)-100:]
	}
}

func (m *Model) loadOrgs() {
	orgs, err := m.app.ListOrgs()
	if err != nil {
		m.addStatus(fmt.Sprintf("Error loading orgs: %v", err))
		return
	}
	m.orgList = orgs
}

func (m *Model) loadChannels() {
	if len(m.orgList) == 0 {
		m.channelList = nil
		return
	}
	orgID := m.orgList[m.selectedOrg].OrgID
	channels, err := m.app.ListChannels(orgID)
	if err != nil {
		m.addStatus(fmt.Sprintf("Error loading channels: %v", err))
		return
	}
	m.channelList = channels
	m.selectedChan = 0
	m.loadMessages()
}

func (m *Model) loadMessages() {
	if len(m.channelList) == 0 {
		m.messages = nil
		m.chatView.SetContent("")
		return
	}
	channelID := m.channelList[m.selectedChan].ChannelID
	msgs, err := m.app.ListMessages(channelID, 100, 0)
	if err != nil {
		m.addStatus(fmt.Sprintf("Error loading messages: %v", err))
		return
	}

	m.messages = nil
	for i := len(msgs) - 1; i >= 0; i-- {
		msg := msgs[i]
		sender := msg.SenderPeerID
		if len(sender) > 12 {
			sender = sender[:12]
		}
		m.messages = append(m.messages, MessageItem{
			Sender:    sender,
			Content:   msg.Content,
			Timestamp: msg.CreatedAt.Format("15:04"),
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

	orgPanel := m.renderOrgPanel()
	channelPanel := m.renderChannelPanel()
	chatPanel := m.renderChatPanel()
	statusBar := m.renderStatusBar()

	topRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.JoinVertical(lipgloss.Top, orgPanel, channelPanel),
		chatPanel,
	)

	input := InputStyle.Render(m.input.View())
	if !m.inputMode {
		input = DimmedStyle.Render(m.input.View())
	}

	body := lipgloss.JoinVertical(lipgloss.Left, topRow, input, statusBar)

	return AppStyle.Render(body)
}

func (m *Model) renderOrgPanel() string {
	if len(m.orgList) == 0 {
		return OrgPanelStyle.Render("No organizations\n\nCreate one:\n/org create <name>")
	}

	var items []string
	for i, org := range m.orgList {
		name := org.Name
		if len(name) > 16 {
			name = name[:16]
		}
		if i == m.selectedOrg {
			items = append(items, SelectedStyle.Render("> "+name))
		} else {
			items = append(items, "  "+name)
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, items...)
	return OrgPanelStyle.Render(TitleStyle.Render("Orgs") + "\n" + content)
}

func (m *Model) renderChannelPanel() string {
	if len(m.channelList) == 0 {
		return ChannelPanelStyle.Render("No channels\n\nCreate one:\n/channel create <name>")
	}

	var items []string
	for i, ch := range m.channelList {
		name := ch.Name
		if len(name) > 20 {
			name = name[:20]
		}
		prefix := "# "
		if ch.ChannelType == "announcement" {
			prefix = "! "
		}
		if i == m.selectedChan {
			items = append(items, SelectedStyle.Render("> "+prefix+name))
		} else {
			items = append(items, "  "+prefix+name)
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, items...)
	return ChannelPanelStyle.Render(TitleStyle.Render("Channels") + "\n" + content)
}

func (m *Model) renderChatPanel() string {
	messages := m.renderMessages()

	header := ""
	if len(m.channelList) > 0 {
		ch := m.channelList[m.selectedChan]
		header = TitleStyle.Render(" # "+ch.Name+" ") + "\n"
	}

	chatContent := header + "\n" + messages
	m.chatView.SetContent(chatContent)

	return ChatPanelStyle.Render(m.chatView.View())
}

func (m *Model) renderMessages() string {
	if len(m.messages) == 0 {
		return "  No messages yet. Start typing!"
	}

	var items []string
	for _, msg := range m.messages {
		line := fmt.Sprintf("%s %s: %s",
			TimeStyle.Render(msg.Timestamp),
			SenderStyle.Render(msg.Sender),
			msg.Content,
		)
		items = append(items, line)
	}

	return lipgloss.JoinVertical(lipgloss.Left, items...)
}

func (m *Model) renderStatusBar() string {
	left := StatusStyle.Render(m.statusText)

	info := fmt.Sprintf("Peers: %d", len(m.peerList))
	if m.showHelp {
		info = "? Help"
	}
	right := StatusStyle.Copy().Align(lipgloss.Right).Render(info)

	return lipgloss.JoinHorizontal(lipgloss.Left, left, right)
}

func (m Model) helpView() string {
	return `Commands:
  /help             Show this help
  /myaddr           Show your shareable multiaddress
  /connect <addr>   Connect to a peer
  /disconnect       Disconnect all peers
  /peers            List known peers
  /org create       Create an organization
  /channel create   Create a channel
  /dm <peer>        Open direct message
  /profile          Show your profile
  /quit             Quit

Keys:
  Tab        Toggle input/navigation mode
  Arrows     Navigate (navigation mode)
  Enter      Send message
  ?          Toggle help
  P          Toggle peers
  Ctrl+C     Quit`
}

func (m Model) peersView() string {
	if len(m.peerList) == 0 {
		return "No peers connected."
	}
	var items []string
	for _, p := range m.peerList {
		items = append(items, fmt.Sprintf("  %s (%s)", p.DisplayName, p.Status))
	}
	return strings.Join(items, "\n")
}
