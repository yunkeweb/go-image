# Gamma 伽马

`Gamma` 对每个 RGB 通道施加幂曲线。`gamma` 必须 `> 0`。小于 1 压暗中间调；大于 1 提亮中间调。

## 功能概述

每个通道映射为 `(c/255)^(1/gamma) * 255`。Alpha 不变。非法 `gamma`（`<= 0`）由修饰器写入延迟错误。

## 签名

```go
func (img *Image) Gamma(gamma float64) *Image
```

## 参数说明

| 名称 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `gamma` | `float64` | 必填 | 必须 `> 0`。`1` 保持原图 |

## 示例：包级提亮中间调

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Gamma(1.2)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("gamma.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## 示例：裁切后压暗

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg", goimage.WithAutoOrientation(true)).
		Cover(800, 450).
		Gamma(0.85)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("moody.png"); err != nil {
		log.Fatal(err)
	}
}
```

## 注意细节

- Gamma 是全局曲线，不会像色调映射那样保护高光。
- 摄影类处理建议在 `1.0` 附近小幅调整。
