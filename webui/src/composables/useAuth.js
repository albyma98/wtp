import { ref } from 'vue'

export const isLoggedIn = ref(!!localStorage.getItem('authUUID'))

export const checkLoginStatus = () => {
	isLoggedIn.value = !!localStorage.getItem('authUUID')
}
