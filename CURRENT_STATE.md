# 雨落 API 当前状态

> 本文件是 `yuluoapi` 项目的磁盘状态入口。后续判断项目现状时，以本文件、Git 提交和对应证据文档为准，不以聊天记录或模型记忆为准。
>
> 最近核验时间：2026-08-26（Asia/Shanghai）

## 1. 项目标识

- 项目名称：雨落 API
- GitHub 仓库：`https://github.com/gwenliu1025/yuluoapi`
- 本地工作区：`E:\gwenliu\YuLuo API`
- 仓库可见性：公开
- 默认分支：`main`
- GitHub 仓库类型：独立仓库，`isFork=false`

## 2. 来源关系

当前工作区的 Git 远端必须保持为：

```text
origin   https://github.com/gwenliu1025/yuluoapi.git
legacy   https://github.com/gwenliu1025/sub2api.git
upstream https://github.com/Wei-Shaw/sub2api.git
```

含义：

- `origin`：雨落 API 的唯一默认推送目标。
- `legacy`：原有二开项目，用于回溯和选择性同步。
- `upstream`：原作者仓库，用于获取上游更新。
- 不得把雨落 API 的提交直接推送到 `legacy` 或 `upstream`。

## 3. 不可变基线

- 基线标签：`v0.1.179`
- 标签对象：`9cdcbff81611c902bfd1cd57ed5f25fb7654535c`
- 标签对应提交：`92f353939ad50946cb709a92cf7568d673aa0924`
- 原二开基线：`gwenliu1025/sub2api` 的 `v0.1.179`
- 建仓完成点：`ea20e29d41fd81338ee11262fe901465dae80fa4`

`v0.1.179` 必须继续指向 `92f353939ad50946cb709a92cf7568d673aa0924`。后续二开从 `main` 继续，不移动、不覆盖该标签。

建仓完成点相对基线只增加以下治理文档，不修改基线业务源码：

- `docs/superpowers/specs/2026-08-26-yuluoapi-repository-design.md`
- `docs/superpowers/plans/2026-08-26-yuluoapi-repository-import.md`

## 4. 已完成事项

1. 从 `gwenliu1025/sub2api` 的 `release/v0.1.179-fork` 克隆完整祖先历史。
2. 将本地工作分支调整为 `main`。
3. 配置 `origin`、`legacy`、`upstream` 三个远端。
4. 创建公开独立仓库 `gwenliu1025/yuluoapi`。
5. 首次导入前关闭 GitHub Actions，避免旧标签触发历史发布工作流。
6. 推送 `main` 与原样的 `v0.1.179` 标签。
7. 恢复 GitHub Actions；当前状态为 `enabled=true`、`allowed_actions=all`。
8. 验证远端分支、标签、基线树、工作区和测试结果。

## 5. 已验证结果

### 仓库引用

```text
origin/main                    ea20e29d41fd81338ee11262fe901465dae80fa4
origin/v0.1.179 tag object     9cdcbff81611c902bfd1cd57ed5f25fb7654535c
origin/v0.1.179 peeled commit  92f353939ad50946cb709a92cf7568d673aa0924
```

### 基线一致性

执行：

```powershell
git diff --exit-code 'v0.1.179^{}' 92f353939
```

结果：退出码 `0`，基线标签对应源码树未发生变化。

### 后端测试

执行：

```powershell
Set-Location backend
go test ./...
```

结果：退出码 `0`，全部 Go 包测试通过。

### 前端测试

执行：

```powershell
Set-Location frontend
pnpm install --frozen-lockfile
pnpm test:run
```

结果：

```text
Test Files  240 passed (240)
Tests       1678 passed (1678)
```

### 工作区

建仓验收时执行 `git diff --check` 成功，`main` 与 `origin/main` 一致，工作区无未提交文件。

## 6. 明确未做事项

以下内容不属于本次建仓工作，尚未实施：

- 未修改前端品牌、名称、Logo、颜色、首页或管理后台界面。
- 未把 Go module、前端包名和源码内 `Sub2API` 标识批量重命名为雨落 API。
- 未创建雨落 API 的新版本标签或 GitHub Release。
- 未发布新的 GHCR 镜像。
- 未部署、迁移或重启任何生产服务。
- 未修改 `gwenliu1025/sub2api` 和 `Wei-Shaw/sub2api`。

## 7. 已知事项

首次推送完整历史时，GitHub 对对象 `9ecc014cf2bfe868d0fc978b3df229dd3c7d924d` 给出 `59.23 MB` 大文件提示。GitHub 已接收该对象；这不是推送失败，但后续历史治理时应保留该证据，不应在未评估影响时重写历史。

首次推送包含约 `98.74 MiB` 的历史包，HTTPS 长连接完成较慢。远端引用已通过 `git ls-remote` 独立核验。

## 8. 后续工作规则

开始任何二开前先执行：

```powershell
Set-Location 'E:\gwenliu\YuLuo API'
git status --short --branch
git remote -v
git fetch origin --prune
git fetch legacy --prune
git fetch upstream --prune
```

停止条件：

- 工作区出现来源不明的修改；
- `origin` 不再指向 `gwenliu1025/yuluoapi`；
- `v0.1.179^{}` 不再等于 `92f353939ad50946cb709a92cf7568d673aa0924`；
- 计划将提交推送到 `legacy` 或 `upstream`。

前端二开应使用独立功能分支，从最新 `origin/main` 创建；合入前至少执行前端测试、类型检查、构建和 `git diff --check`。

## 9. 文档索引

- 当前状态入口：`CURRENT_STATE.md`
- 建仓设计：`docs/superpowers/specs/2026-08-26-yuluoapi-repository-design.md`
- 导入计划：`docs/superpowers/plans/2026-08-26-yuluoapi-repository-import.md`
- 建仓证据：`docs/operations/2026-08-26-yuluoapi-repository-bootstrap-evidence.md`

每次完成重要变更后，应更新 `CURRENT_STATE.md`，并在 `docs/operations/` 下新增带日期的证据记录；不得用聊天记录替代项目文档。
