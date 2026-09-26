# 从 PHP 迁移

Intervention Image v3（PHP）到 go-image 的对照。

## 驱动

PHP 的 `ImageManager::gd()` / `ImageManager::imagick()` 变为单一的 `Manager`，底层是 `image.Image`。调用 `New()` 或包级辅助函数。

## 异常

抛出的 `RuntimeException` 等变为 `img.Err()`。编码/保存同时返回 `error`。

## 空尺寸

PHP `?int $width = null` 对应 Go 的 `0`（未指定）。例如 `Scale(300, 0)` 保持宽高比。

## 闭包

PHP 绘制/字体闭包变为 `func(*Drawable)` 和 `func(*Font)`（或直接传 `*Font`）。

## 对齐

九点枢轴名称与 PHP 一致：`center`、`top`、`top-left`、`top-right`、`left`、`right`、`bottom`、`bottom-left`、`bottom-right`。

## 旋转

`Rotate` 为逆时针，对齐 PHP GD `imagerotate`。

## Cover 与 Fit

Intervention Image v3 的 `cover()` 对应 `Cover`。v2 的 `fit()` 映射为 `Cover`。

## 不支持的格式

仅 Imagick 支持的 AVIF、HEIC、JPEG 2000 会设置延迟 `ErrNotSupported`。

## 示例

PHP：

```php
$manager = ImageManager::gd();
$image = $manager->read('photo.jpg')
    ->cover(400, 300, 'center')
    ->greyscale();
$image->toJpeg(85)->save('out.jpg');
```

Go：

```go
img := goimage.Read("photo.jpg").
    Cover(400, 300, "center").
    Greyscale()
if err := img.Err(); err != nil {
    log.Fatal(err)
}
if err := img.ToJPEG(goimage.EncodeOptions{Quality: 85}).Save("out.jpg"); err != nil {
    log.Fatal(err)
}
```

仓库中的 [MIGRATION_CHECKLIST.md](https://github.com/yunkeweb/go-image/blob/main/MIGRATION_CHECKLIST.md) 列出了完整对照。
