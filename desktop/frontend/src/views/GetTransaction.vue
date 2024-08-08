<script>
import { GetTransaction, UpdateTransaction, DeleteTransaction, UndeleteTransaction } from 'wailsjs/go/services/transactionService.js'
import { GetAccounts } from 'wailsjs/go/services/accountService.js'
import { GetTransactionNote, CreateTransactionNote, DeleteTransactionNote, UpdateTransactionNote } from 'wailsjs/go/services/noteService.js'
import { UpdateSplit, CreateSplit, DeleteSplit, UndeleteSplit } from 'wailsjs/go/services/splitsService.js'

export default {
  props: ['id'],
  data() {
    return {
      transaction: {},
      splits: [],
      accounts: [],
      note: {
        id: 0,
        note: "",
      },
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
        deleted: data.deleted
      }

      for (const split of data.splits) {
        let s = {
          id: split.id,
          amount: split.amount,
          account: { ...split.account },
        }

        // Use account_id to keep track of account.id. Im unable to figure
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

      const transactionNoteResp = await GetTransactionNote(Number(this.id))
      if (!transactionNoteResp.success) {
        $message.error(transactionNoteResp.msg)
        return
      }
      if ('note' in transactionNoteResp.data) {
        this.note = transactionNoteResp.data
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
            if (typeof split.account_id === "undefined") {
              continue
            }

            if (split.id === 0) {
              const { success, msg, data } = await CreateSplit(this.transaction.id, split.account_id, split.amount)
              if (!success) {
                $message.error(msg)
                return
              }
              continue
            }

            const { success, msg, data } = await UpdateSplit(split.id, split.account_id, Number(split.amount))
            if (!success) {
              $message.error(msg)
              return
            }
        }

        if (this.note.note.length != 0) {
          const newTransactionNoteResp = await CreateTransactionNote(this.transaction.id, this.note.note)
          if (!newTransactionNoteResp.success) {
            $message.error(newTransactionNoteResp.msg)
            return
          }
        }

        if (this.note.note.length == 0 && this.note.id != 0) {
          const deleteTransactionNoteResp = await DeleteTransactionNote(this.note.id)
          if (!deleteTransactionNoteResp.success) {
            $message.error(deleteTransactionNoteResp.msg)
            return
          }
        }

        if (this.note.id != 0) {
          const updateTransactionNoteResp = await UpdateTransactionNote(this.transaction.id, this.note.note)
          if (!updateTransactionNoteResp.success) {
            $message.error(updateTransactionNoteResp.msg)
            return
          }
        }

        updateTransactionMessage.success(msg)
        this.getTransaction()
        return
    },

    async deleteTransaction() {
        const { success, msg } = await DeleteTransaction(Number(this.transaction.id))
        if (!success) {
            $message.error(msg)
            return
        }

        const result = await this.deleteSplits()
        if (result instanceof Error) {
            $message.error(result.msg)
            return
        }

        $message.success(msg)
        this.getTransaction()
        return
    },

    async undeleteTransaction() {
        const { success, msg } = await UndeleteTransaction(Number(this.transaction.id))
        if (!success) {
            $message.error(msg)
            return
        }

        const result = await this.undeleteSplits()
        if (result instanceof Error) {
            $message.error(result.msg)
            return
        }

        $message.success(msg)
        this.getTransaction()
        return
    },

    async deleteSplits() {
        for (const split of this.splits) {
            const { success, msg } = await DeleteSplit(Number(split.id))
            if (!success) {
                return new Error(msg)
            }
        }
    },

    async undeleteSplits() {
        for (const split of this.splits) {
            const { success, msg } = await UndeleteSplit(Number(split.id))
            if (!success) {
                return new Error(msg)
            }
        }
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
                                        <label class="small mb-1">Account</label>
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
                                  <label class="small mb-1">Notes</label>
                                  <div class="mb-3">
                                    <textarea class="form-control" type="text" rows=2 v-model="note.note"></textarea>
                                  </div>
                                </div>
                                <div class="row">
                                    <div class="col-auto">
                                        <button class="btn btn-danger btn-primary me-2" type="button" @click="undeleteTransaction" v-if="transaction.deleted">Undelete</button>
                                        <button class="btn btn-danger btn-primary me-2" type="button" @click="deleteTransaction" v-else>Delete</button>
                                    </div>
                                    <div class="col-auto">
                                        <button class="btn btn-primary" type="submit" v-if="transaction.deleted == false">Save changes</button>
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
