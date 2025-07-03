import {createRouter, createWebHashHistory} from 'vue-router'
import HomeView from '../views/HomeView.vue'
import LoginView from '../views/LoginView.vue'
import ConversationListView from '../views/ConversationListView.vue'
import ConversationDetailView from '../views/ConversationDetailView.vue'
import MyAccountView from '../views/MyAccountView.vue'
import NewDirectConversation from '../views/NewDirectConversation.vue'
import NewGroupConversation from '../views/NewGroupConversation.vue'
import LogoutView from '../views/LogoutView.vue'

const router = createRouter({
	history: createWebHashHistory(import.meta.env.BASE_URL),
	routes: [
		{path: '/', component: HomeView},
		{path: '/link1', component: HomeView},
		{path: '/link2', component: HomeView},
		{path: '/some/:id/link', component: HomeView},
		{path: '/session', component: LoginView},
		{path: '/conversations', component: ConversationListView},
		{path: '/conversations/:id', component: ConversationDetailView},
		{path: '/user/me', component: MyAccountView},
		{path: '/new-direct-conversation', component: NewDirectConversation},
		{path: '/new-group-conversation', component: NewGroupConversation},
		{path: '/logout', component: LogoutView}

	]
})

export default router
