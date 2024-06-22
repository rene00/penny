<script>
import { GetTransaction, UpdateTransaction } from 'wailsjs/go/services/transactionService.js'
import { GetAccounts } from 'wailsjs/go/services/accountService.js'
import { UpdateSplit, CreateSplit } from 'wailsjs/go/services/splitsService.js'

export default {
  props: ['id'],
  data() {
    return {
      transaction: {},
      splits: [],
      accounts: [],
    }
  },
  created() {
    this.$watch(
      () => this.$route.params,
      () => {
        this.getAccounts();
        this.getTransaction();
      },
      { immediate: true },
    )
  },
  beforeRouteUpdate: function(to, from, next) {
    this.getTransaction();
    next();
  },
  methods: {
    invertSign(value) {
      return -value;
    },

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
    async getTransaction() {
      let transaction = {}
      let splits = []
      const { success, msg, data } = await GetTransaction(Number(this.id))

      if (!success) {
        $message.error(msg)
        return
      }


      transaction = {
        id: data.id,
        memo: data.memo,
        date: data.date,
      }

      for (const split of data.splits) {
        let s = {
          id: split.id,
          amount: split.amount,
          account: { ...split.account },
        }

        // Use account_id to keep track of account.id. I'm unable to figure
        // out how to mutate nested account object.
        s.account_id = 0
        if (s.account && s.account.id) {
          s.account_id = s.account.id
        }

        splits.push(s)
      }

      if (splits.length === 1) {
        let s = {
          id: 0,
          amount: this.invertSign(splits[0].amount),
          account: 0,
        }
        splits.push(s)
      }

      this.transaction = transaction
      this.splits = splits
    },

    async updateTransaction() {
        const { success, msg, data } = await UpdateTransaction(Number(this.transaction.id), this.transaction.memo, this.transaction.date)
        if (!success) {
            $message.error(msg)
            return
        }
        let updateTransactionMessage = $message

        for (const split of this.splits) {
            if (split.id === 0 && split.account_id !== 0) {
              const { success, msg, data } = await CreateSplit(this.transaction.id, split.account_id, split.amount)
              if (!success) {
                $message.error(msg)
                return
              }
              continue
            }

            const { success, msg, data } = await UpdateSplit(split.id, split.account.id, Number(split.amount))
            console.log("success:" + success + " msg:" + msg + " data: " + data)
            if (!success) {
              $message.error(msg)
              return
            }
        }

        updateTransactionMessage.success(msg)
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
                            <h6 class="m-0 font-weight-bold text-primary">Transaction</h6>
                        </div>
                        <div class="card-body">
                            <form v-on:submit.prevent="updateTransaction">
                                <div class="mb-3">
                                    <label class="small mb-1">
                                      Date
                                    </label>
                                    <input class="form-control" type="text" v-model="transaction.date">
                                </div>
                                <div class="mb-3">
                                    <label class="small mb-1">
                                      Memo
                                    </label>
                                    <input class="form-control" type="text" v-model="transaction.memo">
                                </div>

                                <div v-for="(split, index) in splits" class="row gx-3 mb-3">
                                    <div class="col-md-6">
                                        <label class="small mb-1" for="inputFirstName">Account</label>
                                        <div class="mb-3">
                                            <select class="form-select" aria-label="Type" v-model="splits[index].account_id">
                                                <option v-for="account in accounts" :key="account.id" :value="account.id">{{ account.name }}</option>
                                            </select>
                                        </div>
                                    </div>
                                    <div class="col-md-6">
                                        <label class="small mb-1">Amount</label>
                                        <input class="form-control" v-model="splits[index].amount" name="amount" type="number" value="0.00" step="any">
                                    </div>
                                </div>
                                <div class="mb-3">
                                    <button class="btn btn-primary" type="submit">Save changes</button>
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
