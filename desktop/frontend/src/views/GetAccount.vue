<script>
import { GetAccount, GetChildAccountTypes, UpdateAccount } from 'wailsjs/go/services/accountService.js'

export default {
  props: ['id'],
  data() {
    return {
      active: true,
      accountTypes: [],
      account: {},
    }
  },
  created() {
    this.$watch(
      () => this.$route.params,
      () => {
        this.getAccount();
        this.getChildAccountTypes();
      },
      { immediate: true },
    )
  },
  beforeRouteUpdate: function(to, from, next) {
    this.getAccount();
    this.getChildAccountTypes();
    if (this.active) {
      next(false);
    } else {
      next();
    }
    // next();
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

    async getAccount() {
      let account = {}
      const { success, msg, data } = await GetAccount(Number(this.id))
      if (!success) {
        $message.error(msg)
        return
      }
      account = {
        id: data.id,
        name: data.name,
        description: "",
        amount: data.amount,
        account_type: {
          id: data.account_type.id,
          name: data.account_type.name,
        },
      }
      if (Object.hasOwn(data, 'description')) {
          account.description = data.description
      }
      this.account = account
    },

    async updateAccount(){
        const { success, msg, data } = await UpdateAccount(Number(this.account.id), this.account.name, this.account.description, Number(this.account.account_type.id))
        if (!success) {
            $message.error(msg)
            return
        }
        $message.success(msg)
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
                            <h6 class="m-0 font-weight-bold text-primary">Account</h6>
                        </div>
                        <div class="card-body">
                            <form v-on:submit.prevent="updateAccount">
                                <div class="mb-3">
                                    <label class="small mb-1">
                                      Name
                                    </label>
                                    <input class="form-control" type="text" v-model="account.name">
                                </div>
                                <div class="mb-3">
                                    <label class="small mb-1">
                                      Description
                                    </label>
                                    <input class="form-control" type="text" v-model="account.description">
                                </div>
                                <div class="mb-3">
                                    <label class="small mb-1">
                                      Amount
                                    </label>
                                    <input class="form-control" readonly type="text" v-model="account.amount">
                                </div>
                                <div class="mb-1">
                                    <label class="small mb-1">
                                      Type
                                    </label>
                                </div>
                                <div class="mb-3">
                                    <select class="form-select" aria-label="Type" v-model="account.account_type.id">
                                        <option v-for="accountType in accountTypes" :value="accountType.id" :selected="accountType.id == account.account_type.id">{{ accountType.name }}</option>
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

    <!-- Scroll to Top Button-->
    <a class="scroll-to-top rounded" href="#page-top">
        <i class="fas fa-angle-up"></i>
    </a>

    <!-- Logout Modal-->
    <div class="modal fade" id="logoutModal" tabindex="-1" role="dialog" aria-labelledby="exampleModalLabel"
        aria-hidden="true">
        <div class="modal-dialog" role="document">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title" id="exampleModalLabel">Ready to Leave?</h5>
                    <button class="close" type="button" data-dismiss="modal" aria-label="Close">
                        <span aria-hidden="true">×</span>
                    </button>
                </div>
                <div class="modal-body">Select "Logout" below if you are ready to end your current session.</div>
                <div class="modal-footer">
                    <button class="btn btn-secondary" type="button" data-dismiss="modal">Cancel</button>
                    <a class="btn btn-primary" href="login.html">Logout</a>
                </div>
            </div>
        </div>
    </div>
</template>
