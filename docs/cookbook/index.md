# Cookbook

Production-shaped recipes. Each page is a full program: package-level calls plus functional options, delayed `Err()`, and encode to disk or HTTP.

| Recipe | What it shows |
|--------|----------------|
| [Avatar Crop](/cookbook/avatar) | Square `Cover` with a top anchor, WebP + JPEG fallback |
| [Dynamic Watermark](/cookbook/watermark) | Logo `Place` plus timestamp `Text` |
| [Concurrent WebP Thumbnails](/cookbook/webp-thumbnails) | Worker pool, `Clone`, `sync.Pool`-friendly pipelines |
| [HTTP Handler](/cookbook/http) | `Decode` an upload, `WriteTo` the response |

Related API: [Cover](/modifying/cover), [Place](/modifying/place), [Text](/modifying/text), [Encoders](/concepts/encoding).
