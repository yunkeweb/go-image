import { defineConfig } from 'vitepress'

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
          { text: 'Guide', link: '/guide/getting-started' },
          { text: 'API', link: '/api/manager' },
          { text: 'Recipes', link: '/recipes/' },
        ],
        sidebar: {
          '/guide/': [
            {
              text: 'Guide',
              items: [
                { text: 'Getting started', link: '/guide/getting-started' },
                { text: 'Errors', link: '/guide/errors' },
                { text: 'Formats', link: '/guide/formats' },
                { text: 'Design', link: '/guide/design' },
              ],
            },
          ],
          '/api/': [
            {
              text: 'API',
              items: [
                { text: 'Package API', link: '/api/manager' },
                { text: 'Image', link: '/api/image' },
                { text: 'Geometry', link: '/api/geometry' },
                { text: 'Effects', link: '/api/effects' },
                { text: 'Drawing', link: '/api/drawing' },
                { text: 'Encoding', link: '/api/encoding' },
                { text: 'Animation', link: '/api/animation' },
                { text: 'Color', link: '/api/color' },
              ],
            },
          ],
          '/recipes/': [
            {
              text: 'Recipes',
              items: [{ text: 'Examples', link: '/recipes/' }],
            },
          ],
        },
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
          { text: '指南', link: '/zh/guide/getting-started' },
          { text: 'API', link: '/zh/api/manager' },
          { text: '示例', link: '/zh/recipes/' },
        ],
        sidebar: {
          '/zh/guide/': [
            {
              text: '指南',
              items: [
                { text: '快速开始', link: '/zh/guide/getting-started' },
                { text: '错误处理', link: '/zh/guide/errors' },
                { text: '格式', link: '/zh/guide/formats' },
                { text: '设计理念', link: '/zh/guide/design' },
              ],
            },
          ],
          '/zh/api/': [
            {
              text: 'API',
              items: [
                { text: '包级 API', link: '/zh/api/manager' },
                { text: 'Image', link: '/zh/api/image' },
                { text: '几何变换', link: '/zh/api/geometry' },
                { text: '效果', link: '/zh/api/effects' },
                { text: '绘制', link: '/zh/api/drawing' },
                { text: '编码', link: '/zh/api/encoding' },
                { text: '动画', link: '/zh/api/animation' },
                { text: '颜色', link: '/zh/api/color' },
              ],
            },
          ],
          '/zh/recipes/': [
            {
              text: '示例',
              items: [{ text: '代码示例', link: '/zh/recipes/' }],
            },
          ],
        },
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
