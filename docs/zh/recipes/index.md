# 示例

可运行片段。链式调用后检查 `Err()`，再检查编码/保存错误。

## 铺满目标框的缩略图

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Read("photo.jpg").Cover(400, 300, "center")
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(goimage.EncodeOptions{Quality: 85}).Save("thumb.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## 水印

```go
img := goimage.Read("photo.jpg").
	Place("logo.png", "bottom-right", 16, 16, 70)
if err := img.Err(); err != nil {
	log.Fatal(err)
}
if err := img.ToPNG().Save("watermarked.png"); err != nil {
	log.Fatal(err)
}
```

## 从字节生成 Data URI

```go
img := goimage.Create(1, 1).Fill("#ff00ff")
enc := img.ToPNG()
if err := enc.Err(); err != nil {
	log.Fatal(err)
}
fmt.Println(enc.ToDataURI())
```

## 读取 Base64，写出 WebP

```go
img := goimage.Read(base64.StdEncoding.EncodeToString(pngBytes)).
	Scale(128, 0)
if err := img.Err(); err != nil {
	log.Fatal(err)
}
if err := img.ToWebP().Save("out.webp"); err != nil {
	log.Fatal(err)
}
```

## 动画 GIF

```go
anim := goimage.Animate(func(a *goimage.Animation) {
	a.SetLoops(0)
	for i := 0; i < 8; i++ {
		frame := goimage.Create(64, 64).Fill("#111827")
		frame.DrawCircle(8+i*6, 32, func(d *goimage.Drawable) {
			d.SetRadius(6).SetBackground("#38bdf8")
		})
		a.AddImage(frame, 0.08)
	}
})
if err := anim.Err(); err != nil {
	log.Fatal(err)
}
if err := anim.ToGIF().Save("dot.gif"); err != nil {
	log.Fatal(err)
}
```

## 旋转并压平后输出 JPEG

```go
img := goimage.Read("photo.jpg").
	Rotate(45, "#ffffff").
	SetBlendingColor("#ffffff")
if err := img.Err(); err != nil {
	log.Fatal(err)
}
if err := img.ToJPEG().Save("rotated.jpg"); err != nil {
	log.Fatal(err)
}
```
