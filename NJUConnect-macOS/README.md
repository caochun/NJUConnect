# NJUConnect macOS 菜单栏应用

类似 Tailscale 的原生 macOS 菜单栏应用，用于管理 NJUConnect VPN 连接。

## 功能

- 🎯 菜单栏图标，点击显示控制面板
- ✅ 显示连接状态（已连接/未连接）
- ⚙️ 配置服务器、用户名、密码、SOCKS5 端口
- 🔌 一键连接/断开
- 💾 自动保存设置

## 构建步骤

### 1. 编译 njuconnect 二进制

```bash
cd /Users/chun/Develop/NJUConnect
go build -o njuconnect .
```

### 2. 使用 Xcode 构建应用

1. 打开 Xcode
2. File → New → Project
3. 选择 macOS → App
4. 项目名称: `NJUConnect-macOS`
5. Interface: SwiftUI
6. Language: Swift
7. 将以下文件添加到项目:
   - NJUConnectApp.swift
   - MenuView.swift
   - SettingsView.swift
   - ConnectionManager.swift
   - Info.plist

### 3. 添加 njuconnect 二进制到应用包

1. 将编译好的 `njuconnect` 拖入 Xcode 项目
2. 确保在 "Target Membership" 中勾选
3. 在 Build Phases → Copy Bundle Resources 中确认 njuconnect 已添加

### 4. 配置项目设置

- Deployment Target: macOS 13.0+
- Bundle Identifier: com.njuconnect.macos
- Signing: 使用你的开发者证书

### 5. 构建并运行

```bash
# 在 Xcode 中按 Cmd+R 运行
# 或使用命令行:
xcodebuild -project NJUConnect-macOS.xcodeproj -scheme NJUConnect-macOS
```

## 使用方法

1. 启动应用后，菜单栏会出现网络图标
2. 点击图标打开控制面板
3. 点击"设置"配置服务器信息
4. 点击"连接"建立 VPN 连接
5. 连接成功后显示绿色状态

## 默认配置

- 服务器: vpn.nju.edu.cn:443
- SOCKS5 端口: :1080

## 注意事项

- 应用以 LSUIElement 模式运行（无 Dock 图标）
- 设置会自动保存到 UserDefaults
- njuconnect 进程在应用退出时自动终止
