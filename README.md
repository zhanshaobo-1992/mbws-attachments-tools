# mbws-attachments-tools

为 MBWS 插件附件库准备的 **`ziplist`** / **`unzip`** 多平台包。

仓库：[zhanshaobo-1992/mbws-attachments-tools](https://github.com/zhanshaobo-1992/mbws-attachments-tools)

## 附件约定

| 附件名 | 版本（建议） | 入口 | 用途 |
| --- | --- | --- | --- |
| ziplist | 1.0.0 | ziplist / ziplist.exe | 按 UTF-8 列出 ZIP 中央目录条目（一行一个） |
| unzip | 6.0.0 | unzip / unzip.exe | 解压：`unzip -o <zip> -d <dir>`；回退列举：`unzip -Z1` |

包根目录**只能有可执行入口**（与 ffmpeg 附件布局一致）。

### 重要：darwin unzip 必须用可移植二进制

**不要**再打包 macOS `/usr/bin/unzip`。那是 Apple 平台签名二进制，拷到附件缓存目录后会被 AMFI 直接杀掉（常见 exit `137` / `-1`、无 stderr），宿主侧表现为「附件未就绪 / 解压失败」。

本仓库的 `unzip` 与 `ziplist` 一样，用 **Go 交叉编译的静态二进制**（`archive/zip`），可放在任意路径执行。

## 本地构建

```bash
# 需要 Go ≥ 1.23
export PATH="/path/to/go/bin:$PATH"   # 若未装 Go
./scripts/build-ziplist.sh
./scripts/build-unzip.sh
```

产物在 `dist/ziplist/`、`dist/unzip/`。

说明：

- **ziplist**：Go 交叉编译，无 CGO。
- **unzip**：Go 交叉编译，无 CGO；CLI 对齐 Info-ZIP 子集（`-o` / `-d` / `-Z1`）。
- `scripts/package-unzip.sh` 已弃用，仅转发到 `build-unzip.sh`。

## 发布到 GitHub Release

```bash
# 1. 推送源码
git add -A && git commit -m "Replace darwin system unzip with portable Go unzip"
git push -u origin HEAD:main

# 2. 打 tag（触发 Actions 构建并上传 Release 资产），或本地构建后手动上传
TAG=v1.0.1
git tag "$TAG" && git push origin "$TAG"

# 手动上传示例（本地已 ./scripts/build-*.sh）：
gh release create "$TAG" \
  --title "ziplist 1.0.0 + portable unzip 6.0.0" \
  --notes "Portable Go unzip (fixes macOS AMFI kill of relocated /usr/bin/unzip)." \
  dist/ziplist/ziplist-1.0.0-darwin-arm64.tar.gz \
  dist/ziplist/ziplist-1.0.0-darwin-x64.tar.gz \
  dist/ziplist/ziplist-1.0.0-win-x64.zip \
  dist/ziplist/ziplist-1.0.0-linux-x64.tar.gz \
  dist/unzip/unzip-6.0.0-darwin-arm64.tar.gz \
  dist/unzip/unzip-6.0.0-darwin-x64.tar.gz \
  dist/unzip/unzip-6.0.0-win-x64.zip \
  dist/unzip/unzip-6.0.0-linux-x64.tar.gz
```

也可在 GitHub 网页：Releases → Draft a new release → 上传上述文件。

## 填「录入附件（官方直链转存）」

Release 资产直链形如：

`https://github.com/zhanshaobo-1992/mbws-attachments-tools/releases/download/<tag>/<filename>`

### ziplist @ 1.0.0

| 字段 | 值 |
| --- | --- |
| 附件名 | ziplist |
| 版本 | 1.0.0 |
| 用途说明 | 直读 ZIP 中央目录，UTF-8 列出条目名（避免 unzip -Z1 中文变 ?） |
| macOS Apple Silicon | .../ziplist-1.0.0-darwin-arm64.tar.gz |
| macOS Intel | .../ziplist-1.0.0-darwin-x64.tar.gz |
| Windows x64 | .../ziplist-1.0.0-win-x64.zip |
| Linux x64（选填） | .../ziplist-1.0.0-linux-x64.tar.gz |

### unzip @ 6.0.0（可移植 Go 实现，覆盖旧 darwin 坏包）

| 字段 | 值 |
| --- | --- |
| 附件名 | unzip |
| 版本 | 6.0.0 |
| 用途说明 | 可移植 ZIP 解压（插件 spawn：unzip -o / -Z1）；非系统 /usr/bin/unzip |
| macOS Apple Silicon | .../unzip-6.0.0-darwin-arm64.tar.gz |
| macOS Intel | .../unzip-6.0.0-darwin-x64.tar.gz |
| Windows x64 | .../unzip-6.0.0-win-x64.zip |
| Linux x64（选填） | .../unzip-6.0.0-linux-x64.tar.gz |

转存到 OSS 后请**强制刷新/覆盖**既有 `unzip@6.0.0` 的 darwin 平台 hash，否则宿主仍会下载旧的系统拷贝包。

## 插件侧声明

转存成功、OSS index 可访问后，在插件 `mbws.config.ts`：

```ts
attachments: [
  { name: "ffprobe", version: "6.1.1" },
  { name: "ffmpeg", version: "6.1.1" },
  { name: "ziplist", version: "1.0.0" },
  { name: "unzip", version: "6.0.0" },
],
```

本地 `pnpm dev` 可用 `devPath` / `MBWS_DEV_ATTACHMENT_UNZIP` 指到本机可执行文件做联调。

## 许可

- `ziplist` / `unzip` 源码：本仓库，MIT（见 `LICENSE`）。
- 旧版曾再分发系统 Info-ZIP / GnuWin32 二进制；现已改为自研 Go 实现，不再捆绑上游二进制。
