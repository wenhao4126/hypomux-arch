#!/bin/sh
set -eu

systemctl disable --now hypomux-core.service 2>/dev/null || true
systemctl daemon-reload
printf '%s\n' 'HypoMux service stopped. User configuration under ~/.hypomux was preserved.'
