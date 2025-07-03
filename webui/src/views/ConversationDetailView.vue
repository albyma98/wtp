<template>
<div class="container py-4">
    <div v-if="loading" class="text-muted">
        Caricamento conversazione...
    </div>
    <div v-else-if="errormsg" class="alert alert-danger">
        {{ errormsg }}
    </div>
    <div v-else>
        <!-- intestazione conversazione -->
        <div class="mb-4 flex items-center gap-3">
            <img :src="conversationPhoto || defaultPhoto" class="rounded-circle border" style="width: 64px; height: 64px; object-fit: cover" />
            <div>
                <h2 class="text-xl font-semibold">
                    {{ conversationTitle }}
                </h2>
                <div v-if="!conversation?.isDirect" class="mt-1 flex items-center gap-2">
                    <template v-if="!showPhotoInput">
                        <button class="btn btn-sm btn-outline-primary" @click="startPhotoChange">Cambia foto</button>
                    </template>
                    <template v-else>
                        <input type="file" accept="image/*" @change="onGroupPhotoSelected" ref="groupPhotoInput" class="form-control form-control-sm" />
                        <button class="btn btn-primary btn-sm" :disabled="!newGroupPhoto" @click="confirmGroupPhoto">Conferma</button>
                        <button class="btn btn-secondary btn-sm" @click="cancelGroupPhoto">Annulla</button>
                    </template>
                    <button class="btn btn-sm btn-outline-secondary" @click="fetchMembers">Lista membri</button>
                    <button class="btn btn-sm btn-outline-primary" @click="openAddMembers">Aggiungi membri</button>
                    <button class="btn btn-sm btn-outline-danger" @click="leaveGroup">Lascia gruppo</button>
                </div>
            </div>
        </div>

        <div v-if="membersModal" class="modal fade show d-block" tabindex="-1">
            <div class="modal-dialog">
                <div class="modal-content">
                    <div class="modal-header">
                        <h5 class="modal-title">Members</h5>
                        <button type="button" class="btn-close" @click="membersModal = false"></button>
                    </div>
                    <div class="modal-body">
                        <ul>
                            <li v-for="m in membersList" :key="m">{{ m }}</li>
                        </ul>
                    </div>
                </div>
            </div>
        </div>
        <div class="modal-backdrop fade show" v-if="membersModal"></div>
        <div v-if="addMembersModal" class="modal fade show d-block" tabindex="-1">
            <div class="modal-dialog">
                <div class="modal-content">
                    <div class="modal-header">
                        <h5 class="modal-title">Add Members</h5>
                        <button type="button" class="btn-close" @click="addMembersModal = false"></button>
                    </div>
                    <div class="modal-body">
                        <select multiple class="form-select mb-3" v-model="selectedAddMembers">
                            <option v-for="u in availableUsers" :key="u.uuid" :value="u.uuid">{{ u.username }}</option>
                        </select>
                        <button class="btn btn-primary me-2" @click="confirmAddMembers">Conferma</button>
                        <button class="btn btn-secondary" @click="addMembersModal = false">Annulla</button>
                    </div>
                </div>
            </div>
        </div>
        <div class="modal-backdrop fade show" v-if="addMembersModal"></div>
        <!-- messaggi -->
        <div class="space-y-3 mb-24">
            <MessageItem v-for="msg in messages" :key="msg.ID" :message="msg" :isMine="msg.UUIDSender === currentUserUUID" :usernameSender="msg.usernameSender" @delete="deleteMessage" @forwarded="addForwarded" @reply="replyTo = $event" />
        </div>

        <!-- barra invio messaggio -->
        <div class="fixed bottom-0 left-0 right-0 border-t bg-white py-2">
            <div class="container">
                <div v-if="replyTo" class="mb-2 p-2 bg-light rounded flex justify-between items-center">
                    <span class="text-sm">Risposta a: {{ replyTo.Content || 'Foto' }}</span>
                    <button class="btn-close" @click="replyTo = null"></button>
                </div>
                <div class="flex items-center gap-2">
                    <input v-model="newMessage" type="text" class="flex-1 form-control" placeholder="Scrivi un messaggio" />
                    <input type="file" accept="image/*" @change="handlePhoto" ref="photoInput" /><br />
                    <button class="btn btn-primary" @click="sendMessage">
                        Invia
                    </button>
                </div>
            </div>
        </div>
    </div>
</div>
</template>

<script>
import MessageItem from "@/components/MessageItem.vue";

export default {
    components: {
        MessageItem,
    },
    data() {
        return {
            conversation: null,
            messages: [],
            loading: false,
            errormsg: null,
            defaultPhoto: __IMG_URL__ + "/default-avatar.png",
            newMessage: "",
            photoDataUrl: null,
            showPhotoInput: false,
            newGroupPhoto: null,
            totMsg: null,
            seenQueue: [],
            processingSeen: false,
            membersModal: false,
            membersList: [],
            addMembersModal: false,
            availableUsers: [],
            selectedAddMembers: [],
            replyTo: null,
        };
    },
    computed: {
        conversationTitle() {
            if (!this.conversation) return "";
            return this.conversation.isDirect ?
                this.conversation.usernamePeer || "Utente" :
                this.conversation.groupName || "Gruppo";
        },
        conversationPhoto() {
            if (!this.conversation) return null;
            const photo = this.conversation.isDirect ?
                this.conversation.photoUrlPeer :
                this.conversation.groupPhoto;
            return photo ? __IMG_URL__ + photo : null;
        },
        currentUserUUID() {
            return localStorage.getItem("authUUID");
        },
    },
    methods: {
        async fetchConversation(showLoading = false) {
            if (showLoading) this.loading = true;
            this.errormsg = null;
            const id = this.$route.params.id;
            try {
                const res = await this.$axios.get(`/conversations/${id}`);
                this.conversation = res.data.conversationDetail;

                const myId = localStorage.getItem("authUUID");

                const recipientCount = (this.conversation.numberMembers || 1) - 1;
                const newMsgs = res.data.messages.map((m) => {
                    const delivered = m.delivered || [];
                    const seen = m.seen || [];
                    let status = null;
                    if (m.UUIDSender === myId) {
                        if (seen.length === recipientCount) {
                            status = "read";
                        } else if (delivered.length === recipientCount) {
                            status = "delivered";
                        }
                    } else {
                        if (seen.includes(myId)) {
                            status = "read";
                        } else if (delivered.includes(myId)) {
                            status = "delivered";
                        }
                    }
                    return {
                        ...m,
                        IDForwardedFrom: m.IDForwardedFrom ?? m.idForwardedFrom ?? null,
                        status,
                    };
                });
                this.messages = newMsgs;
                this.markMessagesAsRead();
            } catch (err) {
                const msg = err.response?.data?.error;
                if (msg) {
                    this.errormsg = msg;
                }
            }
            if (showLoading) this.loading = false;
        },
        formatDate(timestamp) {
            const date = new Date(timestamp);
            return date.toLocaleString([], {
                dateStyle: "short",
                timeStyle: "short",
            });
        },
        handlePhoto(event) {
            const file = event.target.files[0];
            if (!file) {
                this.photoDataUrl = null;
                return;
            }
            const reader = new FileReader();
            reader.onload = (e) => {
                this.photoDataUrl = e.target.result;
            };
            reader.readAsDataURL(file);
        },
        startPhotoChange() {
            this.showPhotoInput = true;
        },
        onGroupPhotoSelected(event) {
            const file = event.target.files[0];
            this.newGroupPhoto = file || null;
        },
        cancelGroupPhoto() {
            this.showPhotoInput = false;
            this.newGroupPhoto = null;
            if (this.$refs.groupPhotoInput) this.$refs.groupPhotoInput.value = null;
        },
        async confirmGroupPhoto() {
            if (!this.newGroupPhoto) return;
            const id = this.$route.params.id;
            const form = new FormData();
            form.append('photo', this.newGroupPhoto);
            try {
                const res = await this.$axios.put(`/conversations/${id}/photo`, form);
                this.conversation.groupPhoto = res.data.groupPhoto;
                this.cancelGroupPhoto();
            } catch (err) {
                this.errormsg = err.response?.data?.error || 'Errore aggiornamento foto';
            }
        },
        async fetchMembers() {
            const id = this.$route.params.id;
            try {
                const res = await this.$axios.get(`/conversations/${id}/members`);
                this.membersList = res.data.members || [];
                this.membersModal = true;
            } catch (err) {
                this.errormsg = err.response?.data?.error || 'Errore recupero membri';
            }
        },
        async openAddMembers() {
            try {
                const res = await this.$axios.get('/user/all');
                this.availableUsers = res.data.users || res.data;
                this.selectedAddMembers = [];
                this.addMembersModal = true;
            } catch (err) {
                this.errormsg = err.response?.data?.error || 'Errore caricamento utenti';
            }
        },
        async confirmAddMembers() {
            if (!this.selectedAddMembers.length) {
                this.addMembersModal = false;
                return;
            }
            const id = this.$route.params.id;
            try {
                await this.$axios.post(`/conversations/${id}/members`, { members: this.selectedAddMembers });
                await this.fetchMembers();
                this.addMembersModal = false;
                this.selectedAddMembers = [];
            } catch (err) {
                this.errormsg = err.response?.data?.error || 'Errore aggiunta membri';
            }
        },
        async leaveGroup() {
            const id = this.$route.params.id;
            try {
                await this.$axios.delete(`/conversations/${id}/members/me`);
                this.$router.push('/conversations');
            } catch (err) {
                this.errormsg = err.response?.data?.error || 'Errore uscita dal gruppo';
            }
        },
        async sendMessage() {
            const id = this.$route.params.id;
            if (!this.newMessage && !this.photoDataUrl) return;
            const body = {
                type: this.photoDataUrl ? "photo" : "text",
                content: this.newMessage || null,
                mediaUrl: this.photoDataUrl || null,
                idRepliesTo: this.replyTo?.ID || null,
            };
            try {
                const res = await this.$axios.post(
                    `/conversations/${id}/messages`,
                    body,
                );
                const msg = { ...res.data, status: "sent" };
                this.messages.push(msg);
                this.newMessage = "";
                this.photoDataUrl = null;
                this.replyTo = null;
                if (this.$refs.photoInput) this.$refs.photoInput.value = null;
            } catch (err) {
                this.errormsg =
                    err.response?.data?.error || "Errore invio messaggio";
            }
        },
        async deleteMessage(idMsg) {
            try {
                await this.$axios.delete(`/messages/${idMsg}`);
                this.messages = this.messages.filter((m) => m.ID !== idMsg);
            } catch (err) {
                this.errormsg =
                    err.response?.data?.error ||
                    "Errore eliminazione messaggio";
            }
        },
        addForwarded(msg) {
            const normalized = {
                ...msg,
                IDForwardedFrom: msg.IDForwardedFrom ?? msg.idForwardedFrom ?? null,
                status: "sent",
            };
            this.messages.push(normalized);
        },
        markMessagesAsRead() {
            const myId = this.currentUserUUID;
            this.messages.forEach((m) => {
                if (m.UUIDSender === myId || m.status === "read") return;
                if (!this.seenQueue.includes(m.ID)) {
                    this.seenQueue.push(m.ID);
                }
            });
            this.processSeenQueue();
        },
        async processSeenQueue() {
            if (this.processingSeen) return;
            this.processingSeen = true;
            while (this.seenQueue.length) {
                const id = this.seenQueue.shift();
                try {
                    await this.$axios.put(`/messages/${id}/status`, { seen: true });
                    const msg = this.messages.find((m) => m.ID === id);
                    if (msg) msg.status = "read";
                } catch (e) {
                    console.error(e);
                }
                await new Promise((r) => setTimeout(r, 200));
            }
            this.processingSeen = false;
        },
    },
    mounted() {
        this.fetchConversation(true);

        this.pollingInterval = setInterval(() => {
            this.fetchConversation();
        }, 4000);
    },
    beforeUnmount() {
        if (this.pollingInterval) {
            clearInterval(this.pollingInterval);
        }
    },
};
</script>
