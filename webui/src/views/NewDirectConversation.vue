<template>
<div class="container py-5">
    <div class="row justify-content-center">
        <div class="col-md-8">
            <h2 class="mb-4">Cerca Utenti</h2>

            <div class="mb-3">
                <input v-model="searchText" @keyup.enter="searchUsers(searchText)" placeholder="Cerca username" class="form-control" />
                <button class="btn btn-primary mt-2" @click="searchUsers(searchText)">
                    Cerca
                </button>
            </div>

            <div v-if="loading" class="text-muted mb-3">Caricamento...</div>
            <div v-if="errormsg" class="alert alert-danger">{{ errormsg }}</div>

            <UserItem v-for="u in users" :key="u.uuid" :user="formatUser(u)" @open="openConversation" />
        </div>
    </div>
</div>
</template>

<script>
import UserItem from '../components/UserItem.vue'

export default {
    components: {
        UserItem
    },
    data() {
        return {
            users: [],
            conversations: [],
            searchText: '',
            loading: false,
            errormsg: null
        }
    },
    methods: {
        async searchUsers(query) {
            this.loading = true
            this.errormsg = null
            try {
                const res = await this.$axios.get(`/user?search=${encodeURIComponent(query)}`)
                this.users = res.data
            } catch (err) {
                this.errormsg = err.response?.data?.error || 'Errore nel caricamento degli utenti'
            }
            this.loading = false
        },
        async openConversation(uuid) {
            const user = this.users.find(u => u.uuid === uuid)
            if (user) {
                if (Array.isArray(this.conversations)) {
                    const existing = this.conversations.find(
                        c => c.isDirect && c.peerUsername === user.username
                    )
                    if (existing) {
                        this.$router.push(`/conversations/${existing.id}`)
                        return
                    }
                }
            }
            try {
                const response = await this.$axios.post('/conversations', {
                    isDirect: true,
                    groupName: null,
                    groupPhoto: null,
                    members: [uuid]
                })
                const newConversationId = response.data.ID
                this.$router.push(`/conversations/${newConversationId}`)
            } catch (error) {
                console.error(
                    'Errore nella creazione della conversazione:',
                    error.response?.data || error.message
                )
            }
        },
        async fetchConversations() {
            try {
                const res = await this.$axios.get('/conversations')
                this.conversations = res.data.conversations
            } catch (err) {
                console.error('Errore caricamento conversazioni:', err.response?.data || err.message)
            }
        },
        formatUser(u) {
            return {
                uuid: u.uuid,
                username: u.username,
                photo_url: u.photoUrl
            }
        }
    },
    mounted() {
        this.fetchConversations()
    }
}
</script>
