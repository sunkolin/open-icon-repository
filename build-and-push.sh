#!/bin/bash

# Docker Hub 用户名和镜像名称 - 请修改为你自己的配置
DOCKER_USERNAME="sunkolin"
IMAGE_NAME="open-icon-repository"
IMAGE_TAG="latest"

# 完整的镜像名称
FULL_IMAGE_NAME="${DOCKER_USERNAME}/${IMAGE_NAME}:${IMAGE_TAG}"

echo "=========================================="
echo "开始构建 Docker 镜像..."
echo "镜像名称: ${FULL_IMAGE_NAME}"
echo "=========================================="

# 构建镜像
docker build -t ${FULL_IMAGE_NAME} .

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ 镜像构建成功！"
    echo ""
    echo "=========================================="
    echo "开始推送镜像到 Docker Hub..."
    echo "=========================================="
    
    # 登录 Docker Hub
    docker login
    
    # 推送镜像
    docker push ${FULL_IMAGE_NAME}
    
    if [ $? -eq 0 ]; then
        echo ""
        echo "✅ 镜像推送成功！"
        echo ""
        echo "使用以下命令拉取镜像："
        echo "docker pull ${FULL_IMAGE_NAME}"
        echo ""
        echo "使用以下命令运行镜像："
        echo "docker run -d -p 6024:6024 --name open-icon-repository ${FULL_IMAGE_NAME}"
    else
        echo "❌ 镜像推送失败"
        exit 1
    fi
else
    echo "❌ 镜像构建失败"
    exit 1
fi
