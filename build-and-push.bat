@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

:: Docker Hub 用户名和镜像名称 - 请修改为你自己的配置
set DOCKER_USERNAME=your-username
set IMAGE_NAME=rabbit-icon
set IMAGE_TAG=latest

:: 完整的镜像名称
set FULL_IMAGE_NAME=%DOCKER_USERNAME%/%IMAGE_NAME%:%IMAGE_TAG%

echo ==========================================
echo 开始构建 Docker 镜像...
echo 镜像名称: %FULL_IMAGE_NAME%
echo ==========================================
echo.

:: 构建镜像
docker build -t %FULL_IMAGE_NAME% .

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✅ 镜像构建成功！
    echo.
    echo ==========================================
    echo 开始推送镜像到 Docker Hub...
    echo ==========================================
    
    :: 登录 Docker Hub
    docker login
    
    :: 推送镜像
    docker push %FULL_IMAGE_NAME%
    
    if %ERRORLEVEL% EQU 0 (
        echo.
        echo ✅ 镜像推送成功！
        echo.
        echo 使用以下命令拉取镜像：
        echo docker pull %FULL_IMAGE_NAME%
        echo.
        echo 使用以下命令运行镜像：
        echo docker run -d -p 8080:8080 --name rabbit-icon %FULL_IMAGE_NAME%
    ) else (
        echo ❌ 镜像推送失败
        exit /b 1
    )
) else (
    echo ❌ 镜像构建失败
    exit /b 1
)

endlocal
