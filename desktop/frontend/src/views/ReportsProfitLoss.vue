<script>

import { GetProfitLossReport } from 'wailsjs/go/services/reportService.js'

export default {
  data() {
    return {
      report: {
        end_date: null,
      },
    }
  },
  created() {
    this.$watch(
      () => this.$route.params,
      () => {
        this.getProfitLossReport()
      },
      { immediate: true },
    )
  },
  beforeRouteUpdate: function(to, from, next) {
    this.report = this.getProfitLossReport();
    next();
  },
  methods: {
    async getProfitLossReport() {
      const { success, msg, data } = await GetProfitLossReport()
      if (!success) {
        $message.error(msg)
        return
      }
      this.report = data
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
                            <h6 class="m-0 font-weight-bold text-primary">Reports</h6>
                        </div>
                        <div class="card-body">
                            <div class="mb-3">
                                <div>
                                    <label>
                                      <h1>Profit and Loss</h1>
                                    </label>
                                </div>
                                <div>
                                    <label>
                                      <p>Start: {{ report.start_date }}</p>
                                    </label>
                                </div>
                                <div>
                                    <label>
                                      <p>End: {{ report.end_date }}</p>
                                    </label>
                                </div>
                                <div>
                                  <div>
                                    <label>
                                      <h2>Income</h2>
                                    </label>
                                  </div>
                                  <div v-for="account in report.income">
                                    <label>
                                      <p>Account: {{ account.name }}: <router-link :to="{ name: 'get-transactions-account', params: { accountid: account.id }}">{{ account.amount }}</router-link></p>
                                    </label>
                                  </div>
                                </div>
                                <div>
                                  <div>
                                    <label>
                                      <h2>Expenses</h2>
                                    </label>
                                  </div>
                                  <div v-for="account in report.expense">
                                    <label>
                                      <p>Account: <router-link :to="{ name: 'get-account', params: { id: account.id }}">{{ account.name }}</router-link>: <router-link :to="{ name: 'get-transactions-account', params: { accountid: account.id }}">{{ account.amount }}</router-link></p>
                                    </label>
                                  </div>
                                </div>
                                <div>
                                  <label>
                                    <p>Net: {{ report.net }}</p>
                                  </label>
                                </div>
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

    <!-- Scroll to Top Button-->
    <a class="scroll-to-top rounded" href="#page-top">
        <i class="fas fa-angle-up"></i>
    </a>
</template>
