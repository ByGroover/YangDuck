package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yangduck/yduck/internal/config"
	"github.com/yangduck/yduck/internal/recipe"
	s "github.com/yangduck/yduck/internal/tui/styles"
)

type BrowseView struct {
	registry    *recipe.Registry
	config      *config.Config
	installFlow *InstallFlow

	categories    []browseCategory
	activeTab     int
	cursor        int
	page          int
	pageSize      int
	items         []recipe.Recipe
	totalItems    int
	totalPages    int
	installStatus map[string]bool

	searchMode  bool
	searchInput string
}

type browseCategory struct {
	label   string
	recType recipe.RecipeType
	count   int
}

func NewBrowseView(reg *recipe.Registry, cfg *config.Config, flow *InstallFlow, category string) *BrowseView {
	counts := reg.CountByType()

	allCats := []browseCategory{
		{label: "全部", recType: "", count: reg.Count()},
		{label: "CLI 工具", recType: recipe.TypeCLITool, count: counts[recipe.TypeCLITool]},
		{label: "MCP", recType: recipe.TypeMCP, count: counts[recipe.TypeMCP]},
		{label: "Skill", recType: recipe.TypeSkill, count: counts[recipe.TypeSkill]},
		{label: "Rule", recType: recipe.TypeRule, count: counts[recipe.TypeRule]},
		{label: "Command", recType: recipe.TypeCommand, count: counts[recipe.TypeCommand]},
		{label: "套餐", recType: recipe.TypeBundle, count: counts[recipe.TypeBundle]},
	}

	// Only keep categories that have items (always keep "全部")
	var cats []browseCategory
	for _, c := range allCats {
		if c.count > 0 || c.recType == "" {
			cats = append(cats, c)
		}
	}

	activeTab := 0
	if category != "" {
		for i, c := range cats {
			if string(c.recType) == category {
				activeTab = i
				break
			}
		}
	}

	bv := &BrowseView{
		registry:    reg,
		config:      cfg,
		installFlow: flow,
		categories:  cats,
		activeTab:   activeTab,
		pageSize:    10,
		page:        1,
	}
	bv.loadItems()
	return bv
}

func (bv *BrowseView) SetSearchMode(keyword string) {
	bv.searchMode = true
	bv.searchInput = keyword
	if keyword != "" {
		bv.filterBySearch()
	}
}

func (bv *BrowseView) loadItems() {
	cat := bv.categories[bv.activeTab]
	result := bv.registry.List(recipe.ListOptions{
		Type:     cat.recType,
		Page:     bv.page,
		PageSize: bv.pageSize,
		SortBy:   "popularity",
	})
	bv.items = result.Items
	bv.totalItems = result.Total
	bv.totalPages = result.TotalPages
	bv.refreshInstallStatus()
}

func (bv *BrowseView) filterBySearch() {
	if bv.searchInput == "" {
		bv.loadItems()
		return
	}
	results := bv.registry.Search(bv.searchInput)
	bv.items = results
	bv.totalItems = len(results)
	bv.totalPages = 1
	bv.page = 1
	bv.cursor = 0
	bv.refreshInstallStatus()
}

func (bv *BrowseView) refreshInstallStatus() {
	bv.installStatus = make(map[string]bool)
	for _, rec := range bv.items {
		installed, _ := bv.installFlow.IsRecipeInstalled(rec)
		bv.installStatus[rec.ID] = installed
	}
}

func (bv *BrowseView) Init() tea.Cmd {
	return nil
}

func (bv *BrowseView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if bv.searchMode {
			return bv.updateSearch(msg)
		}
		return bv.updateNormal(msg)
	}
	return bv, nil
}

func (bv *BrowseView) updateSearch(msg tea.KeyMsg) (View, tea.Cmd) {
	switch msg.String() {
	case "esc":
		bv.searchMode = false
		bv.searchInput = ""
		bv.loadItems()
		return bv, nil
	case "enter":
		bv.searchMode = false
		return bv, nil
	case "backspace":
		runes := []rune(bv.searchInput)
		if len(runes) > 0 {
			bv.searchInput = string(runes[:len(runes)-1])
			bv.filterBySearch()
		}
		return bv, nil
	default:
		r := msg.Runes
		if len(r) == 1 {
			bv.searchInput += string(r)
			bv.filterBySearch()
		}
		return bv, nil
	}
}

func (bv *BrowseView) updateNormal(msg tea.KeyMsg) (View, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if bv.cursor > 0 {
			bv.cursor--
		} else if bv.page > 1 {
			bv.page--
			bv.loadItems()
			bv.cursor = len(bv.items) - 1
		}
	case "down", "j":
		if bv.cursor < len(bv.items)-1 {
			bv.cursor++
		} else if bv.page < bv.totalPages {
			bv.page++
			bv.loadItems()
			bv.cursor = 0
		}
	case "left", "shift+tab":
		if bv.activeTab > 0 {
			bv.activeTab--
			bv.page = 1
			bv.cursor = 0
			bv.loadItems()
		}
	case "right", "tab":
		if bv.activeTab < len(bv.categories)-1 {
			bv.activeTab++
			bv.page = 1
			bv.cursor = 0
			bv.loadItems()
		}
	case "/":
		bv.searchMode = true
		bv.searchInput = ""
	case "enter":
		if bv.cursor < len(bv.items) {
			item := bv.items[bv.cursor]
			return bv, func() tea.Msg {
				return SwitchViewMsg{Target: TargetDetail, RecipeID: item.ID}
			}
		}
	case "esc":
		return bv, func() tea.Msg {
			return SwitchViewMsg{Target: TargetHome}
		}
	case "H", "h":
		return bv, func() tea.Msg {
			return SwitchViewMsg{Target: TargetHome}
		}
	case "I", "i":
		return bv, func() tea.Msg {
			return SwitchViewMsg{Target: TargetInstalled}
		}
	}
	return bv, nil
}

func (bv *BrowseView) View() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(s.TitleStyle.Render("  🔍 浏览配方"))
	b.WriteString("\n\n")

	// Tab bar with separator
	var tabs []string
	for i, cat := range bv.categories {
		label := fmt.Sprintf("%s (%d)", cat.label, cat.count)
		if i == bv.activeTab {
			tabs = append(tabs, s.ActiveTabStyle.Render(label))
		} else {
			tabs = append(tabs, s.InactiveTabStyle.Render(label))
		}
	}
	b.WriteString("  " + strings.Join(tabs, s.DescStyle.Render(" │ ")))
	b.WriteString("\n")
	b.WriteString(s.DescStyle.Render("  ────────────────────────────────────────────"))
	b.WriteString("\n\n")

	if bv.searchMode {
		b.WriteString(fmt.Sprintf("  %s %s▌\n\n",
			s.SearchPromptStyle.Render("搜索:"),
			s.SearchInputStyle.Render(bv.searchInput),
		))
	}

	if len(bv.items) == 0 {
		b.WriteString(s.EmptyStateStyle.Render("暂无配方"))
		b.WriteString("\n")
	} else {
		for i, rec := range bv.items {
			selected := i == bv.cursor && !bv.searchMode

			status := ""
			if bv.installStatus[rec.ID] {
				status = s.InstalledBadge.Render(" ✓")
			}

			if selected {
				b.WriteString(s.CursorStyle.Render("  ▸ "))
				b.WriteString(s.SelectedItemStyle.Render(rec.Name))
				b.WriteString(status)
				b.WriteString("\n")
				b.WriteString(s.DescStyle.Render(fmt.Sprintf("    %s  %s",
					s.TypeIcon(string(rec.Type)),
					truncate(rec.Description, 55),
				)))
			} else {
				b.WriteString(fmt.Sprintf("    %s ", s.TypeIcon(string(rec.Type))))
				b.WriteString(s.NormalItemStyle.Render(rec.Name))
				b.WriteString(status)
				b.WriteString("\n")
				b.WriteString(s.DescStyle.Render(fmt.Sprintf("      %s",
					truncate(rec.Description, 55),
				)))
			}
			b.WriteString("\n")

			if i < len(bv.items)-1 {
				b.WriteString("\n")
			}
		}
	}

	if bv.totalPages > 1 {
		b.WriteString("\n")
		b.WriteString(s.PageInfoStyle.Render(
			fmt.Sprintf("  第 %d/%d 页 · 共 %d 项", bv.page, bv.totalPages, bv.totalItems),
		))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	help := "  ←→ 切换分类  ↑↓ 浏览  Enter 详情  / 搜索  Esc 返回"
	b.WriteString(s.HelpBarStyle.Render(help))

	return b.String()
}

func truncate(str string, max int) string {
	runes := []rune(str)
	if len(runes) <= max {
		return str
	}
	return string(runes[:max-1]) + "…"
}
