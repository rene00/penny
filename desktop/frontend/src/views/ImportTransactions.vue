<script>
import { ref } from 'vue';
import { ImportTransactions } from 'wailsjs/go/services/transactionService.js'

export default {
  data() {
    return {
      fileInput: ref(null),
      fileName: null,
      fileContents: null,
    }
  },
  methods: {
    onPickFile() {
      this.$refs.fileInput.click()
    },
    onFilePicked(event) {
        const files = event.target[0].files
        this.fileName = files[0].name
        const fileReader = new FileReader()
        fileReader.addEventListener('load', () => {
            this.fileContents = fileReader.result
            ImportTransactions(this.fileContents);
        })
        fileReader.readAsDataURL(files[0])
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
                            <h6 class="m-0 font-weight-bold text-primary">Import Transactions</h6>
                        </div>
                        <div class="card-body">
                            <form v-on:submit.prevent="onFilePicked($event)">
                                <div class="mb-3">
                                    <label class="small mb-1">
                                      File
                                    </label>
                                    <input type="file" ref="fileInput">
                                </div>
                                <div class="mb-3">
                                    <button>Import</button>
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
