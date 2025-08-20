#!/bin/sh
set -e

# Ensure data directories exist (support both /app/data and /data)
mkdir -p /app/data
mkdir -p /data

# Try to chown the directories to the teamspeak user. If the mount forbids chown,
# ignore the error but let the subsequent permission checks proceed.
chown -R teamspeak:teamspeak /app/data 2>/dev/null || true
chown -R teamspeak:teamspeak /data 2>/dev/null || true

# Ensure directories are writable by the teamspeak user (fallback: add group/world write)
chmod -R u+rwX,g+rwX,o+rwX /app/data || true
chmod -R u+rwX,g+rwX,o+rwX /data || true

# Handle the case where /app/data is a mount point that prevents chown/chmod
# Create a fallback directory that is definitely writable
if [ ! -w "/app/data" ]; then
    mkdir -p /tmp/data
    chmod 777 /tmp/data
    export DATA_DIR="/tmp/data"
else
    export DATA_DIR="/app/data"
fi

# If su-exec is available, use it to drop privileges, otherwise fallback to su
if command -v su-exec >/dev/null 2>&1; then
  exec su-exec teamspeak:teamspeak "$@"
else
  exec su -s /bin/sh teamspeak -c "$*"
fi