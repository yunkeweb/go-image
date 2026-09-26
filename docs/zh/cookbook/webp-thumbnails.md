# 高并发 WebP 缩略图

目录导入应当并行解码、覆盖裁剪、编码。每个 goroutine 拥有自己的 `Image`。像素热路径没有互斥锁。丢弃的 NRGBA 缓冲归还 `sync.Pool`（上限 16 MiB）。

## 流水线

1. 列出路径（或对象存储 key）。
2. Worker pool：`Open` → `Cover(400, 300)` → `Sharpen(8)` → `ToWebP().Save`。
3. 每张图检查 `Err()`；收集失败，不要停掉整个池。
4. 同一来源要出两种尺寸时，第二次 `Cover` 之前先 `Clone()`。

本库 WebP 编码是无损 VP8L — 适合 UI 资源与中间缓存。照片体积请再写一份 `ToJPEG(80)`。

## 示例：包级 worker pool

```go
package main

import (
	"log"
	"path/filepath"
	"sync"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	paths := []string{"a.jpg", "b.jpg", "c.jpg"}
	jobs := make(chan string)
	var wg sync.WaitGroup

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				img := goimage.Open(path).Cover(400, 300).Sharpen(8)
				if err := img.Err(); err != nil {
					log.Println(path, err)
					continue
				}
				out := "thumbs/" + filepath.Base(path) + ".webp"
				if err := img.ToWebP().Save(out); err != nil {
					log.Println(out, err)
				}
			}
		}()
	}
	for _, p := range paths {
		jobs <- p
	}
	close(jobs)
	wg.Wait()
}
```

## 示例：解码一次，Clone + WithConfig 出两种尺寸

```go
package main

import (
	"log"
	"sync"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	cfg := goimage.DefaultConfig
	cfg.AutoOrientation = true
	cfg.DecodeAnimation = false
	base := goimage.Open("hero.jpg", goimage.WithConfig(cfg))
	if err := base.Err(); err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_ = base.Clone().Cover(400, 300, goimage.WithAnchor("center")).
			ToWebP().Save("hero-sm.webp")
	}()
	go func() {
		defer wg.Done()
		_ = base.Clone().Cover(1200, 630, goimage.WithAnchor("center")).
			ToWebP().Save("hero-og.webp")
	}()
	wg.Wait()
}
```

## 注意细节

- 不要从两个 goroutine 对 `base` 调用 `Cover`。先 `Clone`。
- 目录全是静图时，`WithDecodeAnimation(false)` 跳过多余 GIF 帧。
- 对象池自动复用；你不必导入 `internal/pool`。
