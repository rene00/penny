<template>
    <!-- Page Wrapper -->
    <div id="wrapper">
        <Sidebar/>
    </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import Sidebar from './components/Sidebar.vue'

export default defineComponent({
    components: {
        Sidebar,
    },
})
</script>


