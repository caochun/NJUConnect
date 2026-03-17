# 进展报告 2026-03-17

## 已完成

### 1. 依赖全面升级
将所有依赖从 2022 年版本升级至最新，解决 Go 1.26 无法编译的问题。

| 依赖 | 旧版本 | 新版本 |
|------|--------|--------|
| Go | 1.19 | 1.26 |
| gvisor | 2022.9 snapshot | 2025.3 snapshot |
| fyne | v2.3.0 | v2.7.3 |
| utls | v1.2.0 | v1.8.2 |
| tailscale | v1.34.2 | v1.96.1 |
| otp | v1.4.0 | v1.5.0 |
| golang.org/x/* | 2022 版 | 2025/2026 版 |

### 2. API 适配
- **core/tun_stack.go**: `bufferv2` 包重命名为 `buffer`; `tcpip.Address()` 改为 `tcpip.AddrFromSlice()`; 新增 `Close()`, `SetMTU()`, `SetLinkAddress()`, `SetOnCloseAction()`, `ParseHeader()` 以满足新版 gvisor `LinkEndpoint` 接口
- **core/socks.go**: `tcpip.Address()` 改为 `tcpip.AddrFromSlice()` (2 处)

### 3. Bug 修复
- **gui/component/adapter.go**: 添加 `runtime.KeepAlive(client)` 防止 GC 回收 client 对象导致 queryConn 被关闭 (参考 [PR #16](https://github.com/lyc8503/EasierConnect/pull/16))

### 4. 连接稳定性改进
- **core/protocol.go**: 为所有 TLS 连接启用 TCP Keepalive (30s 间隔)，防止空闲连接被 NAT/防火墙丢弃
- **core/protocol.go**: RX/TX 重连增加指数退避延迟 (2s, 4s, 6s, 8s, 10s)，避免重试瞬间耗完

### 5. .gitignore 更新
添加 `build_assets/`, `.vscode/`, `.DS_Store`, `*.dmg`

## 已知问题

### recv EOF 断连 ([Issue #18](https://github.com/lyc8503/EasierConnect/issues/18))
连接建立约 1-2 分钟后 recv 流收到 EOF，重连失败。这是一个上游已知但未修复的问题，原项目因深信服官方要求已停止维护。可能的后续方向：
- 实现断连后自动重新登录
- 抓包对比官方 EasyConnect 客户端的协议行为
- 研究是否存在应用层心跳机制
