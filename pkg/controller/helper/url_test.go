package helper

import (
	"net/url"
	"testing"
)

func TestURLFromDomain(t *testing.T) {
	cases := []struct {
		testName    string
		domain      string
		expectedErr bool
		expectedURL url.URL
	}{
		{"Invalid domain", ":foo", true, url.URL{}},
		{
			"Valid domain", "http://somedomain.example.com", false,
			url.URL{
				Scheme: "https",
				Host:   "somedomain.example.com",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			res, err := URLFromDomain(tc.domain)
			if !tc.expectedErr {
				ok(subT, err)
				equals(subT, tc.expectedURL, *res)
			} else {
				assert(subT, err != nil, "error should not be nil")
			}
		})
	}
}
