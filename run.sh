#!/bin/bash
set -u
set -e
set -x

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
METADATA_FILE="pool_meta"
DATA_FILE="pool_data"
POOL_NAME="test_pool"
IMAGECACHE_DIR="$SCRIPT_DIR/imagecache"
STATE_DIR="$SCRIPT_DIR/state"

function init() {
	if [ -n "$METADATA_DEV" ]; then
		echo "already init. exit."
		exit 1
	fi
	local METADATA_DEV
	local DATA_DEV
	fallocate -l 1M "$METADATA_FILE"
	fallocate -l 2G "$DATA_FILE"
	METADATA_DEV="$(losetup -f --show "$METADATA_FILE" 2>/dev/null)"
	DATA_DEV="$(losetup -f --show "$DATA_FILE" 2>/dev/null)"
	dmsetup create --verifyudev "$POOL_NAME" --table "0 4194304 thin-pool ${METADATA_DEV} ${DATA_DEV} 2048 32768"
	mkdir -p "$IMAGECACHE_DIR"
	mkdir -p "$STATE_DIR"
	ln -s "$IMAGECACHE_DIR" "$STATE_DIR"/"imagecache"
}


function deinit() {
	set +e
	dmsetup remove "$POOL_NAME"
	if [ -z "$METADATA_DEV" ]; then
		losetup -d $METADATA_DEV
		losetup -d $DATA_DEV
	fi
	rm "$METADATA_FILE"
	rm "$DATA_FILE"
	set -e
}


function run() {
	go run ./src
}

function clean() {
	rm -rf ./state
}


pushd "$SCRIPT_DIR"

METADATA_DEV="$(losetup -j "$METADATA_FILE" 2>/dev/null | awk -F':' '{print $1}' | head -n1)"
DATA_DEV="$(losetup -j "$DATA_FILE" 2>/dev/null | awk -F':' '{print $1}' | head -n1)"

cmd="$1"; shift

if [ -z "$METADATA_DEV" ] && [ ! "$cmd" -eq "init" ]; then
	echo "Run init first. Exit.";
	exit 1;
fi

$cmd "$@"
