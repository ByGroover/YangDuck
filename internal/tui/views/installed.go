package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yangduck/yduck/internal/config"
	"github.com/yangduck/yduck/internal/recipe"
	s "github.com/yangduck/yduck/internal/tui/styles"
)

type InstalledView struct {
	registry    *recipe.Registry
	config      *config.Config
	installFlow *InstallFlow

	items  []installedItem
	cursor int
}

type installedItem struct {
	recipe  recipe.Recipe
	version string
}

func NewInstalledView(reg *recipe.Registry, cfg *config.Config, flow *InstallFlow) *InstalledView {
	iv := &InstalledView{
		registry:    reg,
		config:      cfg,
		installFlow: flow,
	}
	iv.loadInstalled()
	return iv
}

func (iv *InstalledView) loadInstalled() {
	iv.items = nil
	for _, rec := range iv.registry.All() {
		if rec.Type == recipe.TypeBundle {
			continue
		}
		installed, ver := iv.installFlow.IsRecipeInstalled(rec)
		if installed {
			iv.items = append(iv.items, installedItem{recipe: rec, version: ver})
		}
	}
}

func (iv *InstalledView) Init() tea.Cmd {
	return nil
}

func (iv *InstalledView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if iv.cursor > 0 {
				iv.cursor--
			}
		case "down", "j":
			if iv.cursor < len(iv.items)-1 {
				iv.cursor++
			}
		case "enter":
			if iv.cursor < len(iv.items) {
				item := iv.items[iv.cursor]
				return iv, func() tea.Msg {
					return SwitchViewMsg{Target: TargetDetail, RecipeID: item.recipe.ID}
				}
			}
		case "esc":
			return iv, func() tea.Msg {
				return SwitchViewMsg{Target: TargetHome}
			}
		case "B", "b":
			return iv, func() tea.Msg {
				return SwitchViewMsg{Target: TargetBrowse}
			}
		case "H", "h":
			return iv, func() tea.Msg {
				return SwitchViewMsg{Target: TargetHome}
			}
		}
	}
	return iv, nil
}

func (iv *InstalledView) View() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(s.TitleStyle.Render("  📋 已安装"))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("  共 %d 个已安装配方\n\n", len(iv.items)))

	if len(iv.items) == 0 {
		b.WriteString(s.EmptyStateStyle.Render("还没有安装任何配方"))
		b.WriteString("\n")
		b.WriteString(s.HintStyle.Render("  按 B 浏览可用配方，或按 H 回到首页"))
		b.WriteString("\n")
	} else {
		for i, item := range iv.items {
			cursor := "  "
			style := s.NormalItemStyle
			if i == iv.cursor {
				cursor = s.CursorStyle.Render("▸ ")
				style = s.SelectedItemStyle
			}

			verStr := ""
			if item.version != "" {
				verStr = s.DescStyle.Render(" v" + item.version)
			}

			line := fmt.Sprintf("%s%s %s%s — %s %s",
				cursor,
				s.TypeIcon(string(item.recipe.Type)),
				style.Render(item.recipe.Name),
				verStr,
				s.DescStyle.Render(truncate(item.recipe.Description, 40)),
				s.InstalledBadge.Render("✓"),
			)
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(s.HelpBarStyle.Render("  ↑↓ 浏览  Enter 详情  B 浏览  H 首页  Esc 返回"))

	return b.String()
}
