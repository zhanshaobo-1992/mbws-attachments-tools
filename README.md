# mbws-attachments-tools

为 MBWS 插件附件库准备的 **`ziplist`** / **`unzip`** 多平台包。

仓库：[zhanshaobo-1992/mbws-attachments-tools](https://github.com/zhanshaobo-1992/mbws-attachments-tools)

## 附件约定

| 附件名 | 版本（建议） | 入口 | 用途 |
|---|---|---|---|
| `ziplist` | `1.0.0` | `ziplist` / `ziplist.exe` | 按 UTF-8 列出 ZIP 中央目录条目（一行一个） |
| `unzip` | `6.0.0` | `unzip` / `unzip.exe` | 解压：`unzip -o <zip> -d <dir>`；回退列举：`unzip -Z1` |

包根目录**只能有可执行入口**（与 ffmpeg 附件布局一致）。

## 本地构建

```bash
# 需要 Go（ziplist）与 curl/unzip/tar/zip
export PATH="/path/to/go/bin:$PATH"   # 若未装 Go
./scripts/build-ziplist.sh
./scripts/package-unzip.sh
```

产物在 `dist/ziplist/`、`dist/unzip/`。

说明：

- **ziplist**：Go 交叉编译，无 CGO。
- **unzip（darwin）**：打包本机 `/usr/bin/unzip`（常见为 universal）。
- **unzip（win）**：从 [GnuWin32 unzip 5.51](https://sourceforge.net/projects/gnuwin32/files/unzip/) 取出 `unzip.exe` 重打扁平包；版本号槽位用 `6.0.0` 便于与 MBWS semver 对齐（见包内说明）。
- **unzip（linux）**：可选，设置 `UNZIP_LINUX_BIN=/path/to/unzip` 后再跑打包脚本。

## 发布到 GitHub Release

```bash
# 1. 推送源码
git add -A && git commit -m "Add ziplist/unzip packaging for MBWS attachments"
git push -u origin main

# 2. 创建 Release 并上传资产（需 gh 已登录）
TAG=v1.0.0
gh release create "$TAG" \
  --title "ziplist 1.0.0 + unzip 6.0.0" \
  --notes "MBWS attachment binaries for ZIP list/extract." \
  dist/ziplist/ziplist-1.0.0-darwin-arm64.tar.gz \
  dist/ziplist/ziplist-1.0.0-darwin-x64.tar.gz \
  dist/ziplist/ziplist-1.0.0-win-x64.zip \
  dist/ziplist/ziplist-1.0.0-linux-x64.tar.gz \
  dist/unzip/unzip-6.0.0-darwin-arm64.tar.gz \
  dist/unzip/unzip-6.0.0-darwin-x64.tar.gz \
  dist/unzip/unzip-6.0.0-win-x64.zip
```

也可在 GitHub 网页：Releases → Draft a new release → 上传上述文件。

## 填「录入附件（官方直链转存）」

Release 资产直链形如：

`https://github.com/zhanshaobo-1992/mbws-attachments-tools/releases/download/<tag>/<filename>`

### ziplist @ 1.0.0

| 字段 | 值 |
|---|---|
| 附件名 | `ziplist` |
| 版本 | `1.0.0` |
| 用途说明 | 直读 ZIP 中央目录，UTF-8 列出条目名（避免 unzip -Z1 中文变 ?） |
| macOS Apple Silicon | `.../ziplist-1.0.0-darwin-arm64.tar.gz` |
| macOS Intel | `.../ziplist-1.0.0-darwin-x64.tar.gz` |
| Windows x64 | `.../ziplist-1.0.0-win-x64.zip` |
| Linux x64（选填） | `.../ziplist-1.0.0-linux-x64.tar.gz` |

### unzip @ 6.0.0

| 字段 | 值 |
|---|---|
| 附件名 | `unzip` |
| 版本 | `6.0.0` |
| 用途说明 | ZIP 解压（插件 spawn：unzip -o / -Z1） |
| macOS Apple Silicon | `.../unzip-6.0.0-darwin-arm64.tar.gz` |
| macOS Intel | `.../unzip-6.0.0-darwin-x64.tar.gz` |
| Windows x64 | `.../unzip-6.0.0-win-x64.zip` |

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

## 许可

- `ziplist` 源码：本仓库，MIT（见 `LICENSE`）。
- `unzip` 二进制：来自系统 Info-ZIP / GnuWin32，遵循其上游许可；再分发请自行确认合规。
