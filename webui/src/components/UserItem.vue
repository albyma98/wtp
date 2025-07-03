<template>
<div class="flex items-center mb-4 p-2 bg-white rounded-lg shadow cursor-pointer hover:bg-gray-100" @click="$emit('open', user.uuid)">
    <img :src="computedPhotoUrl" alt="Avatar" class="rounded-circle border" style="width: 64px; height: 64px; object-fit: cover" />
    <div class="flex-1">
        <div class="font-bold"><b>{{ user.username }}</b></div>
    </div>
</div>
</template>

<script>
export default {
    props: {
        user: {
            type: Object,
            required: true
        }
    },
    computed: {
        computedPhotoUrl() {
            return this.user.photo_url ?
                __IMG_URL__ + this.user.photo_url :
                __IMG_URL__ + 'default-avatar.png'
        }
    }
}
</script>
