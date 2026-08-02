# Arch Linux Package Validation

## Environment

- Distribution and version:
- Kernel:
- Desktop session: GNOME / KDE / other
- GTK4/WebKitGTK:
- Go/Node/pnpm:
- systemd:
- `/dev/net/tun` present:
- User belongs to `hypomux` group:

## Build

- [ ] `makepkg --verifysource`
- [ ] `makepkg -o`
- [ ] `makepkg -f`
- [ ] `namcap PKGBUILD`
- [ ] `namcap hypomux-*.pkg.tar.zst`
- [ ] Package file list contains no `.exe` or `.dll`.

## Runtime

- [ ] GUI starts as a normal user.
- [ ] `hypomux-core.service` starts and creates `/run/hypomux/hypomux-core.sock`.
- [ ] GUI connects to the core through the Unix socket.
- [ ] Adapter list shows IPv4, gateway, DNS, metric and interface type.
- [ ] Bound TCP diagnostics report the selected source address.
- [ ] System proxy mode saves and restores the gsettings snapshot.
- [ ] TUN preflight detects missing `/dev/net/tun` and route conflicts before changes.
- [ ] TUN mode starts and stops without leaving `HypoMux-Tun` or default routes.
- [ ] Killing the core triggers cleanup and leaves proxy mode usable.
- [ ] Uninstall stops the service and preserves `~/.hypomux`.
