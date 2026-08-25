# 雨落 API 独立仓库设计

> 状态：已获用户确认，待规格复核后执行仓库创建与推送。
>
> 日期：2026-08-26

## 目标

在 `gwenliu1025` 账号下创建公开的 `yuluoapi` 独立仓库，以现有二开仓库 `gwenliu1025/sub2api` 的 `v0.1.179` 为代码基线，同时保留与现有二开仓库及原作者仓库 `Wei-Shaw/sub2api` 的可同步关系，为后续雨落 API 前端二开提供独立演进空间。

## 基线证据

- 基线标签：`v0.1.179`
- 基线提交：`92f353939ad50946cb709a92cf7568d673aa0924`
- 基线来源：`https://github.com/gwenliu1025/sub2api.git`
- 原作者上游：`https://github.com/Wei-Shaw/sub2api.git`
- 目标仓库：`https://github.com/gwenliu1025/yuluoapi.git`

基线标签必须继续指向原提交；导入过程不得重写提交历史、修改基线源码或覆盖原有 `sub2api` 仓库。

## 仓库关系

新仓库使用以下 Git 远端命名：

```text
origin   -> https://github.com/gwenliu1025/yuluoapi.git
legacy   -> https://github.com/gwenliu1025/sub2api.git
upstream -> https://github.com/Wei-Shaw/sub2api.git
```

`origin` 是雨落 API 的唯一推送目标；`legacy` 用于对比和回溯现有二开；`upstream` 用于获取原作者更新。由于个人账号已经拥有同一上游网络中的 `sub2api` Fork，新仓库采用独立仓库，不强行制造第二个 GitHub Fork 关系。提交血缘、远端配置和仓库文档共同提供可审计的来源绑定。

## 分支与标签

- `main` 初始指向 `92f353939`，保留完整祖先提交历史。
- `v0.1.179` 原样推送到新仓库，并继续指向 `92f353939`。
- 不复制旧仓库无关的远端分支，避免把历史实验分支带入新项目。
- 后续前端二开提交只推送到 `yuluoapi`，不回写 `legacy` 或 `upstream`。

## GitHub 导入控制

1. 检查 `gwenliu1025/yuluoapi` 不存在；若已存在，立即停止，不覆盖。
2. 创建公开空仓库，不自动生成 README、LICENSE 或 `.gitignore`，避免产生额外根提交。
3. 在首次推送前暂时禁用 GitHub Actions，防止导入 `v0.1.179` 标签触发旧 Release 工作流或发布错误镜像。
4. 推送 `main` 和 `v0.1.179` 后，以提交、标签和文件树哈希进行验证。
5. 验证通过后恢复 Actions；首次启用不补触发历史标签事件。

## 验证与停止条件

必须同时满足以下条件才算完成：

- `git rev-parse v0.1.179^{}` 等于 `92f353939ad50946cb709a92cf7568d673aa0924`。
- `git ls-remote origin refs/heads/main` 与本地 `main` 提交一致。
- `git ls-remote origin refs/tags/v0.1.179^{}` 与基线提交一致。
- `git diff --exit-code v0.1.179^{}..92f353939` 通过，证明标签树无意外变化。
- GitHub API 返回仓库名称、可见性和默认分支正确，且 Actions 状态恢复为启用。
- 工作区只保留预期的设计文档提交，没有未跟踪文件或敏感信息。

任一检查失败时停止推送后的后续操作，保留失败输出和回滚点；不删除或覆盖旧仓库。

## 回滚

建仓后若验证失败，删除新仓库前先保留本地完整克隆和失败证据；若仅远端引用错误，则修复引用并重新验证。任何回滚都不得影响 `gwenliu1025/sub2api`、`Wei-Shaw/sub2api` 或现有生产部署。
