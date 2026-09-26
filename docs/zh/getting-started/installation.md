# 安装

go-image 是纯 Go 模块。使用标准工具链即可编译，不需要 C 编译器，只依赖 Go 标准库与官方 `golang.org/x/image`。

## 环境要求

| 项目 | 取值 |
|------|------|
| Go | 1.22 及以上 |
| CGO | 关闭即可。库不会调用 libvips、ImageMagick 或 libwebp 的 C 绑定 |
| 额外模块 | `golang.org/x/image`（JPEG 扩展、WebP VP8L、BMP、TIFF、字体、`draw` 重采样） |

## 安装命令

```bash
go get github.com/yunkeweb/go-image
```

生产环境建议钉死版本：

```bash
go get github.com/yunkeweb/go-image@v0.1.7
```

## 导入

调用方只导入一个模块路径。公开类型（`Image`、`Color`、`Config`、`EncodedImage`）和全部修饰方法都在包 `goimage` 中：

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg")
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	log.Println(img.Width(), img.Height())
}
```

`modifier`、`encoder`、`internal/pool` 属于实现细节。应用代码始终使用根模块导入。

## 核对工具链

```bash
go version          # go1.22 或更新
go env CGO_ENABLED  # 0 即可；本模块不需要 CGO
```

## 下一步

- [快速开始](/zh/getting-started/quickstart) — 打开、覆盖裁剪、锐化、编码
- [核心设计](/zh/getting-started/design) — 包级 API、Functional Options、`sync.Pool`
