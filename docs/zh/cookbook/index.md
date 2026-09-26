# 实战场景

面向生产的配方。每页都是完整程序：包级调用 + Functional Options、延迟 `Err()`，以及写到磁盘或 HTTP。

| 配方 | 展示内容 |
|------|----------|
| [头像裁剪](/zh/cookbook/avatar) | 顶部锚点的方形 `Cover`，WebP + JPEG 回退 |
| [动态水印](/zh/cookbook/watermark) | Logo `Place` 加时间戳 `Text` |
| [高并发 WebP 缩略图](/zh/cookbook/webp-thumbnails) | worker pool、`Clone`、对 `sync.Pool` 友好的流水线 |
| [HTTP Handler](/zh/cookbook/http) | `Decode` 上传、`WriteTo` 响应 |

相关 API：[Cover](/zh/modifying/cover)、[Place](/zh/modifying/place)、[Text](/zh/modifying/text)、[编解码器](/zh/concepts/encoding)。
