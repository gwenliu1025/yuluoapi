# 雨落 API 在线更新代理迁移证据

> 日期：2026-09-04（Asia/Shanghai）
>
> 范围：故障只读调查、海外站通用 updater 迁移、雨落配置适配和测试；当前阶段不重新安装生产 updater，不调用 Prepare/Activate，不切换应用镜像。

## 1. 生产故障证据

- 页面当前版本为 `v0.1.180`，能够发现最新版 `v0.2.0`。
- `POST /api/v1/admin/system/update` 返回 HTTP 503，后端日志为 `UPDATE_AGENT_PERMISSION_DENIED` / `permission denied while connecting to update agent`。
- `sub2api-updater.service` 为 `active/enabled`；宿主机通过 Socket 请求 `/v1/health` 返回 `ready=true`，`/v1/status` 返回 `idle`。
- `/run/sub2api-updater/updater.sock` 为 `0660 root:root`。
- `/etc/sub2api-updater/config.json` 的非敏感权限字段为 `socket_gid=0`、`allowed_uids=[0]`。
- 容器 PID 1 `/app/sub2api` 的实际 UID/GID 为 `1000:1000`；以 `1000:1000` 检查 Socket 得到 `SOCKET_WRITE=no`。

结论：失败发生在应用连接 Unix Socket 时，早于镜像拉取和切换。雨落 GitHub/GHCR 地址正确，不是本次故障原因。

## 2. 最早失效点

首次部署把 updater 文件直接安装到服务器，但没有把其实现纳入雨落仓库。安装参数又使用了 root UID/GID，而镜像入口脚本实际会把应用主进程降权到 `1000:1000`。因此服务器上存在一个健康但只允许 root 调用的 updater，管理端无法访问。

## 3. 仓库迁移

从海外站已经运行验证的 updater 工作树迁入：

- `deploy/updater/sub2api_updater.py`
- `deploy/updater/updater_core.py`
- `deploy/updater/install.sh`
- `deploy/updater/sub2api-updater.service`
- `deploy/updater/config.example.json`
- `deploy/updater/tests/test_sub2api_updater.py`
- `deploy/updater/tests/test_updater_core.py`

通用实现、安装器、systemd unit 和核心测试保持同源；雨落的配置样例、包装测试与中文 README 独立绑定：

```text
image_repository=ghcr.io/gwenliu1025/yuluoapi
image_source=https://github.com/gwenliu1025/yuluoapi
compose_directory=/opt/yuluoapi/deploy
compose_file=/opt/yuluoapi/deploy/docker-compose.yml
environment_file=/opt/yuluoapi/.env
app_uid=1000
socket_gid=1000
allowed_uids=[0,1000]
```

海外站与国内雨落站只共享 updater 机制，不共享仓库、镜像、部署目录或运行状态。

## 4. 验收入口

- 本地：`python -X utf8 -m unittest discover -s deploy/updater/tests -p 'test_*.py'`。
- CI：`.github/workflows/backend-ci.yml` 的 `updater` job。
- 发布门禁：`.github/scripts/test-release-policy.sh` 检查 updater 文件、雨落仓库/GHCR、部署目录和 `1000:1000` 权限契约。
- 生产安装和只读验收：`deploy/updater/README.md`。

## 5. 生产边界

本阶段未修改 `/etc/sub2api-updater/config.json`、systemd unit、`.env` 或 Compose；未重启 updater、应用、PostgreSQL 或 Redis；未调用 Prepare/Activate。生产仍运行 `ghcr.io/gwenliu1025/yuluoapi:0.1.180`。
