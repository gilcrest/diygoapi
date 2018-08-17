package response

// Info struct should be used for all responses
type Info struct {
	RequestID  string `json:"id"`
	RequestURL string `json:"url"`
}

// NewInfo is a constructor for the Info struct
func NewInfo() (*Info, error) {
	info := new(Info)
	info.RequestID = "fakeID"
	info.RequestURL = "fakeURL"
	return info, nil
}
