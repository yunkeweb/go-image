---
layout: home
hero:
  name: go-image
  text: 面向 Go 的流式图像处理
  tagline: 包级 API、Functional Options、仅标准库编解码。无 CGO。
  actions:
    - theme: brand
      text: 快速开始
      link: /zh/guide/getting-started
    - theme: alt
      text: API 参考
      link: /zh/api/manager
features:
  - title: 地道的 Go
    details: 包级 Open、Decode、New。每个入口都支持 Functional Options。用强类型替代 any。
  - title: 天生适合并发
    details: 每个 *Image 拥有自己的像素与延迟错误。sync.Pool 回收不超过 16 MiB 的 NRGBA 缓冲。
  - title: 无 CGO 编码
    details: JPEG、PNG、GIF（含动画）、无损 WebP、BMP、TIFF。
---
