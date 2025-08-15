package handler

// HTTP Header constants
const (
	HtmxRequestHeader = "true"
	HtmxRequest       = "HX-Request"
	HtmxRedirect      = "HX-Redirect"
	HtmxTrigger       = "HX-Trigger"
	HtmxTarget        = "HX-Target"
	HtmxSwap          = "HX-Swap"

	ContentTypeJSON = "application/json"
)

// Route constants
const (
	RouteHome     = "/"
	RouteLogin    = "/auth/login"
	RouteRegister = "/auth/register"
	RouteLogout   = "/auth/logout"
	RouteProfile  = "/profile"
)

// Response messages
const (
	MsgLoginSuccess    = "Login successful"
	MsgLogoutSuccess   = "Logout successful"
	MsgRegisterSuccess = "Registration successful"
)
