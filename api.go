package soopchat

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/tidwall/gjson"
)

const (
	dataUrl    = "https://live.sooplive.co.kr/afreeca/player_live_api.php?bjId=%s"
	loginUrl   = "https://login.sooplive.co.kr/app/LoginAction.php"
	user_agent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:122.0) Gecko/20100101 Firefox/122.0"
)

func (c *Client) setSocketData() error {
	data := url.Values{}
	data.Set("bid", c.Token.StreamerID)
	data.Set("player_type", "html5")

	resp, err := c.httpClient.R().
		SetFormData(map[string]string{
			"bid":         c.Token.StreamerID,
			"player_type": "html5",
		}).
		Post(fmt.Sprintf(dataUrl, c.Token.StreamerID))
	if err != nil {
		return err
	}

	jsonResult := gjson.GetBytes(resp.Body(), "CHANNEL")
	result := jsonResult.Get("RESULT").Int()
	switch result {
	case -6:
		return errors.New("login required")
	}

	domain := jsonResult.Get("CHDOMAIN").String()
	port := jsonResult.Get("CHPT").Int() + 1

	c.socketAddress = fmt.Sprintf("wss://%s:%d/Websocket", domain, port)
	c.Token.chatRoom = jsonResult.Get("CHATNO").String()

	return nil
}

func (c *Client) login() error {
	resp, err := c.httpClient.R().
		SetFormData(map[string]string{
			"szWork":     "login",
			"szType":     "json",
			"szUid":      c.Token.Identifier.ID,
			"szPassword": c.Token.Identifier.Password,
		}).
		SetHeader("User-Agent", user_agent).
		Post(loginUrl)
	if err != nil {
		return err
	}

	result := gjson.GetBytes(resp.Body(), "RESULT").Bool()
	if !result {
		return errors.New("login failed")
	}

	return nil
}
