# Trim 去边

`Trim` 把画布收缩到与角点颜色不同的像素包围盒。`tolerance` 是 0–100 的颜色距离。动图返回 `ErrNotSupported`。

## 功能概述

参考色是 `(0, 0)` 处的像素。与该颜色距离落在 `tolerance` 内的像素视为边框并剥掉。适合截图或扫描件上的均匀白边。

## 签名

```go
func (img *Image) Trim(tolerance int) *Image
```

## 参数说明

| 名称 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `tolerance` | `int` | 必填 | 相对角点像素的颜色距离 0–100。`0` 只去掉完全相同的像素 |

## 示例：包级紧致去边

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("scan.png").Trim(0)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("tight.png"); err != nil {
		log.Fatal(err)
	}
}
```

## 示例：白边留白后做容差去边

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("logo.png").
		Pad(400, 400, goimage.WithBackground("#ffffff")).
		Trim(8) // 允许白边里轻微的 JPEG/PNG 噪点
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("logo-trim.png"); err != nil {
		log.Fatal(err)
	}
}
```

## 注意细节

- GIF 动图不能 Trim（`ErrNotSupported`）。
- 整图颜色均匀时，结果可能接近 1×1；调用后检查 `Err()` 与 `Size()`。
