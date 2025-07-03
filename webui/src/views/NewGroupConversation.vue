<template>
<div class="container py-5">
    <div class="row justify-content-center">
        <div class="col-md-8">
            <h2 class="mb-4">Crea nuovo gruppo</h2>

            <div v-if="loading" class="text-muted mb-3">Caricamento...</div>
            <div v-if="errormsg" class="alert alert-danger">{{ errormsg }}</div>

            <div class="mb-3">
                <label class="form-label">Nome Gruppo</label>
                <input v-model="groupName" class="form-control" />
            </div>

            <div class="mb-3">
                <label class="form-label">Foto</label>
                <input type="file" class="form-control" accept="image/*" @change="handleGroupPhotoUpload" />
            </div>

            <div class="mb-3">
                <label class="form-label">Membri</label>
                <div v-for="u in users" :key="u.uuid" class="form-check">
                    <input class="form-check-input" type="checkbox" :id="u.uuid" :value="u.uuid" v-model="selectedMembers" />
                    <label class="form-check-label" :for="u.uuid">{{ u.username }}</label>
                </div>
            </div>

            <button class="btn btn-primary" @click="createConversation">Crea Gruppo</button>
        </div>
    </div>
</div>
</template>

<script>
export default {
    data() {
        return {
            users: [],
            selectedMembers: [],
            groupName: '',
            groupPhotoFile: null,
            loading: false,
            errormsg: null
        }
    },
    methods: {
        async fetchUsers() {
            this.loading = true
            this.errormsg = null
            try {
                const res = await this.$axios.get('/user/all')
                this.users = res.data
            } catch (err) {
                this.errormsg = err.response?.data?.error || 'Errore nel caricamento degli utenti'
            }
            this.loading = false
        },
        handleGroupPhotoUpload(event) {
            const file = event.target.files[0]
            this.groupPhotoFile = file || null
        },
        async createConversation() {
            if (!this.selectedMembers.length) {
                this.errormsg = 'Seleziona almeno un membro'
                return
            }
            this.loading = true
            this.errormsg = null
            try {
                const res = await this.$axios.post('/conversations', {
                    isDirect: false,
                    groupName: this.groupName || null,
                    groupPhoto: null,
                    members: this.selectedMembers
                })
                const id = res.data.ID || res.data
                if (this.groupPhotoFile) {
                    const form = new FormData()
                    form.append('photo', this.groupPhotoFile)
                    await this.$axios.put(`/conversations/${id}/photo`, form)
                }
                this.$router.push(`/conversations/${id}`)
            } catch (err) {
                this.errormsg = err.response?.data?.error || 'Errore creazione gruppo'
            }
            this.loading = false
        }
    },
    mounted() {
        this.fetchUsers()
    }
}
</script>
