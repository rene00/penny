<script>
import { GetAccountMatch, UpdateAccountMatch, DeleteAccountMatch, UndeleteAccountMatch, CreateAccountMatchRegex, DeleteAccountMatchRegex, UpdateAccountMatchRegex } from 'wailsjs/go/services/accountMatchService.js'
import { GetAccounts } from 'wailsjs/go/services/accountService.js'

export default {
  props: ['id'],
  data() {
    return {
      newRegex: false,
      account: {},
      accounts: [],
      accountMatch: {
        sourceAccount: {},
        destinationAccount: {},
        regexes: [],
      },
    }
  },
  created() {
    this.$watch(
      () => this.$route.params,
      () => {
        this.getAccounts()
        this.getAccountMatch()
      },
      { immediate: true },
    )
  },
  beforeRouteUpdate: function(to, from, next) {
    this.accounts = this.getAccounts();
    this.getAccountMatch();
    next();
  },
  methods: {
    async addNewRegex() {
      let accountMatch = this.accountMatch
      accountMatch.regexes.push({id:0, regex:""})
      this.accountMatch = accountMatch
      this.newRegex = true
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
    async getAccountMatch() {
      const { success, msg, data } = await GetAccountMatch(Number(this.id))
      if (!success) {
        $message.error(msg)
        return
      }
      let accountMatch = {
        id: data.id,
        name: data.name,
        description: data.description,
        deleted: data.deleted,
        regexes: data.regexes,
        sourceAccount: {
          id: data.source_account.id,
          name: data.source_account.name,
        },
        destinationAccount: {
          id: data.destination_account.id,
          name: data.destination_account.name,
        },
      }
      this.accountMatch = accountMatch
    },
    async updateAccountMatch(){
      const { success, msg, data } = await UpdateAccountMatch(this.accountMatch.id, this.accountMatch.sourceAccount.id, this.accountMatch.destinationAccount.id, this.accountMatch.name, this.accountMatch.description)
      if (!success) {
          $message.error(msg)
          return
      }
      for (const regex of this.accountMatch.regexes) {
        if (regex.id == 0 && regex.regex != "") {
          const newRegexResp = await CreateAccountMatchRegex(this.accountMatch.id, regex.regex);
          if (!newRegexResp.success) {
            $message.error(newRegexResp.msg)
            return
          }
          continue
        }

        if (regex.regex == "") {
          const deleteRegexResp = await DeleteAccountMatchRegex(regex.id)
          if (!deleteRegexResp.success) {
            $message.error(deleteRegexResp)
            return
          }
          continue
        }

        const updateRegexResp = await UpdateAccountMatchRegex(regex.id, regex.regex);
        if (!updateRegexResp.success) {
          $message.error(updateRegexResp.msg)
          return
        }
      }

      $message.success(msg)
      this.getAccountMatch()
      return
    },
    async deleteAccountMatch(){
        const { success, msg } = await DeleteAccountMatch(Number(this.accountMatch.id))
        if (!success) {
            $message.error(msg)
            return
        }
        $message.success(msg)
        this.getAccountMatch()
        return
    },
    async undeleteAccountMatch(){
        const { success, msg } = await UndeleteAccountMatch(Number(this.accountMatch.id))
        if (!success) {
            $message.error(msg)
            return
        }
        $message.success(msg)
        this.getAccountMatch()
        return
    }
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
                            <h6 class="m-0 font-weight-bold text-primary">Account Match</h6>
                        </div>
                        <div class="card-body">
                            <form v-on:submit.prevent="updateAccountMatch">
                                <div class="mb-3">
                                    <label class="small mb-1">
                                      Name
                                    </label>
                                    <input class="form-control" type="text" v-model="accountMatch.name">
                                </div>
                                <div class="mb-3">
                                    <label class="small mb-1">
                                      Description
                                    </label>
                                    <input class="form-control" type="text" v-model="accountMatch.description">
                                </div>
                                <div class="row gx-3 mb-3">
                                    <div class="col-md-6">
                                        <label class="mall mb-1">Source Account</label>
                                        <div class="mb-3">
                                            <select class="form-select" aria-label="Type" v-model="accountMatch.sourceAccount.id">
                                                <option v-for="account in accounts" :value="account.id">{{ account.name }}</option>
                                            </select>
                                        </div>
                                    </div>
                                </div>
                                <div class="row gx-3 mb-3">
                                    <div class="col-md-6">
                                        <label class="mall mb-1">Destination Account</label>
                                        <div class="mb-3">
                                            <select class="form-select" aria-label="Type" v-model="accountMatch.destinationAccount.id">
                                              <option v-for="account in accounts" :value="account.id">{{ account.name }}</option>
                                            </select>
                                        </div>
                                    </div>
                                </div>
                                <div v-for="(regex, idx) in accountMatch.regexes" class="row gx-3 mb-3">
                                  <div class="col-md-6">
                                    <label class="small mb-1">Filter</label>
                                    <div class="mb-3">
                                        <input class="form-control" type="text" v-model="regex.regex">
                                    </div>
                                  </div>
                                </div>
                                <div class="row">
                                    <div class="col-auto">
                                        <button class="btn btn-danger btn-primary me-2" type="button" @click="undeleteAccountMatch" v-if="accountMatch.deleted">Undelete</button>
                                        <button class="btn btn-danger btn-primary me-2" type="button" @click="deleteAccountMatch" v-else>Delete</button>
                                    </div>
                                    <div class="col-auto">
                                        <button class="btn btn-primary" type="button" @click="addNewRegex">Add new Filter</button>
                                    </div>
                                    <div class="col-auto">
                                        <button class="btn btn-primary" type="submit" v-if="accountMatch.deleted == false">Save changes</button>
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

    <!-- Scroll to Top Button-->
    <a class="scroll-to-top rounded" href="#page-top">
        <i class="fas fa-angle-up"></i>
    </a>
</template>
