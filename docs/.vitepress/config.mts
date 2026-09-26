import { defineConfig } from 'vitepress'

function sidebarEN() {
  return [
    {
      text: 'Getting Started',
      items: [
        { text: 'Installation', link: '/getting-started/installation' },
        { text: 'Quickstart', link: '/getting-started/quickstart' },
        { text: 'Core Design', link: '/getting-started/design' },
      ],
    },
    {
      text: 'Concepts',
      items: [
        { text: 'Image', link: '/concepts/image' },
        { text: 'Colors', link: '/concepts/colors' },
        { text: 'Encoders & Decoders', link: '/concepts/encoding' },
        { text: 'Error Handling', link: '/concepts/errors' },
      ],
    },
    {
      text: 'Modifying Images',
      items: [
        {
          text: 'Geometry',
          collapsed: false,
          items: [
            { text: 'Resize', link: '/modifying/resize' },
            { text: 'Scale', link: '/modifying/scale' },
            { text: 'Cover / Fit', link: '/modifying/cover' },
            { text: 'Contain / Pad', link: '/modifying/contain' },
            { text: 'Crop', link: '/modifying/crop' },
            { text: 'Resize Canvas', link: '/modifying/canvas' },
            { text: 'Trim', link: '/modifying/trim' },
            { text: 'Rotate', link: '/modifying/rotate' },
            { text: 'Flip / Flop / Orient', link: '/modifying/flip' },
          ],
        },
        {
          text: 'Effects',
          collapsed: false,
          items: [
            { text: 'Greyscale', link: '/modifying/greyscale' },
            { text: 'Invert', link: '/modifying/invert' },
            { text: 'Brightness', link: '/modifying/brightness' },
            { text: 'Contrast', link: '/modifying/contrast' },
            { text: 'Gamma', link: '/modifying/gamma' },
            { text: 'Colorize', link: '/modifying/colorize' },
            { text: 'Pixelate', link: '/modifying/pixelate' },
            { text: 'Blur', link: '/modifying/blur' },
            { text: 'Sharpen', link: '/modifying/sharpen' },
            { text: 'Blend Transparency', link: '/modifying/blend' },
            { text: 'Reduce Colors', link: '/modifying/reduce-colors' },
          ],
        },
        {
          text: 'Drawing',
          collapsed: false,
          items: [
            { text: 'Pixel', link: '/modifying/pixel' },
            { text: 'Fill', link: '/modifying/fill' },
            { text: 'Shapes', link: '/modifying/shapes' },
            { text: 'Place', link: '/modifying/place' },
            { text: 'Text', link: '/modifying/text' },
          ],
        },
      ],
    },
    {
      text: 'Animation',
      items: [
        { text: 'Overview', link: '/animation/' },
        { text: 'Frames & Disposal', link: '/animation/frames' },
      ],
    },
    {
      text: 'Cookbook',
      items: [
        { text: 'Overview', link: '/cookbook/' },
        { text: 'Avatar Crop', link: '/cookbook/avatar' },
        { text: 'Dynamic Watermark', link: '/cookbook/watermark' },
        { text: 'Concurrent WebP Thumbnails', link: '/cookbook/webp-thumbnails' },
        { text: 'HTTP Handler', link: '/cookbook/http' },
      ],
    },
  ]
}

function sidebarZH() {
  return [
    {
      text: '入门指南',
      items: [
        { text: '安装', link: '/zh/getting-started/installation' },
        { text: '快速开始', link: '/zh/getting-started/quickstart' },
        { text: '核心设计', link: '/zh/getting-started/design' },
      ],
    },
    {
      text: '核心概念',
      items: [
        { text: 'Image 结构体', link: '/zh/concepts/image' },
        { text: '颜色系统', link: '/zh/concepts/colors' },
        { text: '编解码器', link: '/zh/concepts/encoding' },
        { text: '错误处理', link: '/zh/concepts/errors' },
      ],
    },
    {
      text: '图像处理 API',
      items: [
        {
          text: '几何变换',
          collapsed: false,
          items: [
            { text: 'Resize 缩放', link: '/zh/modifying/resize' },
            { text: 'Scale 等比缩放', link: '/zh/modifying/scale' },
            { text: 'Cover / Fit 覆盖', link: '/zh/modifying/cover' },
            { text: 'Contain / Pad 包含', link: '/zh/modifying/contain' },
            { text: 'Crop 裁剪', link: '/zh/modifying/crop' },
            { text: 'ResizeCanvas 画布', link: '/zh/modifying/canvas' },
            { text: 'Trim 去边', link: '/zh/modifying/trim' },
            { text: 'Rotate 旋转', link: '/zh/modifying/rotate' },
            { text: 'Flip / Flop / Orient', link: '/zh/modifying/flip' },
          ],
        },
        {
          text: '特效滤镜',
          collapsed: false,
          items: [
            { text: 'Greyscale 灰度', link: '/zh/modifying/greyscale' },
            { text: 'Invert 反色', link: '/zh/modifying/invert' },
            { text: 'Brightness 亮度', link: '/zh/modifying/brightness' },
            { text: 'Contrast 对比度', link: '/zh/modifying/contrast' },
            { text: 'Gamma 伽马', link: '/zh/modifying/gamma' },
            { text: 'Colorize 着色', link: '/zh/modifying/colorize' },
            { text: 'Pixelate 像素化', link: '/zh/modifying/pixelate' },
            { text: 'Blur 模糊', link: '/zh/modifying/blur' },
            { text: 'Sharpen 锐化', link: '/zh/modifying/sharpen' },
            { text: 'BlendTransparency', link: '/zh/modifying/blend' },
            { text: 'ReduceColors 减色', link: '/zh/modifying/reduce-colors' },
          ],
        },
        {
          text: '绘制与合成',
          collapsed: false,
          items: [
            { text: 'Pixel 像素', link: '/zh/modifying/pixel' },
            { text: 'Fill 填充', link: '/zh/modifying/fill' },
            { text: 'Shapes 图形', link: '/zh/modifying/shapes' },
            { text: 'Place 水印', link: '/zh/modifying/place' },
            { text: 'Text 文字', link: '/zh/modifying/text' },
          ],
        },
      ],
    },
    {
      text: '动图专题',
      items: [
        { text: '概述', link: '/zh/animation/' },
        { text: '帧控制与 Disposal', link: '/zh/animation/frames' },
      ],
    },
    {
      text: '实战场景',
      items: [
        { text: '概述', link: '/zh/cookbook/' },
        { text: '头像裁剪', link: '/zh/cookbook/avatar' },
        { text: '动态水印', link: '/zh/cookbook/watermark' },
        { text: '高并发 WebP 缩略图', link: '/zh/cookbook/webp-thumbnails' },
        { text: 'HTTP Handler', link: '/zh/cookbook/http' },
      ],
    },
  ]
}

export default defineConfig({
  title: 'go-image',
  description: 'Fluent image processing library for Go',
  base: '/go-image/',
  lastUpdated: true,
  ignoreDeadLinks: false,
  locales: {
    root: {
      label: 'English',
      lang: 'en-US',
      themeConfig: {
        nav: [
          { text: 'Getting Started', link: '/getting-started/installation' },
          { text: 'Concepts', link: '/concepts/image' },
          { text: 'Modifying Images', link: '/modifying/resize' },
          { text: 'Animation', link: '/animation/' },
          { text: 'Cookbook', link: '/cookbook/' },
        ],
        sidebar: sidebarEN(),
        outline: [2, 3],
        socialLinks: [
          { icon: 'github', link: 'https://github.com/yunkeweb/go-image' },
        ],
        search: { provider: 'local' },
        footer: {
          message: 'MIT License.',
          copyright: 'Copyright © 2026 yunkeweb',
        },
      },
    },
    zh: {
      label: '简体中文',
      lang: 'zh-CN',
      link: '/zh/',
      themeConfig: {
        nav: [
          { text: '入门', link: '/zh/getting-started/installation' },
          { text: '概念', link: '/zh/concepts/image' },
          { text: '图像处理', link: '/zh/modifying/resize' },
          { text: '动图', link: '/zh/animation/' },
          { text: '实战', link: '/zh/cookbook/' },
        ],
        sidebar: sidebarZH(),
        outline: [2, 3],
        socialLinks: [
          { icon: 'github', link: 'https://github.com/yunkeweb/go-image' },
        ],
        search: { provider: 'local' },
        footer: {
          message: 'MIT 许可证。',
          copyright: 'Copyright © 2026 yunkeweb',
        },
      },
    },
  },
})
