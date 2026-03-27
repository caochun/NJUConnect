import SwiftUI

struct SettingsView: View {
    @ObservedObject var connectionManager: ConnectionManager
    @Environment(\.dismiss) var dismiss

    var body: some View {
        VStack(spacing: 20) {
            Text("连接设置")
                .font(.title2)
                .bold()

            Form {
                TextField("服务器", text: $connectionManager.server)
                TextField("端口", text: $connectionManager.port)
                TextField("用户名", text: $connectionManager.username)
                SecureField("密码", text: $connectionManager.password)
                TextField("SOCKS5 端口", text: $connectionManager.socksPort)
            }
            .textFieldStyle(.roundedBorder)

            HStack {
                Button("取消") {
                    dismiss()
                }
                .keyboardShortcut(.cancelAction)

                Button("保存") {
                    connectionManager.saveSettings()
                    dismiss()
                }
                .keyboardShortcut(.defaultAction)
            }
        }
        .padding()
        .frame(width: 400, height: 350)
    }
}
