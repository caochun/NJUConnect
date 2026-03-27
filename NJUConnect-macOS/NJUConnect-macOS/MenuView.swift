import SwiftUI

struct MenuView: View {
    @ObservedObject var connectionManager: ConnectionManager
    @State private var showSettings = false
    @State private var showSMSInput = false

    var body: some View {
        VStack(spacing: 0) {
            // Header
            HStack {
                Image(systemName: connectionManager.isConnected ? "checkmark.circle.fill" : "circle")
                    .foregroundColor(connectionManager.isConnected ? .green : .gray)
                Text(connectionManager.isConnected ? "已连接" : "未连接")
                    .font(.headline)
                Spacer()
            }
            .padding()

            Divider()

            // Connection info
            if connectionManager.isConnected {
                VStack(alignment: .leading, spacing: 8) {
                    InfoRow(label: "服务器", value: connectionManager.server)
                    InfoRow(label: "用户", value: connectionManager.username)
                    InfoRow(label: "SOCKS5", value: connectionManager.socksPort)
                }
                .padding()
            }

            Divider()

            // Actions
            VStack(spacing: 0) {
                MenuButton(title: "设置", icon: "gearshape") {
                    showSettings = true
                }

                if connectionManager.isConnected {
                    MenuButton(title: "断开连接", icon: "xmark.circle") {
                        connectionManager.disconnect()
                    }
                } else {
                    MenuButton(title: "连接", icon: "play.circle") {
                        connectionManager.connect()
                    }
                }

                Divider()

                MenuButton(title: "退出", icon: "power") {
                    NSApplication.shared.terminate(nil)
                }
            }

            Spacer()
        }
        .frame(width: 300, height: 400)
        .sheet(isPresented: $showSettings) {
            SettingsView(connectionManager: connectionManager)
        }
        .sheet(isPresented: $showSMSInput) {
            SMSCodeView(isPresented: $showSMSInput) { code in
                connectionManager.submitSMSCode(code)
            }
        }
        .onReceive(connectionManager.$needsSMSCode) { needs in
            showSMSInput = needs
        }
    }
}

struct InfoRow: View {
    let label: String
    let value: String

    var body: some View {
        HStack {
            Text(label)
                .foregroundColor(.secondary)
            Spacer()
            Text(value)
                .font(.system(.body, design: .monospaced))
        }
    }
}

struct MenuButton: View {
    let title: String
    let icon: String
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack {
                Image(systemName: icon)
                    .frame(width: 20)
                Text(title)
                Spacer()
            }
            .padding(.horizontal)
            .padding(.vertical, 8)
            .contentShape(Rectangle())
        }
        .buttonStyle(PlainButtonStyle())
        .onHover { hovering in
            if hovering {
                NSCursor.pointingHand.push()
            } else {
                NSCursor.pop()
            }
        }
    }
}
