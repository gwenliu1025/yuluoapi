# Docker 生产更新

## 运行契约

- 部署目录：`/opt/yuluoapi`
- Compose：`/opt/yuluoapi/deploy/docker-compose.yml`
- 环境文件：`/opt/yuluoapi/.env`
- 应用容器：`sub2api`
- 数据服务：`sub2api-postgres`、`sub2api-redis`
- 更新服务：`sub2api-updater.service`
- Socket：`/run/sub2api-updater/updater.sock`
- `UPDATE_REPO=gwenliu1025/yuluoapi`
- `UPDATE_IMAGE_REPOSITORY=ghcr.io/gwenliu1025/yuluoapi`
- `UPDATE_MODE=docker_agent`

任何运行事实先在目标主机只读回查，不用本文件中的名称推断当前状态。

## 更新前

1. 记录应用镜像、容器 ID、启动时间、健康状态、重启次数和无关容器基线。
2. 备份 PostgreSQL custom dump、Redis 数据、`.env`、Compose、updater 配置/状态，并生成 SHA-256 清单。
3. 用 `pg_restore --list` 和校验清单验证备份可读。
4. 大版本或迁移量较大时，用隔离网络和备份数据预演迁移；不开放宿主机端口，结束后删除临时容器、网络、卷和临时环境文件。

## 更新代理只读检查

```bash
systemctl is-active sub2api-updater.service
curl --unix-socket /run/sub2api-updater/updater.sock http://localhost/v1/health
curl --unix-socket /run/sub2api-updater/updater.sock http://localhost/v1/status
```

Prepare 只拉取并验证精确镜像；Activate 只重建 `sub2api`。用户要求自行更新时，不调用两者。

## 更新后

```bash
docker inspect sub2api --format '{{.Config.Image}}|{{.State.Status}}|{{.State.Health.Status}}|{{.RestartCount}}'
curl -fsS http://127.0.0.1:8080/health
curl -fsS https://sub.yuluocloud.com/health
```

继续核对容器内版本/提交、匿名鉴权预期、启动错误、真实成功请求、上一精确镜像和无关容器 ID/启动时间/重启次数。失败时按已验证回滚入口恢复，不把超时直接解释为应用崩溃。
