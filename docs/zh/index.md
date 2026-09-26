---
layout: home
hero:
  name: go-image
  text: 现代化 Go 图像处理库
  tagline: 零 CGO。包级 API、Functional Options、sync.Pool 复用 NRGBA 缓冲。支持 JPEG、PNG、GIF、WebP、BMP、TIFF。
  actions:
    - theme: brand
      text: 快速开始
      link: /zh/getting-started/installation
    - theme: alt
      text: 实战场景
      link: /zh/cookbook/
features:
  - title: 地道的 Go
    details: Open、Decode、New 都是包级函数。Resize 用变长 height。Cover / Crop 用 WithAnchor、WithBackground。调用方只导入一个模块路径。
  - title: 无锁流水线
    details: 每个 Image 自持像素与延迟错误。多文件用独立 goroutine 处理，同一张图先 Clone 再并发。热路径没有互斥锁。
  - title: 自有缓冲与对象池
    details: 解码经 draw.Draw 拷贝到 NRGBA。丢弃的缓冲在 16 MiB 上限内归还 sync.Pool，大图不会长期占用巨大切片。
  - title: 无 CGO 编码
    details: JPEG、PNG、动画 GIF、无损 WebP、BMP、TIFF。EncodedImage.WriteTo 可直接写入任意 io.Writer，包括 HTTP Handler。
---
