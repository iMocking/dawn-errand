import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { watch } from 'vue'
import App from './App.vue'
import router, { refreshTitle } from './router'
import { useSiteStore } from './stores/site'
import './style.css'

const app = createApp(App)
app.use(createPinia())
app.use(router)

// 语言切换或公共数据重载完成后刷新文档标题（需在 pinia 安装后注册）
watch(
  () => [useSiteStore().locale, useSiteStore().common],
  () => refreshTitle(),
)

app.mount('#app')
