<script>
import { ref } from 'vue';
import { GetAttachments } from 'wailsjs/go/services/attachmentService.js'
import { useRouter } from 'vue-router';

export default {
  setup() {
    const router = useRouter();
    return {
      router
    }
  },

  data() {
    return {
      searchValue: ref(),
      searchFields: [],
      attachments: [],
      headers: [
        {
          text: 'Name',
          value: 'name',
        },
      ]
    }
  },

  created() {
    this.$watch(
      () => this.$route.params,
      () => {
        this.getAttachments()
      },
      { immediate: true },
    )
  },

  methods: {
    async getAttachments() {
      let attachments = [];
      const { success, msg, data } = await GetAttachments()
      if (!success) {
          $message.error(msg)
          return
      }
      for (const attachment of data) {
        attachments.push(
          {
            id: attachment.id,
            name: attachment.name,
            size: attachment.size,
          }
        )
      }
      this.attachments = attachments
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
                            <h6 class="m-0 font-weight-bold text-primary">Attachments</h6>
                        </div>
                        <div class="card-body">
                          <div class="mb-3">
                            <span>search:</span>
                            <input type="text" v-model="searchValue">
                              <EasyDataTable :headers="headers" :items="attachments" :search-field="searchField" :search-value="searchValue">
                                <template #item-name="{ name, id }">
                                  <div>
                                    <router-link :to="{ name: 'get-attachment', params: { id: id }}">{{ name }}</router-link>
                                  </div>
                                </template>
                              </EasyDataTable>
                            </div>
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
