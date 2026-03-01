package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yangduck/yduck/internal/recipe"
)

type View interface {
	Init() tea.Cmd
	Update(tea.Msg) (View, tea.Cmd)
	View() string
}

const (
	TargetHome = iota
	TargetBrowse
	TargetDetail
	TargetInstalled
)

type SwitchViewMsg struct {
	Target     int
	Category   string
	RecipeID   string
	SearchTerm string
}

type InstallRecipeMsg struct {
	Recipe recipe.Recipe
}

type ToggleModeMsg struct{}
