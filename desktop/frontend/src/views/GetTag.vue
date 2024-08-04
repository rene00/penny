<script>
import { GetTag, UpdateTag, DeleteTag, UndeleteTag, CreateTagRegex, DeleteTagRegex, UpdateTagRegex } from 'wailsjs/go/services/tagService.js'
import { GetAccounts } from 'wailsjs/go/services/accountService.js'

export default {
  props: ['id'],
  data() {
    return {
      accounts: [],
      tag: {
        regexes: [],
      },
    }
  },
  created() {
    this.$watch(
      () => this.$route.params,
      () => {
        this.getAccounts()
        this.getTag()
      },
      { immediate: true },
    )
  },
  beforeRouteUpdate: function(to, from, next) {
    this.accounts = this.getAccounts();
    this.getTag();
    next();
  },
  methods: {
    async addNewRegex() {
      this.tag.regexes.push({id:0, regex:""})
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
    async getTag() {
      const { success, msg, data } = await GetTag(Number(this.id))
      if (!success) {
        $message.error(msg)
        return
      }
      let tag = {
        id: data.id,
        name: data.name,
        description: data.description,
        amount: data.amount,
        deleted: data.deleted,
        regexes: data.regexes,
      }
      this.tag = tag
    },
    async updateTag(){
      const updateTagResp = await UpdateTag(this.tag.id, this.tag.name, this.tag.description)
      if (!updateTagResp.success) {
          $message.error(updateTagResp.msg)
          return
      }

      for (const regex of this.tag.regexes) {
        if (regex.id == 0 && regex.regex != "") {
          const newRegexResp = await CreateTagRegex(this.tag.id, regex.regex);
          if (!newRegexResp.success) {
            $message.error(newRegexResp.msg)
            return
          }
          continue
        }

        if (regex.regex == "") {
          const deleteRegexResp = await DeleteTagRegex(regex.id)
          if (!deleteRegexResp.success) {
            $message.error(deleteRegexResp)
            return
          }
          continue
        }

        const updateRegexResp = await UpdateTagRegex(regex.id, regex.regex)
        if (!updateRegexResp.success) {
          $message.error(updateRegexResp.msg)
          return
        }
      }

      $message.success(updateTagResp.msg)
      this.getTag()
      return
    },
    async deleteTag(){
        const { success, msg } = await DeleteTag(Number(this.tag.id))
        if (!success) {
            $message.error(msg)
            return
        }
        $message.success(msg)
        this.getTag()
        return
    },
    async undeleteTag(){
        const { success, msg } = await UndeleteTag(Number(this.tag.id))
        if (!success) {
            $message.error(msg)
            return
        }
        $message.success(msg)
        this.getTag()
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
                            <h6 class="m-0 font-weight-bold text-primary">Tag</h6>
                        </div>
                        <div class="card-body">
                            <form v-on:submit.prevent="updateTag">
                                <div class="mb-3">
                                    <label class="small mb-1">
                                      Name
                                    </label>
                                    <input class="form-control" type="text" v-model="tag.name">
                                </div>
                                <div class="mb-3">
                                    <label class="small mb-1">
                                      Description
                                    </label>
                                    <input class="form-control" type="text" v-model="tag.description">
                                </div>
                                <div class="mb-3">
                                    <label class="small mb-1">
                                      Amount
                                    </label>
                                    <input class="form-control" readonly type="text" v-model="tag.amount">
                                </div>
                                <div v-for="(regex, idx) in tag.regexes" class="row gx-3 mb-3">
                                  <div class="col-md-6">
                                    <label class="small mb-1">Filter</label>
                                    <div class="mb-3">
                                        <input class="form-control" type="text" v-model="regex.regex">
                                    </div>
                                  </div>
                                </div>
                                <div class="row">
                                    <div class="col-auto">
                                        <button class="btn btn-danger btn-primary me-2" type="button" @click="undeleteTag" v-if="tag.deleted">Undelete</button>
                                        <button class="btn btn-danger btn-primary me-2" type="button" @click="deleteTag" v-else>Delete</button>
                                    </div>
                                    <div class="col-auto">
                                        <button class="btn btn-primary" type="button" @click="addNewRegex">Add new Filter</button>
                                    </div>
                                    <div class="col-auto">
                                        <button class="btn btn-primary" type="submit" v-if="tag.deleted == false">Save changes</button>
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
