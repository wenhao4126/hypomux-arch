#!/bin/sh
set -eu

systemd-sysusers /usr/lib/sysusers.d/hypomux.conf
systemd-tmpfiles --create /usr/lib/tmpfiles.d/hypomux.conf
systemctl daemon-reload
printf '%s\n' 'HypoMux installed. Add desktop users to the hypomux group, then enable hypomux-core.service.'
