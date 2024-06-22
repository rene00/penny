<script>
import { CreateAccount, GetChildAccountTypes } from 'wailsjs/go/services/accountService.js'

export default {
  data() {
    return {
      accountTypes: [],
      account: {}
    }
  },
  created() {
    this.$watch(
      () => this.$route.params,
      () => {
        this.getChildAccountTypes()
      },
      { immediate: true },
    )
  },
  methods: {
    async getChildAccountTypes() {
      let accountTypes = [];
      const { success, msg, data } = await GetChildAccountTypes()
      if (!success) {
        $message.error(msg)
        return
      }

      for (const accountType of data) {
        accountTypes.push({
          id: accountType.id,
          name: accountType.name,
        })
      }

      this.accountTypes = accountTypes
    },

    async createAccount() {
      const { success, msg, data } = await CreateAccount(this.account.name, this.account.description, this.account.accountTypeID)
      if (!success) {
        $message.error(msg)
      } else {
        $message.success(msg)
        this.$router.push({name: 'get-account', params: { id: data.id }})
      }
      return 
    },

  },
  beforeRouteUpdate: function(to, from, next) {
    this.accountTypes = this.getChildAccountTypes();
    next();
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
                            <h6 class="m-0 font-weight-bold text-primary">Create Account</h6>
                        </div>
                        <div class="card-body">
                            <form v-on:submit.prevent="createAccount">
                                <div class="mb-3">
                                    <label class="small mb-1">
                                      Name
                                    </label>
                                    <input class="form-control" type="text" required v-model="account.name">
                                </div>
                                <div class="mb-3">
                                    <label class="small mb-1">
                                      Description
                                    </label>
                                    <input class="form-control" type="text" v-model="account.description">
                                </div>
                                <div class="mb-1">
                                    <label class="small mb-1">
                                      Type
                                    </label>
                                </div>
                                <div class="mb-3">
                                    <select class="form-select" aria-label="Type" v-model="account.accountTypeID">
                                        <option selected="" disabled="">Select a type:</option>
                                        <option v-for="accountType in accountTypes" :value="accountType.id">{{ accountType.name }}</option>
                                    </select>
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
