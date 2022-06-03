package genesis

permissions: [_pingV1Get, _loggerV1Get, _loggerV1Put, _orgsV1Post, _orgsV1Put, _orgsV1Delete, _orgsV1Get, _orgsV1GetByExtlID, _appsV1Post, _permissionsV1Post, _permissionsV1Get]
roles: [_sysAdmin]

_pingV1Get: #Permission & {
	resource:    "/api/v1/ping"
	operation:   "GET"
	description: "allows for calling the ping service to determine if system is up and running"
	active:      true
}

_loggerV1Get: #Permission & {
	resource:    "/api/v1/logger"
	operation:   "GET"
	description: "allows for reading the logger state"
	active:      true
}

_loggerV1Put: #Permission & {
	resource:    "/api/v1/logger"
	operation:   "PUT"
	description: "allows for updating the logger state"
	active:      true
}

_orgsV1Post: #Permission & {
	resource:    "/api/v1/orgs"
	operation:   "POST"
	description: "allows for creating an organization"
	active:      true
}

_orgsV1Put: #Permission & {
	resource:    "/api/v1/orgs"
	operation:   "PUT"
	description: "allows for updating an organization"
	active:      true
}

_orgsV1Delete: #Permission & {
	resource:    "/api/v1/orgs"
	operation:   "DELETE"
	description: "allows for deleting an organization"
	active:      true
}

_orgsV1Get: #Permission & {
	resource:    "/api/v1/orgs"
	operation:   "GET"
	description: "allows for finding all organizations"
	active:      true
}

_orgsV1GetByExtlID: #Permission & {
	resource:    "/api/v1/orgs/{extlID}"
	operation:   "GET"
	description: "allows for finding an organization by external ID"
	active:      true
}

_appsV1Post: #Permission & {
	resource:    "/api/v1/apps"
	operation:   "POST"
	description: "allows for creating an app"
	active:      true
}

_appsV1Post: #Permission & {
	resource:    "/api/v1/apps"
	operation:   "POST"
	description: "allows for creating an app"
	active:      true
}

//   {PathTemplate: pathPrefix + registerV1PathRoot, HTTPMethods: []string{http.MethodPost}},

_permissionsV1Post: #Permission & {
	resource:    "/api/v1/permissions"
	operation:   "POST"
	description: "allows for creating a permission"
	active:      true
}

_permissionsV1Get: #Permission & {
	resource:    "/api/v1/permissions"
	operation:   "GET"
	description: "allows for finding all permissions"
	active:      true
}

_sysAdmin: #Role & {
	role_cd:          "sysAdmin"
	role_description: "System administrator role."
	active:           true
	permissions: [_pingV1Get, _loggerV1Get, _loggerV1Put, _orgsV1Post, _orgsV1Put, _orgsV1Delete, _orgsV1Get, _orgsV1GetByExtlID, _appsV1Post, _permissionsV1Post, _permissionsV1Get]
}
