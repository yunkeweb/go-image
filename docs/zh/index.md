---
layout: home
hero:
  name: go-image
  text: 面向 Go 的流式图像处理
  tagline: Intervention Image 的惯用 Go 移植。仅使用标准库与 golang.org/x/image。
  actions:
    - theme: brand
      text: 快速开始
      link: /zh/guide/getting-started
    - theme: alt
      text: API 参考
      link: /zh/api/manager
features:
  - title: 链式调用
    details: 修改器返回 *Image。第一次失败保存在对象上，通过 Err() 读取。
  - title: 几何与效果
    details: Resize、Cover、Contain、Crop、Rotate、Blur、Sharpen、Greyscale 等 Intervention Image v3 公共 API。
  - title: 无 CGO 编码
    details: JPEG、PNG、GIF（含动画）、无损 WebP、BMP、TIFF。无需 GD 或 Imagick。
---
