#!/bin/sh
PAK_DIR="$(dirname "$0")"
cd "$PAK_DIR" || exit 1

export LD_LIBRARY_PATH=$PAK_DIR/resources/lib:$LD_LIBRARY_PATH

LOG_DIR="${LOGS_PATH:-/mnt/SDCARD/.userdata/$PLATFORM/logs}"
mkdir -p "$LOG_DIR"

./pak-store 2>>"$LOG_DIR/pak_store.stderr.log"
