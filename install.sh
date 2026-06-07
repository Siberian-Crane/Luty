#!/bin/bash
# install.sh - 上传到 GitHub 仓库
REPO="Siberian-Crane/Luty"
INSTALL_DIR="$HOME/.local/bin"

echo "正在下载 luty..."

# 检测系统
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case $ARCH in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
esac

# 获取最新版本下载链接
URL=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | \
      grep -o "https://[^\"]*luty-${OS}-${ARCH}[^\"]*" | head -1)

# 下载
mkdir -p "$INSTALL_DIR"
curl -L "$URL" -o "$INSTALL_DIR/luty"
chmod +x "$INSTALL_DIR/luty"

# 添加到 PATH（如果还没有）
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc 2>/dev/null || true
    export PATH="$INSTALL_DIR:$PATH"
fi

echo "✅ 安装成功！请重新打开终端，然后运行: luty --help"