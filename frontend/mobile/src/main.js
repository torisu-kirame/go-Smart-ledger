import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { Capacitor } from '@capacitor/core'
import { StatusBar, Style } from '@capacitor/status-bar'
import 'vant/lib/index.css'
import './styles/global.css'
import App from './App.vue'
import router from './router'
import { loadApiBaseFromStorage } from './utils/apiBase'

async function initNativeChrome() {
  if (!Capacitor.isNativePlatform()) return
  try {
    await StatusBar.setStyle({ style: Style.Light })
    await StatusBar.setBackgroundColor({ color: '#1a56db' })
  } catch {
    /* optional plugin */
  }
}

async function bootstrap() {
  await loadApiBaseFromStorage()
  await initNativeChrome()
  const app = createApp(App)
  app.use(createPinia())
  app.use(router)
  app.mount('#app')
}

bootstrap()
