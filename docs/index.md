---
layout: home
hero:
  name: go-image
  text: Fluent image processing in Go
  tagline: Zero CGO. Package-level APIs, functional options, and pooled NRGBA buffers. JPEG, PNG, GIF, WebP, BMP, and TIFF.
  actions:
    - theme: brand
      text: Get Started
      link: /getting-started/installation
    - theme: alt
      text: Cookbook
      link: /cookbook/
features:
  - title: Idiomatic Go
    details: Open, Decode, and New live at package level. Variadic height on Resize. WithAnchor and WithBackground on Cover and Crop. One import path.
  - title: Lock-free pipelines
    details: Each Image owns its pixels and delayed error. Process files in parallel goroutines, or Clone before sharing one source. No mutex on the hot path.
  - title: Owned, pooled buffers
    details: Decode copies into NRGBA with draw.Draw. Discarded buffers return to sync.Pool up to 16 MiB so large frames do not pin huge slices.
  - title: Encode without CGO
    details: JPEG, PNG, animated GIF, lossless WebP, BMP, and TIFF. EncodedImage.WriteTo streams bytes to any io.Writer, including HTTP handlers.
---
