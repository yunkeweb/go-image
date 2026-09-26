# HTTP Handler

`Decode` 读取任意 `io.Reader`（上传 body）。`EncodedImage.WriteTo` 实现 `io.WriterTo`，把字节流式写入 `http.ResponseWriter`。用 `MimeType()` 设置 `Content-Type`。

## 流水线

1. 限制请求体（`http.MaxBytesReader`）。
2. `Decode(r, WithAutoOrientation(true), WithDecodeAnimation(false))`。
3. 按需要的变体 `Cover` / `Sharpen`。
4. `ToJPEG(85)` 或 `ToWebP()`。
5. `enc.Err()` 时返回 400/500；否则 `WriteTo(w)`。

## 示例：包级缩略图接口

```go
package main

import (
	"log"
	"net/http"

	goimage "github.com/yunkeweb/go-image"
)

func thumbnail(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 12<<20)
	img := goimage.Decode(r.Body).Cover(400, 300).Sharpen(8)
	if err := img.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	enc := img.ToJPEG(85)
	if enc.Err() != nil {
		http.Error(w, enc.Err().Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", enc.MimeType())
	if _, err := enc.WriteTo(w); err != nil {
		log.Println("write", err)
	}
}

func main() {
	http.HandleFunc("/thumb", thumbnail)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## 示例：Functional Options + WebP 查询参数

```go
package main

import (
	"net/http"

	goimage "github.com/yunkeweb/go-image"
)

func variant(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 12<<20)
	img := goimage.Decode(r.Body,
		goimage.WithAutoOrientation(true),
		goimage.WithDecodeAnimation(false),
	).Cover(800, 600, goimage.WithAnchor("center"))
	if err := img.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var enc goimage.EncodedImage
	if r.URL.Query().Get("fmt") == "webp" {
		enc = img.ToWebP()
	} else {
		enc = img.ToJPEG(85)
	}
	if enc.Err() != nil {
		http.Error(w, enc.Err().Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", enc.MimeType())
	_, _ = enc.WriteTo(w)
}
```

## 注意细节

- 若 `enc.Err() != nil`，`WriteTo` 直接返回该错误，不写字节。
- 优先 `Decode`，不要 `DecodeBytes(io.ReadAll(...))`，让库内部读一次 body。
- JSON 内嵌图片可用 `enc.ToDataURI()`。
