<script setup>
import { RouterLink, RouterView } from 'vue-router'
</script>
<script>
import { isLoggedIn, checkLoginStatus } from './composables/useAuth'

export default {
  mounted() {
    checkLoginStatus()
    window.addEventListener('storage', checkLoginStatus)
  }
}
</script>

<template>
	<header class="navbar navbar-dark sticky-top bg-dark flex-md-nowrap p-0 shadow">
		<a class="navbar-brand col-md-3 col-lg-2 me-0 px-3 fs-6" href="#/">Example App</a>
		<button class="navbar-toggler position-absolute d-md-none collapsed" type="button" data-bs-toggle="collapse" data-bs-target="#sidebarMenu" aria-controls="sidebarMenu" aria-expanded="false" aria-label="Toggle navigation">
			<span class="navbar-toggler-icon"></span>
		</button>
	</header>

	<div class="container-fluid">
		<div class="row">
			<nav id="sidebarMenu" class="col-md-3 col-lg-2 d-md-block bg-light sidebar collapse">
				<div class="position-sticky pt-3 sidebar-sticky">
					<h6 class="sidebar-heading d-flex justify-content-between align-items-center px-3 mt-4 mb-1 text-muted text-uppercase">
						<span>General</span>
					</h6>
					<ul class="nav flex-column">
						<li class="nav-item">
							<RouterLink :to="isLoggedIn ? '/user/me' : '/session'" class="nav-link">
								<svg class="feather"><use href="/feather-sprite-v4.29.0.svg#home"/></svg>
								{{isLoggedIn ? 'Il mio Account' : 'Login'}} 
							</RouterLink>
						</li>
						<li v-if="isLoggedIn" class="nav-item">
							<RouterLink to="/conversations" class="nav-link">
								<svg class="feather"><use href="/feather-sprite-v4.29.0.svg#layout"/></svg>
								Chat
							</RouterLink>
						</li>
						<li v-if="isLoggedIn" class="nav-item">
							<RouterLink to="/new-direct-conversation" class="nav-link">
								<svg class="feather"><use href="/feather-sprite-v4.29.0.svg#plus-circle"/></svg>
								Inizia Chat
							</RouterLink>
						</li>
						<li v-if="isLoggedIn" class="nav-item">
							<RouterLink to="/new-group-conversation" class="nav-link">
								<svg class="feather"><use href="/feather-sprite-v4.29.0.svg#plus-circle"/></svg>
								Crea Nuovo Gruppo
							</RouterLink>
						</li>
					</ul>

					<h6 v-if="isLoggedIn" class="sidebar-heading d-flex justify-content-between align-items-center px-3 mt-4 mb-1 text-muted text-uppercase">
						<span>Secondary menu</span>
					</h6>
					<ul v-if="isLoggedIn" class="nav flex-column">
						<li class="nav-item">
							<RouterLink to="/logout" class="nav-link">
								<svg class="feather"><use href="/feather-sprite-v4.29.0.svg#log-out"/></svg>
								LogOut
							</RouterLink>
						</li>
					</ul>
				</div>
			</nav>

			<main class="col-md-9 ms-sm-auto col-lg-10 px-md-4">
				<RouterView />
			</main>
		</div>
	</div>
</template>

<style>
</style>
