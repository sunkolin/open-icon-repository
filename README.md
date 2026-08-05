### 开源图标仓库

#### 欢迎提交图标
- 欢迎提交图标，要求格式为1024x1024及以上的图片

#### docker-compose部署

```yaml
services:
  open-icon-repository:
    image: sunkolin/open-icon-repository:latest
    container_name: open-icon-repository
    ports:
      - "6024:6024"
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:6024"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s
```