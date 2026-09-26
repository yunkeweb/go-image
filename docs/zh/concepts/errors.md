# 错误处理

go-image 把第一次失败记在 `Image` 上，链式调用保持流畅。该模型对 goroutine 安全：每个 `*Image` 自持错误与像素缓冲。哨兵错误用 `errors.Is` 判断。

## 功能概述

修改方法返回同一个 `*Image`。若 `Err()` 已有值，后续修饰变成空操作。编码辅助返回 `EncodedImage`；检查 `enc.Err()` 或 `Save` / `WriteTo` 的返回值。几何方法收到 `0` 或负数尺寸时写入 `ErrInvalidDimensions`（同时包裹 `ErrGeometry`）。

## 签名

```go
func (img *Image) Err() error
func (e EncodedImage) Err() error
```

## 哨兵错误

| 变量 | 典型原因 |
|------|----------|
| `ErrRuntime` | 空图像或异常状态 |
| `ErrDecoder` | 无法读取或空输入 |
| `ErrEncoder` | 编码失败或 writer 为 nil |
| `ErrGeometry` | 非法尺寸 |
| `ErrInvalidDimensions` | 宽或高为 `0` 或负数 |
| `ErrColor` | 无法解析的颜色 |
| `ErrInput` | 参数错误（动画下标、空水印） |
| `ErrNotSupported` | AVIF / HEIC / JPEG 2000，或对动图 Trim |
| `ErrNotWritable` | 文件系统写入失败 |
| `ErrAnimation` | 空动画构建器或空帧 |
| `ErrFont` | 字体文件加载失败 |
| `ErrDriver` | 驱动级失败 |

## 示例：Open 上的延迟解码错误

```go
package main

import (
	"errors"
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("missing.jpg").Cover(200, 200)
	if err := img.Err(); err != nil {
		if errors.Is(err, goimage.ErrDecoder) {
			log.Fatal("could not decode:", err)
		}
		log.Fatal(err)
	}
}
```

## 示例：非法 Resize 触发 ErrInvalidDimensions

```go
package main

import (
	"errors"
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.New(64, 64).Fill("#334155").Resize(0) // 0 从不表示“自动”
	if err := img.Err(); err != nil {
		if errors.Is(err, goimage.ErrInvalidDimensions) {
			log.Fatal("width/height must be >= 1:", err)
		}
		log.Fatal(err)
	}
}
```

## 注意细节

- 链结束后检查一次 `Err()` 即可。中间检查可选。
- `EncodedImage.WriteTo` 在已有编码错误时直接返回该错误，不写字节。
- `errors.Is` 有效，因为根包哨兵与 `internal/errs` 是同一组值。
