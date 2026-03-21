#!/bin/bash

set -euo pipefail

VERSION="$1"
PACKAGE_NAME="@mailgo-cli/mailgo"
MAIN_PACKAGE_DIR="npm-package"
PLATFORM_PACKAGES_DIR="platform-packages"

rm -rf "$MAIN_PACKAGE_DIR" "$PLATFORM_PACKAGES_DIR"

mkdir -p "$MAIN_PACKAGE_DIR/bin" "$PLATFORM_PACKAGES_DIR"

declare -A PLATFORM_MAP=(
    ["mailgo_${VERSION}_Darwin_x86_64"]="darwin-x64"
    ["mailgo_${VERSION}_Darwin_arm64"]="darwin-arm64"
    ["mailgo_${VERSION}_Linux_x86_64"]="linux-x64"
    ["mailgo_${VERSION}_Linux_arm64"]="linux-arm64"
    ["mailgo_${VERSION}_Windows_x86_64"]="win32-x64"
    ["mailgo_${VERSION}_Windows_arm64"]="win32-arm64"
)

declare -A OS_MAP=(
    ["darwin-x64"]="darwin"
    ["darwin-arm64"]="darwin"
    ["linux-x64"]="linux"
    ["linux-arm64"]="linux"
    ["win32-x64"]="win32"
    ["win32-arm64"]="win32"
)

declare -A CPU_MAP=(
    ["darwin-x64"]="x64"
    ["darwin-arm64"]="arm64"
    ["linux-x64"]="x64"
    ["linux-arm64"]="arm64"
    ["win32-x64"]="x64"
    ["win32-arm64"]="arm64"
)

OPTIONAL_DEPS=""
for archive in dist/*.tar.gz dist/*.zip; do
    if [ -f "$archive" ]; then
        archive_name=$(basename "$archive")
        archive_name="${archive_name%.tar.gz}"
        archive_name="${archive_name%.zip}"

        platform_key="${PLATFORM_MAP[$archive_name]:-}"

        if [ -n "$platform_key" ]; then
            echo "Processing $archive for platform: $platform_key"

            platform_package_dir="$PLATFORM_PACKAGES_DIR/mailgo-$platform_key"
            mkdir -p "$platform_package_dir/bin"

            if [[ "$archive" == *.tar.gz ]]; then
                tar -xzf "$archive" -C "$platform_package_dir/bin"
            else
                unzip -j "$archive" -d "$platform_package_dir/bin"
            fi

            for doc_file in README.md README README.txt LICENSE LICENSE.md LICENSE.txt; do
                if [ -f "$platform_package_dir/bin/$doc_file" ]; then
                    mv "$platform_package_dir/bin/$doc_file" "$platform_package_dir/"
                fi
            done

            chmod +x "$platform_package_dir/bin/"* 2>/dev/null || true

            os_value="${OS_MAP[$platform_key]}"
            cpu_value="${CPU_MAP[$platform_key]}"

            files_array='["bin/"]'
            for doc_file in README.md README README.txt LICENSE LICENSE.md LICENSE.txt; do
                if [ -f "$platform_package_dir/$doc_file" ]; then
                    files_array="${files_array%]}, \"$doc_file\"]"
                fi
            done

            binary_name="mailgo"
            if [[ "$os_value" == "win32" ]]; then
                binary_name="mailgo.exe"
            fi

            cat > "$platform_package_dir/package.json" << EOF
{
  "name": "$PACKAGE_NAME-$platform_key",
  "version": "$VERSION",
  "description": "Platform-specific binary for $PACKAGE_NAME ($platform_key)",
  "os": ["$os_value"],
  "cpu": ["$cpu_value"],
  "bin": {
    "mailgo": "bin/$binary_name"
  },
  "files": $files_array,
  "repository": {
    "type": "git",
    "url": "https://github.com/Hrid-a/mailgo.git"
  },
  "author": "Hrid-a",
  "license": "MIT"
}
EOF

            if [ -n "$OPTIONAL_DEPS" ]; then
                OPTIONAL_DEPS="$OPTIONAL_DEPS,"
            fi
            OPTIONAL_DEPS="$OPTIONAL_DEPS\"$PACKAGE_NAME-$platform_key\": \"$VERSION\""
        fi
    fi
done

cat > "$MAIN_PACKAGE_DIR/bin/mailgo" << EOF
#!/usr/bin/env node

const { execFileSync } = require('child_process')

const packageName = '$PACKAGE_NAME'

const platformPackages = {
  'darwin-x64':   \`\${packageName}-darwin-x64\`,
  'darwin-arm64': \`\${packageName}-darwin-arm64\`,
  'linux-x64':    \`\${packageName}-linux-x64\`,
  'linux-arm64':  \`\${packageName}-linux-arm64\`,
  'win32-x64':    \`\${packageName}-win32-x64\`,
  'win32-arm64':  \`\${packageName}-win32-arm64\`
}

function getBinaryPath() {
  const platformKey = \`\${process.platform}-\${process.arch}\`
  const platformPackageName = platformPackages[platformKey]

  if (!platformPackageName) {
    console.error(\`mailgo: unsupported platform \${platformKey}\`)
    process.exit(1)
  }

  try {
    const binaryName = process.platform === 'win32' ? 'mailgo.exe' : 'mailgo'
    return require.resolve(\`\${platformPackageName}/bin/\${binaryName}\`)
  } catch (e) {
    console.error(\`mailgo: platform package \${platformPackageName} not found\`)
    process.exit(1)
  }
}

try {
  const binaryPath = getBinaryPath()
  execFileSync(binaryPath, process.argv.slice(2), { stdio: 'inherit' })
} catch (error) {
  process.exit(error.status ?? 1)
}
EOF

chmod +x "$MAIN_PACKAGE_DIR/bin/mailgo"

cat > "$MAIN_PACKAGE_DIR/package.json" << EOF
{
  "name": "$PACKAGE_NAME",
  "version": "$VERSION",
  "description": "Verify email deliverability via syntax checks and live SMTP handshake — catch bad addresses before you send",
  "bin": {
    "mailgo": "bin/mailgo"
  },
  "optionalDependencies": {
    $OPTIONAL_DEPS
  },
  "keywords": ["email", "smtp", "cli", "deliverability", "email-verification", "email-validation", "catch-all", "bulk-email", "mailbox", "go"],
  "author": "Hrid-a",
  "license": "MIT",
  "repository": {
    "type": "git",
    "url": "https://github.com/Hrid-a/mailgo.git"
  },
  "bugs": {
    "url": "https://github.com/Hrid-a/mailgo/issues"
  },
  "homepage": "https://github.com/Hrid-a/mailgo",
  "engines": {
    "node": ">=14.0.0"
  },
  "files": [
    "bin/",
    "README.md"
  ]
}
EOF

first_platform_dir=$(ls -1d "$PLATFORM_PACKAGES_DIR"/mailgo-* 2>/dev/null | head -1 || echo "")
if [ -n "$first_platform_dir" ] && [ -f "$first_platform_dir/README.md" ]; then
    cp "$first_platform_dir/README.md" "$MAIN_PACKAGE_DIR/"
fi
