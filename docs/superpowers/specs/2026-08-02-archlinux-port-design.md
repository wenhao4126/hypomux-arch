# Arch Linux Desktop Port Design

## Goal

让 HypoMux 在 Arch Linux 上提供可安装、可启动的桌面 GUI，并支持系统代理模式与基于 Linux TUN/netlink 的虚拟网卡模式。Windows 版本继续使用现有 Windows 实现，Linux 版本不携带 Windows 专属运行时。

## Scope

### Included

- Linux `amd64` 原生构建与 Arch `PKGBUILD`。
- React/Wails GUI 在 Linux 上的窗口、托盘、主题、用户目录和开机自启适配。
- GUI 与核心之间的 Unix domain socket IPC。
- 由 systemd 管理的 `hypomux-core.service`，运行网络核心并获得最小必要的 `CAP_NET_ADMIN`。
- Linux 网卡枚举、源地址绑定、网关/DNS/接口指标读取和绑定 TCP/UDP 诊断。
- 使用 Linux TUN 设备、sing-box Linux 运行时和 netlink 路由完成 TUN 模式。
- Linux 本地系统代理后端，优先支持 GNOME/KDE 常见的 `gsettings` 设置路径；无法识别桌面环境时明确报告不支持。
- `.desktop` 文件、图标、systemd unit、polkit 规则和 Arch 包安装/卸载脚本。

### Excluded

- 把 Windows WFP 代码移植成同名实现。Linux 使用明确的路由、防泄漏检查和 TUN 生命周期模型。
- 继续分发 `sing-box.exe`、`wintun.dll` 或 `libcronet.dll`。
- 通过 GUI 直接获取 root 权限。GUI 始终普通用户运行。
- 同时支持所有 Linux 桌面环境和发行版。首版目标是 Arch Linux、systemd、主流 GNOME/KDE 会话。
- Linux 软件自动更新安装器。Arch 更新由 pacman/AUR 管理。

## Architecture

```text
React/Wails GUI (普通用户)
        |
        | Unix domain socket, user/group ACL
        v
hypomux-core.service (systemd, CAP_NET_ADMIN)
        |
        +-- hypomux-engine: proxy, DNS, connection scheduling
        +-- sing-box: Linux TUN sidecar
        +-- /dev/net/tun and netlink routes
```

现有 protocol v1 保持不变。Windows 继续使用 named pipe 和 Windows Service；Linux 使用独立的 `*_linux.go` 实现，公共业务代码只依赖平台抽象。核心服务的 socket 路径、运行时目录和配置目录使用标准 Linux 路径，避免写入安装目录。

### Privilege boundary

systemd unit 使用专用服务用户或受控服务身份，并只授予 TUN、路由所需能力。socket 不监听 TCP，不允许远程访问。安装时创建服务所需的组/目录；运行时由 polkit/systemd 管理服务启动，卸载时只清理本包创建的 unit、polkit 文件和缓存，不删除用户配置。

### TUN lifecycle

启动前只读检查：核心可用、sing-box 可用、`/dev/net/tun` 可用、选定网卡有源地址、没有冲突的默认路由/TUN、当前配置可生成并通过 sing-box check。启动过程按“核心代理 -> 出站池 -> TUN -> 路由”顺序进行；任一步失败，按逆序停止并恢复 HypoMux 创建的资源。停止、崩溃和服务重启都只清理精确标记为 HypoMux 所有的 TUN、路由和 DNS 状态。

严格路由的 Linux 实现不依赖 WFP，而是检查默认路由、策略路由、route metric 和 TUN 路由表的一致性；不满足条件时将严格路由降级为关闭，并向 GUI 返回结构化原因，不静默接管网络。

### Adapter metadata

Linux 实现读取 `net.Interfaces`、`/sys/class/net/<name>` 和 netlink 信息，填充现有 `AdapterView`：接口名称、MAC、IPv4/IPv6、ifindex、网关、DNS、metric、接口类型和运行状态。源地址绑定继续由 Go socket 完成；Linux 不使用 Windows 的 `IP_UNICAST_IF`。

### Desktop integration

Linux 平台实现补齐窗口关闭/托盘、原生主题降级、用户配置目录、开机自启和 WebKitGTK 可用性提示。前端不新增平台专属网络逻辑，只消费现有服务和结构化状态。

## Packaging

新增 Arch 包至少包含：

- `PKGBUILD`：从源码构建前端、桌面端和引擎，使用系统 Go/Node/pnpm/Wails 构建依赖。
- `hypomux-core.service`：安装到 `/usr/lib/systemd/system/`。
- `hypomux.desktop` 和图标：安装到 `/usr/share/applications/` 与 `/usr/share/icons/hicolor/`。
- polkit 规则或 action：只授权当前登录用户管理 HypoMux 核心服务。
- Linux 运行时目录与配置目录创建/升级/卸载逻辑。

包不从仓库 `bin/` 复制 Windows 运行时。sing-box Linux 依赖使用 Arch 包依赖或构建阶段的明确版本化依赖，不能下载未经校验的二进制。

## Error handling

- 权限、设备节点、路由冲突、运行时缺失和桌面集成失败都返回结构化错误。
- TUN 启动失败必须保证 GUI 能继续使用普通代理模式。
- 不识别的桌面环境只禁用系统代理集成，不阻止核心和 TUN 使用。
- 服务停止/崩溃后的恢复操作必须幂等，并记录可导出的诊断日志。

## Verification

- `go test ./...` 和 `go vet ./...`：engine 与 desktop Linux 构建路径。
- 前端 TypeScript/Vite production build。
- `go build` 生成 GUI 和 `hypomux-engine` Linux `amd64` 可执行文件。
- 在 Arch 或等价 systemd 环境验证：安装、启动 GUI、socket 连接、systemd 服务、代理模式、TUN 启停、异常退出恢复和卸载。
- `makepkg --verifysource`、`makepkg -o`、`makepkg -f`，并检查包文件列表不包含 `.exe`/`.dll`。
- 保留 Windows 构建和现有 Windows 测试路径，不以 Linux 适配替换 Windows 行为。

## Delivery boundaries

实现按以下可独立验证的阶段推进：

1. Linux 构建基线、Wails GUI 平台适配和 Unix socket IPC。
2. Linux 网卡/DNS/诊断与系统代理。
3. Linux TUN、netlink 路由、恢复和 systemd 核心服务。
4. Arch `PKGBUILD`、桌面文件、polkit、安装卸载验证和文档。
