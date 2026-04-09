import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import dialogPlugin from './plugins/dialog'

// Import global text overflow styles
import './styles/text-overflow.css'

const app = createApp(App)
app.use(router)
app.use(dialogPlugin)
app.mount('#app')
