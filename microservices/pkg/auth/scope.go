package authdomain

type Scope string

const (
	ScopeServerRead    Scope = "server:read"
	ScopeServerCreate  Scope = "server:create"
	ScopeServerUpdate  Scope = "server:update"
	ScopeServerDelete  Scope = "server:delete"
	ScopeServerImport  Scope = "server:import"
	ScopeServerExport  Scope = "server:export"
	ScopeMonitorRead   Scope = "monitor:read"
	ScopeMonitorReport Scope = "monitor:report"
)
