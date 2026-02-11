# ImageCLI

ImageCLI 是一个基于 Go + libvips 的图像处理命令行工具。支持基础图像处理、批量操作和 AI 功能（OCR、图像生成、视觉识别）。

## 依赖

- Go >= 1.23.12
- libvips >= 8.13.0

macOS:

```bash
brew install vips
```

Linux (Debian/Ubuntu):

```bash
sudo apt-get update
sudo apt-get install -y libvips libvips-dev pkg-config
```

## 构建

```bash
go build -o image-cli main.go
```

## 安装

```bash
curl -fsSL https://raw.githubusercontent.com/kiry163/image-cli/main/scripts/install.sh | bash
```


## 配置文件

默认读取 `~/.config/image-cli/config.yaml`（若未显式指定且文件不存在则忽略）。

示例配置：`config/config.example.yaml`

初始化默认配置：

```bash
image-cli config init
```

覆盖已有配置：

```bash
image-cli config init --overwrite
```

## 当前支持的命令

### formats

查看当前环境支持的输入/输出格式以及可转换组合。

```bash
image-cli formats
```

筛选示例：

```bash
image-cli formats --from png
image-cli formats --to webp
image-cli formats --from png --to webp
```

### info

查看图像基础信息（格式、尺寸、文件大小）。

```bash
image-cli info input.jpg
```

### version

查看版本信息。

```bash
image-cli version
image-cli --version
```

### convert

格式转换，支持指定质量与输出格式。

```bash
image-cli convert input.jpg output.webp
image-cli convert input.jpg output.png --format png
image-cli convert input.jpg output.webp --quality 80
image-cli convert input.png output.ico --format ico
image-cli convert input.png output.ico --format ico --ico-sizes 256,128,64
```

说明: `ico` 输出依赖 ImageMagick（`magick` 或 `convert`），并要求 PNG 输出可用。默认尺寸为 256,128,64,48,32,16。

### compress

压缩图片，支持最大体积与激进压缩。

```bash
image-cli compress input.jpg --quality 75 --output ./output/
image-cli compress input.jpg --max-size 200KB --output ./output/
image-cli compress input.jpg --max-size 1MB --aggressive --output ./output/
```

### resize

调整尺寸，支持 px/% 与 fit 模式。

```bash
image-cli resize input.jpg output.jpg --width 800
image-cli resize input.jpg output.jpg --height 600
image-cli resize input.jpg output.jpg --width 50% --height 50%
image-cli resize input.jpg output.jpg --width 800 --height 600 --fit cover
```

### rotate

旋转/翻转。

```bash
image-cli rotate input.jpg output.jpg --degrees 90
image-cli rotate input.jpg output.jpg --flip
image-cli rotate input.jpg output.jpg --flop
```

### watermark

添加图片或文字水印。

```bash
image-cli watermark input.jpg logo.png output.jpg --opacity 0.6 --scale 0.2
image-cli watermark input.jpg output.jpg --text "Sample" --font-size 24 --gravity southeast
image-cli watermark input.jpg output.jpg --text "Sample" --font-size 24 --font "Arial" --color "#ffffff"
image-cli watermark input.jpg output.jpg --text "Sample" --font-size 24 --font-file "/path/to/font.ttf"
image-cli watermark input.jpg output.jpg --text "Sample" --offset-x 10 --offset-y -10
image-cli watermark input.jpg output.jpg --text "Sample" --font-size 24 --stroke-color black --stroke-width 2
image-cli watermark input.jpg output.jpg --text "Sample" --font-size 24 --stroke-color black --stroke-width 2 --stroke-mode 8dir
image-cli watermark input.jpg output.jpg --text "Sample" --font-size 24 --background "#000000" --color "#ffffff"
```

说明: 文字水印默认使用内置字体，亦可通过 `--font-file` 指定字体文件。

### batch

批量处理（支持通配符/目录，默认保留相对路径结构）。

```bash
image-cli batch convert "./images/*.png" --to webp --quality 85 --output ./output/
image-cli batch compress "./images/*.jpg" --quality 80 --max-size 200KB --output ./output/
image-cli batch resize "./images" --width 800 --height 600 --fit cover --output ./output/
image-cli batch rotate "./images" --degrees 90 --output ./output/
image-cli batch watermark "./images" --logo logo.png --opacity 0.6 --output ./output/
image-cli batch watermark "./images" --text "Sample" --font-size 24 --font "Arial" --color "#ffffff" --stroke-color black --stroke-width 2 --output ./output/
image-cli batch watermark "./images" --text "Sample" --font-size 24 --font-file "/path/to/font.ttf" --output ./output/
```

### ocr（OCR 文字识别）

使用 DeepSeek OCR API 从图片中提取文字内容。

```bash
image-cli ocr document.jpg
image-cli ocr document.png --mode markdown
image-cli ocr scan.jpg --mode text --output result.txt
```

配置：需在 `config.yaml` 中设置 `ocr.api_key` 或使用环境变量 `OCR_API_KEY`。

### generate（AI 图像生成）

使用智谱 AI CogView 模型根据文本描述生成图像。

```bash
image-cli generate "一只可爱的小猫咪"
image-cli generate "夕阳下的海景" --model cogview-3-flash --size 1440x720
image-cli generate "科幻城市" --quality hd --output ./output/futuristic.png
```

配置：需在 `config.yaml` 中设置 `image_generation.api_key` 或使用环境变量 `IMAGE_GENERATION_API_KEY`。

支持的模型：`cogview-3-flash`(免费)、`glm-image`、`cogview-4`。

### recognize（AI 图片识别）

使用智谱 AI GLM-4V 模型对图片进行视觉理解和分析。

```bash
image-cli recognize photo.jpg
image-cli recognize chart.png --prompt "分析这张图表的数据趋势"
image-cli recognize product.jpg --prompt "这是什么商品？请给出类别和特点"
```

配置：需在 `config.yaml` 中设置 `vision.api_key` 或使用环境变量 `IMAGE_VISION_API_KEY`。

支持的模型：`glm-4v-flash`(免费)、`glm-4.6v`。

## 全局参数

```bash
--config, -c     配置文件路径 (默认 ~/.config/image-cli/config.yaml)
--verbose, -v    详细输出
--quiet          静默模式
--recursive      目录递归处理 (默认 true)
--no-recursive   关闭递归
--conflict       冲突策略: skip|overwrite|rename (默认 skip)
--version, -V    显示版本
```
