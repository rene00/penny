<script>
import { ref } from 'vue';
import { GetTags } from 'wailsjs/go/services/tagService.js'
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
      tags: [],
      headers: [
        {
          text: 'Name',
          value: 'name',
        },
        {
          text: 'Description',
          value: 'description',
        },
        {
          text: 'Amount',
          value: 'amount',
        },
      ]
    }
  },

  created() {
    this.$watch(
      () => this.$route.params,
      () => {
        this.getTags()
      },
      { immediate: true },
    )
  },

  methods: {
    async getTags() {
      let tags = [];
      const { success, msg, data } = await GetTags()
      if (!success) {
          $message.error(msg)
          return
      }
      for (const tag of data) {
        tags.push(
          {
            id: tag.id,
            name: tag.name,
            description: tag.description,
            amount: tag.amount,
          }
        )
      }
      this.tags = tags
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
                            <h6 class="m-0 font-weight-bold text-primary">Tags</h6>
                        </div>
                        <div class="card-body">
                          <div class="mb-3">
                            <span>search:</span>
                            <input type="text" v-model="searchValue">
                              <EasyDataTable :headers="headers" :items="tags" :search-field="searchField" :search-value="searchValue">
                                <template #item-name="{ name, id }">
                                  <div>
                                    <router-link :to="{ name: 'get-tag', params: { id: id }}">{{ name }}</router-link>
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
