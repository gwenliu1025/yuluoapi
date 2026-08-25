# 雨落 API 独立仓库导入实施计划

> **For agentic workers:** 使用当前会话内联执行本计划；每项完成后执行对应验证。

**Goal:** 将现有二开 `v0.1.179` 基线导入为公开独立仓库 `gwenliu1025/yuluoapi`，并保留 `legacy` 与 `upstream` 来源关系。

**Architecture:** 目标 GitHub 仓库先创建为空仓库并临时关闭 Actions。通过现有本地克隆推送 `main` 与精确的 `v0.1.179` 标签；新仓库 `origin` 作为唯一推送端，`legacy` 指向现有二开仓库，`upstream` 指向原作者仓库。基线标签保持原提交，设计文档作为其后的文档提交，不改源码基线。

**Tech Stack:** Git、GitHub CLI (`gh`)、GitHub REST API、PowerShell。

**Spec:** `docs/superpowers/specs/2026-08-26-yuluoapi-repository-design.md`

## Global Constraints

- 目标仓库：`gwenliu1025/yuluoapi`，公开、空初始化，不覆盖已有仓库。
- 基线提交：`92f353939ad50946cb709a92cf7568d673aa0924`，标签：`v0.1.179`。
- 不修改、不删除 `gwenliu1025/sub2api` 或 `Wei-Shaw/sub2api`。
- 首次推送期间关闭 Actions，完成引用验收后恢复。
- 不上传密钥、生产配置或发行资产。

---

### Task 1: 检查本地基线与目标仓库

**Files:**
- Read: `docs/superpowers/specs/2026-08-26-yuluoapi-repository-design.md`

- [ ] **Step 1: Verify local commit, tag, remotes and clean state**

Run:

```powershell
git status --short --branch
git rev-parse v0.1.179^{}
git rev-parse HEAD
git remote -v
git diff --check
```

Expected: `v0.1.179^{}` equals `92f353939ad50946cb709a92cf7568d673aa0924`; only the committed design document is above the tag; remotes are `legacy` and `upstream`; no uncommitted changes.

- [ ] **Step 2: Verify target does not exist**

Run:

```powershell
gh repo view gwenliu1025/yuluoapi --json nameWithOwner,url 2>&1
```

Expected: GitHub reports that the repository cannot be resolved. If it exists, stop without modifying it.

### Task 2: Create empty GitHub repository and block import-time workflows

**Files:**
- Modify: GitHub repository metadata for `gwenliu1025/yuluoapi`

- [ ] **Step 1: Create empty public repository**

Run:

```powershell
gh api --method POST user/repos -f name=yuluoapi -f description='雨落 API：基于 Sub2API 二开基线的独立 API 网关项目' -F private=false -F has_issues=true -F has_projects=true -F has_wiki=true
```

Expected: JSON contains `full_name: gwenliu1025/yuluoapi`, `private: false`, and `default_branch: null` or no branch yet.

- [ ] **Step 2: Disable Actions before pushing refs**

Run:

```powershell
gh api --method PUT repos/gwenliu1025/yuluoapi/actions/permissions -F enabled=false
```

Expected: HTTP success with Actions disabled.

- [ ] **Step 3: Add new origin and verify remotes**

Run:

```powershell
git remote add origin https://github.com/gwenliu1025/yuluoapi.git
git remote -v
```

Expected: `origin` points only to `yuluoapi`; `legacy` and `upstream` remain unchanged.

### Task 3: Push the baseline and repository documentation

**Files:**
- Push: local `main` and annotated tag `v0.1.179`

- [ ] **Step 1: Push main**

Run:

```powershell
git push -u origin main
```

Expected: remote `main` points to local `HEAD` (`b2703ad59` at execution time).

- [ ] **Step 2: Push exact baseline tag**

Run:

```powershell
git push origin refs/tags/v0.1.179
```

Expected: remote tag `v0.1.179` points to commit `92f353939`.

### Task 4: Restore Actions and verify remote identity

**Files:**
- Modify: GitHub Actions repository permission

- [ ] **Step 1: Re-enable Actions**

Run:

```powershell
gh api --method PUT repos/gwenliu1025/yuluoapi/actions/permissions -F enabled=true
```

Expected: HTTP success with Actions enabled.

- [ ] **Step 2: Verify repository metadata, refs and tree**

Run:

```powershell
gh repo view gwenliu1025/yuluoapi --json nameWithOwner,url,visibility,defaultBranchRef,isFork,description
git ls-remote origin refs/heads/main refs/tags/v0.1.179 refs/tags/v0.1.179^{}
git diff --exit-code v0.1.179^{}..92f353939
git status --short --branch
```

Expected:

- Repository is public, named `gwenliu1025/yuluoapi`, and `isFork` is `false`.
- Remote `main` equals local `HEAD`; peeled tag equals `92f353939`.
- Baseline tree comparison exits 0.
- Working tree is clean.

- [ ] **Step 3: Record final remotes and commit IDs**

Run:

```powershell
git remote -v
git rev-parse HEAD
git rev-parse v0.1.179^{}
```

Expected: `origin`, `legacy`, and `upstream` are present; final IDs are recorded in the completion report.

