import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve, dirname } from 'path'
import { fileURLToPath } from 'url'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { AntDesignVueResolver } from 'unplugin-vue-components/resolvers'
 
const __dirname = dirname(fileURLToPath(import.meta.url)) 
export default defineConfig( () => ({
  // base: command === 'build' ? '/etlApi/' : '/',
  base: '/',
  plugins: [
    vue(),
    // Ant Design Vue 自动导入配置
    AutoImport({
      resolvers: [AntDesignVueResolver()],
      imports: ['vue', 'vue-router', 'pinia'],
      dts: 'auto-imports.d.ts',
    }),
    Components({
      resolvers: [AntDesignVueResolver({
        importStyle: false, // css in js
      })],
      dts: 'components.d.ts',
    }),
  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  // server: {
  //   port: 3000,
  //   open: true,
  //   proxy: {
  //     '/api': {
  //       target: 'http://localhost:8080',
  //       changeOrigin: true,
  //       rewrite: (path) => path.replace(/^\/api/, ''),
  //     },
  //   },
  // },
}))

