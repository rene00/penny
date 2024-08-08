<script>
import { CreateTag, CreateTagRegex } from 'wailsjs/go/services/tagService.js'

export default {
  data() {
    return {
      tag: {},
      regex: "",
    }
  },
  methods: {
    async createTag() {
      const tagResponse = await CreateTag(this.tag.name, this.tag.description)

      if (!tagResponse.success) {
        $message.error(tagResponse.msg)
        return
      }

      if (this.regex.length > 0) {
        const tagRegexResponse = await CreateTagRegex(tagResponse.data.id, this.regex)
        if (!tagRegexResponse.success) {
          $message.error(tagRegexResponse.msg)
          return
        }
      }

      $message.success(tagResponse.msg)
      this.$router.push({name: 'get-tag', params: { id: tagResponse.data.id }})
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
                            <h6 class="m-0 font-weight-bold text-primary">Create Tag</h6>
                        </div>
                        <div class="card-body">
                            <form v-on:submit.prevent="createTag">
                                <div class="mb-3">
                                    <label class="small mb-1">
                                      Name
                                    </label>
                                    <input class="form-control" type="text" required v-model="tag.name">
                                </div>
                                <div class="mb-3">
                                    <label class="small mb-1">
                                      Description
                                    </label>
                                    <input class="form-control" type="text" v-model="tag.description">
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
