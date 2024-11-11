<script>
import { GetAccount, GetChildAccountTypes, UpdateAccount, DeleteAccount, UndeleteAccount } from 'wailsjs/go/services/accountService.js'

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
        description: data.description,
        amount: data.amount,
        amount2: data.amount2,
        account_type: {
          id: data.account_type.id,
          name: data.account_type.name,
        },
        deleted: data.deleted,
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

    async deleteAccount(){
        const { success, msg } = await DeleteAccount(Number(this.account.id))
        if (!success) {
            $message.error(msg)
            return
        }
        $message.success(msg)
        this.getAccount()
        return
    },

    async undeleteAccount(){
        const { success, msg } = await UndeleteAccount(Number(this.account.id))
        if (!success) {
            $message.error(msg)
            return
        }
        $message.success(msg)
        this.getAccount()
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
                                    <div>
                                      <router-link :to="{ name: 'get-transactions-account', params: { accountid: account.id }}">{{ account.amount }}, {{ account.amount2 }}</router-link>
                                    </div>
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
                                <div class="row">
                                    <div class="col-auto">
                                        <button class="btn btn-danger btn-primary me-2" type="button" @click="undeleteAccount" v-if="account.deleted">Undelete</button>
                                        <button class="btn btn-danger btn-primary me-2" type="button" @click="deleteAccount" v-else>Delete</button>
                                    </div>
                                    <div class="col-auto">
                                        <button class="btn btn-primary" type="submit" v-if="account.deleted == false">Save changes</button>
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
