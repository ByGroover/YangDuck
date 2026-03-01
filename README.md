# 🐤 YangDuck (yduck)

快速配置你的 Mac 开发环境 + AI 编码工具。

一键安装 CLI 工具、配置 MCP 服务器、Cursor Skills/Commands/Rules，面向新手的傻瓜式体验。

## 安装

### Homebrew（推荐）

```bash
brew install tc6-01/yduck/yduck

# 更新到最新版
brew update && brew upgrade yduck
```

### 一键脚本

```bash
curl -fsSL https://raw.githubusercontent.com/tc6-01/YangDuck/master/install.sh | sh
```

### 手动安装

```bash
# macOS arm64 (Apple Silicon)
curl -L -o /usr/local/bin/yduck https://github.com/tc6-01/YangDuck/releases/latest/download/yduck-darwin-arm64
chmod +x /usr/local/bin/yduck

# macOS amd64 (Intel)
curl -L -o /usr/local/bin/yduck https://github.com/tc6-01/YangDuck/releases/latest/download/yduck-darwin-amd64
chmod +x /usr/local/bin/yduck
```

## 快速开始

```bash
# 进入交互式界面（推荐新手）
yduck

# 环境检查
yduck doctor
```

## 命令参考

### 全局选项

| 选项 | 说明 |
|------|------|
| `-v, --verbose` | 启用详细日志输出 |

### `yduck`

无参数运行，进入交互式 TUI 界面。新手首次使用会进入引导流程（选择身份、AI 工具偏好）。

### `yduck install`

安装配方（CLI 工具、MCP、Skill、Command、Rule）。

```bash
yduck install fzf                       # 安装单个工具
yduck install fzf ripgrep bat           # 安装多个工具
yduck install claude-code               # 安装 Claude Code CLI
yduck install --bundle cli-essentials   # 安装套餐
yduck install mysql-mcp                 # 安装 MCP（会通过 stdin 收集连接参数）
```

安装 MCP 配方时，如果尚未设置 AI 工具偏好，会自动询问你使用 Cursor 还是 Claude Code。

### `yduck search`

按关键字搜索配方，直接打印匹配结果。

```bash
yduck search database
yduck search mcp
```

### `yduck list`

列出可用配方。

```bash
yduck list                # 列出所有配方
yduck list --installed    # 仅显示已安装的
yduck list --bundles      # 仅显示套餐
```

### `yduck doctor`

检查开发环境状态（Homebrew、Node.js、Cursor、Claude Code）。

### `yduck config`

管理配置。

```bash
yduck config mode advanced       # 切换到高级模式
yduck config mode beginner       # 切换回新手模式
yduck config editor cursor       # 设置 AI 工具为 Cursor
yduck config editor claude-code  # 设置 AI 工具为 Claude Code
yduck config editor both         # 设置两个都用
```

### `yduck recipe generate`

自动采集工具信息并生成配方 YAML。

```bash
yduck recipe generate fzf                                    # 生成 CLI 工具配方
yduck recipe generate --type mcp @some/mcp-pkg               # 生成 MCP 配方
yduck recipe generate --from-brewfile ~/Brewfile              # 从 Brewfile 批量生成
yduck recipe generate --from-mcp-config .cursor/mcp.json     # 从 MCP 配置批量生成
```

### `yduck update`

更新配方索引（远程仓库配置后启用）。

## 功能

### CLI 工具安装

通过 Homebrew 或 npm 安装常用命令行工具，安装后自动展示使用教程（新手模式）：

- fzf, ripgrep, bat, eza, jq, lazygit, tldr, httpie, fd
- **Claude Code** — Anthropic 官方 AI 编码助手 CLI（通过 npm 安装）

### MCP 服务器配置

一键配置 MCP 服务器到 **Cursor** 和 **Claude Code**：

- **MySQL MCP** — 让 AI 直接操作数据库
- **Filesystem MCP** — 让 AI 读写文件
- **Fetch MCP** — 让 AI 访问网页和 API

安装目标根据用户配置自动选择：

| 目标 | 配置路径 | 作用范围 |
|------|----------|----------|
| Cursor | `.cursor/mcp.json` | 当前项目 |
| Claude Code | `~/.claude.json` | 全局 |

### Cursor 扩展

支持安装 Cursor 的 Skill、Command 和 Rule 配方。

### 套餐安装

按场景组合的工具包，一键安装多个工具：

- **CLI 必备工具包** (`cli-essentials`) — fzf, ripgrep, bat, eza, jq, fd, lazygit, tldr
- **AI 入门套餐** (`ai-starter`) — Fetch MCP, Filesystem MCP

### 配方生成器

自动采集工具信息 + AI 生成配方，批量创建无需手写。支持从 Brewfile 和 MCP 配置文件批量导入。

## 支持的 AI 工具

| 工具 | 安装 | MCP 配置 |
|------|------|----------|
| Cursor | — | ✓ 写入 `.cursor/mcp.json` |
| Claude Code | ✓ `yduck install claude-code` | ✓ 写入 `~/.claude.json` |

首次安装 MCP 配方时会询问你使用的 AI 工具（Cursor / Claude Code / 两个都用），选择会保存到 `~/.yduck/config.yaml`，也可以通过 `yduck config editor` 随时修改。

## 双模式

- **新手模式**（默认）：逐步引导、术语解释、安装后教程
- **高级模式**：精简输出、批量安装、跳过引导

## 交互模式 vs 命令行模式

| 入口 | 模式 | 说明 |
|------|------|------|
| `yduck` | TUI 交互界面 | 浏览、搜索、安装，全在界面里完成 |
| `yduck <子命令>` | 纯文本输出 | 直接执行，适合知道自己要干啥的用户 |

## 配方贡献

配方是 YAML 文件，定义了工具的安装方式和使用引导。欢迎贡献！

1. Fork 仓库
2. 在 `recipes/` 下添加 YAML 配方（可用 `yduck recipe generate` 自动生成初稿）
3. 确保符合 `schemas/` 中的 JSON Schema
4. 提交 PR

配方变更合入 master 后会通过 GitHub Actions 自动构建发布。

## 项目结构

```
yduck-cli/
├── cmd/
│   ├── yduck/             # 主程序入口
│   └── tuidebug/          # TUI 调试工具
├── internal/
│   ├── config/            # 配置管理
│   ├── generator/         # 配方生成器（采集 + AI 生成）
│   ├── installer/         # 安装器（Brew、Npm、MCP、Skill、Command、Rule、Bundle）
│   ├── log/               # 日志
│   ├── quickstart/        # 新手引导
│   ├── recipe/            # 配方加载与校验
│   └── tui/               # 交互式界面
├── recipes/               # 配方源文件
│   ├── bundles/           # 套餐
│   ├── cli-tools/         # CLI 工具配方
│   ├── mcps/              # MCP 服务器配方
│   └── skills/            # Cursor Skill 配方
├── schemas/               # JSON Schema（cli-tool, mcp, skill, bundle）
├── Formula/               # Homebrew Formula 模板
├── .github/workflows/     # CI/CD（自动构建 + 发布 + 更新 Homebrew tap）
├── scripts/               # 构建脚本
└── install.sh             # 一键安装脚本
```

## 开发

```bash
go build ./cmd/yduck/           # 构建
go run ./cmd/yduck/             # 运行
go run scripts/build-index.go   # 构建配方索引
```

## 发布流程

代码推送到 master 后，GitHub Actions 自动：

1. 构建 macOS arm64 / amd64 二进制
2. 创建 GitHub Release 并上传构建产物
3. 更新 Homebrew tap（`tc6-01/homebrew-yduck`）

## 致谢

- [Cobra](https://github.com/spf13/cobra) — CLI 框架
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI 框架
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — 终端样式
- [Huh](https://github.com/charmbracelet/huh) — 终端表单
- [Zap](https://go.uber.org/zap) — 结构化日志
- [gojsonschema](https://github.com/xeipuuv/gojsonschema) — JSON Schema 校验

## License

MIT
