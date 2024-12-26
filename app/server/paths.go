package server

const (
	RouteHome         string = "/"
	RouteAuth         string = "/auth"
	RouteAuthRequired string = "/auth-required"
	RouteAudit        string = "/audit"
	RouteStatus       string = "/status"
)

var pages = map[string]string{
	"home":         RouteHome,
	"auth":         RouteAuth,
	"authRequired": RouteAuthRequired,
	"audit":        RouteAudit,
	"status":       RouteStatus,
}
