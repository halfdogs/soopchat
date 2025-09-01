package soopchat

import (
	"errors"
	"fmt"

	"github.com/tidwall/gjson"
)

// Error
var (
	ErrLogin      = errors.New("login failed")
	ErrStatusCode = errors.New("status code not 200")
)

// API Domain
const (
	apiDataURL   string = "https://live.sooplive.co.kr/afreeca/player_live_api.php?bjId=%s"
	apiLoginURL  string = "https://login.sooplive.co.kr/app/LoginAction.php"
	apiUserAgent string = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:122.0) Gecko/20100101 Firefox/122.0"
)

func (api *apiService) login(identifier identifier) error {
	resp, err := api.httpClient.R().
		SetFormData(map[string]string{
			"szWork":     "login",
			"szType":     "json",
			"szUid":      identifier.ID,
			"szPassword": identifier.Password,
		}).
		SetHeader("User-Agent", apiUserAgent).
		Post(apiLoginURL)
	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		return ErrLogin
	}

	if result := gjson.GetBytes(resp.Body(), "RESULT").Bool(); !result {
		return ErrLogin
	}

	return nil
}

func (api *apiService) setSocketData(client *Client) ([]byte, error) {
	resp, err := api.httpClient.R().
		SetFormData(map[string]string{
			"bid":         client.channel.streamer,
			"player_type": "html5",
		}).
		Post(fmt.Sprintf(apiDataURL, client.channel.streamer))
	if err != nil {
		return []byte{}, err
	}

	if resp.StatusCode() != 200 {
		return []byte{}, ErrStatusCode
	}

	return resp.Body(), nil
}
