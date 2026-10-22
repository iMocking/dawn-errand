import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import { permission } from './permission/v-permission'
import './style.css'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.directive('permission', permission)
app.mount('#app')
