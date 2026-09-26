---
layout: home
hero:
  name: go-image
  text: Fluent image processing in Go
  tagline: Package-level APIs, functional options, and stdlib-only codecs. No CGO.
  actions:
    - theme: brand
      text: Get Started
      link: /guide/getting-started
    - theme: alt
      text: API Reference
      link: /api/manager
features:
  - title: Idiomatic Go
    details: Open, Decode, and New at package level. Functional options on every entry point. Strong types instead of any.
  - title: Concurrent by default
    details: Each *Image owns its pixels and delayed error. sync.Pool recycles NRGBA buffers up to 16 MiB.
  - title: Encode without CGO
    details: JPEG, PNG, GIF (including animation), lossless WebP, BMP, and TIFF.
---
