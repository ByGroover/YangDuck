package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/yangduck/yduck/internal/config"
	"github.com/yangduck/yduck/internal/generator"
	"github.com/yangduck/yduck/internal/installer"
	ylog "github.com/yangduck/yduck/internal/log"
	"github.com/yangduck/yduck/internal/quickstart"
	"github.com/yangduck/yduck/internal/recipe"
	"github.com/yangduck/yduck/internal/tui"
)

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#32CD32")).Bold(true)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6347")).Bold(true)
	mutedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))
)

var verbose bool

func main() {
	reg := loadRegistry()
	cfg := config.Load()

	root := &cobra.Command{
		Use:   "yduck",
		Short: "🐤 YangDuck — 快速配置你的 Mac 开发环境",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			ylog.Init(verbose)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			app := tui.NewApp(reg, cfg)
			return app.Run()
		},
	}
	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "启用详细日志输出")

	root.AddCommand(
		installCmd(reg, cfg),
		searchCmd(reg),
		listCmd(reg),
		doctorCmd(),
		configCmd(cfg),
		updateCmd(),
		recipeCmd(),
	)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
	ylog.Sync()
}

func loadRegistry() *recipe.Registry {
	ylog.Init(false)
	reg := recipe.NewRegistry()
	recipes, err := recipe.LoadFromFS(recipe.EmbeddedRecipes, "embedded")
	if err == nil {
		reg.Add(recipes...)
		ylog.S.Debugw("loaded embedded recipes", "count", len(recipes))
	} else {
		ylog.S.Warnw("failed to load embedded recipes", "error", err)
	}
	cacheDir := config.CacheDir()
	if cached, err := recipe.LoadFromDir(cacheDir); err == nil {
		reg.Add(cached...)
		ylog.S.Debugw("loaded cached recipes", "dir", cacheDir, "count", len(cached))
	}
	return reg
}

func installCmd(reg *recipe.Registry, cfg *config.Config) *cobra.Command {
	var bundleFlag string
	cmd := &cobra.Command{
		Use:   "install [recipe...]",
		Short: "安装配方",
		RunE: func(cmd *cobra.Command, args []string) error {
			if bundleFlag != "" {
				return installBundle(reg, cfg, bundleFlag)
			}
			brew := installer.NewBrewInstaller()
			npm := installer.NewNpmInstaller()
			mcp := installer.NewMCPInstaller()
			skill := installer.NewSkillInstaller()
			for _, id := range args {
				rec, ok := reg.Get(id)
				if !ok {
					ylog.S.Warnw("recipe not found", "id", id)
					fmt.Println(errorStyle.Render("✗ 配方未找到: " + id))
					continue
				}
				ylog.S.Infow("installing recipe", "id", rec.ID, "type", rec.Type)
				switch rec.Type {
				case recipe.TypeCLITool:
					if rec.Install == nil {
						continue
					}
					if rec.Install.Method == "npm" {
						if installed, _ := npm.IsInstalled(rec.Install.Package); installed {
							fmt.Println(successStyle.Render("✓ " + rec.Name + " 已安装"))
							continue
						}
						fmt.Printf("正在安装 %s...\n", rec.Name)
						if err := npm.Install(rec.Install.Package); err != nil {
							ylog.S.Errorw("npm install failed", "package", rec.Install.Package, "error", err)
							fmt.Println(errorStyle.Render("✗ " + err.Error()))
							continue
						}
						fmt.Println(successStyle.Render("✓ " + rec.Name + " 安装完成"))
					} else {
						if installed, _ := brew.IsInstalled(rec.Install.Package); installed {
							ylog.S.Debugw("already installed", "package", rec.Install.Package)
							fmt.Println(successStyle.Render("✓ " + rec.Name + " 已安装"))
							continue
						}
						fmt.Printf("正在安装 %s...\n", rec.Name)
						if err := brew.Install(rec.Install.Package); err != nil {
							ylog.S.Errorw("brew install failed", "package", rec.Install.Package, "error", err)
							fmt.Println(errorStyle.Render("✗ " + err.Error()))
							continue
						}
						_ = brew.RunPostInstall(rec.Install.PostInstall)
						ylog.S.Infow("installed", "package", rec.Install.Package)
						fmt.Println(successStyle.Render("✓ " + rec.Name + " 安装完成"))
					}
					if cfg.IsBeginner() {
						quickstart.Show(rec)
					}
				case recipe.TypeMCP:
					if cfg.Editor == "" {
						askEditor(cfg)
					}
					var targets []string
					if rec.Targets != nil && rec.Targets.Cursor != nil && cfg.ShouldInstallFor("cursor") {
						targets = append(targets, "cursor")
					}
					if rec.Targets != nil && rec.Targets.ClaudeCode != nil && cfg.ShouldInstallFor("claude-code") {
						targets = append(targets, "claude-code")
					}
					promptValues := collectPrompts(rec.Prompts)
					for _, t := range targets {
						if err := mcp.Install(&rec, t, promptValues); err != nil {
							ylog.S.Errorw("mcp install failed", "target", t, "recipe", rec.ID, "error", err)
							fmt.Println(errorStyle.Render("✗ " + t + ": " + err.Error()))
						} else {
							ylog.S.Infow("mcp configured", "target", t, "recipe", rec.ID)
							fmt.Println(successStyle.Render("✓ " + rec.Name + " 已配置到 " + t))
							if t == "cursor" {
								fmt.Println(mutedStyle.Render("  ℹ 配置已写入 .cursor/mcp.json，仅在当前项目中生效。如需在其他项目使用，请在对应项目目录重新运行安装。"))
							}
							if t == "claude-code" {
								fmt.Println(mutedStyle.Render("  ℹ 配置已写入 ~/.claude.json，全局生效。"))
							}
						}
					}
				case recipe.TypeSkill, recipe.TypeCommand, recipe.TypeRule:
					if err := skill.Install(&rec); err != nil {
						ylog.S.Errorw("skill install failed", "recipe", rec.ID, "error", err)
						fmt.Println(errorStyle.Render("✗ " + err.Error()))
					} else {
						ylog.S.Infow("skill installed", "recipe", rec.ID)
						fmt.Println(successStyle.Render("✓ " + rec.Name + " 安装完成"))
					}
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&bundleFlag, "bundle", "", "安装套餐")
	return cmd
}

func installBundle(reg *recipe.Registry, cfg *config.Config, id string) error {
	rec, ok := reg.Get(id)
	if !ok {
		return fmt.Errorf("套餐未找到: %s", id)
	}
	bi := installer.NewBundleInstaller(reg)
	var targets []string
	if cfg.ShouldInstallFor("cursor") {
		targets = append(targets, "cursor")
	}
	if cfg.ShouldInstallFor("claude-code") {
		targets = append(targets, "claude-code")
	}
	result, err := bi.Install(&rec, nil, targets)
	if err != nil {
		return err
	}
	for _, id := range result.Installed {
		fmt.Println(successStyle.Render("✓ " + id))
	}
	for id, e := range result.Failed {
		fmt.Println(errorStyle.Render("✗ " + id + ": " + e.Error()))
	}
	return nil
}

func searchCmd(reg *recipe.Registry) *cobra.Command {
	return &cobra.Command{
		Use:   "search <keyword>",
		Short: "搜索配方",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			results := reg.Search(args[0])
			if len(results) == 0 {
				fmt.Println("未找到匹配的配方")
				return
			}
			typeNames := recipeTypeNames()
			for _, r := range results {
				fmt.Printf("  %s  %s  %s\n", typeNames[r.Type], r.ID, mutedStyle.Render(r.Description))
			}
		},
	}
}

func listCmd(reg *recipe.Registry) *cobra.Command {
	var installedFlag bool
	var bundlesFlag bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出可用配方",
		Run: func(cmd *cobra.Command, args []string) {
			typeNames := recipeTypeNames()

			if bundlesFlag {
				for _, b := range reg.Bundles() {
					fmt.Printf("  📦 %s  %s  (%d 个工具)\n", b.ID, mutedStyle.Render(b.Description), len(b.Includes))
				}
				return
			}

			brew := installer.NewBrewInstaller()
			npm := installer.NewNpmInstaller()
			for _, r := range reg.All() {
				if installedFlag {
					switch r.Type {
					case recipe.TypeCLITool:
						if r.Install == nil {
							continue
						}
						if r.Install.Method == "npm" {
							if ok, _ := npm.IsInstalled(r.Install.Package); !ok {
								continue
							}
						} else {
							if ok, _ := brew.IsInstalled(r.Install.Package); !ok {
								continue
							}
						}
					default:
						continue
					}
				}
				fmt.Printf("  %s  %s  %s\n", typeNames[r.Type], r.ID, mutedStyle.Render(r.Description))
			}
		},
	}
	cmd.Flags().BoolVar(&installedFlag, "installed", false, "仅显示已安装的")
	cmd.Flags().BoolVar(&bundlesFlag, "bundles", false, "仅显示套餐")
	return cmd
}

func recipeTypeNames() map[recipe.RecipeType]string {
	return map[recipe.RecipeType]string{
		recipe.TypeCLITool: "🔧 CLI",
		recipe.TypeMCP:     "🔌 MCP",
		recipe.TypeSkill:   "📝 Skill",
		recipe.TypeCommand: "⌨️  Command",
		recipe.TypeRule:    "📏 Rule",
		recipe.TypeBundle:  "📦 Bundle",
	}
}

func doctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "检查开发环境状态",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println()
			fmt.Println("🔍 环境检查")
			fmt.Println()
			checks := []struct {
				name  string
				check func() (bool, string)
			}{
				{"Homebrew", func() (bool, string) {
					out, err := exec.Command("brew", "--version").Output()
					if err != nil {
						return false, ""
					}
					return true, strings.Split(strings.TrimSpace(string(out)), "\n")[0]
				}},
				{"Node.js", func() (bool, string) {
					out, err := exec.Command("node", "--version").Output()
					if err != nil {
						return false, ""
					}
					return true, strings.TrimSpace(string(out))
				}},
				{"Cursor", func() (bool, string) {
					paths := []string{"/Applications/Cursor.app", os.ExpandEnv("$HOME/Applications/Cursor.app")}
					for _, p := range paths {
						if _, err := os.Stat(p); err == nil {
							return true, "已安装"
						}
					}
					return false, ""
				}},
				{"Claude Code", func() (bool, string) {
					_, err := exec.LookPath("claude")
					if err != nil {
						return false, ""
					}
					return true, "已安装"
				}},
			}
			for _, c := range checks {
				ok, info := c.check()
				if ok {
					fmt.Printf("  %s  %s %s\n", successStyle.Render("✓"), c.name, mutedStyle.Render(info))
				} else {
					fmt.Printf("  %s  %s\n", errorStyle.Render("✗"), c.name)
				}
			}
			fmt.Println()
		},
	}
}

func configCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "管理配置",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "mode <beginner|advanced>",
		Short: "切换模式",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "beginner":
				_ = cfg.SetMode(config.ModeBeginner)
				fmt.Println(successStyle.Render("✓ 已切换到新手模式"))
			case "advanced":
				_ = cfg.SetMode(config.ModeAdvanced)
				fmt.Println(successStyle.Render("✓ 已切换到高级模式"))
			default:
				return fmt.Errorf("无效模式: %s（可选: beginner, advanced）", args[0])
			}
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "editor <cursor|claude-code|both>",
		Short: "设置 AI 工具",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			names := map[string]string{"cursor": "Cursor", "claude-code": "Claude Code", "both": "Cursor + Claude Code"}
			name, ok := names[args[0]]
			if !ok {
				return fmt.Errorf("无效选项: %s（可选: cursor, claude-code, both）", args[0])
			}
			_ = cfg.SetEditor(config.Editor(args[0]))
			fmt.Println(successStyle.Render("✓ AI 工具: " + name))
			return nil
		},
	})
	return cmd
}

func askEditor(cfg *config.Config) {
	fmt.Println()
	fmt.Println("你使用的 AI 编码工具是？")
	fmt.Println("  1) Cursor")
	fmt.Println("  2) Claude Code")
	fmt.Println("  3) 两个都用")
	fmt.Print("请选择 [1/2/3] (默认: 3): ")

	scanner := bufio.NewScanner(os.Stdin)
	choice := "3"
	if scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		if input != "" {
			choice = input
		}
	}

	editors := map[string]config.Editor{"1": config.EditorCursor, "2": config.EditorClaudeCode, "3": config.EditorBoth}
	names := map[string]string{"1": "Cursor", "2": "Claude Code", "3": "Cursor + Claude Code"}
	e, ok := editors[choice]
	if !ok {
		e = config.EditorBoth
		choice = "3"
	}
	_ = cfg.SetEditor(e)
	fmt.Println(successStyle.Render("✓ AI 工具: " + names[choice]))
	fmt.Println()
}

func collectPrompts(prompts []recipe.Prompt) map[string]string {
	if len(prompts) == 0 {
		return nil
	}
	values := make(map[string]string)
	scanner := bufio.NewScanner(os.Stdin)
	for _, p := range prompts {
		for {
			if p.Default != "" {
				fmt.Printf("%s (默认: %s): ", p.Ask, p.Default)
			} else {
				fmt.Printf("%s: ", p.Ask)
			}
			val := ""
			if scanner.Scan() {
				val = strings.TrimSpace(scanner.Text())
			}
			if val == "" {
				val = p.Default
			}
			if val == "" {
				fmt.Println(errorStyle.Render("  请输入一个值"))
				continue
			}
			values[p.Key] = val
			break
		}
	}
	return values
}

func updateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "更新配方索引",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("ℹ 配方更新功能将在远程仓库配置后启用")
		},
	}
}

func recipeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recipe",
		Short: "配方管理",
	}

	var recipeType string
	var fromBrewfile string
	var fromMCPConfig string
	genCmd := &cobra.Command{
		Use:   "generate [name]",
		Short: "自动生成配方",
		RunE: func(cmd *cobra.Command, args []string) error {
			gen, err := generator.New("recipes")
			if err != nil {
				return err
			}

			if fromBrewfile != "" {
				generated, err := gen.GenerateFromBrewfile(fromBrewfile)
				if err != nil {
					return err
				}
				fmt.Printf("\n✓ 已生成 %d 个配方\n", len(generated))
				return nil
			}

			if fromMCPConfig != "" {
				generated, err := gen.GenerateFromMCPConfig(fromMCPConfig)
				if err != nil {
					return err
				}
				fmt.Printf("\n✓ 已生成 %d 个配方\n", len(generated))
				return nil
			}

			if len(args) == 0 {
				return fmt.Errorf("请指定工具名称，例如: yduck recipe generate fzf")
			}

			name := args[0]
			fmt.Printf("🔍 正在采集 %s 的信息...\n", name)

			switch recipeType {
			case "mcp":
				rec, err := gen.GenerateMCP(name)
				if err != nil {
					return err
				}
				fmt.Println(successStyle.Render(fmt.Sprintf("✓ 已生成 recipes/mcps/%s.yaml", rec.ID)))
			default:
				rec, err := gen.GenerateCLITool(name)
				if err != nil {
					return err
				}
				fmt.Println(successStyle.Render(fmt.Sprintf("✓ 已生成 recipes/cli-tools/%s.yaml", rec.ID)))
			}
			return nil
		},
	}
	genCmd.Flags().StringVar(&recipeType, "type", "cli-tool", "配方类型 (cli-tool, mcp)")
	genCmd.Flags().StringVar(&fromBrewfile, "from-brewfile", "", "从 Brewfile 批量生成")
	genCmd.Flags().StringVar(&fromMCPConfig, "from-mcp-config", "", "从 mcp.json 批量生成")

	cmd.AddCommand(genCmd)
	return cmd
}
