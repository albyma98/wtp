<template>
<div class="container py-5">
    <div class="row justify-content-center">
        <div class="col-md-8">
            <h2 class="mb-4">Le tue conversazioni</h2>

            <div v-if="loading" class="text-muted mb-3">Caricamento...</div>
            <div v-if="errormsg" class="alert alert-danger">{{ errormsg }}</div>

            <ConversationItem v-for="conv in conversations" :key="conv.id" :conversation="formatConversation(conv)" @open="openConversation" />
        </div>
    </div>
</div>
</template>

<script>
import ConversationItem from '@/components/ConversationItem.vue'

export default {
    components: {
        ConversationItem
    },
    data() {
        return {
            conversations: [],
            loading: false,
            errormsg: null
        }
    },
    methods: {
        async fetchConversations(showLoading = false) {
            if (showLoading) this.loading = true;
            this.errormsg = null
            try {
                const res = await this.$axios.get('/conversations')
                console.log("RESPONSE COMPLETA:", res.data)
                this.conversations = res.data.conversations
            } catch (err) {
                this.errormsg = err.response?.data?.error || 'Errore nel caricamento delle conversazioni'
            }
            if (showLoading) this.loading = false;
        },
        openConversation(id) {
            this.$router.push(`/conversations/${id}`)
        },
        formatConversation(conv) {
            return {
                id: conv.id,
                title: conv.isDirect ? conv.peerUsername || 'Utente' : conv.groupName || 'Gruppo',
                photo_url: conv.isDirect ? conv.peerPhoto : conv.groupPhoto,
                last_message: conv.lastMessageText || conv.lastMessageType ? {
                    text: conv.lastMessageText,
                    type: conv.lastMessageType
                } : null,
                lastTimestamp: conv.timestampLastMessage
            }
        }
    },
    mounted() {
        this.fetchConversations(true);
        this.pollingInterval = setInterval(() => {
            this.fetchConversations();
        }, 4000);
    },
    beforeUnmount() {
        if (this.pollingInterval) {
            clearInterval(this.pollingInterval);
        }
    },
}
</script>
