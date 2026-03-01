package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yangduck/yduck/internal/config"
	"github.com/yangduck/yduck/internal/recipe"
	s "github.com/yangduck/yduck/internal/tui/styles"
)

type homeDataLoaded struct {
	bundles       []recipe.Recipe
	popular       []recipe.Recipe
	recent        []recipe.Recipe
	allItems      []homeItem
	installStatus map[string]bool
}

type HomeView struct {
	registry    *recipe.Registry
	config      *config.Config
	installFlow *InstallFlow

	bundles  []recipe.Recipe
	popular  []recipe.Recipe
	recent   []recipe.Recipe

	installStatus map[string]bool

	cursor   int
	allItems []homeItem
	loading  bool
}

type homeItem struct {
	recipe  recipe.Recipe
	section string
}

func NewHomeView(reg *recipe.Registry, cfg *config.Config, flow *InstallFlow) *HomeView {
	return &HomeView{
		registry:      reg,
		config:        cfg,
		installFlow:   flow,
		installStatus: make(map[string]bool),
		loading:       true,
	}
}

func (h *HomeView) Init() tea.Cmd {
	reg := h.registry
	flow := h.installFlow
	return func() tea.Msg {
		featured := reg.Featured()
		var bundles []recipe.Recipe
		for _, r := range featured {
			if r.Type == recipe.TypeBundle {
				bundles = append(bundles, r)
			}
		}

		result := reg.List(recipe.ListOptions{SortBy: "popularity", PageSize: 5})
		popular := result.Items
		recent := reg.RecentlyAdded(5)

		var allItems []homeItem
		for _, b := range bundles {
			allItems = append(allItems, homeItem{recipe: b, section: "bundle"})
		}
		for _, r := range popular {
			allItems = append(allItems, homeItem{recipe: r, section: "popular"})
		}
		for _, r := range recent {
			allItems = append(allItems, homeItem{recipe: r, section: "recent"})
		}

		installStatus := make(map[string]bool)
		for _, item := range allItems {
			installed, _ := flow.IsRecipeInstalled(item.recipe)
			installStatus[item.recipe.ID] = installed
		}

		return homeDataLoaded{
			bundles:       bundles,
			popular:       popular,
			recent:        recent,
			allItems:      allItems,
			installStatus: installStatus,
		}
	}
}

func (h *HomeView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case homeDataLoaded:
		h.bundles = msg.bundles
		h.popular = msg.popular
		h.recent = msg.recent
		h.allItems = msg.allItems
		h.installStatus = msg.installStatus
		h.loading = false
		return h, nil

	case tea.KeyMsg:
		if h.loading {
			return h, nil
		}
		switch msg.String() {
		case "up", "k":
			if h.cursor > 0 {
				h.cursor--
			}
		case "down", "j":
			if h.cursor < len(h.allItems)-1 {
				h.cursor++
			}
		case "enter":
			if h.cursor < len(h.allItems) {
				item := h.allItems[h.cursor]
				return h, func() tea.Msg {
					return SwitchViewMsg{Target: TargetDetail, RecipeID: item.recipe.ID}
				}
			}
		case "B", "b":
			return h, func() tea.Msg {
				return SwitchViewMsg{Target: TargetBrowse}
			}
		case "S", "s", "/":
			return h, func() tea.Msg {
				return SwitchViewMsg{Target: TargetBrowse, SearchTerm: "/"}
			}
		case "I", "i":
			return h, func() tea.Msg {
				return SwitchViewMsg{Target: TargetInstalled}
			}
		case "M", "m":
			return h, func() tea.Msg {
				return ToggleModeMsg{}
			}
		}
	}
	return h, nil
}

func (h *HomeView) View() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(s.BannerStyle.Render(s.DuckBanner))
	b.WriteString("\n")

	if h.loading {
		b.WriteString(s.DescStyle.Render("  正在检查已安装工具..."))
		b.WriteString("\n")
		return b.String()
	}

	if h.config.IsBeginner() {
		b.WriteString(s.WelcomeStyle.Render("  👋 欢迎！浏览下方推荐内容，或按快捷键导航。"))
		b.WriteString("\n\n")
	}

	idx := 0

	if len(h.bundles) > 0 {
		b.WriteString(s.SectionStyle.Render("  📦 推荐套餐"))
		b.WriteString("\n")
		for _, bundle := range h.bundles {
			cursor := "  "
			style := s.NormalItemStyle
			if idx == h.cursor {
				cursor = s.CursorStyle.Render("▸ ")
				style = s.SelectedItemStyle
			}
			status := ""
			if h.installStatus[bundle.ID] {
				status = s.InstalledBadge.Render(" ✓")
			}
			line := fmt.Sprintf("%s%s %s — %s (%d 个工具)%s",
				cursor,
				s.TypeIcon(string(bundle.Type)),
				style.Render(bundle.Name),
				s.DescStyle.Render(bundle.Description),
				len(bundle.Includes),
				status,
			)
			b.WriteString(line)
			b.WriteString("\n")
			idx++
		}
		b.WriteString("\n")
	}

	if len(h.popular) > 0 {
		b.WriteString(s.SectionStyle.Render("  🔥 热门配方"))
		b.WriteString("\n")
		for _, rec := range h.popular {
			cursor := "  "
			style := s.NormalItemStyle
			if idx == h.cursor {
				cursor = s.CursorStyle.Render("▸ ")
				style = s.SelectedItemStyle
			}
			status := ""
			if h.installStatus[rec.ID] {
				status = s.InstalledBadge.Render(" ✓")
			}
			line := fmt.Sprintf("%s%s %s — %s%s",
				cursor,
				s.TypeIcon(string(rec.Type)),
				style.Render(rec.Name),
				s.DescStyle.Render(rec.Description),
				status,
			)
			b.WriteString(line)
			b.WriteString("\n")
			idx++
		}
		b.WriteString("\n")
	}

	if len(h.recent) > 0 {
		b.WriteString(s.SectionStyle.Render("  🆕 最近新增"))
		b.WriteString("\n")
		for _, rec := range h.recent {
			cursor := "  "
			style := s.NormalItemStyle
			if idx == h.cursor {
				cursor = s.CursorStyle.Render("▸ ")
				style = s.SelectedItemStyle
			}
			status := ""
			if h.installStatus[rec.ID] {
				status = s.InstalledBadge.Render(" ✓")
			}
			line := fmt.Sprintf("%s%s %s — %s%s",
				cursor,
				s.TypeIcon(string(rec.Type)),
				style.Render(rec.Name),
				s.DescStyle.Render(rec.Description),
				status,
			)
			b.WriteString(line)
			b.WriteString("\n")
			idx++
		}
	}

	// 当前模式
	modeLabel := "🌱 新手模式"
	if !h.config.IsBeginner() {
		modeLabel = "⚡ 高级模式"
	}
	b.WriteString(s.DescStyle.Render(fmt.Sprintf("  当前: %s", modeLabel)))
	b.WriteString("\n\n")
	b.WriteString(renderHelpBar())

	return b.String()
}

func renderHelpBar() string {
	keys := []struct{ key, desc string }{
		{"B", "浏览"},
		{"S", "搜索"},
		{"I", "已安装"},
		{"M", "切换模式"},
		{"Q", "退出"},
	}
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s %s",
			lipgloss.NewStyle().Bold(true).Foreground(s.ColorPrimary).Render(k.key),
			lipgloss.NewStyle().Foreground(s.ColorMuted).Render(k.desc),
		))
	}
	return lipgloss.NewStyle().Foreground(s.ColorMuted).MarginTop(1).Render(
		"  " + strings.Join(parts, "  "),
	)
}
