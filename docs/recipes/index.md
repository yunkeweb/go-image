# Recipes

Runnable snippets. Check `Err()` after a chain, then check encode/save errors.

## Thumbnail that covers a box

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Cover(400, 300, "center")
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("thumb.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## Watermark

```go
img := goimage.Open("photo.jpg").
	Place(goimage.Open("logo.png"), "bottom-right", 16, 16, 70)
if err := img.Err(); err != nil {
	log.Fatal(err)
}
if err := img.ToPNG().Save("watermarked.png"); err != nil {
	log.Fatal(err)
}
```

## Data URI from bytes

```go
img := goimage.New(1, 1).Fill("#ff00ff")
enc := img.ToPNG()
if err := enc.Err(); err != nil {
	log.Fatal(err)
}
fmt.Println(enc.ToDataURI())
```

## Decode bytes, write WebP

```go
img := goimage.DecodeBytes(pngBytes).
	Scale(128, 0)
if err := img.Err(); err != nil {
	log.Fatal(err)
}
if err := img.ToWebP().Save("out.webp"); err != nil {
	log.Fatal(err)
}
```

## Animated GIF

```go
anim := goimage.Animate(func(a *goimage.Animation) {
	a.SetLoops(0)
	for i := 0; i < 8; i++ {
		frame := goimage.New(64, 64).Fill("#111827")
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

## Rotate and flatten for JPEG

```go
img := goimage.Open("photo.jpg").
	Rotate(45, "#ffffff").
	SetBlendingColor("#ffffff")
if err := img.Err(); err != nil {
	log.Fatal(err)
}
if err := img.ToJPEG().Save("rotated.jpg"); err != nil {
	log.Fatal(err)
}
```
