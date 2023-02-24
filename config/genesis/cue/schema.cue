package genesis

// Objects included in output. All object definitions are in the schema.cue file. The permissions and roles are in the
// auth.cue file, but they are only included as part of the request.json file due to their inclusion below (because of
// the leading underscore).
user: #User
org:  #Org
permissions: [_pingV1Get, _loggerV1Get, _loggerV1Put, _orgsV1Post, _orgsV1Put, _orgsV1Delete, _orgsV1Get,
	_orgsV1GetByExtlID, _appsV1Post, _permissionsV1Post, _permissionsV1Get, _permissionsV1Delete,
	_moviesV1Post, _moviesV1UpdateByExtlID, _moviesV1DeleteByExtlID, _moviesV1FindByExtlID, _moviesV1FindAll]
roles: [_sysAdmin, _movieAdmin]

#User: {
	provider: #Oauth2Provider
	token:    !="" // must be specified and non-empty
}

#Oauth2Provider: "google"

#Org: {
	name:        !="" // must be specified and non-empty
	description: !="" // must be specified and non-empty
	kind:        #OrgKinds
	app:         #App
}

#OrgKinds: "genesis" | "test" | "standard"

#App: {
	name:                      !="" // must be specified and non-empty
	description:               !="" // must be specified and non-empty
	oauth2_provider:           #Oauth2Provider
	oauth2_provider_client_id: !="" // must be specified and non-empty
}

// Auth is the permissions and roles required for the Role Based Access Control (RBAC) setup of the app
#Auth: {
	permissions: [...#Permission]
	roles: [...#Role]
}

// Role is a job function or title which defines an authority level.
#Role: {
	// A human-readable code which represents the role.
	role_cd: !="" // must be specified and non-empty
	// A longer description of the role.
	role_description: !="" // must be specified and non-empty
	// A boolean denoting whether the role is active (true) or not (false).
	active: bool
	// A list of permissions that the role allows
	permissions: [...#Permission]
}

// Permission stores an approval of a mode of access to a resource.
#Permission: {
	// A human-readable string which represents a resource (e.g. an HTTP route or document, etc.).
	resource: !="" // must be specified and non-empty
	// A string representing the action taken on the resource (e.g. POST, GET, edit, etc.)
	operation: !="" // must be specified and non-empty
	// A description of what the permission is granting, e.g. "grants ability to edit a billing document".
	description: !="" // must be specified and non-empty
	// A boolean denoting whether the permission is active (true) or not (false).
	active: bool
}
