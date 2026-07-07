#!/bin/bash
# ─────────────────────────────────────────────────
# Kitex + RocketMQ 电商微服务 — 代码生成脚本
#
# 前置条件:
#   1. go install github.com/cloudwego/kitex/tool/cmd/kitex@latest
#   2. go install github.com/cloudwego/thriftgo@latest
#
# 使用:
#   chmod +x generate.sh && ./generate.sh
#   然后 cd user && go mod tidy && go build
# ─────────────────────────────────────────────────

set -e

echo "⏳ 生成用户服务代码..."
cd user
kitex -module user -service user ../idl/user.thrift
cd ..

echo "⏳ 生成商品服务代码..."
cd product
kitex -module product -service product ../idl/product.thrift
cd ..

echo "⏳ 生成订单服务代码..."
cd order
kitex -module order -service order ../idl/order.thrift
cd ..

echo ""
echo "✅ 代码生成完成！"
echo ""
echo "构建各服务:"
echo "  cd user    && go mod tidy && go build -o user_svc ."
echo "  cd product && go mod tidy && go build -o product_svc ."
echo "  cd order   && go mod tidy && go build -o order_svc ."
echo ""
echo "启动顺序:"
echo "  ./user/user_svc     # :9001"
echo "  ./product/product_svc  # :9002"
echo "  ./order/order_svc   # :9003"
