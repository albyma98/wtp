import axios from "axios";

const instance = axios.create({
	baseURL: __API_URL__,
	timeout: 1000 * 5
});

// Interceptor per aggiungere il token in automatico
instance.interceptors.request.use(config => {
  const UUID = localStorage.getItem('authUUID')
  if (UUID) {
    config.headers.Authorization = `Bearer ${UUID}`
  }
  return config
})


// Interceptor per gestire gli errori globalmente
instance.interceptors.response.use(
  response => response,
  error => {
    const status = error.response?.status

    // Errore di autenticazione → rimanda al login
    if (status === 401) {
      localStorage.removeItem('authUUID')
      router.push('/session')
    }

    // Altri errori li rilanciamo per essere gestiti dove serve
    return Promise.reject(error)
  }
)

instance.interceptors.request.use(config => {
  console.log("Axios REQUEST:", config)
  return config
})

// ✅ Interceptor risposte
instance.interceptors.response.use(
  response => {
    console.log("✅ Axios RESPONSE:", response)
    return response
  },
  error => {
    console.error("❌ Axios ERROR RESPONSE:", error.response || error)
    return Promise.reject(error)
  }
)

export default instance;
