package random

import (
	"testing"
)

func TestGenerateRandomBytes(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"typical", args{15}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GenerateRandomBytes(tt.args.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateRandomBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestCryptoString(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"typical", args{15}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sg := StringGenerator{}
			_, err := sg.CryptoString(tt.args.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("CryptoString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
