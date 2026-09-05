# 雨落 API 仓库协作规则

## 新任务入口

1. 每次开始新任务，必须先阅读根目录 `CURRENT_STATE.md`，确认项目身份、当前完成状态、不可变基线、远端关系、停止条件和已知风险。
2. 再根据任务类型读取相关规格；不得只依赖聊天记录、历史记忆或文件名推断需求。
3. 若任务已有用户确认的实施计划，实施前还必须读取该计划，并把计划与对应规格共同作为范围边界。

## 按任务类型选择文档

- 首页、品牌、默认 `/home`、首页文案、首页模型展示或雨滴/涟漪动效：必须阅读 `docs/superpowers/specs/2026-08-27-yuluo-homepage-brand-design.md`；执行首页改造时同时阅读 `docs/superpowers/plans/2026-08-27-yuluo-homepage-implementation.md`。
- 版本同步、GitHub Release/GHCR、Docker 更新、生产核验或故障排查：必须阅读 `.codex/skills/yuluoapi-ops/SKILL.md`，再按任务只加载其中一个或少量相关参考；不得把其他站点或附属服务的运行事实带入雨落 API。
- 仓库来源、历史、分支、标签、远端或同步关系：阅读 `docs/superpowers/specs/2026-08-26-yuluoapi-repository-design.md` 及相关 `docs/operations/` 证据。
- 已有 OpenSpec change 覆盖的功能：读取对应 `openspec/changes/<change-name>/` 下的 `proposal.md`、`design.md`、`specs/`、`tasks.md` 和验证材料；只实施当前任务涉及的 change 与任务项。
- 本机雨落后台管理连接：读取 `.codex/skills/yuluoapi-ops/SKILL.md` 的“本机管理员 API”，凭据只从该入口指定的用户环境变量读取。
- 其他任务：优先查找同模块最近的已确认规格、计划、OpenSpec 和操作证据；没有适用文档时，先明确范围再修改。

## 仓库与远端约束

必须保留以下远端语义：

```text
origin   https://github.com/gwenliu1025/yuluoapi.git
legacy   https://github.com/gwenliu1025/sub2api.git
upstream https://github.com/Wei-Shaw/sub2api.git
```

- `origin` 是雨落 API 的唯一默认推送目标。
- `legacy` 仅用于回溯原有二开和选择性同步，`upstream` 仅用于获取原作者更新。
- 不得将雨落 API 的二开提交推送到 `legacy` 或 `upstream`。
- 不得移动或覆盖 `v0.1.179` 标签，不得在未评估影响时重写既有历史。

## 外部副作用

- 未得到用户在当前任务中的明确授权，不得自动提交、推送、创建或修改远端仓库、修改 GitHub Actions、发布 Release/GHCR、部署、迁移、重启服务或变更生产环境。
- 执行任何已获授权的外部副作用前，必须按 `CURRENT_STATE.md` 和对应规格复核目标、来源、回滚点与停止条件。
