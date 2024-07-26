<script>
import { CreateTransaction } from 'wailsjs/go/services/transactionService.js'
import { CreateSplit } from 'wailsjs/go/services/splitsService.js'
import { GetAccounts } from 'wailsjs/go/services/accountService.js'
import VueDatePicker from '@vuepic/vue-datepicker'
import '@vuepic/vue-datepicker/dist/main.css'
import { ref } from 'vue'

export default {
  components: { VueDatePicker },
  setup() {
    const date = ref(new Date());
    const format = (date) => {
      const day = date.getDate();
      const month = date.getMonth() + 1;
      const year = date.getFullYear();
      return `${day}/${month}/${year}`;
    }

    return {
      date,
      format,
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
  data() {
    return {
      transaction: {
        memo: null,
      },
      splits: [
        {
          account: {
            id: null,
          },
          amount: 0.00,
        },
        {
          account: {
            id: null,
          },
          amount: 0.00,
        },
      ],
      accounts: [],
    }
  },
  computed: {
    amountInverse() {
      return this.transaction.amount * -1;
    },
  },
  beforeRouteUpdate: function(to, from, next) {
    this.accounts = this.getAccounts();
    next();
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

    async createTransaction() {
      const transactionResponse = await CreateTransaction(this.transaction.memo, this.date)

      let transaction = transactionResponse.data

      this.splits.forEach((split, index) => {
        CreateSplit(transaction.id, split.account.id, split.amount)
      });

      if (!transactionResponse.success) {
        $message.error(transactionResponse.msg)
      } else {
        $message.success(transactionResponse.msg)
        this.$router.push({name: 'get-transaction', params: { id: transaction.id }})
      }
      return
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
                            <h6 class="m-0 font-weight-bold text-primary">Create Transaction</h6>
                        </div>
                        <div class="card-body">
                            <form v-on:submit.prevent="createTransaction">
                                <div class="mb-3">
                                    <label class="small mb-1">
                                      Date
                                    </label>
                                    <VueDatePicker v-model="date" :format="format" :preview-format="format" :enable-time-picker="false"></VueDatePicker>
                                </div>

                                <div class="mb-3">
                                    <label class="small mb-1">
                                      Memo
                                    </label>
                                    <input class="form-control" type="text" required v-model="transaction.memo">
                                </div>

                                <div class="row gx-3 mb-3">
                                    <!-- Form Group (increase account)-->
                                    <div class="col-md-6">
                                        <label class="small mb-1" for="inputFirstName">Account</label>
                                        <div class="mb-3">
                                            <select class="form-select" aria-label="Type" v-model="splits[0].account.id">
                                                <option selected="" disabled="">Select an account:</option>
                                                <option v-for="account in accounts" :value="account.id">{{ account.name }}</option>
                                            </select>
                                        </div>
                                    </div>

                                    <!-- Form Group (increase amount)-->
                                    <div class="col-md-6">
                                        <label class="small mb-1">Amount</label>
                                        <input class="form-control" v-model="splits[0].amount" name="amount" type="number" value="0.00" step="any">
                                    </div>
                                </div>

                                <!-- <div v-for="(split, index) in splits" :key="index" v-show="splits.length >= 2"> -->
                                <div class="row gx-3 mb-3" v-show="splits[0].amount != '0.00'">
                                  <!-- Form Group (increase account)-->
                                  <div class="col-md-6">
                                      <label class="small mb-1" for="inputFirstName">Account</label>
                                      <div class="mb-3">
                                          <select class="form-select" aria-label="Type" v-model="splits[1].account.id">
                                              <option selected="" disabled="">Select an account:</option>
                                              <!-- <option v-for="account in accounts.filter(account => account.account_type.name === 'Current Liability')" :value="account.id">{{ account.name }}</option> -->
                                              <option v-for="account in accounts" :value="account.id">{{ account.name }}</option>
                                          </select>
                                      </div>
                                  </div>

                                  <!-- Form Group (increase amount)-->
                                  <div class="col-md-6">
                                      <label class="small mb-1">Amount</label>
                                      <input class="form-control" v-model="splits[1].amount" name="amount" type="number" step="any">
                                  </div>

                                </div>
                                <!-- </div> -->

                                <div class="row gx-3 jjmb-3">
                                  <div class="col-md-6" v-show="splits[1].amount != '0.00' && splits[0].amount != '0.00'">
                                    <button class="btn btn-primary" type="submit">Create</button>
                                  </div>
                                </div>
                            </form>
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
