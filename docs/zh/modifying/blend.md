# Blend Transparency 压扁透明

`BlendTransparency` 把每个像素合成到不透明背景上。传入 `nil` 时使用图像的 `BlendingColor`（默认白色）。JPEG 编码本身也会压扁；若希望 PNG 里也能看到压扁结果，请先调用本方法。

## 功能概述

数学是标准 “over” 合成：`out = src.RGB * src.A + bg.RGB * (1 - src.A)`，随后 alpha 变为 255。适合在忽略透明通道的算法之前，或发给无法显示透明的客户端之前使用。

## 签名

```go
func (img *Image) BlendTransparency(col any) *Image
```

## 参数说明

| 名称 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `col` | `any` | `nil` 时用图像的 `BlendingColor` | 十六进制、名称、`Color` 或 `nil` |

非法颜色写入延迟 `ErrColor`。

## 示例：包级白底压扁

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("logo.png").BlendTransparency("#ffffff")
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(90).Save("logo.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## 示例：从 Open 选项读取混合色

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("logo.png", goimage.WithBlendingColor("#0f172a")).
		BlendTransparency(nil) // 使用配置中的颜色
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("on-slate.png"); err != nil {
		log.Fatal(err)
	}
}
```

## 注意细节

- 调用后 `PickColor` 处处报告 alpha 255。
- 部分覆盖的 GIF 帧变为不透明；编码时可能选择 `DisposalNone`。
