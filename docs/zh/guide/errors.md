# 错误处理

PHP Intervention Image 抛出异常。go-image 把第一次失败保存在 `Image` 上，以便继续链式调用。

## 签名

```go
func (img *Image) Err() error
```

## 行为

1. 修改器返回同一个 `*Image`。
2. 若 `Err()` 已有值，后续修改器为空操作。
3. 编码方法返回 `EncodedImage`。使用 `enc.Err()` 或 `Save` 的返回值。
4. 可用 `errors.Is` 判断哨兵错误。

```go
img := goimage.Read("missing.jpg").Cover(200, 200, "center")
if err := img.Err(); err != nil {
    if errors.Is(err, goimage.ErrDecoder) {
        log.Fatal("could not decode:", err)
    }
    log.Fatal(err)
}
```

## 哨兵错误

| 变量 | 典型原因 |
|------|----------|
| `ErrRuntime` | 空图像或意外状态 |
| `ErrDecoder` | 无法读取输入 |
| `ErrEncoder` | 编码或保存路径失败 |
| `ErrGeometry` | 非法尺寸 |
| `ErrInvalidDimensions` | 宽度与高度同时为 `0` |
| `ErrColor` | 无法解析的颜色 |
| `ErrInput` | 参数错误（动画索引等） |
| `ErrNotSupported` | AVIF / HEIC / JPEG 2000，或未知格式 |
| `ErrNotWritable` | 文件系统写入失败 |
| `ErrAnimation` | 动画构建器没有帧 |
| `ErrFont` | 字体文件加载失败 |
| `ErrDriver` | 驱动层失败 |

## EncodedImage

```go
enc := img.ToWebP()
if err := enc.Err(); err != nil {
    log.Fatal(err)
}
_ = enc.Bytes()
_ = enc.MimeType()
_ = enc.ToDataURI()
```
