# 雨落 API 仓库初始化证据记录

> 记录日期：2026-08-26（Asia/Shanghai）
>
> 记录范围：新仓库创建、基线导入、来源绑定、远端验收和本地测试。

## 1. 输入与目标

目标是在不改动原有二开仓库的前提下，新建 `gwenliu1025/yuluoapi`，以 `gwenliu1025/sub2api` 的 `v0.1.179` 为不可变基线，后续独立进行雨落 API 前端二开，并保留访问原二开仓库及原作者仓库的同步通道。

## 2. 建仓前核验

建仓前确认：

- `E:\gwenliu\YuLuo API` 是空目录。
- GitHub CLI 登录账号为 `gwenliu1025`，具备 `repo` 权限。
- `gwenliu1025/yuluoapi` 当时不存在。
- `gwenliu1025/sub2api` 是 `Wei-Shaw/sub2api` 的公开 Fork。
- `v0.1.179` 的 peeled commit 为 `92f353939ad50946cb709a92cf7568d673aa0924`。
- 现有本地发布工作区 `release/v0.1.179-fork` 干净，并与该标签一致。

## 3. 本地导入

执行的核心操作：

```powershell
git clone --branch release/v0.1.179-fork --single-branch https://github.com/gwenliu1025/sub2api.git .
git remote rename origin legacy
git remote add upstream https://github.com/Wei-Shaw/sub2api.git
git branch -M main
```

核验输出：

```text
HEAD       92f353939ad50946cb709a92cf7568d673aa0924
exact tag  v0.1.179
```

随后增加并提交设计与实施计划文档：

```text
b2703ad59 docs: 记录雨落 API 独立仓库设计
ea20e29d4 docs: 添加雨落 API 仓库导入计划
```

## 4. GitHub 仓库创建与导入保护

通过 GitHub API 创建：

```text
full_name       gwenliu1025/yuluoapi
visibility      PUBLIC
default_branch  main
isFork          false
```

目标仓库为空时，先执行：

```powershell
gh api --method PUT repos/gwenliu1025/yuluoapi/actions/permissions -F enabled=false
```

核验结果：

```json
{"enabled":false,"sha_pinning_required":false}
```

这样处理是为了防止导入旧的 `v0.1.179` 标签时触发历史 Release 工作流。

## 5. 推送事实

新增远端：

```powershell
git remote add origin https://github.com/gwenliu1025/yuluoapi.git
```

推送对象：

```text
main
标签：v0.1.179
```

完整历史包约 `98.74 MiB`，HTTPS 推送连接完成较慢。第一次推送实际已在 GitHub 建立远端引用；后续确认命令再次推送 `main` 时返回：

```text
cannot lock ref 'refs/heads/main': reference already exists
```

这表示远端分支已经存在，不是数据丢失。没有执行强制推送。随后使用 `git ls-remote` 独立确认远端引用正确。

GitHub 同时给出既有历史对象提示：

```text
File 9ecc014cf2bfe868d0fc978b3df229dd3c7d924d is 59.23 MB
larger than GitHub's recommended maximum file size of 50.00 MB
```

该对象已被 GitHub 接收。本次没有使用 Git LFS，也没有重写历史。

## 6. 远端最终证据

`git ls-remote origin` 的关键结果：

```text
ea20e29d41fd81338ee11262fe901465dae80fa4  refs/heads/main
9cdcbff81611c902bfd1cd57ed5f25fb7654535c  refs/tags/v0.1.179
92f353939ad50946cb709a92cf7568d673aa0924  refs/tags/v0.1.179^{}
```

GitHub Actions 恢复后核验：

```json
{"enabled":true,"allowed_actions":"all","sha_pinning_required":false}
```

导入完成时查询 Actions 运行记录得到空数组：

```json
[]
```

因此导入期间没有触发历史工作流。

## 7. 测试证据

### 后端

命令：

```powershell
Set-Location backend
go test ./...
```

结果：退出码 `0`。`internal/service` 等主要测试包均通过。

### 前端

首次直接运行测试时，因新克隆工作区尚无 `node_modules`，命令未能启动 `vitest`。随后按锁文件安装依赖：

```powershell
Set-Location frontend
pnpm install --frozen-lockfile
pnpm test:run
```

最终结果：

```text
Test Files  240 passed (240)
Tests       1678 passed (1678)
Duration    21.68s
```

测试输出包含既有的 Vue、i18n、模拟错误路径和 Browserslist 提示，但测试进程退出码为 `0`。

## 8. 最终远端关系

```text
legacy  https://github.com/gwenliu1025/sub2api.git
origin  https://github.com/gwenliu1025/yuluoapi.git
upstream https://github.com/Wei-Shaw/sub2api.git
```

`main` 已设置为跟踪 `origin/main`。`legacy` 与 `upstream` 只用于获取和对比，不是雨落 API 的默认推送目标。

## 9. 影响边界

本次操作只影响：

- 本地目录 `E:\gwenliu\YuLuo API`；
- 新仓库 `gwenliu1025/yuluoapi`；
- 新仓库的 Actions 权限状态和 Git 引用。

本次没有修改原有二开仓库、原作者仓库、GHCR 镜像、Release、生产数据库、生产容器、DNS 或线上服务。

## 10. 复核命令

任何后续执行者都可以使用以下命令重新核验，不依赖聊天历史：

```powershell
Set-Location 'E:\gwenliu\YuLuo API'
git status --short --branch
git remote -v
git rev-parse HEAD
git rev-parse 'v0.1.179^{}'
git ls-remote origin refs/heads/main 'refs/tags/v0.1.179' 'refs/tags/v0.1.179^{}'
gh repo view gwenliu1025/yuluoapi --json nameWithOwner,url,visibility,defaultBranchRef,isFork,description
gh api repos/gwenliu1025/yuluoapi/actions/permissions
```

