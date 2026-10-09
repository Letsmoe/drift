#!/usr/bin/env bash
# Builds the gomobile AAR of core/mobile into apps/android/app/libs/drift-core.aar.
# Usage: scripts/build-android-core.sh [android-sdk-dir]
set -euo pipefail

repoRoot="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
outputAar="$repoRoot/apps/android/app/libs/drift-core.aar"

# Without an argument, take the first SDK with an NDK: a system SDK in
# ANDROID_HOME often lacks one while Android Studio's ~/Android/Sdk has it.
pickSdkDir() {
	local candidate
	for candidate in "${ANDROID_HOME:-}" "$HOME/Android/Sdk"; do
		if [[ -n "$candidate" && -d "$candidate/ndk" ]]; then
			echo "$candidate"
			return
		fi
	done
	echo "${ANDROID_HOME:-$HOME/Android/Sdk}"
}

sdkDir="${1:-$(pickSdkDir)}"

if [[ ! -d "$sdkDir" ]]; then
	echo "Android SDK not found at $sdkDir (pass it as an argument or set ANDROID_HOME)" >&2
	exit 1
fi

if [[ -z "${ANDROID_NDK_HOME:-}" ]]; then
	latestNdk="$(find "$sdkDir/ndk" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | sort -V | tail -n 1 || true)"
	if [[ -z "$latestNdk" ]]; then
		echo "No NDK under $sdkDir/ndk; install one or set ANDROID_NDK_HOME" >&2
		exit 1
	fi
	export ANDROID_NDK_HOME="$latestNdk"
fi
export ANDROID_HOME="$sdkDir"

export PATH="$(go env GOPATH)/bin:$PATH"
if ! command -v gomobile >/dev/null || ! command -v gobind >/dev/null; then
	echo "gomobile and gobind are required; install them with:" >&2
	echo "  go install golang.org/x/mobile/cmd/gomobile@latest golang.org/x/mobile/cmd/gobind@latest" >&2
	exit 1
fi

if [[ -f "$outputAar" ]]; then
	newerSource="$(find "$repoRoot/core" -type f \( -name '*.go' -o -name go.mod -o -name go.sum \) \
		! -name '*_test.go' -newer "$outputAar" -print -quit)"
	if [[ -z "$newerSource" ]]; then
		echo "drift-core.aar is up to date"
		exit 0
	fi
fi

mkdir -p "$(dirname "$outputAar")"
cd "$repoRoot/core"
gomobile bind \
	-target=android/arm64,android/amd64 \
	-androidapi 26 \
	-javapkg dev.letsmoe.drift \
	-ldflags='-s -w' \
	-trimpath \
	-o "$outputAar" \
	./mobile
echo "built $outputAar"
