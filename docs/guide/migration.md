# PHP migration

Mapping from Intervention Image v3 (PHP) to go-image.

## Drivers

PHP `ImageManager::gd()` / `ImageManager::imagick()` become a single `Manager` over `image.Image`. Call `New()` or the package-level helpers.

## Exceptions

Thrown `RuntimeException` and friends become `img.Err()`. Encode/save also return `error`.

## Null sizes

PHP `?int $width = null` is Go `0` (unspecified). Example: `Scale(300, 0)` keeps aspect ratio.

## Closures

PHP draw/font closures become `func(*Drawable)` and `func(*Font)` (or a `*Font` value).

## Positions

The 9-point pivot names match PHP: `center`, `top`, `top-left`, `top-right`, `left`, `right`, `bottom`, `bottom-left`, `bottom-right`.

## Rotate

`Rotate` is counter-clockwise, matching PHP GD `imagerotate`.

## Cover vs Fit

Intervention Image v3 `cover()` is implemented as `Cover`. v2 `fit()` maps to `Cover`.

## Unsupported formats

Imagick-only AVIF, HEIC, and JPEG 2000 set delayed `ErrNotSupported`.

## Example

PHP:

```php
$manager = ImageManager::gd();
$image = $manager->read('photo.jpg')
    ->cover(400, 300, 'center')
    ->greyscale();
$image->toJpeg(85)->save('out.jpg');
```

Go:

```go
img := goimage.Read("photo.jpg").
    Cover(400, 300, "center").
    Greyscale()
if err := img.Err(); err != nil {
    log.Fatal(err)
}
if err := img.ToJPEG(85).Save("out.jpg"); err != nil {
    log.Fatal(err)
}
```

See also [MIGRATION_CHECKLIST.md](https://github.com/yunkeweb/go-image/blob/main/MIGRATION_CHECKLIST.md) in the repository.
