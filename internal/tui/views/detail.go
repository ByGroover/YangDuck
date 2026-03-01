package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yangduck/yduck/internal/config"
	"github.com/yangduck/yduck/internal/recipe"
	s "github.com/yangduck/yduck/internal/tui/styles"
)

type DetailView struct {
	registry    *recipe.Registry
	config      *config.Config
	installFlow *InstallFlow

	recipe        recipe.Recipe
	related       []recipe.Recipe
	installed     bool
	version       string
	cursor        int // 0 = install action, 1+ = related items
	includeStatus map[string]bool
}

func NewDetailView(reg *recipe.Registry, cfg *config.Config, flow *InstallFlow, recipeID string) *DetailView {
	rec, _ := reg.Get(recipeID)
	installed, ver := flow.IsRecipeInstalled(rec)
	related := reg.Related(recipeID, 3)

	includeStatus := make(map[string]bool)
	if rec.Type == recipe.TypeBundle {
		for _, id := range rec.Includes {
			if sub, ok := reg.Get(id); ok {
				inst, _ := flow.IsRecipeInstalled(sub)
				includeStatus[id] = inst
			}
		}
	}

	return &DetailView{
		registry:      reg,
		config:        cfg,
		installFlow:   flow,
		recipe:        rec,
		related:       related,
		installed:     installed,
		version:       ver,
		includeStatus: includeStatus,
	}
}

func (d *DetailView) RecipeID() string {
	return d.recipe.ID
}

func (d *DetailView) Init() tea.Cmd {
	return nil
}

func (d *DetailView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if d.cursor > 0 {
				d.cursor--
			}
		case "down", "j":
			maxCursor := len(d.related)
			if d.cursor < maxCursor {
				d.cursor++
			}
		case "enter":
			if d.cursor == 0 {
				rec := d.recipe
				return d, func() tea.Msg {
					return InstallRecipeMsg{Recipe: rec}
				}
			}
			relIdx := d.cursor - 1
			if relIdx < len(d.related) {
				rel := d.related[relIdx]
				return d, func() tea.Msg {
					return SwitchViewMsg{Target: TargetDetail, RecipeID: rel.ID}
				}
			}
		case "esc":
			return d, func() tea.Msg {
				return SwitchViewMsg{Target: TargetBrowse}
			}
		case "H", "h":
			return d, func() tea.Msg {
				return SwitchViewMsg{Target: TargetHome}
			}
		}
	}
	return d, nil
}

func (d *DetailView) View() string {
	var b strings.Builder
	rec := d.recipe

	b.WriteString("\n")

	icon := s.TypeIcon(string(rec.Type))
	b.WriteString(fmt.Sprintf("  %s %s\n", icon, s.DetailTitleStyle.Render(rec.Name)))
	b.WriteString("\n")

	typeNames := map[recipe.RecipeType]string{
		recipe.TypeCLITool: "CLI 工具",
		recipe.TypeMCP:     "MCP 服务器",
		recipe.TypeSkill:   "Skill",
		recipe.TypeCommand: "Command",
		recipe.TypeRule:    "Rule",
		recipe.TypeBundle:  "套餐",
	}

	b.WriteString(fmt.Sprintf("  %s %s\n",
		s.DetailLabelStyle.Render("类型"),
		s.DetailValueStyle.Render(typeNames[rec.Type]),
	))
	b.WriteString(fmt.Sprintf("  %s %s\n",
		s.DetailLabelStyle.Render("描述"),
		s.DetailValueStyle.Render(rec.Description),
	))
	if rec.Difficulty != "" {
		b.WriteString(fmt.Sprintf("  %s %s\n",
			s.DetailLabelStyle.Render("难度"),
			s.DetailValueStyle.Render(rec.Difficulty),
		))
	}

	if len(rec.Tags) > 0 {
		var tags []string
		for _, t := range rec.Tags {
			tags = append(tags, s.TagStyle.Render(t))
		}
		b.WriteString(fmt.Sprintf("  %s %s\n",
			s.DetailLabelStyle.Render("标签"),
			strings.Join(tags, " "),
		))
	}

	if d.installed {
		status := s.InstalledBadge.Render("✓ 已安装")
		if d.version != "" {
			status += s.DescStyle.Render(" (v" + d.version + ")")
		}
		b.WriteString(fmt.Sprintf("  %s %s\n",
			s.DetailLabelStyle.Render("状态"),
			status,
		))
	} else {
		b.WriteString(fmt.Sprintf("  %s %s\n",
			s.DetailLabelStyle.Render("状态"),
			s.DescStyle.Render("未安装"),
		))
	}

	if rec.Type == recipe.TypeBundle && len(rec.Includes) > 0 {
		b.WriteString("\n")
		b.WriteString(s.SectionStyle.Render("  📋 包含工具"))
		b.WriteString("\n")
		for _, id := range rec.Includes {
			sub, ok := d.registry.Get(id)
			if ok {
				status := "  "
				if d.includeStatus[id] {
					status = s.InstalledBadge.Render("✓ ")
				}
				b.WriteString(fmt.Sprintf("    %s%s %s — %s\n",
					status,
					s.TypeIcon(string(sub.Type)),
					sub.Name,
					s.DescStyle.Render(sub.Description),
				))
			}
		}
	}

	if len(rec.Quickstart) > 0 {
		b.WriteString("\n")
		b.WriteString(s.SectionStyle.Render("  🚀 快速上手"))
		b.WriteString("\n")
		for i, qs := range rec.Quickstart {
			b.WriteString(fmt.Sprintf("    %d. %s\n", i+1, s.StepStyle.Render(qs.Title)))
			if qs.Command != "" {
				b.WriteString(fmt.Sprintf("       %s\n", s.CmdStyle.Render("$ "+qs.Command)))
			}
			b.WriteString(fmt.Sprintf("       %s\n", s.NoteStyle.Render(qs.Explain)))
		}
	}

	b.WriteString("\n")
	actionLabel := "安装"
	if d.installed {
		actionLabel = "重新安装"
	}
	if d.cursor == 0 {
		b.WriteString(s.CursorStyle.Render("  ▸ ") + s.SelectedItemStyle.Render(actionLabel) + "\n")
	} else {
		b.WriteString("    " + s.NormalItemStyle.Render(actionLabel) + "\n")
	}

	if len(d.related) > 0 {
		b.WriteString("\n")
		b.WriteString(s.SectionStyle.Render("  💡 相关推荐"))
		b.WriteString("\n")
		for i, rel := range d.related {
			cursor := "    "
			style := s.NormalItemStyle
			if i+1 == d.cursor {
				cursor = s.CursorStyle.Render("  ▸ ")
				style = s.SelectedItemStyle
			}
			b.WriteString(fmt.Sprintf("%s%s %s — %s\n",
				cursor,
				s.TypeIcon(string(rel.Type)),
				style.Render(rel.Name),
				s.DescStyle.Render(truncate(rel.Description, 40)),
			))
		}
	}

	b.WriteString("\n")
	b.WriteString(s.HelpBarStyle.Render("  Enter 安装  ↑↓ 选择  Esc 返回浏览"))

	return b.String()
}
