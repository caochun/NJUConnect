#!/bin/bash
# 抓取 NJUConnect 开源实现的流量
# 使用方法：sudo ./capture_njuconnect.sh

echo "开始抓包 NJUConnect..."
echo "请在另一个终端运行 NJUConnect 并连接"
echo "按 Ctrl+C 停止抓包"
echo ""

sudo tcpdump -i any -s 65535 -w njuconnect.pcap host vpn.nju.edu.cn

echo ""
echo "抓包已保存到 njuconnect.pcap"
