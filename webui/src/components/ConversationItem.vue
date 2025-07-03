<template>
<div class="flex items-center mb-4 p-2 bg-white rounded-lg shadow cursor-pointer hover:bg-gray-100" @click="$emit('open', conversation.id)">
    <img :src="computedPhotoUrl" alt="Avatar" class="rounded-circle border" style="width: 64px; height: 64px; object-fit: cover" />
    <div class="flex-1">
        <div class="font-bold"><b>{{ conversation.title }}</b></div>
        <div class="text-gray-500 text-sm truncate">
            <template v-if="conversation.last_message?.type === 'photo'">
                <svg class="feather">
                    <use href="/feather-sprite-v4.29.0.svg#image" />
                </svg>
            </template>
            {{
          conversation.last_message?.text ||
            (conversation.last_message?.type === 'photo' ? 'Photo' : 'Nessun messaggio')
        }}
        </div>
        <span v-if="conversation.last_message">
            {{ formatTime(conversation.lastTimestamp) }}
        </span>
    </div>
</div>
</template>

<script>
export default {
    props: {
        conversation: {
            type: Object,
            required: true
        }
    },
    methods: {
        formatTime(timestamp) {
            const date = new Date(timestamp)
            return date.toLocaleString([], {
                dateStyle: 'short',
                timeStyle: 'short'
            })
        }
    },
    computed: {
        computedPhotoUrl() {
            return this.conversation.photo_url ?
                __IMG_URL__ + this.conversation.photo_url :
                __IMG_URL__ + 'default-avatar.png'
        }
    }
}
</script>
