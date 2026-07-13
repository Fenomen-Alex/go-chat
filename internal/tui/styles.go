package tui

import "github.com/charmbracelet/lipgloss"

var (
	PrivateChannelIcon = "\U0001F512"

	AppStyle = lipgloss.NewStyle().Padding(0, 1)

	ChannelPanelStyle = lipgloss.NewStyle().
				Width(32).
				Border(lipgloss.RoundedBorder())

	ChannelPanelFocusedStyle = lipgloss.NewStyle().
					Width(32).
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("#57F287"))

	DMPanelStyle = lipgloss.NewStyle().
			Width(32).
			Border(lipgloss.RoundedBorder())

	DMPanelFocusedStyle = lipgloss.NewStyle().
				Width(32).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#00E5FF"))

	ChatPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder())

	InputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FEE75C"))

	DimmedInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#3a3a3a"))

	StatusStyle = lipgloss.NewStyle().
			Height(1).
			Padding(0, 1)

	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#5865F2")).
			Padding(0, 1).
			Bold(true)

	DMTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#00E5FF")).
			Padding(0, 1).
			Bold(true)

	SelectedChannelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#57F287")).
				Padding(0, 1).
				Bold(true)

	ChannelItemStyle = lipgloss.NewStyle().Padding(0, 1)

	SelectedDMStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#00E5FF")).
			Padding(0, 1).
			Bold(true)

	DMItemStyle = lipgloss.NewStyle().Padding(0, 1)

	senderPalette = []string{
		"#57F287", // green
		"#00E5FF", // cyan
		"#FF73FA", // pink
		"#FFA657", // orange
		"#5865F2", // blue
		"#B077FF", // purple
		"#ED4245", // red
		"#00FFAA", // teal
	}

	SelfSenderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FEE75C"))

	TimeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#808080")).
			Width(8).
			Align(lipgloss.Right)

	DimmedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#808080"))

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ED4245"))

	ModeBadgeInput = lipgloss.NewStyle().
			Background(lipgloss.Color("#FEE75C")).
			Foreground(lipgloss.Color("#000000")).
			Padding(0, 1).
			Bold(true)

	ModeBadgeNav = lipgloss.NewStyle().
			Background(lipgloss.Color("#5865F2")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			Bold(true)

	HelpStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Width(60)

	ChannelHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#ED4245")).
				Padding(0, 1).
				Bold(true)

	DMHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#00E5FF")).
			Padding(0, 1).
			Bold(true)

)

func senderColor(s string) lipgloss.Color {
	h := 0
	for _, c := range s {
		h = h*31 + int(c)
	}
	if h < 0 {
		h = -h
	}
	return lipgloss.Color(senderPalette[h%len(senderPalette)])
}
