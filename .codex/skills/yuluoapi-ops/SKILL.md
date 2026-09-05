---
name: yuluoapi-ops
description: 用于雨落 API 的版本同步、GitHub Release/GHCR 发布、Docker 更新、生产核验和故障排查；不用于其他站点或附属服务。
---

# 雨落 API 运维

## 权威入口

1. 先读取仓库根目录 `CURRENT_STATE.md`，以当前 Git、发布和运行证据为准。
2. 版本同步同时读取 `docs/superpowers/specs/2026-08-26-yuluoapi-repository-design.md`；首页相关冲突再读取品牌首页规格和实施计划。
3. 只加载当前任务需要的参考：
   - 版本同步、提交、标签、Release 与 GHCR：`references/release-and-assets.md`
   - Docker 更新、备份、预热、切换与回滚：`references/production-update.md`
   - 502、上游错误、账号与计费排查：`references/incident-triage.md`

## 雨落 API 固定身份

```text
本地仓库：E:\gwenliu\YuLuo API
GitHub：  gwenliu1025/yuluoapi
镜像：    ghcr.io/gwenliu1025/yuluoapi:X.Y.Z
生产目录：/opt/yuluoapi
公网入口：https://sub.yuluocloud.com
```

远端语义固定为 `origin=yuluoapi`、`legacy=gwenliu1025/sub2api`、`upstream=Wei-Shaw/sub2api`。只向 `origin` 写入；`legacy` 和 `upstream` 只读。

## 凭据与连接

- 仓库和 Skill 不保存主机 IP、SSH 密码、私钥、API Key、Cookie、OAuth JSON、Token、TOTP 种子或管理员凭据。
- 连接参数只从当前批准的安全来源读取；推荐变量名为 `YULUOAPI_SSH_HOST`、`YULUOAPI_SSH_USER`、`YULUOAPI_SSH_PORT`、`YULUOAPI_SSH_KEY`、`YULUOAPI_SSH_HOSTKEY_SHA256`。
- 只检查变量是否存在，不打印变量值；固定校验主机公钥后再连接。

### 本机管理员 API

本机 Windows 用户级环境变量 `YULUOAPI_BASE_URL` 与 `YULUOAPI_ADMIN_API_KEY` 是雨落后台连接配置的读取入口；只检查是否存在，不回显密钥，不写入仓库或文档。旧进程通过 `GetEnvironmentVariable(..., 'User')` 读取最新值。变量缺失或鉴权失败时停止管理请求，由用户修正该环境变量，不搜索其它站点凭据。

管理操作沿用 `skills/sub2api-admin/SKILL.md` 的 CLI；仅在当前命令进程映射其通用变量名，避免污染其它站点的用户级配置：

```powershell
$env:SUB2API_BASE_URL = [Environment]::GetEnvironmentVariable('YULUOAPI_BASE_URL', 'User')
$env:SUB2API_ADMIN_API_KEY = [Environment]::GetEnvironmentVariable('YULUOAPI_ADMIN_API_KEY', 'User')
if ($env:SUB2API_BASE_URL -ne 'https://sub.yuluocloud.com' -or -not $env:SUB2API_ADMIN_API_KEY) { throw '雨落管理连接配置缺失或目标不符' }
node skills/sub2api-admin/scripts/sub2api-admin.js groups all
```

持有连接配置不扩大变更授权；先只读确认目标，写入后回读核验，保持渠道启停、支付配置、发布和生产切换各自的授权边界。

## 共同门禁

- 区分代码合入、Git 推送、Release/GHCR 发布、Prepare、Activate 和生产部署；前一步不自动授权后一步。
- 生产变更前先做 PostgreSQL、Redis、`.env`、Compose、updater 状态和容器基线备份并校验。
- 只重建目标应用容器；PostgreSQL、Redis 和其他无关服务不得随应用切换重建。
- 健康检查只证明存活；变更后还需验证版本/提交、鉴权边界、真实业务流量、错误日志、回滚镜像和无关容器不变。
- 结果不确定或副作用接受状态未知时停止，不盲目重试。
