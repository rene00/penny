<script>
import { SendAccountMatchRequest } from 'wailsjs/go/services/accountMatchService.js'
import { SendTagRequest } from 'wailsjs/go/services/tagService.js'

export default {
  data() {
    return {
      processAccountMatches: false,
      processTagMatches: false,
    }
  },
  methods: {
    async runTasks() {


      if (this.processAccountMatches) {
        let resp = await SendAccountMatchRequest()
        if (!resp.success) {
            $message.error(resp.msg)
            return
        }
      }

      if (this.processTagMatches) {
        let resp = await SendTagRequest()
        if (!resp.success) {
          $message.error(resp.msg)
          return
        }
      }

      $message.success("Request sent")
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
                            <h6 class="m-0 font-weight-bold text-primary">Tasks</h6>
                        </div>
                        <div class="card-body">
                            <form v-on:submit.prevent="runTasks">
                                <div class="mb-3">
                                  <div class="form-group">
                                    <div class="checkbox">
                                        <label>
                                        <input id="processAccountMatches" name="processAccountMatches" type="checkbox" v-model="processAccountMatches"> Process Account Matches
                                        </label>
                                    </div>
                                    <div class="checkbox">
                                        <label>
                                        <input id="processTagMatches" name="processTagMatches" type="checkbox" v-model="processTagMatches"> Process Tag Matches
                                        </label>
                                    </div>
                                  </div>
                                </div>
                                <div class="row">
                                    <div class="col-auto">
                                        <button class="btn btn-primary" type="submit">Submit</button>
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
