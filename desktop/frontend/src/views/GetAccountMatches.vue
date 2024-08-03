<script>
import { ref } from 'vue';
import { GetAccountMatches } from 'wailsjs/go/services/accountMatchService.js'
import { useRouter } from 'vue-router';

export default {
  setup() {
    const router = useRouter();
    return {
      router
    }
  },

  data() {
    return {
      searchValue: ref(),
      searchFields: [],
      accountMatches: [],
      headers: [
        {
          text: 'Name',
          value: 'name',
        },
        {
          text: 'Description',
          value: 'description',
        },
        {
          text: 'Source Account',
          value: 'sourceAccount.name',
        },
        {
          text: 'Destination Account',
          value: 'destinationAccount.name',
        }
      ]
    }
  },

  created() {
    this.$watch(
      () => this.$route.params,
      () => {
        this.getAccountMatches()
      },
      { immediate: true },
    )
  },

  methods: {
    async getAccountMatches() {
      let accountMatches = [];
      const { success, msg, data } = await GetAccountMatches()
      if (!success) {
          $message.error(msg)
          return
      }
      for (const accountMatch of data) {
        let am = {
          id: accountMatch.id,
          name: accountMatch.name,
          description: accountMatch.description,
          sourceAccount: {
            id: accountMatch.source_account.id,
            name: accountMatch.source_account.name,
          },
          destinationAccount: {
            id: accountMatch.destination_account.id,
            name: accountMatch.destination_account.name,
          }
        }
        accountMatches.push(am)
      }
      this.accountMatches = accountMatches
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
                            <h6 class="m-0 font-weight-bold text-primary">Account Matches</h6>
                        </div>
                        <div class="card-body">
                          <div class="mb-3">
                            <span>search:</span>
                            <input type="text" v-model="searchValue">
                              <EasyDataTable :headers="headers" :items="accountMatches" :search-field="searchField" :search-value="searchValue">
                                <template #item-name="{ name, id }">
                                  <div>
                                    <router-link :to="{ name: 'get-account-match', params: { id: id }}">{{ name }}</router-link>
                                  </div>
                                </template>
                                <template #item-sourceAccount.name="item">
                                  <div>
                                    <router-link :to="{ name: 'get-account', params: { id: item.sourceAccount.id }}">{{ item.sourceAccount.name }}</router-link>
                                  </div>
                                </template>
                                <template #item-destinationAccount.name="item">
                                  <div>
                                    <router-link :to="{ name: 'get-account', params: { id: item.destinationAccount.id }}">{{ item.destinationAccount.name }}</router-link>
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
