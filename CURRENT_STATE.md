# 雨落 API 当前状态

> 本文件是 `yuluoapi` 项目的磁盘状态入口。后续判断项目现状时，以本文件、Git 提交和对应证据文档为准，不以聊天记录或模型记忆为准。
>
> 最近核验时间：2026-08-31（Asia/Shanghai）

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
origin/main                    85f31f50d54970ff617e708097ed2af938b5649c
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

- 建仓阶段未修改前端品牌、名称、Logo、颜色、首页或管理后台界面；当前仅在独立功能分支形成默认 `/home` 品牌首页候选，尚未合入 `main`。
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

## 10. 默认 `/home` 品牌首页候选（2026-08-27）

- 状态：已在独立工作树 `E:\gwenliu\YuLuo API-worktrees\yuluo-home-brand`、分支 `codex/yuluo-home-brand` 完成实现与验收；尚未提交、推送、发布或部署。
- 基线提交：`85f31f50d54970ff617e708097ed2af938b5649c`。
- 范围：只替换未配置 `home_content` 且未启用 `compact_home_enabled` 时的默认 `/home`；自定义 URL/HTML 首页、简洁首页、配置优先级、其他路由及后端保持不变。
- 内容：雨后青瓷主视觉、固定品牌文案、双 CTA、DeepSeek/智谱 GLM/Kimi/Qwen 原品牌标志、更多模型入口，以及带减少动态效果、离屏暂停和卸载清理的 GSAP 雨滴、动态涟漪与真实照片花瓣动效。
- 页脚边界：只移除了默认首页精选模型区下方的最末小字行；简洁首页页脚及其他页面未改动。
- 2026-08-28 水波修复：确认主视觉背景自带静态水纹，原动画中心与背景落点错位且低对比；现已把三组自动涟漪对齐到背景水面、增强扩散可见度，并加入复用节点池的鼠标/触控点击涟漪，保留 `prefers-reduced-motion`、离屏暂停与卸载清理。
- 验收：前端全量 `241` 个测试文件、`1686` 项测试通过；类型检查、Lint、生产构建和 `git diff --check` 通过；桌面、移动端、深色模式、中英文操作文案、CTA、模型广场跳转及 `prefers-reduced-motion` 已完成浏览器验证。
- 证据：`docs/operations/2026-08-27-yuluo-homepage-brand-evidence.md` 与根目录 `design-qa.md`。

## 11. 前端资产定位与版本汇流路线（2026-08-31）

### 11.1 仓库身份

本仓库是用户自有仓库，不是从他人项目分叉而来的下游。`main` 上的既有代码是用户在售卖海外模型阶段完成的二次创作成果，已完整迁入本仓库，属于必须保留的资产。当前业务方向由海外模型转为国产模型，二创资产在方向调整后继续演进，不得因换方向而丢弃或重置。

### 11.2 三份输入汇流

后续版本由三份输入汇流构建，缺一不可：

1. 线上运行的用户自有二创代码（业务基线，含既有功能与配置体系）。
2. 本次默认 `/home` 品牌首页前端改动（见第 10 节）。
3. 原作者最新版本的更新内容，经 `upstream` 获取后择取。

汇流产物在本仓库内构建，并与原作者版本号建立对应关系，便于追溯某个雨落 API 版本基于哪个原作者版本。此处的语义是「以自有二创为主体、择取原作者更新」，不是把本仓库当作上游的下游做整体合并；`AGENTS.md` 的远端约束继续生效。

### 11.3 前端改动的工程约束

品牌首页前端是要整体替换低版本前端的资产，且后续需反复承接原作者更新，因此前端改动必须优先复用既有令牌、组件与工具函数，减少重放冲突面：

- 颜色取值应走 `tailwind.config.js` 既有 `primary-*` / `accent-*` 令牌，不新开并行色板。
- 功能可见性判断应复用 `frontend/src/utils/featureFlags.ts` 与既有组件写法，不在新组件内另立判断逻辑。
- 不在生产 DOM 中新增仅为测试读取而存在的常量属性。

### 11.4 已知顺序风险

当前顺序是先定前端设计、再择取原作者更新。若原作者改动落在 `HomeView.vue` 或前端目录结构上，本次首页改动需在冲突中重放一次。反向顺序（先择取更新、再做品牌前端）冲突面更小。选择当前顺序时，须在择取更新前先复核原作者对前端目录的改动范围，并在 `docs/operations/` 记录复核结论。
