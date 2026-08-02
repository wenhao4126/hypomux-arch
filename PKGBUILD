# Maintainer: HypoMux contributors
pkgname=hypomux
pkgver=2.5.2
pkgrel=1
pkgdesc='Multi-interface connection scheduler with proxy and Linux TUN modes'
arch=('x86_64')
url='https://github.com/wenhao4126/hypomux-arch'
license=('MIT')
depends=('gtk4' 'webkitgtk-6.0' 'sing-box' 'iproute2' 'polkit' 'systemd')
makedepends=('git' 'go' 'nodejs' 'pnpm')
source=("git+${url}.git#branch=main")
sha256sums=('SKIP')

build() {
  cd "${srcdir}/hypomux-arch"

  mkdir -p "${srcdir}/bin"
  GOBIN="${srcdir}/bin" go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-alpha2.119
  export PATH="${srcdir}/bin:${PATH}"
  pnpm --dir desktop/frontend install --frozen-lockfile
  (cd desktop && wails3 generate bindings -clean=true -ts -i)
  pnpm --dir desktop/frontend build

  go -C engine build -trimpath -buildvcs=false -ldflags="-s -w -X main.version=${pkgver}" \
    -o ../desktop/hypomux-engine ./cmd/hypomux-engine
  go -C desktop build -tags production -trimpath -buildvcs=false -ldflags="-s -w" \
    -o hypomux
}

check() {
  cd "${srcdir}/hypomux-arch"
  go -C engine test ./...
  go -C desktop test ./internal/engineclient ./internal/services ./internal/platform ./internal/startup
}

package() {
  cd "${srcdir}/hypomux-arch"

  install -Dm755 desktop/hypomux "${pkgdir}/usr/lib/hypomux/hypomux"
  install -Dm755 desktop/hypomux-engine "${pkgdir}/usr/lib/hypomux/hypomux-engine"
  install -Dm755 desktop/build/linux/hypomux-core.service \
    "${pkgdir}/usr/lib/systemd/system/hypomux-core.service"
  install -Dm644 desktop/build/linux/hypomux.desktop \
    "${pkgdir}/usr/share/applications/hypomux.desktop"
  install -Dm644 desktop/build/linux/io.hypomux.core.policy \
    "${pkgdir}/usr/share/polkit-1/actions/io.hypomux.core.policy"
  install -Dm644 desktop/build/linux/50-hypomux.rules \
    "${pkgdir}/usr/share/polkit-1/rules.d/50-hypomux.rules"
  install -Dm644 desktop/build/linux/hypomux.sysusers \
    "${pkgdir}/usr/lib/sysusers.d/hypomux.conf"
  install -Dm644 desktop/build/linux/hypomux.tmpfiles \
    "${pkgdir}/usr/lib/tmpfiles.d/hypomux.conf"
  install -Dm644 LICENSE "${pkgdir}/usr/share/licenses/${pkgname}/LICENSE"
  install -Dm644 desktop/build/appicon.png \
    "${pkgdir}/usr/share/icons/hicolor/256x256/apps/hypomux.png"
  install -Dm755 desktop/build/linux/install.sh "${pkgdir}/usr/share/doc/${pkgname}/install.sh"
  install -Dm755 desktop/build/linux/remove.sh "${pkgdir}/usr/share/doc/${pkgname}/remove.sh"

  install -d "${pkgdir}/usr/bin"
  ln -s /usr/lib/hypomux/hypomux "${pkgdir}/usr/bin/hypomux"
}
