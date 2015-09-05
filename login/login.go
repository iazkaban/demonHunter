package login

import (
	"net/http"
	"net/url"
	"regexp"

	"github.com/iazkaban/demonHunter/config"
)

var Header http.Header

func Login() error {
	Header = make(http.Header, 5)
	if len(config.Config.Login.PostValues) == 0 {
		return nil
	}
	v := url.Values{}
	for key, value := range config.Config.Login.PostValues {
		v.Set(key, value)
	}
	response, err := http.PostForm(config.Config.Login.LoginUrl, v)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	for key, value := range response.Header {
		switch key {
		case "Set-Cookie":
			re := regexp.MustCompile(";.+$")
			value[0] = re.ReplaceAllString(value[0], "")
			Header.Set("Cookie", value[0])
		}
	}
	Header.Set("Content-Type", "application/x-www-form-urlencoded")
	/*
		body := make([]byte, 65536)
		response.Body.Read(body)
		fmt.Println(string(body))
	*/
	return nil
}
