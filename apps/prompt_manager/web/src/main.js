import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import './style.css'

// Import components for routing
import ConversationList from './components/ConversationList.vue'
import ConversationDetail from './components/ConversationDetail.vue'

// Define routes
const routes = [
  { 
    path: '/', 
    name: 'conversations',
    component: ConversationList 
  },
  { 
    path: '/conversation/:id', 
    name: 'conversation-detail',
    component: ConversationDetail,
    props: true
  }
]

// Create router
const router = createRouter({
  history: createWebHistory(),
  routes
})

// Create and mount the app
const app = createApp(App)
app.use(router)
app.mount('#app')