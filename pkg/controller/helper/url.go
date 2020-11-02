package helper

import "net/url"

func URLFromDomain(domain string) (*url.URL, error) {
	u, err := url.Parse(domain)
	if err != nil {
		return nil, err
	}
	u.Scheme = "https"
	return u, nil
}
