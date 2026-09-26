---
layout: home
hero:
  name: go-image
  text: Fluent image processing in Go
  tagline: Idiomatic port of Intervention Image. Standard library plus golang.org/x/image only.
  actions:
    - theme: brand
      text: Get Started
      link: /guide/getting-started
    - theme: alt
      text: API Reference
      link: /api/manager
features:
  - title: Fluent chains
    details: Mutating methods return *Image. The first failure is stored on the value and retrieved with Err().
  - title: Geometry and effects
    details: Resize, Cover, Contain, Crop, Rotate, Blur, Sharpen, Greyscale, and the rest of the Intervention Image v3 surface.
  - title: Encode without CGO
    details: JPEG, PNG, GIF (including animation), lossless WebP, BMP, and TIFF. No GD or Imagick.
---
