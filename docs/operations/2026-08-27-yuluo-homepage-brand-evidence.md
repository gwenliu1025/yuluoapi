# 雨落 API 默认首页品牌化验证证据

> 日期：2026-08-27（Asia/Shanghai）
>
> 分支：`codex/yuluo-home-brand`
>
> 工作树：`E:\gwenliu\YuLuo API-worktrees\yuluo-home-brand`

## 1. 范围结论

- 只替换 `HomeView.vue` 的默认 `v-else` 展示分支。
- `home_content` 自定义 URL/HTML 分支与 `compact_home_enabled` 简洁首页分支保持原优先级和原结构。
- 未修改后端 API、数据库、鉴权、计费、控制台、管理后台或其他路由行为。
- 未提交、推送、发布或部署。

## 2. 实现要点

- 固定展示“雨落 API”“企业级 API 服务网关”及用户确认的两行品牌母题。
- 主 CTA 继续按登录态进入 `/login` 或对应控制台；次 CTA 与完整模型入口进入 `/model-plaza`。
- 精选区使用 DeepSeek、智谱 GLM、Kimi、Qwen 原品牌标志；Kimi 标志取自仓库既有依赖 `@lobehub/icons` 的 Kimi 图标资源。
- 使用 GSAP 驱动真实雨滴与涟漪图片资源，仅持续改变 `transform` 与 `opacity`；页面不可见或离屏时暂停，卸载时清理。
- `prefers-reduced-motion: reduce` 下不创建循环雨滴和涟漪，静态内容与导航保持完整。
- 只移除默认首页精选模型区下方的最末小字行。

## 3. 自动化验证

在 `frontend` 目录执行：

```powershell
pnpm test:run
pnpm typecheck
pnpm lint:check
pnpm build
```

结果：

```text
Test Files  241 passed (241)
Tests       1686 passed (1686)
typecheck   exit 0
lint:check  exit 0
build       exit 0（1036 modules transformed）
```

首页定向测试：

```text
HomeView.compact.spec.ts  8 passed
HomeView.brand.spec.ts    8 passed
合计                       16 passed
```

仓库根目录执行 `git diff --check`，结果为退出码 `0`。

构建日志保留了仓库原有的 TypeScript ESLint 版本提示、Browserslist 数据提示、动态/静态导入与大 chunk 提示；均未导致失败。

## 4. 浏览器验证

本地预览：`http://127.0.0.1:3001/home`

- 桌面视口：`1440×1024`，浅色与深色均通过。
- 移动端响应式宽度：`390`，导航、文字、CTA 和两列模型区均可读可操作。
- “浏览模型”点击后路径为 `/model-plaza`；未登录“立即开始”链接为 `/login`。
- 通过真实语言切换器切换到英文后，导航、主题提示、登录/控制台、CTA 与模型区操作文案同步切换；三项固定品牌文案仍按规格保留中文。
- 深色模式成功切换暗色背景，Kimi 黑色品牌标志在暗色底板上自动反相以保持对比度。
- 模拟 `prefers-reduced-motion: reduce` 后，雨滴数 `0`、涟漪数 `0`，`data-reduced-motion="true"`。
- 组合状态测试证明页面隐藏和组件离屏任一成立时均保持暂停，恢复条件不会互相覆盖；观察器在卸载时断开。
- 无页面运行时异常；因本地未启动后端，`/setup/status` 与 `/api/v1/settings/public` 返回 `500`，属于预览环境依赖缺失，不影响静态首页验收。

## 5. 视觉证据

- 用户确认稿：`docs/operations/homepage-visual/reference-yuluo-home-v3.png`
- 浅色桌面：`docs/operations/homepage-visual/yuluo-home-1440x1024-light.png`
- 深色桌面：`docs/operations/homepage-visual/yuluo-home-1440x1024-dark.png`
- 移动端：`docs/operations/homepage-visual/yuluo-home-390x844-light.png`
- 并排对照：`docs/operations/homepage-visual/comparison-reference-vs-implementation.png`

最终视觉结论见根目录 `design-qa.md`：`final result: passed`。

## 6. 最终只读审查闭环

最终审查未发现 P0/P1，并提出 3 个 P2 与 1 个 P3，已全部闭环：

1. 将页面隐藏与元素相交状态合并为单一暂停条件，补充组合状态及观察器清理测试。
2. 用明确的移动端隐藏类替代跨雨滴/涟漪共享的 `nth-of-type` 计数，保持 2 个雨滴和 2 个涟漪参与移动端动效。
3. 保留固定中文品牌文案，其余导航与操作文案接入现有 i18n，并补充中英文回归测试与真实切换验证。
4. 恢复 pnpm 锁文件中与 GSAP 无关的 Rollup Linux `libc` 元数据，最终锁文件差异只包含 GSAP。
5. 复核发现移动端隐藏节点仍创建 GSAP 循环；已改用 `gsap.matchMedia()`，移动端只为实际显示的 2 个雨滴和 2 个涟漪创建动画，并补充运行节点数量测试与浏览器实测。

上述修复后的最终只读复核未发现任何 P0、P1、P2 或 P3 问题。
