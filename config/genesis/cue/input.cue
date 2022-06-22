package genesis

// The "genesis" user - the first user to create the system and is
// given the sysAdmin role (which has all permissions). This user is
// added to the Principal org and the user initiated org created below.
user: email:      "otto.maddox@gmail.com"
user: first_name: "Otto"
user: last_name:  "Maddox"

// The first organization created which can actually transact
// (e.g. is not the principal or test org)
org: name:        "Movie Makers Unlimited"
org: description: "An organization dedicated to creating movies in a demo app."
org: kind:        "standard"

// The initial app created along with the Organization created above
org: app: name:        "Movie Makers App"
org: app: description: "The first app dedicated to creating movies in a demo app."
