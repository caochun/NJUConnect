import Foundation
import Combine

class ConnectionManager: ObservableObject {
    @Published var isConnected = false
    @Published var needsSMSCode = false
    @Published var server = ""
    @Published var port = "443"
    @Published var username = ""
    @Published var password = ""
    @Published var socksPort = ":1080"

    private var process: Process?
    private var inputPipe: Pipe?
    private let defaults = UserDefaults.standard

    init() {
        loadSettings()
    }

    func loadSettings() {
        server = defaults.string(forKey: "server") ?? "vpn.nju.edu.cn"
        port = defaults.string(forKey: "port") ?? "443"
        username = defaults.string(forKey: "username") ?? ""
        password = defaults.string(forKey: "password") ?? ""
        socksPort = defaults.string(forKey: "socksPort") ?? ":1080"
    }

    func saveSettings() {
        defaults.set(server, forKey: "server")
        defaults.set(port, forKey: "port")
        defaults.set(username, forKey: "username")
        defaults.set(password, forKey: "password")
        defaults.set(socksPort, forKey: "socksPort")
    }

    func connect() {
        guard !isConnected else { return }

        let njuconnectPath = Bundle.main.path(forResource: "njuconnect", ofType: nil) ?? "./njuconnect"

        process = Process()
        inputPipe = Pipe()
        let outputPipe = Pipe()

        process?.executableURL = URL(fileURLWithPath: njuconnectPath)
        process?.arguments = [
            "-server", server,
            "-port", port,
            "-username", username,
            "-password", password,
            "-socks-bind", socksPort
        ]
        process?.standardInput = inputPipe
        process?.standardOutput = outputPipe

        // 监听输出，检测是否需要短信验证码
        outputPipe.fileHandleForReading.readabilityHandler = { [weak self] handle in
            let data = handle.availableData
            if let output = String(data: data, encoding: .utf8) {
                if output.contains("Please enter your sms code") {
                    DispatchQueue.main.async {
                        self?.needsSMSCode = true
                    }
                }
            }
        }

        process?.terminationHandler = { [weak self] _ in
            DispatchQueue.main.async {
                self?.isConnected = false
                self?.needsSMSCode = false
            }
        }

        do {
            try process?.run()
            isConnected = true
        } catch {
            print("Failed to start njuconnect: \(error)")
        }
    }

    func submitSMSCode(_ code: String) {
        guard let inputPipe = inputPipe else { return }
        let data = (code + "\n").data(using: .utf8)!
        inputPipe.fileHandleForWriting.write(data)
        needsSMSCode = false
    }

    func disconnect() {
        process?.terminate()
        process = nil
        isConnected = false
    }
}
