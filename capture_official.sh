#!/bin/bash
# 抓取官方 EasyConnect 客户端的流量
# 使用方法：sudo ./capture_official.sh

echo "开始抓包官方客户端..."
echo "请在另一个终端启动官方 EasyConnect 并连接"
echo "按 Ctrl+C 停止抓包"
echo ""

sudo tcpdump -i any -s 65535 -w official_client.pcap host vpn.nju.edu.cn

echo ""
echo "抓包已保存到 official_client.pcap"
