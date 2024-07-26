<script>
import { CreateAccountMatch, CreateAccountMatchRegex } from 'wailsjs/go/services/accountMatchService.js'
import { GetAccounts } from 'wailsjs/go/services/accountService.js'

export default {
  data() {
    return {
      sourceAccount: {},
      destinationAccount: {},
      accounts: [],
      accountMatch: {},
      regex: "",
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
    async createAccountMatch() {
      const accountMatchResponse = await CreateAccountMatch(this.accountMatch.name, this.accountMatch.description, this.sourceAccount.id, this.destinationAccount.id, this.regex)

      if (!accountMatchResponse.success) {
        $message.error(accountMatchResponse.msg)
        return
      }

      if (this.regex.length > 0) {
      	const accountMatchRegexResponse = await CreateAccountMatchRegex(accountMatchResponse.data.id, this.regex)
      	if (!accountMatchRegexResponse.success) {
          $message.error(accountMatchRegexResponse.msg)
	        return
        }
      }

      $message.success(accountMatchResponse.msg)
      this.$router.push({name: 'get-account-match', params: { id: accountMatchResponse.data.id }})
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
                            <h6 class="m-0 font-weight-bold text-primary">Create Account Match</h6>
                        </div>
                        <div class="card-body">
                            <form v-on:submit.prevent="createAccountMatch">
                                <div class="mb-3">
                                    <label class="small mb-1">
                                      Name
                                    </label>
                                    <input class="form-control" type="text" required v-model="accountMatch.name">
                                </div>
                                <div class="mb-3">
                                    <label class="small mb-1">
                                      Description
                                    </label>
                                    <input class="form-control" type="text" v-model="accountMatch.description">
                                </div>
                                <div class="row gx-3 mb-3">
                                    <div class="col-md-6">
                                        <label class="small mb-1">Source Account</label>
                                        <div class="mb-3">
                                            <select class="form-select" aria-label="Type" required v-model="sourceAccount.id">
                                                <option selected="" disabled="">Select an account:</option>
                                                <option v-for="account in accounts" :value="account.id">{{ account.name }}</option>
                                            </select>
                                        </div>
                                    </div>
                                </div>
                                <div class="row gx-3 mb-3">
                                    <div class="col-md-6">
                                        <label class="small mb-1">Destination Account</label>
                                        <div class="mb-3">
                                            <select class="form-select" aria-label="Type" required v-model="destinationAccount.id">
                                                <option selected="" disabled="">Select an account:</option>
                                                <option v-for="account in accounts" :value="account.id">{{ account.name }}</option>
                                            </select>
                                        </div>
                                    </div>
                                </div>
                                <div class="mb-3">
                                    <label class="small mb-1">
                                      Filter
                                    </label>
                                    <div class="mb-3">
                                        <input class="form-control" type="text" required v-model="regex">
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
