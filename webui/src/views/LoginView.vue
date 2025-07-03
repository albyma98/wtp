<template>
<div class="container py-5">
    <div class="row justify-content-center">
        <div class="col-md-6">
            <h2 class="mb-4">Login</h2>

            <ErrorMsg v-if="errormsg" :msg="errormsg" />

            <form @submit.prevent="handleLogin">
                <div class="mb-3">
                    <label class="form-label">Username</label>
                    <input v-model="username" type="text" class="form-control" required />
                </div>
                <button type="submit" class="btn btn-primary" :disabled="loading">
                    {{ loading ? 'Accesso...' : 'Login' }}
                </button>
            </form>
        </div>
    </div>
</div>
</template>

<script>
import { checkLoginStatus } from '@/composables/useAuth'

export default {
    data() {
        return {
            username: '',
            loading: false,
            errormsg: null
        }
    },
    methods: {
        async handleLogin() {
            this.loading = true
            this.errormsg = null
            try {
                const response = await this.$axios.post('/session', {
                    username: this.username
                })

                localStorage.setItem('authUUID', response.data.uuid)
                checkLoginStatus()
                this.$router.push('/conversations') // reindirizza dopo il login
            } catch (err) {
                this.errormsg = err.response?.data?.message || 'Errore durante il login'
            }
            this.loading = false
        }
    }
}
</script>
