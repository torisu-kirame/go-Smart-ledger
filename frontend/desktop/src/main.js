import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import { useAuthStore } from './stores/auth'
import './assets/global.css'
import { initTheme } from './utils/theme'

initTheme() // applies saved theme + accent

const app = createApp(App)
const pinia = createPinia()
app.use(pinia)
app.use(router)

const auth = useAuthStore()
auth.bootstrap().finally(() => app.mount('#app'))
