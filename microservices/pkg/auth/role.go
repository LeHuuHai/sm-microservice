package authdomain

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
	RoleGuest Role = "guest"
)

func (r Role) String() string {
	return string(r)
}

func (r Role) Scopes() []Scope {
	switch r {
	case RoleAdmin:
		return []Scope{
			ScopeServerRead,
			ScopeServerCreate,
			ScopeServerUpdate,
			ScopeServerDelete,
			ScopeServerImport,
			ScopeServerExport,
			ScopeMonitorRead,
			ScopeMonitorReport,
		}

	case RoleUser:
		return []Scope{
			ScopeServerRead,
			ScopeServerCreate,
			ScopeServerUpdate,
			ScopeServerDelete,
			ScopeServerImport,
			ScopeServerExport,
		}

	case RoleGuest:
		return []Scope{
			ScopeServerRead,
		}

	default:
		return []Scope{}
	}
}
