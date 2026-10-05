import './style/element_visiable.scss'
import 'element-plus/theme-chalk/dark/css-vars.css'
import { createApp } from 'vue'
import ElementPlus from 'element-plus'

import 'element-plus/dist/index.css'
// 导入Smart-NESMA设计系统 - 仅限Smart-NESMA组件使用，不影响全局样式
import './styles/index.scss'
// 引入gin-vue-admin前端初始化相关内容
import './core/gin-vue-admin'
// 引入封装的router
import router from '@/router/index'
import '@/permission'
import run from '@/core/gin-vue-admin.js'
import auth from '@/directive/auth'
import { store } from '@/pinia'
import App from './App.vue'

// 引入Vue3-Beautiful-Chat
import Chat from 'vue3-beautiful-chat'
import 'element-plus/dist/index.css'

const app = createApp(App)
app.config.productionTip = false

app.use(run).use(ElementPlus).use(store).use(auth).use(router).use(Chat).mount('#app')
export default app
