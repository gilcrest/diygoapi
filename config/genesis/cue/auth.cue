package genesis

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

_permissionsV1Delete: #Permission & {
	resource:    "/api/v1/permissions"
	operation:   "DELETE"
	description: "allows for deleting a permission"
	active:      true
}

_moviesV1Post: #Permission & {
	resource:    "/api/v1/movies"
	operation:   "POST"
	description: "allows for creating a movie"
	active:      true
}

_moviesV1UpdateByExtlID: #Permission & {
	resource:    "/api/v1/movies/{extlID}"
	operation:   "PUT"
	description: "allows for updating a movie"
	active:      true
}

_moviesV1DeleteByExtlID: #Permission & {
	resource:    "/api/v1/movies/{extlID}"
	operation:   "DELETE"
	description: "allows for deleting a movie"
	active:      true
}

_moviesV1FindByExtlID: #Permission & {
	resource:    "/api/v1/movies/{extlID}"
	operation:   "GET"
	description: "allows for finding a unique movie"
	active:      true
}

_moviesV1FindAll: #Permission & {
	resource:    "/api/v1/movies"
	operation:   "GET"
	description: "allows for finding all movies"
	active:      true
}

_sysAdmin: #Role & {
	role_cd:          "sysAdmin"
	role_description: "System administrator role."
	active:           true
	permissions: [_pingV1Get, _loggerV1Get, _loggerV1Put, _orgsV1Post, _orgsV1Put, _orgsV1Delete, _orgsV1Get, _orgsV1GetByExtlID, _appsV1Post,
		_permissionsV1Post, _permissionsV1Get, _permissionsV1Delete, _moviesV1Post, _moviesV1UpdateByExtlID, _moviesV1DeleteByExtlID,
		_moviesV1FindByExtlID, _moviesV1FindAll]
}

_movieAdmin: #Role & {
	role_cd:          "movieAdmin"
	role_description: "Users can create, update, delete and read the movie database"
	active:           true
	permissions: [_moviesV1Post, _moviesV1UpdateByExtlID, _moviesV1DeleteByExtlID, _moviesV1FindByExtlID, _moviesV1FindAll]
}
