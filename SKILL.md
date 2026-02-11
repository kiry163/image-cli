---
name: image-cli
description: "Go image processing CLI tool with libvips. Supports convert, compress, resize, rotate, watermark, batch operations, OCR, AI image generation, and visual recognition."
metadata:
  {
    "openclaw":
      {
        "emoji": "🖼️",
        "requires": { "bins": ["image-cli"] },
        "install":
          [
            {
              "id": "install-script",
              "kind": "script",
              "url": "https://raw.githubusercontent.com/kiry163/image-cli/main/scripts/install.sh",
              "label": "Install ImageCLI (curl | bash)",
            },
            {
              "id": "manual-build",
              "kind": "build",
              "label": "Build from source (Go 1.23.12+)",
            },
          ],
      },
  }
---

# ImageCLI Skill 指南

面向 AI Agent 的 ImageCLI 使用指南，包含安装、依赖检查与命令示例。

## 1. 安装

推荐使用安装脚本（自动下载最新版本、检查依赖、执行 init、安装到 bin 目录）：

```bash
curl -fsSL https://raw.githubusercontent.com/kiry163/image-cli/main/scripts/install.sh | bash
```

默认安装到 `/usr/local/bin/image-cli`，如需修改安装目录：

```bash
INSTALL_DIR="$HOME/.local/bin" \
  curl -fsSL https://raw.githubusercontent.com/kiry163/image-cli/main/scripts/install.sh | bash
```

### 依赖要求

- Go >= 1.23.12（仅开发构建需要）
- libvips >= 8.13.0（必须）
- pkg-config（必须）
- ImageMagick（仅文字水印、ICO 转换、ICO 信息读取需要）

macOS 安装依赖：

```bash
brew install vips imagemagick
```

Linux (Debian/Ubuntu) 安装依赖：

```bash
sudo apt-get update
sudo apt-get install -y libvips libvips-dev pkg-config imagemagick
```

## 2. 配置

默认配置路径：`~/.config/image-cli/config.yaml`。

初始化默认配置：

```bash
image-cli config init
```

覆盖已有配置：

```bash
image-cli config init --overwrite
```

## 3. 命令使用指南

### 3.1 version

查看当前版本：

```bash
image-cli version
image-cli --version
```

### 3.2 formats

查看当前环境支持的输入/输出格式与可转换组合：

```bash
image-cli formats
```

过滤示例：

```bash
image-cli formats --from png
image-cli formats --to webp
image-cli formats --from png --to webp
```

### 3.3 info

查看图像信息（格式、尺寸、大小）：

```bash
image-cli info input.jpg
image-cli info input.webp
image-cli info input.ico
```

### 3.4 convert

格式转换。适合做格式统一、质量压缩或 ICO 生成前的预处理。

关键参数：
- `--format, -f` 输出格式（如 jpg/png/webp/avif/ico）
- `--quality, -q` 输出质量（1-100）
- `--overwrite` 覆盖已存在文件
- `--ico-sizes` ICO 尺寸列表（如 256,128,64）

```bash
image-cli convert input.jpg output.webp
image-cli convert input.jpg output.png --format png
image-cli convert input.jpg output.webp --quality 80
```

ICO 输出（依赖 ImageMagick）：

```bash
image-cli convert input.png output.ico --format ico
image-cli convert input.png output.ico --format ico --ico-sizes 256,128,64
```

### 3.5 compress

压缩图片。可指定质量或目标体积，适合批量压缩资源包。

关键参数：
- `--quality, -Q` JPEG/WebP 质量（1-100）
- `--max-size` 目标体积（如 200KB/1MB）
- `--aggressive` 激进压缩（可能更损画质）
- `--output, -o` 输出路径（目录或文件）

```bash
image-cli compress input.jpg --quality 75 --output ./output/
image-cli compress input.jpg --max-size 200KB --output ./output/
image-cli compress input.jpg --max-size 1MB --aggressive --output ./output/
```

### 3.6 resize

调整尺寸。支持 px/% 与 fit 模式，常用于适配不同尺寸资产。

关键参数：
- `--width, -w` 宽度（px 或 %）
- `--height` 高度（px 或 %）
- `--fit, -f` 适配模式（cover/contain/fill/inside/outside）
- `--without-enlargement` 不放大
- `--keep-ratio` 保持比例

```bash
image-cli resize input.jpg output.jpg --width 800
image-cli resize input.jpg output.jpg --height 600
image-cli resize input.jpg output.jpg --width 50% --height 50%
image-cli resize input.jpg output.jpg --width 800 --height 600 --fit cover
```

### 3.7 rotate

旋转/翻转：

```bash
image-cli rotate input.jpg output.jpg --degrees 90
image-cli rotate input.jpg output.jpg --flip
image-cli rotate input.jpg output.jpg --flop
```

### 3.8 watermark（图片水印）

图片水印适合品牌标识或版权保护，位置与透明度可控。

关键参数：
- `--gravity, -g` 水印位置（九宫格）
- `--opacity, -o` 透明度（0-1）
- `--scale, -s` 缩放比例（相对原图短边）
- `--offset-x/--offset-y` 像素偏移

```bash
image-cli watermark input.jpg logo.png output.jpg --opacity 0.6 --scale 0.2
image-cli watermark input.jpg logo.png output.jpg --gravity center --offset-x 10 --offset-y -10
```

### 3.9 watermark（文字水印）

文字水印支持字体、颜色、描边与背景，适合动态标识。默认使用内置字体，可通过 `--font-file` 指定外部字体。

关键参数：
- `--text` 文字内容
- `--font-size` 字号（px）
- `--font` 字体名称
- `--font-file` 字体文件路径
- `--color` 文字颜色（如 #ffffff）
- `--stroke-color` 描边颜色
- `--stroke-width` 描边宽度（px）
- `--stroke-mode` 描边模式（circle/8dir）
- `--background` 背景色（如 #000000 或 none）

```bash
image-cli watermark input.jpg output.jpg --text "Sample" --font-size 24 --gravity southeast
image-cli watermark input.jpg output.jpg --text "Sample" --font-size 24 --font "Arial" --color "#ffffff"
image-cli watermark input.jpg output.jpg --text "Sample" --font-size 24 --font-file "/path/to/font.ttf"
image-cli watermark input.jpg output.jpg --text "Sample" --font-size 24 --stroke-color black --stroke-width 2
image-cli watermark input.jpg output.jpg --text "Sample" --font-size 24 --stroke-color black --stroke-width 2 --stroke-mode 8dir
image-cli watermark input.jpg output.jpg --text "Sample" --font-size 24 --background "#000000" --color "#ffffff"
```

### 3.10 batch

批量处理适合大规模素材改造，支持通配符与目录输入。

关键参数：
- `--output, -o` 输出目录
- `--to` 批量转换目标格式（用于 convert）
- `--quality`/`--max-size`/`--aggressive`（用于 compress）
- `--width`/`--height`/`--fit`（用于 resize）
- `--degrees`/`--flip`/`--flop`（用于 rotate）
- `--logo` 或 `--text`（用于 watermark）

```bash
image-cli batch convert "./images/*.png" --to webp --quality 85 --output ./output/
image-cli batch compress "./images/*.jpg" --quality 80 --max-size 200KB --output ./output/
image-cli batch resize "./images" --width 800 --height 600 --fit cover --output ./output/
image-cli batch rotate "./images" --degrees 90 --output ./output/
image-cli batch watermark "./images" --logo logo.png --opacity 0.6 --output ./output/
image-cli batch watermark "./images" --text "Sample" --font-size 24 --font "Arial" --color "#ffffff" --stroke-color black --stroke-width 2 --output ./output/
```

### 3.11 ocr（OCR 文字识别）

使用 DeepSeek OCR API 从图片中提取文字内容，支持多种识别模式。

关键参数：
- `--mode, -m` 识别模式：`free`(默认), `markdown`, `text`, `figure`, `detail`
- `--output, -o` 输出文件路径（默认输出到控制台）

```bash
# 基础识别
image-cli ocr document.jpg

# 输出为 Markdown 格式
image-cli ocr document.png --mode markdown

# 保存到文件
image-cli ocr scan.jpg --mode text --output result.txt
```

### 3.12 generate（AI 图像生成）

使用智谱 AI CogView 模型根据文本描述生成图像。支持多种模型和尺寸。

关键参数：
- `--model, -m` 模型：`cogview-3-flash`(免费), `glm-image`, `cogview-4`
- `--size, -s` 图片尺寸：`1024x1024`, `768x1344`, `864x1152` 等
- `--quality, -q` 质量：`standard`(默认), `hd`
- `--output, -o` 输出文件路径

```bash
# 基本使用（免费模型）
image-cli generate "一只可爱的小猫咪"

# 指定模型和尺寸
image-cli generate "夕阳下的海景" --model cogview-3-flash --size 1440x720

# 高质量生成
image-cli generate "科幻城市" --quality hd --output ./output/futuristic.png
```

### 3.13 recognize（AI 图片识别）

使用智谱 AI GLM-4V 模型对图片进行视觉理解和分析，支持图片描述、分类、图表分析等。

关键参数：
- `--model, -m` 模型：`glm-4v-flash`(免费), `glm-4.6v`
- `--prompt, -p` 分析提示词/问题（默认："请描述这张图片的内容"）
- `--output, -o` 输出文件路径

```bash
# 基本描述
image-cli recognize photo.jpg

# 分析图表数据
image-cli recognize chart.png --prompt "分析这张图表的数据趋势"

# 商品分类
image-cli recognize product.jpg --prompt "这是什么商品？请给出类别和特点"

# 保存分析结果
image-cli recognize scan.jpg --prompt "提取图中的文字并总结内容" --output analysis.txt
```

## 4. 全局参数

```bash
--config, -c     配置文件路径 (默认 ~/.config/image-cli/config.yaml)
--verbose, -v    详细输出
--quiet          静默模式
--recursive      目录递归处理 (默认 true)
--no-recursive   关闭递归
--conflict       冲突策略: skip|overwrite|rename (默认 skip)
--version, -V    显示版本
```

## 5. 常见问题

### 5.1 watermarks/ico 相关错误

- 提示需要 ImageMagick：安装 `imagemagick` 后重试。
- 中文水印失败：安装中文字体（如 `fonts-noto-cjk` 或 `fonts-wqy-zenhei`）。

### 5.2 配置文件错误

- 若不需要自定义配置，可不创建 `config.yaml`。
- 使用 `image-cli config init` 自动生成。
