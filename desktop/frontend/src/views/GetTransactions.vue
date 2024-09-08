<script>
import { ref } from 'vue';
import { GetTransactions, GetTransactionsAccount, GetTransactionsTag, GetTotalBalance } from 'wailsjs/go/services/transactionService.js'
import { GetAccountByName } from 'wailsjs/go/services/accountService.js'
import { useRouter } from 'vue-router';

export default {
  props: [
    'accountid',
    'tagid',
  ],
  setup() {
    const router = useRouter();
    return {
      router
    }
  },

  data() {
    return {
      searchValue: ref(),
      searchFields: ['memo', 'amount'],
      transactions: [],
      totalBalance: ref(),
      imbalancedAccountID: ref(),
      headers: [
        {
          text: 'Date',
          value: 'date',
        },
        {
          text: 'Memo',
          value: 'memo',
        },
        {
          text: 'Account',
          value: 'display.account.name',
        },
        {
          text: 'Amount',
          value: 'amount',
        },
      ]
    }
  },

  created() {
    this.$watch(
      () => this.$route.params,
      () => {
        this.getTransactions()
        this.getTotalBalance()
        this.getImbalancedAccountID()
      },
      { immediate: true },
    )
  },

  methods: {
    async getImbalancedAccountID() {
      const getAccountByNameResp = await GetAccountByName("Imbalanced")
      if (!getAccountByNameResp.success) {
        return
      }
      this.imbalancedAccountID = getAccountByNameResp.data.id
    },
    async getTotalBalance() {
      const getTotalBalanceResp = await GetTotalBalance()
      if (!getTotalBalanceResp.success) {
        $message.error(getTotalBalanceResp.msg)
        return
      }
      this.totalBalance = getTotalBalanceResp.data
    },
    async getTransactions() {
      let transactions = [];

			let success = false
			let msg = ""
			let data = {}

      if (typeof this.accountid !== 'undefined') {
        const getTransactionsAccountResp = await GetTransactionsAccount(Number(this.accountid))
				success = getTransactionsAccountResp.success
				msg = getTransactionsAccountResp.msg
				data = getTransactionsAccountResp.data
        if (!success) {
            $message.error(msg)
            return
        }
      } else if (typeof this.tagid !== 'undefined') {
        const getTransactionsTagResp = await GetTransactionsTag(Number(this.tagid))
				success = getTransactionsTagResp.success
				msg = getTransactionsTagResp.msg
				data = getTransactionsTagResp.data
        if (!success) {
            $message.error(msg)
            return
        }

			} else {
				const getTransactionsResp = await GetTransactions()
				success = getTransactionsResp.success
				msg = getTransactionsResp.msg
				data = getTransactionsResp.data
				if (!success) {
						$message.error(msg)
						return
				}
      }

      for (const transaction of data) {
        let t = {
          id: transaction.id,
          memo: transaction.memo,
          date: transaction.date,
          amount: transaction.amount,
          display: {
            account: {
              name: transaction.display.account.name,
              id: transaction.display.account.id
            }
          }
        }
        transactions.push(t)
      }
      this.transactions = transactions
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
                            <h6 class="m-0 font-weight-bold text-primary">Transactions</h6>
                        </div>
                        <div class="card-body">
                          <div class="mb-3">
                            <span>search:</span>
                            <input type="text" v-model="searchValue">
                              <EasyDataTable :headers="headers" :items="transactions" :search-field="searchField" :search-value="searchValue">

                                <template #item-memo="{ memo, id }">
                                  <div>
                                    <router-link :to="{ name: 'get-transaction', params: { id: id }}">{{ memo }}</router-link>
                                  </div>
                                </template>

                                <template #item-display.account.name="item">
                                  <div>
                                    <router-link :to="{ name: 'get-account', params: { id: item.display.account.id }}">{{ item.display.account.name }}</router-link>
                                  </div>
                                </template>
                              </EasyDataTable>
                            </div>
                        </div>

                      <div class="col-md-6">
                          <label class="small mb-1">Total Balance:</label>
                          <label class="small mb-1"><router-link :to="{ name: 'get-account', params: { id: this.imbalancedAccountID }}">{{ totalBalance }}</router-link></label>
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
