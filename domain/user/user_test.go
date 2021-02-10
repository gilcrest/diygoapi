package user

import "testing"

func TestUser_IsValid(t *testing.T) {
	type fields struct {
		Email        string
		LastName     string
		FirstName    string
		FullName     string
		HostedDomain string
		PictureURL   string
		ProfileLink  string
	}

	otto := fields{
		Email:        "otto.maddox@helpinghandacceptanceco.com",
		LastName:     "Maddox",
		FirstName:    "Otto",
		FullName:     "Otto Maddox",
		HostedDomain: "",
		PictureURL:   "",
		ProfileLink:  "",
	}

	noEmail := fields{
		Email:        "",
		LastName:     "Maddox",
		FirstName:    "Otto",
		FullName:     "Otto Maddox",
		HostedDomain: "",
		PictureURL:   "",
		ProfileLink:  "",
	}

	noLastName := fields{
		Email:        "otto.maddox@helpinghandacceptanceco.com",
		LastName:     "",
		FirstName:    "Otto",
		FullName:     "Otto Maddox",
		HostedDomain: "",
		PictureURL:   "",
		ProfileLink:  "",
	}

	noFirstName := fields{
		Email:        "otto.maddox@helpinghandacceptanceco.com",
		LastName:     "Maddox",
		FirstName:    "",
		FullName:     "Otto Maddox",
		HostedDomain: "",
		PictureURL:   "",
		ProfileLink:  "",
	}

	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"typical", otto, true},
		{"no email", noEmail, false},
		{"no last name", noLastName, false},
		{"no first name", noFirstName, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := User{
				Email:        tt.fields.Email,
				LastName:     tt.fields.LastName,
				FirstName:    tt.fields.FirstName,
				FullName:     tt.fields.FullName,
				HostedDomain: tt.fields.HostedDomain,
				PictureURL:   tt.fields.PictureURL,
				ProfileLink:  tt.fields.ProfileLink,
			}
			if got := u.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
