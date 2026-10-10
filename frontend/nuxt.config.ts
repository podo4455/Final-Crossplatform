export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  ssr: false,
  devtools: { enabled: false },
  css: ['~/assets/css/main.css'],
  nitro: { output: { publicDir: 'dist' } },
  router: { options: { hashMode: true } },
  vite: { server: { hmr: { host: 'localhost' } } },
})
