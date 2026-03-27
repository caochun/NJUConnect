import SwiftUI

struct SMSCodeView: View {
    @Binding var isPresented: Bool
    @State private var smsCode = ""
    let onSubmit: (String) -> Void

    var body: some View {
        VStack(spacing: 20) {
            Text("短信验证")
                .font(.title2)
                .bold()

            Text("请输入收到的短信验证码")
                .foregroundColor(.secondary)

            TextField("验证码", text: $smsCode)
                .textFieldStyle(.roundedBorder)
                .frame(width: 200)

            HStack {
                Button("取消") {
                    isPresented = false
                }
                .keyboardShortcut(.cancelAction)

                Button("提交") {
                    onSubmit(smsCode)
                    isPresented = false
                }
                .keyboardShortcut(.defaultAction)
                .disabled(smsCode.isEmpty)
            }
        }
        .padding()
        .frame(width: 300, height: 200)
    }
}
