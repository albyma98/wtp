<template>
<div class="relative">
    <div :style="{
    position: 'relative',
    maxWidth: '100%',
    padding: '12px',
    borderRadius: '16px',
    boxShadow: '0 2px 6px rgba(0, 0, 0, 0.1)',
    cursor: 'pointer',
    backgroundColor: isMine ? '#d0eaff' : '#d6f5d6',
    alignSelf: isMine ? 'flex-end' : 'flex-start',
    marginBottom: '10px'
}" @click="toggleMenu">
        <div v-if="!isMine" class="text-xs font-semibold mb-1">
            <b>{{ usernameSender }}</b>
        </div>
        <div v-if="message.IDForwardedFrom" class="text-xxs text-gray-100 mb-1">
            inoltrato
        </div>
        <div v-if="message.replyToMessage" class="reply-preview">
            <template v-if="message.replyToMessage.type === 'text'">
                {{ message.replyToMessage.content }}
            </template>
            <template v-else>
                <img :src="message.replyToMessage.mediaUrl" class="reply-img" />
            </template>
        </div>
        <div class="text-sm text-gray-800 break-words whitespace-pre-wrap">{{ message.Content }}</div>
        <div v-if="message.MediaUrl" class="mt-2">
            <img :src="message.MediaUrl" class="rounded-lg" :style="{
    width: '360px',
    height: '360px',
    objectFit: 'contain'
}" />
        </div>
        <div v-if="reactions.length" class="flex gap-1 mt-1 text-sm">
            <span v-for="(r, i) in reactions" :key="i" class="bg-gray-100 rounded px-1">{{ r.emoji }} {{ r.username }}</span>
        </div>
        <div class="flex justify-end items-center gap-1 text-xs text-gray-500 mt-1">
            <span>{{ formatTime(message.Timestamp) }}</span>
            <span v-if="isMine">
                <template v-if="message.status === 'read'">✔✔</template>
                <template v-else-if="message.status === 'delivered'">✔</template>
            </span>
        </div>
    </div>

    <div v-if="showMenu" :style="{
    position: 'relative',
    maxWidth: '25%',
    padding: '12px',
    borderRadius: '16px',
    boxShadow: '0 2px 6px rgba(0, 0, 0, 0.1)',
    cursor: 'pointer',
    backgroundColor: isMine ? '#d0eaff' : '#d6f5d6',
    alignSelf: isMine ? 'flex-end' : 'flex-start',
    marginBottom: '10px'
}">
        <div class="px-2 py-1 hover:bg-gray-100 cursor-pointer" @click.stop="showPicker = !showPicker">
            Aggiungi Reazione
        </div>
        <div class="px-2 py-1 hover:bg-gray-100 cursor-pointer" @click.stop="emitReply">
            Rispondi
        </div>
        <div class="px-2 py-1 hover:bg-gray-100 cursor-pointer" @click.stop="toggleForwardSelector">
            Inoltra messaggio
        </div>
        <div v-if="isMine" class="px-2 py-1 hover:bg-gray-100 cursor-pointer" @click.stop="emitDelete">
            Elimina messaggio
        </div>
    </div>

    <div v-if="showPicker" :style="{
    position: 'relative',
    maxWidth: '25%',
    padding: '12px',
    borderRadius: '16px',
    boxShadow: '0 2px 6px rgba(0, 0, 0, 0.1)',
    cursor: 'pointer',
    backgroundColor: isMine ? '#d0eaff' : '#d6f5d6',
    alignSelf: isMine ? 'flex-end' : 'flex-start',
    marginBottom: '10px'
}">
        <span v-for="e in emojis" :key="e" class="cursor-pointer" @click.stop="selectEmoji(e)">{{ e }}</span>
    </div>

    <div v-if="showForwardSelector" :style="{
    position: 'relative',
    maxWidth: '25%',
    padding: '12px',
    borderRadius: '16px',
    boxShadow: '0 2px 6px rgba(0, 0, 0, 0.1)',
    cursor: 'pointer',
    backgroundColor: isMine ? '#d0eaff' : '#d6f5d6',
    alignSelf: isMine ? 'flex-end' : 'flex-start',
    marginBottom: '10px'
}">
        <select v-model="selectedConversationId" :disabled="!!selectedUserUUID" class="form-select mb-2">
            <option selected disabled value="">Seleziona gruppo</option>
            <option v-for="g in availableGroups" :key="g.id" :value="g.id">
                {{ g.groupName || 'Gruppo' }}
            </option>
        </select>
        <select v-model="selectedUserUUID" :disabled="!!selectedConversationId" class="form-select mb-2">
            <option selected disabled value="">Seleziona utente</option>
            <option v-for="u in availableUsers" :key="u.uuid" :value="u.uuid">
                {{ u.username }}
            </option>
        </select>
        <button class="btn btn-primary" @click.stop="forwardMessage">Conferma</button>
        <button class="btn btn-secondary" @click.stop="cancelForward">Annulla</button>
    </div>

    <div v-if="myReaction" :style="{
    position: 'relative',
    maxWidth: '10%',
    padding: '12px',
    borderRadius: '16px',
    boxShadow: '0 2px 6px rgba(0, 0, 0, 0.1)',
    cursor: 'pointer',
    backgroundColor: isMine ? '#d0eaff' : '#d6f5d6',
    alignSelf: isMine ? 'flex-end' : 'flex-start',
    marginBottom: '10px'
}" @click.stop="removeReaction">
        {{ myReaction.emoji }} ✖
    </div>
</div>
</template>

<script>
import { useRoute } from 'vue-router'
import axios from '@/services/axios.js'

export default {
    props: {
        message: Object,
        isMine: Boolean,
        usernameSender: String
    },
    emits: ['delete', 'forwarded', 'reply'],
    data() {
        return {
            showMenu: false,
            showPicker: false,
            showForwardSelector: false,
            availableGroups: [],
            availableUsers: [],
            selectedConversationId: null,
            selectedUserUUID: null,
            emojis: ['👍', '❤️', '😂'],
            myReaction: null,
            reactions: [...this.message.reactions || []]
        }
    },
    watch: {
        'message.reactions': {
            immediate: true,
            handler(newReactions) {
                this.reactions = [...newReactions || []]
                this.myReaction = this.reactions.find(r => r.uuidUser === this.currentUserUUID) || null
            }
        }
    },
    computed: {
        route() {
            return useRoute()
        },
        currentUserUUID() {
            return localStorage.getItem('authUUID')
        }
    },
    created() {
        this.myReaction = this.reactions.find(r => r.uuidUser === this.currentUserUUID) || null
    },
    methods: {
        formatTime(timestamp) {
            const date = new Date(timestamp)
            return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
        },
        toggleMenu() {
            this.showMenu = !this.showMenu
            if (!this.showMenu) this.showPicker = false
            this.showForwardSelector = false
        },
        async toggleForwardSelector() {
            if (this.showForwardSelector) {
                this.showForwardSelector = false
                return
            }
            this.showForwardSelector = true
            this.showMenu = false
            this.showPicker = false
            if (!this.availableGroups.length || !this.availableUsers.length) {
                try {
                    const [convRes, userRes] = await Promise.all([
                        axios.get('/conversations'),
                        axios.get('/user/all')
                    ])
                    this.availableGroups = convRes.data.conversations.filter(c => !c.isDirect)
                    const users = userRes.data.users || userRes.data
                    this.availableUsers = users.filter(u => u.uuid !== this.currentUserUUID)
                } catch (err) {
                    console.error('Errore caricamento dati inoltro', err)
                    this.showForwardSelector = false
                }
            }
        },
        async selectEmoji(e) {
            try {
                const res = await axios.post(`/messages/${this.message.ID}/reactions`, { emoji: e })
                this.reactions.push(res.data)
                this.myReaction = res.data
            } catch (err) {
                console.error('Errore aggiunta reazione', err)
            }
            this.showPicker = false
            this.showMenu = false
        },
        async removeReaction() {
            try {
                await axios.delete(`/messages/${this.message.ID}/reactions/me`)
                const idx = this.reactions.findIndex(r => r.uuidUser === this.currentUserUUID)
                if (idx !== -1) this.reactions.splice(idx, 1)
                this.myReaction = null
            } catch (err) {
                console.error('Errore rimozione reazione', err)
            }
        },
        async forwardMessage() {
            let idConv = this.selectedConversationId ? parseInt(this.selectedConversationId) : null
            const uuidUser = this.selectedUserUUID

            if (!idConv && !uuidUser) {
                this.showForwardSelector = false
                this.selectedConversationId = null
                this.selectedUserUUID = null
                return
            }

            try {

                if (!idConv && uuidUser) {
                    const convList = await axios.get('/conversations')
                    const username = (this.availableUsers.find(u => u.uuid === uuidUser) || {}).username
                    const existing = convList.data.conversations.find(c => c.isDirect && (c.peerUsername === username || c.usernamePeer === username))
                    if (existing) {
                        idConv = existing.id
                    } else {
                        const cRes = await axios.post('/conversations', {
                            isDirect: true,
                            groupName: null,
                            groupPhoto: null,
                            members: [uuidUser]
                        })
                        idConv = cRes.data.ID || cRes.data.id || cRes.data
                    }
                }

                const res = await axios.post(`/messages/${this.message.ID}/forward`, {
                    idConversation: idConv
                })
                if (parseInt(this.route.params.id) === idConv) {
                    this.$emit('forwarded', res.data)
                }
            } catch (err) {
                console.error('Errore inoltro messaggio', err)
            }
            this.showForwardSelector = false
            this.showMenu = false
            this.selectedConversationId = null
            this.selectedUserUUID = null
        },
        cancelForward() {
            this.showForwardSelector = false
            this.showMenu = false
            this.selectedConversationId = null
            this.selectedUserUUID = null
        },
        formatConvName(conv) {
            return conv.isDirect ?
                conv.peerUsername || 'Utente' :
                conv.groupName || 'Gruppo'
        },
        emitReply() {
            this.$emit('reply', this.message)
            this.showMenu = false
        },
        emitDelete() {
            this.$emit('delete', this.message.ID)
            this.showMenu = false
        }
    }
}
</script>

<style scoped>
.reply-preview {
    font-size: 0.75rem;
    border-left: 2px solid #bbb;
    padding-left: 4px;
    margin-bottom: 4px;
    color: #555;
}

.reply-img {
    width: 40px;
    height: 40px;
    object-fit: cover;
    border-radius: 4px;
}
</style>
