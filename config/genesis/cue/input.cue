package genesis

// The "genesis" user - the first user to create the system and is
// given the sysAdmin role (which has all permissions). This user is
// added to the Principal org and the user initiated org created below.
// Add the Oauth2 provider (currently only google is supported) and the
// Oauth2 token to be used to create the user.
user: provider: "google"
user: token:    "REPLACE_ME"

// The first organization created which can actually transact
// (e.g. is not the principal or test org)
org: name:        "Movie Makers Unlimited"
org: description: "An organization dedicated to creating movies in a demo app."
org: kind:        "standard"

// The initial app created along with the Organization created above
org: app: name:                      "Movie Makers App"
org: app: description:               "The first app dedicated to creating movies in a demo app."
org: app: oauth2_provider:           "google"
org: app: oauth2_provider_client_id: "REPLACE_ME"
