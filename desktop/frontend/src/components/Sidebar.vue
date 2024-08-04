<script setup></script>

<template>
        <!-- Sidebar -->
        <ul class="navbar-nav bg-gradient-primary sidebar sidebar-dark accordion" id="accordionSidebar">

            <!-- Sidebar - Brand -->
            <router-link :to="{ name: 'get-transactions' }" class="sidebar-brand d-flex align-items-center justify-content-center">
                <div class="sidebar-brand-text mx-3">penny</div>
            </router-link>

            <!-- Divider -->
            <hr class="sidebar-divider my-0">

            <li class="nav-item">
                <a class="nav-link collapsed" href="#" data-toggle="collapse" data-target="#collapseTransactions"
                    aria-expanded="true" aria-controls="collapseTwo">
                    <i class="fas fa-fw fa-cog"></i>
                    <span>Transactions</span>
                </a>
                <div id="collapseTransactions" class="collapse" aria-labelledby="headingTwo" data-parent="#accordionSidebar">
                    <div class="bg-white py-2 collapse-inner rounded">
                        <router-link to="/create-transaction" class="collapse-item">Create</router-link>
                        <router-link to="/get-transactions" class="collapse-item">List</router-link>
                        <router-link to="/import-transactions" class="collapse-item">Import</router-link>
                    </div>
                </div>
            </li>

            <li class="nav-item">
                <a class="nav-link collapsed" href="#" data-toggle="collapse" data-target="#collapseAccounts"
                    aria-expanded="true" aria-controls="collapseAccounts">
                    <i class="fas fa-fw fa-cog"></i>
                    <span>Accounts</span>
                </a>
                <div id="collapseAccounts" class="collapse" aria-labelledby="headingTwo" data-parent="#accordionSidebar">
                    <div class="bg-white py-2 collapse-inner rounded">
                        <router-link to="/create-account" class="collapse-item">Create</router-link>
                        <router-link to="/get-accounts" class="collapse-item">List</router-link>
                    </div>
                </div>
            </li>

            <li class="nav-item">
                <a class="nav-link collapsed" href="#" data-toggle="collapse" data-target="#collapseAccountMatches"
                    aria-expanded="true" aria-controls="collapseAccountMatches">
                    <i class="fas fa-fw fa-cog"></i>
                    <span>Account Matches</span>
                </a>
                <div id="collapseAccountMatches" class="collapse" aria-labelledby="headingTwo" data-parent="#accordionSidebar">
                    <div class="bg-white py-2 collapse-inner rounded">
                        <router-link :to="{ name: 'create-account-match' }" class="collapse-item">Create</router-link>
                        <router-link :to="{ name: 'get-account-matches' }" class="collapse-item">List</router-link>
                    </div>
                </div>
            </li>

            <li class="nav-item">
                <a class="nav-link collapsed" href="#" data-toggle="collapse" data-target="#collapseTags"
                    aria-expanded="true" aria-controls="collapseTags">
                    <i class="fas fa-fw fa-cog"></i>
                    <span>Tags</span>
                </a>
                <div id="collapseTags" class="collapse" aria-labelledby="headingTwo" data-parent="#accordionSidebar">
                    <div class="bg-white py-2 collapse-inner rounded">
                        <router-link :to="{ name: 'create-tag' }" class="collapse-item">Create</router-link>
                        <router-link :to="{ name: 'get-tags' }" class="collapse-item">List</router-link>
                    </div>
                </div>
            </li>

            <!-- Divider -->
            <hr class="sidebar-divider">

            <li class="nav-item">
                <router-link :to="{ name: 'tasks' }" class="nav-link">
                    <i class="fas fa-fw fa-cog"></i>
                    <span>Tasks</span>
                </router-link>
            </li>

        </ul>
        <!-- End of Sidebar -->
</template>

