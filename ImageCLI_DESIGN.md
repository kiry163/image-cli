# ImageCLI 设计文档

## 1. 项目概述

### 1.1 简介
ImageCLI 是一个基于 Go + libvips 的图像处理命令行工具，提供基础的图像处理功能和 AI 增强能力。

### 1.2 设计目标
- **高性能**：Go + libvips，原生编译
- **轻量级**：单二进制部署，零运行时依赖
- **模块化**：基础功能与 AI 功能分离
- **可扩展**：支持自定义 AI 模型接入
- **易集成**：可作为独立 CLI 使用，也可集成到 Agent 中

---

## 2. 核心功能

### 2.1 基础功能（libvips via bimg）

| 功能 | 说明 | 示例 |
|------|------|------|
| 格式转换 | JPG ↔ PNG ↔ WebP ↔ GIF ↔ AVIF | `image-cli convert input.jpg output.webp` |
| 图像压缩 | 质量压缩、尺寸优化 | `image-cli compress photo.jpg --quality 80` |
| 尺寸调整 | 缩放、裁剪、适应 | `image-cli resize image.jpg --width 800` |
| 旋转翻转 | 90° 旋转、水平/垂直翻转 | `image-cli rotate photo.jpg --degrees 90` |
| 简单水印 | 图片叠加定位 | `image-cli watermark photo.jpg logo.png --gravity southeast` |
| 图像信息 | 查看 EXIF，元数据 | `image-cli info image.jpg` |

### 2.2 AI 功能（外部模型）

| 功能 | 实现方式 | 说明 |
|------|----------|------|
| 去除水印 | 调用大模型 API | `image-cli remove-watermark input.jpg` |
| 智能抠图 | 调用大模型 API | `image-cli remove-bg input.jpg` |
| AI 超分辨率 | 调用大模型 API | `image-cli enhance input.jpg --scale 2` |
| 风格迁移 | 调用大模型 API | `image-cli style-transfer input.jpg --style oil` |

---

## 3. 架构设计

### 3.1 整体架构

```
ImageCLI
├── cmd/                           # CLI 命令入口
│   ├── root.go                    # 根命令
│   ├── convert.go                  # 格式转换
│   ├── compress.go                # 图像压缩
│   ├── resize.go                  # 尺寸调整
│   ├── rotate.go                  # 旋转翻转
│   ├── watermark.go               # 水印添加
│   ├── info.go                    # 图像信息
│   ├── remove-watermark.go        # 去除水印（AI）
│   ├── remove-bg.go               # 智能抠图（AI）
│   ├── enhance.go                 # AI 增强
│   ├── style-transfer.go          # 风格迁移（AI）
│   └── batch.go                   # 批量处理
├── internal/                      # 内部实现
│   ├── core/                      # 基础功能（libvips）
│   │   ├── converter.go           # 格式转换
│   │   ├── compressor.go          # 图像压缩
│   │   ├── resizer.go             # 尺寸调整
│   │   ├── rotator.go             # 旋转翻转
│   │   ├── watermarker.go         # 水印添加
│   │   ├── cropper.go             # 裁剪
│   │   └── metadata.go            # 元数据
│   ├── ai/                        # AI 功能
│   │   ├── base.go                # AI 基类
│   │   ├── remover.go             # 去除水印
│   │   ├── background.go           # 智能抠图
│   │   ├── enhancer.go            # 超分辨率
│   │   └── styler.go              # 风格迁移
│   └── batch/                     # 批量处理
│       ├── processor.go           # 批处理引擎
│       └── workers.go             # 并发控制
├── pkg/                           # 公共包
│   ├── config/                    # 配置管理
│   │   ├── config.go              # 配置加载
│   │   └── defaults.go            # 默认配置
│   ├── logger/                    # 日志
│   ├── validator/                 # 参数验证
│   └── file/                      # 文件操作
├── config/
│   └── config.yaml                # 配置文件
├── scripts/
│   ├── build.sh                   # 构建脚本
│   └── install.sh                 # 安装脚本
├── docs/
│   ├── README.md                  # 使用说明
│   └── EXAMPLES.md                # 示例
├── tests/
│   ├── unit/                      # 单元测试
│   └── integration/               # 集成测试
├── go.mod
├── go.sum
└── main.go
```

### 3.2 技术栈

| 组件 | 技术选型 | 说明 |
|------|----------|------|
| 语言 | Go 1.23.12 | 现代化 Go 特性 |
| CLI 框架 | github.com/spf13/cobra | 业界标准 CLI 框架 |
| 配置管理 | github.com/spf13/viper | YAML + 环境变量支持 |
| 图像处理 | github.com/h2non/bimg | libvips Go 绑定 |
| AI 集成 | github.com/sashabaranov/go-openai | OpenAI 官方 SDK |
| YAML 解析 | gopkg.in/yaml.v3 | 结构化配置 |

---

## 4. CLI 命令设计

### 4.1 命令结构

```bash
image-cli <command> <input> [options]
```

#### 4.1.1 输入与通配符约定

- 支持单文件、目录、通配符（由 shell 展开）。
- 目录输入默认递归处理，支持通过参数关闭递归。
- 当输入为通配符或目录时，必须指定 `--output` 目录；默认输出目录为 `./output`。
- 输出文件命名规则：保留原文件名，后缀按目标格式或命令指定。
- 若输出目录不存在则创建；默认不覆盖已存在文件，需 `--overwrite`。

#### 4.1.2 批量输出与冲突处理

- 当输入为多个文件时，输出路径视为目录；单文件时输出可为文件或目录。
- 文件名冲突策略（默认 `skip`）：
  - `skip`：跳过并记录警告
  - `overwrite`：覆盖已有文件
  - `rename`：追加后缀 `_1`, `_2` 直到可用
- 目标目录下保持原始相对路径结构（默认开启，可通过参数关闭）。

### 4.2 命令列表

#### 4.2.0 config - 配置管理

```bash
image-cli config init

Options:
  --overwrite     覆盖已存在配置文件
```

#### 4.2.0.1 formats - 格式能力查询

```bash
image-cli formats

Options:
  --from    输入格式过滤
  --to      输出格式过滤

输出:
  - 支持的输入格式
  - 支持的输出格式
  - 可转换格式组合
```

#### 4.2.0.2 version - 查看版本

```bash
image-cli version
image-cli --version
```

#### 4.2.1 convert - 格式转换

```bash
image-cli convert <input> <output> [options]

Options:
  --format, -f     输出格式 (jpg, png, webp, gif, avif, tiff, pdf, ico)
  --quality, -q    质量 (1-100, 默认 85)
  --overwrite       覆盖已存在文件
  --ico-sizes      ICO 尺寸列表 (如 256,128,64)

说明:
  - `ico` 输出依赖 ImageMagick（使用 `magick` 或 `convert`），且需要 PNG 输出可用
  - `ico` 默认尺寸: 256,128,64,48,32,16

Examples:
  image-cli convert photo.jpg photo.webp
  image-cli convert *.png --format webp --quality 80
  image-cli convert image.png --format pdf
```

#### 4.2.2 compress - 图像压缩

```bash
image-cli compress <input> [options]

Options:
  --quality, -Q    JPEG/WebP 质量 (1-100)
  --max-size       最大文件大小 (如 100KB, 1MB)
  --output, -o     输出路径
  --aggressive     激进压缩（可能降低质量）

Examples:
  image-cli compress photo.jpg --quality 75
  image-cli compress photo.jpg --max-size 100KB --output small/
  image-cli compress *.jpg --quality 80 --output compressed/
```

#### 4.2.3 resize - 尺寸调整

```bash
image-cli resize <input> <output> [options]

Options:
  --width, -w       宽度 (px 或 %)
  --height         高度 (px 或 %)
  --fit, -f        适应模式: cover | contain | fill | inside | outside
  --without-enlargement  不放大（默认）
  --keep-ratio     保持比例（默认）

Examples:
  image-cli resize photo.jpg --width 800 output/
  image-cli resize image.png --height 600 --fit cover
  image-cli resize *.jpg --width 1920 --keep-ratio
```

#### 4.2.4 rotate - 旋转翻转

```bash
image-cli rotate <input> <output> [options]

Options:
  --degrees, -d    旋转角度 (90, 180, 270, -90)
  --flip           水平翻转
  --flop           垂直翻转

Examples:
  image-cli rotate photo.jpg --degrees 90
  image-cli rotate image.png --flip
  image-cli rotate *.jpg --degrees 180
```

#### 4.2.5 watermark - 添加水印

```bash
image-cli watermark <input> <logo> <output> [options]
image-cli watermark <input> <output> --text "Hello" [options]

Options:
  --gravity, -g   位置: northwest | north | northeast | west | center |
                   east | southwest | south | southeast (默认 southeast)
  --opacity, -o    水印透明度 (0-1, 默认 0.5)
  --scale, -s      水印缩放比例 (默认 0.2)
  --offset-x       水平偏移 (px)
  --offset-y       垂直偏移 (px)
  --text           文字水印内容
  --font-size      文字水印字号 (px)
  --font           文字水印字体
  --font-file      文字水印字体文件
  --color          文字水印颜色
  --stroke-color   文字水印描边颜色
  --stroke-width   文字水印描边宽度 (px)
  --stroke-mode    描边模式 (circle|8dir)
  --background     文字水印背景色

Examples:
  image-cli watermark photo.jpg logo.png output/
  image-cli watermark photo.jpg logo.png --gravity center --opacity 0.8
  image-cli watermark *.jpg logo.png --gravity southeast --scale 0.15
  image-cli watermark photo.jpg output/ --text "Sample" --font-size 24
  image-cli watermark photo.jpg output/ --text "Sample" --font-size 24 --font "Arial" --color "#ffffff"
  image-cli watermark photo.jpg output/ --text "Sample" --font-size 24 --font-file "/path/to/font.ttf"
  image-cli watermark photo.jpg output/ --text "Sample" --font-size 24 --stroke-color black --stroke-width 2
  image-cli watermark photo.jpg output/ --text "Sample" --font-size 24 --stroke-color black --stroke-width 2 --stroke-mode 8dir
  image-cli watermark photo.jpg output/ --text "Sample" --font-size 24 --background "#000000" --color "#ffffff"

说明:
  - 文字水印默认使用内置字体，可通过 `--font-file` 指定外部字体
```

#### 4.2.6 info - 查看信息

```bash
image-cli info <input>

Output:
  - 文件名、尺寸、格式
  - EXIF 数据（相机、GPS、日期等）
  - 文件大小、颜色深度
```

#### 4.2.7 remove-watermark - 去除水印（AI）

```bash
image-cli remove-watermark <input> [options]

Options:
  --output, -o     输出路径
  --model, -m      使用模型 (gpt-4o, claude-3-5, gemini-1.5)
  --api-key        API Key (或使用环境变量)
  --format         输出格式 (默认与输入一致)

Examples:
  image-cli remove-watermark photo.jpg
  image-cli remove-watermark photo.jpg --model gpt-4o
  image-cli remove-watermark photo.jpg --output cleaned.jpg
```

#### 4.2.8 remove-bg - 智能抠图（AI）

```bash
image-cli remove-bg <input> [options]

Options:
  --output, -o     输出路径
  --model, -m      使用模型
  --matte          保留边缘细节
  --format         输出格式 (默认 png)

Examples:
  image-cli remove-bg person.jpg
  image-cli remove-bg photo.png --matte
  image-cli remove-bg *.png --output transparent/
```

#### 4.2.9 enhance - AI 图像增强

```bash
image-cli enhance <input> [options]

Options:
  --scale, -s      放大倍数 (2, 4, 8)
  --model, -m      超分辨率模型
  --denoise        降噪
  --sharpen        锐化
  --format         输出格式 (默认与输入一致)

Examples:
  image-cli enhance photo.jpg --scale 2
  image-cli enhance old-photo.png --denoise --sharpen
```

#### 4.2.10 batch - 批量处理

```bash
image-cli batch <command> <pattern> [options]

支持命令:
  batch convert <pattern> --to webp --quality 85
  batch compress <pattern> --quality 75 --max-size 200KB
  batch watermark <pattern> --logo logo.png
  batch resize <pattern> --width 800 --height 600 --fit cover
  batch rotate <pattern> --degrees 90

Examples:
  image-cli batch convert *.png --to webp
  image-cli batch compress *.jpg --quality 80 --output ./compressed/
```

---

## 5. 参数规范

### 5.1 全局参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| --config, -c | 配置文件路径 | `~/.config/image-cli/config.yaml` |
| --verbose, -v | 详细输出 | false |
| --quiet | 静默模式 | false |
| --recursive | 目录递归处理 | true |
| --no-recursive | 关闭递归 | false |
| --conflict | 冲突策略: skip|overwrite|rename | skip |
| --help, -h | 显示帮助 | - |
| --version, -V | 显示版本 | - |

### 5.2 配置文件 (config.yaml)

```yaml
# ImageCLI 配置

# 基础设置
base:
  output_dir: ./output
  overwrite: false
  keep_temp: false
  recursive: true
  conflict: skip

# 压缩设置
compress:
  default_quality: 85
  max_width: 4096
  max_height: 4096

# 水印设置
watermark:
  default_opacity: 0.5
  default_scale: 0.2
  default_gravity: southeast
  default_offset_x: 0
  default_offset_y: 0
  default_font_size: 24
  default_font: ""
  default_font_file: ""
  default_color: white
  default_stroke_color: ""
  default_stroke_width: 0
  default_background: none
  default_stroke_mode: circle

# AI 模型配置
ai:
  default_model: gpt-4o

  # 输出行为
  output:
    default_format: ""   # 空表示与输入一致
    remove_bg_format: png

  models:
    gpt-4o:
      provider: openai
      api_key_env: OPENAI_API_KEY
      endpoint: https://api.openai.com/v1

    claude-3-5-sonnet:
      provider: anthropic
      api_key_env: ANTHROPIC_API_KEY
      endpoint: https://api.anthropic.com

    gemini-1.5-pro:
      provider: google
      api_key_env: GOOGLE_API_KEY
      endpoint: https://generativelanguage.googleapis.com/v1

# 日志设置
logging:
  level: info  # debug, info, warn, error
  format: json  # json, text
```

### 5.3 环境变量

```bash
# API Keys（可在 config.yaml 中配置）
export OPENAI_API_KEY="sk-..."
export ANTHROPIC_API_KEY="sk-ant-..."
export GOOGLE_API_KEY="AIza..."

# 其他设置
export IMAGE_CLI_CONFIG="/path/to/config.yaml"
export IMAGE_CLI_OUTPUT="/path/to/output"
export IMAGE_CLI_RECURSIVE=true

### 5.4 参数优先级

优先级从高到低：命令行参数 > 环境变量 > 配置文件 > 默认值。
```

---

## 6. 依赖和配置

### 6.1 Go 依赖

```go
require (
    github.com/spf13/cobra v1.8.0
    github.com/spf13/viper v1.18.2
    github.com/h2non/bimg v1.2.0
    github.com/sashabaranov/go-openai v1.5.0
    gopkg.in/yaml.v3 v3.0.1
    github.com/lucasb-eyer/go-colorful v1.2.0
    github.com/stretchr/testify v1.8.4
)
```

### 6.2 系统依赖

- **libvips**: >= 8.13.0
  - Debian/Ubuntu: `apt-get install libvips`
  - macOS: `brew install vips`
  - 从源码编译需要: `pkg-config`, `glib2`, `libexif`, `libjpeg`, `libpng`, `libwebp`, `libtiff`

### 6.3 安装要求

- **Go**: >= 1.23.12

---

## 7. 使用示例

### 7.1 基础用法

```bash
# 查看帮助
image-cli --help

# 查看版本
image-cli --version

# 转换格式
image-cli convert input.jpg output.webp

# 压缩图片
image-cli compress photo.jpg --quality 80

# 调整尺寸
image-cli resize photo.jpg --width 800 --output small/

# 添加水印
image-cli watermark photo.jpg logo.png --gravity southeast

# 批量处理
image-cli batch convert *.png --to webp --quality 85
```

### 7.2 AI 功能

```bash
# 去除水印
image-cli remove-watermark photo.jpg

# 智能抠图
image-cli remove-bg person.png --output transparent/

# AI 增强
image-cli enhance old-photo.jpg --scale 2 --denoise
```

### 7.3 复合任务

```bash
# 转换 + 压缩 + 调整尺寸
image-cli convert photo.jpg | \
image-cli compress --quality 80 | \
image-cli resize --width 800

# 使用管道
cat images.txt | xargs -I {} image-cli convert {} --format webp
```

---

## 8. 错误处理

### 8.1 错误代码

| 代码 | 说明 |
|------|------|
| E001 | 无效的输入文件 |
| E002 | 不支持的格式 |
| E003 | 文件不存在 |
| E004 | 权限错误 |
| E005 | 配置错误 |
| E006 | 输出文件已存在 |
| E007 | 参数错误 |
| E008 | 批量处理部分失败 |
| E101 | AI API 调用失败 |
| E102 | API Key 未配置 |
| E103 | 模型不支持 |
| E104 | AI 功能未实现 |
| E900 | 功能未实现 |

### 8.3 退出码约定

- `0` 成功
- `1` 通用错误
- `2` 参数或配置错误
- `3` 输入文件错误
- `4` AI 调用错误

### 8.2 错误输出示例

```bash
$ image-cli convert invalid.jpg output.webp
Error [E001]: 无效的输入文件
  → 文件不存在或无法读取: invalid.jpg

$ image-cli remove-watermark photo.jpg
Error [E102]: API Key 未配置
  → 请设置环境变量或配置文件中的 API Key
  → 示例: export OPENAI_API_KEY="sk-..."
```

---

## 9. 扩展性设计

### 9.1 自定义 AI 模型

```go
// 注册自定义模型
import "github.com/kiry163/image-cli/internal/ai"

func init() {
    ai.RegisterModel(ai.ModelConfig{
        Name:      "my-model",
        Endpoint:  "https://my-api.com/v1/enhance",
        APIKeyEnv: "MY_API_KEY",
        Headers: map[string]string{
            "Authorization": "Bearer {{API_KEY}}",
        },
        Process: func(image []byte, options map[string]interface{}) ([]byte, error) {
            // 自定义处理逻辑
            return enhancedImage, nil
        },
    })
}
```

### 9.2 自定义处理器

```go
// 注册自定义处理器
import "github.com/kiry163/image-cli/internal/core"

func init() {
    core.RegisterHandler(core.HandlerConfig{
        Name: "my-effect",
        Process: func(img *bimg.Image, options map[string]interface{}) (*bimg.Image, error) {
            // 自定义效果
            return img, nil
        },
    })
}
```

---

## 10. 测试

### 10.1 测试命令

```bash
# 运行所有测试
go test ./...

# 运行单元测试
go test ./internal/... -v

# 运行集成测试
go test ./tests/integration/... -v

# 代码覆盖率
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# 代码检查
go vet ./...

# 格式化代码
gofmt -w .
```

### 10.2 基准测试

```bash
# 性能测试
go test ./internal/core/... -bench=.

# 内存使用测试
go test ./internal/core/... -benchmem -bench=BenchmarkResize
```

---

## 11. 性能优化

### 11.1 优化建议

- **批量处理**：使用 goroutine 并行处理
- **内存管理**：及时释放 bimg.Image 实例
- **缓存**：对重复操作启用缓存
- **分块处理**：大文件分块处理

### 11.2 性能指标

| 操作 | 基准 | 目标 |
|------|------|------|
| 格式转换 | 50ms (1MB JPG) | < 100ms |
| 压缩 | 80ms (1MB JPG) | < 150ms |
| 调整尺寸 | 100ms (2MB PNG) | < 200ms |
| AI 处理 | 5s (API) | < 10s |

---

## 12. 构建与发布

### 12.1 构建命令

```bash
# 本地构建
go build -o image-cli main.go

# 跨平台构建
GOOS=linux GOARCH=amd64 go build -o image-cli-linux-amd64 main.go
GOOS=darwin GOARCH=amd64 go build -o image-cli-darwin-amd64 main.go
GOOS=darwin GOARCH=arm64 go build -o image-cli-darwin-arm64 main.go
GOOS=windows GOARCH=amd64 go build -o image-cli.exe main.go

# 使用构建脚本
chmod +x scripts/build.sh
./scripts/build.sh
```

### 12.2 发布流程

```bash
# 1. 更新版本号
# 修改 cmd/root.go 中的 Version 常量

# 2. 生成变更日志
git log --oneline --since="v0.0.0" --until="v0.1.0" > CHANGELOG.md

# 3. 创建标签
git tag -a v0.1.0 -m "Release v0.1.0"

# 4. 构建并发布
goreleaser release --rm-dist
```

---

## 13. 版本历史

### v0.1.0 (计划首发版本)
- ✨ 初始版本发布（未发布）
- ✨ 基础图像处理功能（convert, compress, resize, rotate, watermark, info）
- ✨ 批量处理功能
- ✨ 配置文件支持
- ✨ AI 增强功能（预留）

### 后续版本规划

#### v0.2.0
- [ ] AI 去除水印功能
- [ ] AI 智能抠图功能
- [ ] AI 超分辨率功能

#### v0.3.0
- [ ] 插件系统
- [ ] 自定义处理器注册
- [ ] Web UI 界面（可选）

---

## 14. 贡献指南

### 14.1 开发流程
1. Fork 项目
2. 创建功能分支
3. 编写代码和测试
4. 提交 PR
5. 代码审查
6. 合并

### 14.2 代码规范
- 遵循 Effective Go
- 使用 `gofmt` 格式化代码
- 编写单元测试（覆盖率 > 80%）
- 遵循 Semantic Versioning

### 14.3 提交规范
```
feat: 新功能
fix: 修复 bug
docs: 文档更新
style: 代码格式（不影响功能）
refactor: 重构
perf: 性能优化
test: 测试
chore: 构建/工具链
```

---

## 15. 许可协议

MIT License

---

## 16. 联系

- 作者: kiry
- 项目: https://github.com/kiry163/image-cli
- 问题反馈: https://github.com/kiry163/image-cli/issues

---

## 17. 附录

### A. 安装脚本示例

```bash
#!/bin/bash
# install.sh - 安装脚本

set -e

# 检测系统
OS="$(uname -s)"
ARCH="$(uname -m)"

# 下载对应平台的二进制
case "$OS" in
    Linux)
        case "$ARCH" in
            x86_64)
                URL="https://github.com/kiry163/image-cli/releases/download/v0.1.0/image-cli-linux-amd64"
                ;;
            aarch64)
                URL="https://github.com/kiry163/image-cli/releases/download/v0.1.0/image-cli-linux-arm64"
                ;;
        esac
        ;;
    Darwin)
        case "$ARCH" in
            x86_64)
                URL="https://github.com/kiry163/image-cli/releases/download/v0.1.0/image-cli-darwin-amd64"
                ;;
            arm64)
                URL="https://github.com/kiry163/image-cli/releases/download/v0.1.0/image-cli-darwin-arm64"
                ;;
        esac
        ;;
esac

# 下载并安装
curl -L -o image-cli "$URL"
chmod +x image-cli
sudo mv image-cli /usr/local/bin/

echo "✅ ImageCLI v0.1.0 安装完成！"
```

### B. Makefile 示例

```makefile
.PHONY: all build test clean install uninstall

# 变量
VERSION := 0.1.0
BINARY_NAME := image-cli
BUILD_DIR := build

all: build

build:
	@echo "🏗️  构建 ImageCLI v$(VERSION)..."
	go build -o $(BUILD_DIR)/$(BINARY_NAME) main.go
	@echo "✅ 构建完成: $(BUILD_DIR)/$(BINARY_NAME)"

test:
	@echo "🧪 运行测试..."
	go test ./... -v -cover

clean:
	@echo "🧹 清理..."
	rm -rf $(BUILD_DIR)
	@echo "✅ 清理完成"

install: build
	@echo "📦 安装..."
	sudo mv $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "✅ 安装完成"

uninstall:
	@echo "🗑️  卸载..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "✅ 卸载完成"

release: build
	@echo "🚀 准备发布 v$(VERSION)..."
	# 这里可以添加发布到 GitHub Release 的逻辑
```

---

*文档版本: v2.0.0*
*最后更新: 2026-02-07*
