# 版本同步与仓库资产

## 版本契约

```text
Tag:     vX.Y.Z
App:     X.Y.Z
Image:   ghcr.io/gwenliu1025/yuluoapi:X.Y.Z
Release: https://github.com/gwenliu1025/yuluoapi/releases/tag/vX.Y.Z
```

`backend/cmd/server/VERSION`、部署示例、在线更新仓库、镜像仓库、Release workflow 和标签必须使用同一版本。品牌首页文件与测试必须在上游同步中保留。

## 同步顺序

1. 核对干净工作区、三条远端、不可变 `v0.1.179` 和 `origin/main`。
2. 获取 `origin`、`legacy`、`upstream`，确认目标上游标签和海外站 fork 的已验证合入提交。
3. 以雨落 `main` 为主体合入经过验证的 fork 变更；冲突优先保留雨落仓库身份、品牌首页、更新代理与发布契约，业务代码采用目标版本并以测试证明。
4. 检查海外站仓库、非雨落镜像和非雨落域名是否残留在运行配置、部署文档或发布门禁中；历史来源说明和 Go module 路径不做机械替换。
5. 完成本地测试后提交并推送 `origin/main`，等待 CI 与 Security Scan 成功。
6. 创建指向已验证提交的 annotated tag，推送到 `origin`，等待 Release workflow 完成。

## 资产核验

- 标签 peeled commit 必须等于发布提交。
- Release 不是 draft/prerelease（除非版本本身明确为预发布）。
- 五个平台归档和 `checksums.txt` 齐全，下载资产的 SHA-256 与清单一致。
- GHCR 是 `linux/amd64`、`linux/arm64` 多平台 OCI index；远端拉取后用 `--version` 核对版本、提交和 OCI source。
- 发布完成不代表生产已更新；没有当前明确授权时停在资产可用状态。
