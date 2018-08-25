package auth

import "testing"

func Test_loginTokenExpire(t *testing.T) {
	type args struct {
		seconds2expire int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{"test1", args{dayinsecs}, 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := loginTokenExpire(tt.args.seconds2expire); got != tt.want {
				t.Errorf("loginTokenExpire() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateJWT(t *testing.T) {

	j := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFzZGYiLCJleHAiOjE1MzQ5OTI0MjQsImlzcyI6ImdpbGNyZXN0In0.5cwzt4x7pDNP4D0ZWH3ZXY2ou1xLbFbey7xc6Tz2HIU"

	validateJWT(j)
}
