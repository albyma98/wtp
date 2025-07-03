<template>
<div class="container py-5">
    <div class="row justify-content-center">
        <div class="col-md-8">
            <h2 v-if="loading" class="mb-4">Logout in corso...</h2>
        </div>
    </div>
</div>
</template>

<script>
import { checkLoginStatus } from '@/composables/useAuth'
export default {
    data: function () {
        return {
            loading: true
        }
    },
    methods: {
        async logout() {
            localStorage.removeItem('authUUID')
            checkLoginStatus()
            this.$router.push(`/session`)
        },
    },
    mounted() {
        this.logout()
    }
}
</script>
