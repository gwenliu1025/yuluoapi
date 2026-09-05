# 雨落 API 宿主机 Docker 更新代理

本目录持有雨落 API 在线更新的正式宿主机实现。代理以 `root` 运行，只监听 Unix Socket，不开放 TCP；应用容器不接触 Docker Socket。

## 固定边界

```text
GitHub：        gwenliu1025/yuluoapi
GHCR：          ghcr.io/gwenliu1025/yuluoapi:X.Y.Z
部署目录：      /opt/yuluoapi
Compose：       /opt/yuluoapi/deploy/docker-compose.yml
环境文件：      /opt/yuluoapi/.env
应用容器：      sub2api
应用 UID/GID：  1000:1000
Updater Socket：/run/sub2api-updater/updater.sock
```

海外站可复用本目录的 updater 机制，但必须使用它自己的仓库、镜像和部署参数；不得把两个站点的资产地址写成同一值。

## 工作机制

1. `POST /v1/prepare` 拉取精确版本镜像，并验证来源、版本、架构和镜像身份。
2. `POST /v1/activate` 只重建 `sub2api` 应用容器。
3. 代理检查 Docker 状态和 HTTP 健康状态。
4. 激活失败时恢复上一精确镜像。
5. 状态保存在 `/var/lib/sub2api-updater/state.json`，服务重启后可继续协调中断状态。

Socket 使用 Linux `SO_PEERCRED` 校验调用方 UID。雨落容器内 `/app/sub2api` 以 `1000:1000` 运行，因此必须安装为 `socket_gid=1000`、`allowed_uids=[0,1000]`；`root:root 0660` 会直接产生 `UPDATE_AGENT_PERMISSION_DENIED`。

## Socket API

- `GET /v1/health`：返回就绪状态和协议版本。
- `POST /v1/prepare`：接收 `{"version":"0.2.002"}`，拉取并验证目标镜像。
- `POST /v1/activate`：要求空请求体，返回 `202` 后由后台执行切换。
- `GET /v1/status`：返回最后一次持久化的公开状态。

状态响应固定包含 `state`、`current_image`、`target_image`、`previous_image`、`message` 和 `updated_at` 六个字段。内部 Docker image ID 不通过 API 返回。

## 安装或重新收敛

只从雨落 API 当前仓库检出的本目录执行：

```bash
sudo ./deploy/updater/install.sh \
  --compose-directory /opt/yuluoapi/deploy \
  --compose-file /opt/yuluoapi/deploy/docker-compose.yml \
  --environment-file /opt/yuluoapi/.env \
  --service-name sub2api \
  --container-name sub2api \
  --app-uid 1000 \
  --socket-gid 1000 \
  --expected-architecture amd64 \
  --health-url http://127.0.0.1:8080/health \
  --image-repository ghcr.io/gwenliu1025/yuluoapi \
  --image-source https://github.com/gwenliu1025/yuluoapi \
  --prepare-timeout-seconds 600 \
  --activation-timeout-seconds 120 \
  --poll-interval-seconds 2
```

安装器只复制代理文件、生成 `0600 root:root` 配置、安装并重启 `sub2api-updater.service`，随后验证健康接口。它不修改 `.env` 或 Compose，不调用 Prepare/Activate，也不切换应用镜像。

不得使用 `docker inspect ... .Config.User` 的空值推断应用 UID；当前镜像由入口脚本降权。应读取实际主进程：

```bash
docker exec sub2api sh -c "grep -E '^(Uid|Gid|Groups):' /proc/1/status"
```

## 验收

```bash
systemctl is-active sub2api-updater.service
systemctl is-enabled sub2api-updater.service
stat -c '%A|%a|%U|%G|%u|%g|%n' \
  /run/sub2api-updater/updater.sock
curl --unix-socket /run/sub2api-updater/updater.sock \
  http://localhost/v1/health
curl --unix-socket /run/sub2api-updater/updater.sock \
  http://localhost/v1/status
docker exec -u 1000:1000 sub2api sh -c \
  '[ -w /run/sub2api-updater/updater.sock ]'
```

通过条件：服务为 `active/enabled`；Socket 为 `0660 root:1000`；健康接口返回 `ready=true`；状态可读；应用 UID `1000` 对 Socket 具备读写权限。

## 测试

```bash
python3 -m unittest discover -s deploy/updater/tests -p 'test_*.py'
```

测试只使用临时目录和模拟命令，不连接 Docker 或生产服务。

## 故障与恢复

```bash
systemctl status sub2api-updater.service
journalctl -u sub2api-updater.service -n 100 --no-pager
stat -c '%A %U %G' /run/sub2api-updater/updater.sock
cat /var/lib/sub2api-updater/state.json
```

若状态为 `rollback_failed`，先确认状态文件、精确镜像和部署环境。唯一允许的应用恢复动作是：

```bash
docker compose \
  --project-directory /opt/yuluoapi/deploy \
  -f /opt/yuluoapi/deploy/docker-compose.yml \
  --env-file /opt/yuluoapi/.env \
  up -d --no-deps --force-recreate sub2api
```

不得随应用更新重建或重启 PostgreSQL、Redis 及其他服务。
