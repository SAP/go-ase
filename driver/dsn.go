package driver

import (
	"net/url"
)

//ase://myuser:mypassword@localhost:4901?

type DsnInfo struct {
	Host, Port, Username, Password string
	ConnectProps                   url.Values
}

func parseDSN(dsn string) (*DsnInfo, error) {

	url, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}

	userName := ""
	password := ""
	if url.User != nil {
		userName = url.User.Username()
		password, _ = url.User.Password()
	}

	return &DsnInfo{
		Host:         url.Hostname(),
		Port:         url.Port(),
		Username:     userName,
		Password:     password,
		ConnectProps: url.Query(),
	}, nil
}
