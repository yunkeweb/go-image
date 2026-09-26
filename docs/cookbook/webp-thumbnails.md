# Concurrent WebP Thumbnails

A catalog importer should decode, cover, and encode many files in parallel. Each goroutine owns its `Image`. There is no mutex on the pixel path. Discarded NRGBA buffers return to `sync.Pool` (cap 16 MiB).

## Pipeline

1. List paths (or object-storage keys).
2. Worker pool: `Open` → `Cover(400, 300)` → `Sharpen(8)` → `ToWebP().Save`.
3. Check `Err()` per image; collect failures without stopping the pool.
4. If one source must feed two sizes, `Clone()` before the second `Cover`.

WebP encode in this library is lossless VP8L — great for UI assets and intermediate caches. For photographic weight, add a JPEG sibling with `ToJPEG(80)`.

## Example: package-level worker pool

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

## Example: one decode, two sizes via Clone + WithConfig

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

## Notes

- Do not call `Cover` on `base` from two goroutines. `Clone` first.
- `WithDecodeAnimation(false)` skips extra GIF frames when the catalog is stills-only.
- Pool reuse is automatic; you never import `internal/pool`.
