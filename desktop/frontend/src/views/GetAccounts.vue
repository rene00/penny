<script>
import { ref } from 'vue';
import { useRouter, NavigationFailureType, isNavigationFailure } from 'vue-router';
import { GetAccounts } from 'wailsjs/go/services/accountService.js'

export default {
  setup() {
    const router = useRouter();
    return {
      router
    }
  },
  data() {
    return {
      accounts: [],
      searchField: ["name", "description", "account_type.name"],
      searchValue: ref(),
      headers: [
        {
          text: "Name",
          value: "name",
        },
        {
          text: "Description",
          value: "description",
        },
        {
          text: "Type",
          value: "account_type.name",
        },
      ]
    }
  },

  created() {
    this.$watch(
      () => this.$route.params,
      () => {
        this.getAccounts()
      },
      { immediate: true },
    )
  },
  methods: {
    async getAccounts() {
      let accounts = [];
      const { success, msg, data } = await GetAccounts()
      if (!success) {
          $message.error(msg)
          return
      }
      for (const account of data) {
        accounts.push({
          id: account.id,
          name: account.name,
          description: account.description,
          account_type: {
            id: account.account_type.id,
            name: account.account_type.name,
          },
        })
      }
      this.accounts = accounts
    },

  },
}

</script>

<template>
    <!-- Page Wrapper -->
    <div id="wrapper">
      <Sidebar/>
      <slot/>

        <!-- Content Wrapper -->
        <div id="content-wrapper" class="d-flex flex-column">

            <!-- Main Content -->
            <div id="content">

                <Topbar/>
                <slot/>


                <!-- Begin Page Content -->
                <div class="container-fluid">

                    <div class="card shadow mb-4">
                        <div class="card-header py-3">
                            <h6 class="m-0 font-weight-bold text-primary">Accounts</h6>
                        </div>

                        <div class="card-body">
                          <div class="mb-3">
                            <span>search:</span>
                            <input type="text" v-model="searchValue">
                              <EasyDataTable :headers="headers" :items="accounts" :search-field="searchField" :search-value="searchValue">
                                <template #item-name="{ name, id }">
                                  <div>
                                    <router-link :to="{ name: 'get-account', params: { id: id }}">{{ name }}</router-link>
                                  </div>
                                </template>
                              </EasyDataTable>
                            </div>
                        </div>

                    </div>

                </div>
                <!-- /.container-fluid -->

            </div>
            <!-- End of Main Content -->

            <!-- Footer -->
            <footer class="sticky-footer bg-white">
            </footer>
            <!-- End of Footer -->

        </div>
        <!-- End of Content Wrapper -->

    </div>
    <!-- End of Page Wrapper -->
</template>
