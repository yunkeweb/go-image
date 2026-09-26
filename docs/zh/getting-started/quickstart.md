# 快速开始

本页覆盖最高频的三种调用：`Open` 打开文件、`New` 空白画布、`ToJPEG` / `ToPNG` 编码。每个修改方法都返回同一个 `*Image`，链式调用很短。在链末尾检查一次 `Err()` 即可。

## 打开、覆盖裁剪并编码 JPEG

`Cover` 填满 `400×300` 的盒子，溢出部分按中心裁掉。`ToJPEG(85)` 指定质量 85；省略参数时默认为 80。

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").
		Cover(400, 300). // 填满 400×300，裁掉溢出，Catmull-Rom 重采样
		Sharpen(8)       // 反锐化；8 适合轻量网页缩略图
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}

	enc := img.ToJPEG(85) // 质量 85；ToJPEG() 则为 80
	if err := enc.Save("thumb.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## 用 Functional Options 创建画布

`New` 分配透明 NRGBA 画布。`WithBlendingColor` 会记在图像上，后续编码压扁透明通道（JPEG）时使用。绘制通过 `func(*Drawable)` 回调配置。

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.New(640, 480, goimage.WithBlendingColor("#0f172a")).
		Fill("#1e293b").
		DrawCircle(320, 240, func(d *goimage.Drawable) {
			d.SetRadius(80).SetBackground("#38bdf8")
		})
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("circle.png"); err != nil {
		log.Fatal(err)
	}
}
```

## 入口函数

| 函数 | 输入 | 典型用途 |
|------|------|----------|
| `Open(path, opts...)` | 文件系统路径 | 磁盘文件 |
| `Decode(r, opts...)` | `io.Reader` | HTTP 上传、`os.File` |
| `DecodeBytes(data, opts...)` | 编码后的 `[]byte` | 缓存、对象存储 |
| `DecodeDataURI(uri, opts...)` | `data:image/...;base64,...` | HTML 内嵌 |
| `FromImage(src, opts...)` | `image.Image` | 标准库解码结果 |
| `New(w, h, opts...)` | 空白画布 | 程序生成图形 |
| `Animate(init, opts...)` | 帧构建回调 | 组装 GIF |

选项只作用于**这一次**调用。实现会复制默认 `Config`（`AutoOrientation: true`、`DecodeAnimation: true`、`BlendingColor: "ffffff"`、`Strip: false`），再叠加上你传入的函数。

```go
img := goimage.Open("input.png",
	goimage.WithAutoOrientation(true),
	goimage.WithDecodeAnimation(true),
	goimage.WithBlendingColor("ffffff"),
	goimage.WithStrip(false),
)
```

多文件共用一份配置时用 `NewManager`：

```go
mgr := goimage.NewManager(
	goimage.WithAutoOrientation(true),
	goimage.WithDecodeAnimation(false), // 只要静态首帧
)
thumb := mgr.Open("photo.jpg").Cover(200, 200)
```

## 一行几何变换

`Resize(400)` 保持宽高比。`Cover` 接受 `WithAnchor`。宽或高为 0 / 负数会写入延迟错误 `ErrInvalidDimensions`。

```go
wide := goimage.Open("photo.jpg").Resize(800) // 高度按比例计算
box := goimage.Open("photo.jpg").Cover(400, 300, goimage.WithAnchor("top"))
```

## 下一步

- [核心设计](/zh/getting-started/design) — API 为什么长成这样
- [Image 结构体](/zh/concepts/image) — 查询方法、`Clone`、自有缓冲
- [实战场景](/zh/cookbook/) — 头像、水印、HTTP
