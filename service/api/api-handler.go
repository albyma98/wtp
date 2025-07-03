package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	// Register routes
	rt.router.GET("/", rt.getHelloWorld)
	rt.router.GET("/context", rt.wrap(rt.getContextReply))

	// Special routes
	rt.router.GET("/liveness", rt.liveness)

	// Auth
	rt.router.POST("/session", rt.doLogin)

	// User
	rt.router.GET("/user/me", rt.wrap(rt.getMyUserInfo))
	rt.router.PUT("/user/me/username", rt.wrap(rt.setMyUserName))
	rt.router.PUT("/user/me/photo", rt.wrap(rt.setMyPhoto))
	rt.router.GET("/user/all", rt.wrap(rt.getAllUsers))
	rt.router.GET("/user", rt.wrap(rt.searchUsers))

	// Conversation
	rt.router.GET("/conversations", rt.wrap(rt.getMyConversations))
	rt.router.POST("/conversations", rt.wrap(rt.createConversation))
	rt.router.GET("/conversations/:id", rt.wrap(rt.getConversation))
	rt.router.PUT("/conversations/:id/name", rt.wrap(rt.setGroupName))
	rt.router.PUT("/conversations/:id/photo", rt.wrap(rt.setGroupPhoto))
	rt.router.POST("/conversations/:id/members", rt.wrap(rt.addToGroup))
	rt.router.GET("/conversations/:id/members", rt.wrap(rt.getGroupMembers))
	rt.router.DELETE("/conversations/:id/members/me", rt.wrap(rt.leaveGroup))

	// Message
	rt.router.POST("/conversations/:id/messages", rt.wrap(rt.sendMessage))
	rt.router.DELETE("/messages/:id", rt.wrap(rt.deleteMessage))
	rt.router.POST("/messages/:id/forward", rt.wrap(rt.forwardMessage))

	// Reaction
	rt.router.POST("/messages/:id/reactions", rt.wrap(rt.commentMessage))
	rt.router.DELETE("/messages/:id/reactions/me", rt.wrap(rt.uncommentMessage))

	// Status
	rt.router.PUT("/messages/:id/status", rt.wrap(rt.updateMessageStatus))

	return rt.router
}
