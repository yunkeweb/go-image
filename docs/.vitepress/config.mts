import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'go-image',
  description: 'Idiomatic Go port of Intervention Image',
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
                { text: 'PHP migration', link: '/guide/migration' },
              ],
            },
          ],
          '/api/': [
            {
              text: 'API',
              items: [
                { text: 'Manager', link: '/api/manager' },
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
          message: 'MIT License. Port of Intervention Image by Oliver Vogel.',
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
                { text: '从 PHP 迁移', link: '/zh/guide/migration' },
              ],
            },
          ],
          '/zh/api/': [
            {
              text: 'API',
              items: [
                { text: 'Manager', link: '/zh/api/manager' },
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
          message: 'MIT 许可证。移植自 Oliver Vogel 的 Intervention Image。',
          copyright: 'Copyright © 2026 yunkeweb',
        },
      },
    },
  },
})
