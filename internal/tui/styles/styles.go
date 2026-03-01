package styles

import "github.com/charmbracelet/lipgloss"

var (
	ColorPrimary   = lipgloss.Color("#22d3ee")
	ColorSecondary = lipgloss.Color("#818cf8")
	ColorSuccess   = lipgloss.Color("#32CD32")
	ColorError     = lipgloss.Color("#FF6347")
	ColorMuted     = lipgloss.Color("#808080")
	ColorWhite     = lipgloss.Color("#FFFFFF")
	ColorWarning   = lipgloss.Color("#FFA500")

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			MarginBottom(1)

	DescStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2)

	HintStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Italic(true)

	BannerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1)

	ActiveTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Background(lipgloss.Color("#1a3a4a")).
			Padding(0, 1)

	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Padding(0, 1)

	TabCountStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true)

	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSecondary).
			Padding(1, 2).
			Width(50)

	CardTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSecondary)

	SelectedItemStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorPrimary)

	NormalItemStyle = lipgloss.NewStyle().
			Foreground(ColorWhite)

	InstalledBadge = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	DetailTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorPrimary).
				MarginBottom(1)

	DetailLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorSecondary).
				Width(10)

	DetailValueStyle = lipgloss.NewStyle().
				Foreground(ColorWhite)

	TagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1a1a2e")).
			Background(ColorPrimary).
			Padding(0, 1)

	SectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSecondary).
			MarginTop(1).
			MarginBottom(1)

	HelpBarStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			MarginTop(1)

	HelpKeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	CursorStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	PageInfoStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true)

	SearchPromptStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true)

	SearchInputStyle = lipgloss.NewStyle().
				Foreground(ColorWhite)

	EmptyStateStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true).
			Padding(2, 4)

	WelcomeStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Italic(true).
			PaddingLeft(2)

	StepStyle = lipgloss.NewStyle().Foreground(ColorSecondary).Bold(true)
	NoteStyle = lipgloss.NewStyle().Foreground(ColorSecondary).Italic(true).PaddingLeft(2)
	WarnStyle = lipgloss.NewStyle().Foreground(ColorWarning).Bold(true)
	CmdStyle  = lipgloss.NewStyle().Foreground(ColorPrimary).PaddingLeft(4)
	CheckStyle = lipgloss.NewStyle().Foreground(ColorSuccess)
	CrossStyle = lipgloss.NewStyle().Foreground(ColorError)
)

var DuckBanner = `
      __
     ( o>
     //\
     V_/_
    yduck
`

func TypeIcon(t string) string {
	switch t {
	case "cli-tool":
		return "🔧"
	case "mcp":
		return "🔌"
	case "skill":
		return "📝"
	case "command":
		return "⌨️ "
	case "rule":
		return "📏"
	case "bundle":
		return "📦"
	default:
		return "·"
	}
}
