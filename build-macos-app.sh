#!/bin/bash
# 自动构建 NJUConnect macOS 应用

set -e

echo "🔨 开始构建 NJUConnect macOS 应用..."

# 1. 编译 njuconnect 二进制
echo "📦 编译 njuconnect..."
cd /Users/chun/Develop/NJUConnect
go build -o njuconnect .

# 2. 创建应用包结构
echo "📁 创建应用包..."
APP_DIR="NJUConnect.app"
rm -rf "$APP_DIR"
mkdir -p "$APP_DIR/Contents/MacOS"
mkdir -p "$APP_DIR/Contents/Resources"

# 3. 复制二进制
cp njuconnect "$APP_DIR/Contents/Resources/"

# 4. 编译 Swift 代码
echo "🔧 编译 Swift 代码..."
cd NJUConnect-macOS/NJUConnect-macOS
swiftc -o ../../NJUConnect.app/Contents/MacOS/NJUConnect \
    -target arm64-apple-macos13.0 \
    -framework SwiftUI \
    -framework AppKit \
    NJUConnectApp.swift \
    MenuView.swift \
    SettingsView.swift \
    SMSCodeView.swift \
    ConnectionManager.swift

# 5. 创建正确的 Info.plist
cd /Users/chun/Develop/NJUConnect
cat > "$APP_DIR/Contents/Info.plist" << 'PLIST'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleExecutable</key>
	<string>NJUConnect</string>
	<key>CFBundleIdentifier</key>
	<string>com.njuconnect.macos</string>
	<key>CFBundleName</key>
	<string>NJUConnect</string>
	<key>CFBundlePackageType</key>
	<string>APPL</string>
	<key>CFBundleShortVersionString</key>
	<string>1.0</string>
	<key>CFBundleVersion</key>
	<string>1</string>
	<key>LSMinimumSystemVersion</key>
	<string>13.0</string>
	<key>LSUIElement</key>
	<true/>
</dict>
</plist>
PLIST

echo "✅ 构建完成！"
echo "📍 应用位置: /Users/chun/Develop/NJUConnect/NJUConnect.app"
echo "🚀 运行: open NJUConnect.app"
