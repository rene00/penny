import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
// import './assets/css/main.css';
import './css/sb-admin-2.min.css';
import './vendor/fontawesome-free/css/all.min.css';
import { i18n } from '@/utils/i18n.js'
import { setupDiscreteApi } from '@/utils/discrete.js'
import Sidebar from './components/Sidebar.vue'
import Topbar from './components/Topbar.vue'
import Vue3EasyDataTable from 'vue3-easy-data-table';
import 'vue3-easy-data-table/dist/style.css';

async function setupApp() {
  const app = createApp(App)
  await setupDiscreteApi()
  app.use(i18n)
  app.component('Sidebar', Sidebar)
  app.component('Topbar', Topbar)
  app.component('EasyDataTable', Vue3EasyDataTable)
  app.use(router)
  app.mount('#app')
}

setupApp()
