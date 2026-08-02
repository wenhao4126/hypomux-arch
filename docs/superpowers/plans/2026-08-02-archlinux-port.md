# Arch Linux Desktop Port Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 Arch Linux 上构建并安装完整的 HypoMux GUI、Linux 核心服务、系统代理和 TUN 模式，同时保持 Windows 构建行为不变。

**Architecture:** 保留 protocol v1 和现有 React 前端；Linux GUI 通过 Unix domain socket 连接 systemd 管理的 `hypomux-core`，核心使用 Linux TUN、netlink 和 sing-box。平台差异全部放入 `*_linux.go` 或明确的运行时适配层，Arch 包只依赖 Linux 原生运行时。

**Tech Stack:** Go 1.26、Wails v3、React/TypeScript/Vite、Unix domain socket、Linux netlink、systemd、polkit、Arch `makepkg`。

---

## 文件地图

- `engine/cmd/hypomux-engine/pipe_linux.go`：Linux 核心 IPC socket 监听/认证。
- `engine/cmd/hypomux-engine/service_linux.go`：Linux service 模式和生命周期。
- `engine/internal/platform/*_linux.go`、`engine/internal/tun/*_linux.go`：Linux 身份、TUN 设备、进程清理和路由适配。
- `desktop/internal/engineclient/*_linux.go`：GUI 到核心服务的 socket 客户端和服务启动检查。
- `desktop/internal/services/*_linux.go`：网卡元数据、诊断、系统代理、TUN 预检和自启动。
- `desktop/build/linux/*`：桌面文件、systemd unit、polkit 和图标安装资源。
- `PKGBUILD`：源码构建、依赖声明、安装文件和验证。
- `README.md`、`desktop/README.md`、`engine/README.md`：Linux 开发、运行和权限说明。

## Task 1: Linux 构建基线与 Unix IPC

**Files:**
- Create: `engine/cmd/hypomux-engine/pipe_linux.go`
- Create: `engine/cmd/hypomux-engine/service_linux.go`
- Create: `desktop/internal/engineclient/service_linux.go`
- Create: `desktop/internal/engineclient/privileged_linux.go`
- Modify: `engine/cmd/hypomux-engine/main.go`
- Test: `engine/cmd/hypomux-engine/pipe_linux_test.go`, `desktop/internal/engineclient/service_linux_test.go`

- [ ] 写 Linux socket 认证测试：服务端创建仅当前用户可读写的 Unix socket，客户端发送已有 session token，错误 token 或非预期 PID/UID 被拒绝。
- [ ] 运行 `go test ./cmd/hypomux-engine ./internal/...`（`cd engine`），确认新测试先失败于 Linux transport 未实现。
- [ ] 实现 `connectAuthenticatedPipe` 的 Linux 分支，使用 `net.ListenUnix`/`net.DialUnix`、socket 文件权限 `0600`、一次性 token 握手，并复用现有协议帧。
- [ ] 将 Linux 核心服务接入现有 `serve` 主循环；`service` 命令以前台运行，systemd 负责守护，不新增 Windows Service API。
- [ ] 让 desktop Linux client 优先连接标准 socket；服务不可用时返回结构化的“核心服务未运行”，不请求 GUI root 权限。
- [ ] 运行 `go test ./...`（engine、desktop）和 `GOOS=linux GOARCH=amd64 go build ./cmd/hypomux-engine`，确认 Windows 文件仍由 build tags 排除。
- [ ] 提交：`feat: add linux unix socket core transport`。

## Task 2: Linux GUI 平台与网卡/代理能力

**Files:**
- Create: `desktop/internal/services/adapter_metadata_linux.go`
- Create: `desktop/internal/services/diagnostic_probe_linux.go`
- Create: `desktop/internal/services/system_proxy_linux.go`
- Create: `desktop/internal/platform/autostart_linux.go`
- Create: `desktop/internal/platform/wails/appearance_linux.go`
- Create: `desktop/internal/webview_linux.go`（Linux WebKitGTK availability check）
- Modify: `desktop/internal/services/adapters.go`, `desktop/main.go`
- Test: 对应 `*_linux_test.go`

- [ ] 为接口 metric、默认网关、DNS 和接口类型写 fixture 测试，覆盖有线、无线、无默认网关和 HypoMux-Tun 排除。
- [ ] 运行 Linux 单测确认 metadata、绑定 TCP/UDP 诊断和系统代理实现缺失。
- [ ] 使用 `/sys/class/net`、`/proc/net/route` 或 netlink 实现 `adapterPlatformMetadata`，不得依赖命令输出解析作为唯一数据源。
- [ ] 实现 Linux 绑定诊断：按源地址创建 TCP/UDP socket，返回可展示的错误类别和接口名称。
- [ ] 实现 GNOME/KDE 常见 `gsettings` 代理读写，保存旧值并在停止/异常退出恢复；桌面环境不支持时返回明确错误。
- [ ] 实现 XDG autostart 文件和 Linux GUI 主题/窗口降级；移除 Linux 启动路径对 WebView2 的阻断。
- [ ] 运行 `go test ./...`、`pnpm --dir desktop/frontend build` 和 Linux GUI `go build`，确认前端无需平台专属改动。
- [ ] 提交：`feat: implement linux desktop and proxy adapters`。

## Task 3: Linux TUN、路由和核心服务权限

**Files:**
- Create: `engine/internal/tun/device_linux.go`
- Create: `engine/internal/tun/cleanup_linux.go`
- Create: `engine/internal/tun/process_linux.go`
- Create: `desktop/internal/services/tun_preflight_linux.go`
- Create: `desktop/internal/services/tun_config_linux_test.go`
- Modify: `engine/internal/tun/supervisor.go`, `engine/internal/tun/recover.go`, `desktop/internal/services/tun_config.go`, `desktop/internal/services/tun_preflight.go`
- Test: Linux TUN lifecycle and route ownership tests

- [ ] 写纯函数测试：Linux TUN 名称、路由表/metric、资源 owner 标记和回滚顺序；测试不得修改真实默认路由。
- [ ] 运行单测确认 Linux 预检当前被 `tun_preflight_other.go` 阻断。
- [ ] 使用 `/dev/net/tun` 和 netlink 创建/检查 `HypoMux-Tun`，为 HypoMux 创建的路由保存精确句柄/标记，清理时只删除这些资源。
- [ ] 实现 Linux 进程 containment、SIGTERM/SIGKILL 超时、sing-box check 和 TUN 接口稳定性探测。
- [ ] 将 WFP 字段映射为 Linux route safety 状态，保持现有 JSON 字段兼容，但在 UI 中展示 Linux 具体原因。
- [ ] 将 `resolveSingBox` 从固定 `sing-box.exe` 改为平台运行时解析；Linux 使用 PATH 或包安装路径，Windows 保留现有资源名。
- [ ] 运行无特权单测、root/`CAP_NET_ADMIN` 集成测试（若环境有 `/dev/net/tun`），并验证失败时代理模式仍可启动。
- [ ] 提交：`feat: add linux tun and route lifecycle`。

## Task 4: Arch 软件包与 systemd/polkit 集成

**Files:**
- Create: `PKGBUILD`
- Create: `desktop/build/linux/hypomux-core.service`
- Create: `desktop/build/linux/hypomux.desktop`
- Create: `desktop/build/linux/io.hypomux.core.policy`
- Create: `desktop/build/linux/install.sh`, `desktop/build/linux/remove.sh`
- Create: `desktop/build/linux/icons/hicolor/*/apps/hypomux.png`
- Modify: `.gitignore`, `README.md`, `desktop/README.md`

- [ ] 写 `PKGBUILD` 源码校验和依赖，构建 frontend、desktop GUI、engine，并显式排除 `bin/*.exe`、`bin/*.dll`。
- [ ] 写 package file-list 检查，要求包含 GUI、engine、`.desktop`、systemd unit、polkit action 和图标，且不包含 Windows 二进制。
- [ ] 实现 systemd unit：核心以前台模式运行，设置 `AmbientCapabilities=CAP_NET_ADMIN`、`DeviceAllow=/dev/net/tun`、私有临时目录和 Unix socket 路径。
- [ ] 实现 polkit 规则/服务管理边界，安装卸载脚本只 reload unit、启停本包服务和创建标准目录，不删除用户配置。
- [ ] 运行 `makepkg --verifysource`、`makepkg -o`、`makepkg -f`、`namcap PKGBUILD` 和 `namcap *.pkg.tar.zst`。
- [ ] 在 systemd 环境安装包，验证 GUI -> socket -> core、代理模式、TUN 启停、杀死核心后的恢复和卸载。
- [ ] 提交：`feat: package hypomux for arch linux`。

## Task 5: 文档、回归和发布验收

**Files:**
- Modify: `README.md`, `README_EN.md`, `desktop/README.md`, `engine/README.md`
- Create: `docs/validation/linux/arch-package.md`
- Modify: `.github/workflows/go-engine.yml`, `.github/workflows/linux-desktop.yml`

- [ ] 文档明确 Arch 依赖、安装命令、systemd/polkit 权限、`/dev/net/tun` 要求、代理/TUN 边界和卸载恢复。
- [ ] 添加 Linux 验证记录模板，包含内核、桌面环境、Wails/WebKitGTK、systemd、代理、TUN 和回滚结果。
- [ ] 运行 engine/desktop 全量测试、`go vet`、前端 production build、Linux amd64 build 和 `makepkg` 验证。
- [ ] 运行 Windows build-tag 回归测试；确认 Windows NSIS 任务和 `bin` 资源路径未被 Linux 改动破坏。
- [ ] 执行 `git diff --check`、审查包文件列表和用户配置保留行为，再提交：`docs: document arch linux support`。
