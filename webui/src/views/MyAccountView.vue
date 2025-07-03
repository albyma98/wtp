<template>
<div class="container py-5">
    <div class="row justify-content-center">
        <div class="col-md-8">
            <h2 class="mb-4">Il mio profilo</h2>

            <div v-if="loading" class="text-muted mb-3">Caricamento profilo...</div>
            <div v-if="errormsg" class="alert alert-danger">{{ errormsg }}</div>

            <div v-if="photoUrl" class="mb-4">
                <!-- Foto profilo -->
                <div class="d-flex align-items-center gap-3 mb-3">
                    <img :src="getFullPhotoUrl(photoUrl)" alt="Foto profilo" class="rounded-circle border" style="width: 64px; height: 64px; object-fit: cover" />
                </div>

                <!-- Upload immagine -->
                <div class="mb-3">
                    <label class="form-label">Carica nuova foto profilo</label>
                    <input type="file" class="form-control" accept="image/*" @change="onPhotoSelected" />
                    <button class="btn btn-primary mt-2" :disabled="!selectedPhoto" @click="uploadPhoto">Carica</button>
                </div>

                <!-- Username -->
                <div class="mb-3">
                    <label class="form-label">Username</label>
                    <div class="d-flex align-items-center gap-2">
                        <span v-if="!editingUsername">{{ username }}</span>
                        <input v-else v-model="newUsername" class="form-control form-control-sm" style="width: auto" />
                        <button @click="toggleUsernameEdit" class="btn btn-sm btn-outline-primary">
                            {{ editingUsername ? 'Annulla' : 'Modifica' }}
                        </button>
                        <button v-if="editingUsername" @click="saveUsername" class="btn btn-sm btn-success">
                            Salva
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>
</template>

<script>
export default {
    data() {
        return {
            username: '',
            newUsername: '',
            editingUsername: false,
            photoUrl: '',
            selectedPhoto: null,
            loading: false,
            errormsg: null
        }
    },
    methods: {
        async fetchUser() {
            this.loading = true
            this.errormsg = null
            try {
                const res = await this.$axios.get('/user/me')
                this.username = res.data.username
                this.newUsername = res.data.username
                this.photoUrl = res.data.photoUrl ? __IMG_URL__ + res.data.photoUrl : __IMG_URL__ + '/default-avatar.png'
                console.log(this.username)
            } catch (err) {
                this.errormsg = 'Errore nel caricamento dati utente'
                console.error(err)
            }
            this.loading = false
        },
        toggleUsernameEdit() {
            this.editingUsername = !this.editingUsername
        },
        async saveUsername() {
            try {
                const res = await this.$axios.put(
                    '/user/me/username', { username: this.newUsername })
                this.username = res.data.username
                this.editingUsername = false
            } catch (err) {
                if (err.response?.status === 409) {
                    alert('Username già in uso')
                } else {
                    console.error('Errore durante aggiornamento username:', err)
                }
            }
        },
        onPhotoSelected(event) {
            const file = event.target.files[0]
            this.selectedPhoto = file || null
        },
        async uploadPhoto() {
            if (!this.selectedPhoto) return

            const formData = new FormData()
            formData.append('photo', this.selectedPhoto)

            try {
                const res = await this.$axios.put('/user/me/photo', formData)
                this.photoUrl = __IMG_URL__ + res.data.photoUrl
                this.selectedPhoto = null
            } catch (err) {
                console.error('Errore upload immagine:', err)
            }
        },
        getFullPhotoUrl(photoPath) {
            if (!photoPath) return
            return photoPath
        }
    },
    mounted() {
        this.fetchUser()
    }
}
</script>
