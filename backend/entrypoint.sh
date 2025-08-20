#!/bin/sh
set -e

# Ensure data directory exists
mkdir -p /app/data

# Try to chown the directory to the teamspeak user. If the mount forbids chown,
# ignore the error but let the subsequent permission checks proceed.
chown -R teamspeak:teamspeak /app/data 2>/dev/null || true

# Ensure directory is writable by the teamspeak user (fallback: add group/world write)
chmod -R u+rwX,g+rwX,o+rwX /app/data || true

# If su-exec is available, use it to drop privileges, otherwise fallback to su
if command -v su-exec >/dev/null 2>&1; then
  exec su-exec teamspeak:teamspeak "$@"
else
  exec su -s /bin/sh teamspeak -c "$*"
fi
