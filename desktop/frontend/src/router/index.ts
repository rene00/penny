import { createRouter, createWebHashHistory } from 'vue-router'


const router = createRouter({
  history: createWebHashHistory(import.meta.env.BASE_URL),
  routes: [
    {
        path: '/',
        name: 'home',
        component: () => import('../views/GetTransactions.vue'),
    },
    {
        path: '/create-account',
        name: 'create-account',
        component: () => import('../views/CreateAccount.vue'),
    },
    {
        path: '/get-accounts',
        name: 'get-accounts',
        component: () => import('../views/GetAccounts.vue'),
    },
    {
        path: '/get-account/:id',
        name: 'get-account',
        component: () => import('../views/GetAccount.vue'),
        props: true
    },
    {
        path: '/create-transaction',
        name: 'create-transaction',
        component: () => import('../views/CreateTransaction.vue'),
    },
    {
        path: '/get-transactions',
        name: 'get-transactions',
        component: () => import('../views/GetTransactions.vue')
    },
    {
        path: '/get-transaction/:id',
        name: 'get-transaction',
        component: () => import('../views/GetTransaction.vue'),
        props: true
    },
    {
        path: '/import-transactions',
        name: 'import-transactions',
        component: () => import('../views/ImportTransactions.vue'),
    },
    {
        path: '/get-account-matches',
        name: 'get-account-matches',
        component: () => import('../views/GetAccountMatches.vue'),
    },
    {
        path: '/create-account-match',
        name: 'create-account-match',
        component: () => import('../views/CreateAccountMatch.vue'),
    },
    {
        path: '/get-account-match/:id',
        name: 'get-account-match',
        component: () => import('../views/GetAccountMatch.vue'),
        props: true
    },
    {
        path: '/tasks',
        name: 'tasks',
        component: () => import('../views/Tasks.vue'),
        props: true
    },
  ]
})


export default router
