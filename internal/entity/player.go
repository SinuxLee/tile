package entity

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io/ioutil"
)

type Player struct {
	AccessToken   string `json:"accessToken"`
	RefreshToken  string `json:"refreshToken"`
	HeadImgUrl    string `json:"headImgUrl"`
	UserId        int    `json:"userId"`
	UserIp        string `json:"userIp"`
	NickName      string `json:"nickName"`
	ExpiresIn     int    `json:"expiresIn"`
	LastLoginTime int64  `json:"lastLoginTime"`
}

func (p *Player) Marshal() ([]byte, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	return p.gzipMsg(data)
}

func (p *Player) Unmarshal(data []byte) error {
	data, err := p.unGzipMsg(data)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, p)
}

func (p *Player) gzipMsg(data []byte) ([]byte, error) {
	buff := &bytes.Buffer{}
	gzipWriter := gzip.NewWriter(buff)
	if _, err := gzipWriter.Write(data); err != nil {
		return nil, err
	}

	if err := gzipWriter.Flush(); err != nil {
		return nil, err
	}

	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func (p *Player) unGzipMsg(compress []byte) ([]byte, error) {
	gzipReader, err := gzip.NewReader(bytes.NewReader(compress))
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(gzipReader)
	if err != nil {
		return nil, err
	}

	err = gzipReader.Close()
	if err != nil {
		return nil, err
	}

	return data, nil
}
