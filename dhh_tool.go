package dhhclient

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/big"
	"net/http"
	"regexp"
	"strconv"

	"github.com/kalrey/zlog"
)

const (
	DHH_LOGIN_URL = "https://login.taobao.com/member/login.jhtml"
	DHH_CSRF_URL  = "https://dahanghai.taobao.com/"
)

type DHHTool struct {
	initViewData map[string]interface{}
	rsaModulus   string
	rsaExponent  int
}

func InitDHHTool() *DHHTool {
	dhh := &DHHTool{}
	err := dhh.loadingTaobaoInitViewData()
	if err != nil {
		zlog.Logger.Error("Init DHHTool failed, verbose info: " + err.Error())
		return nil
	}
	return dhh
}

func (this *DHHTool) GenPassword2(plaintext string) (string, error) {
	return this.rsa_tb(this.rsaModulus, this.rsaExponent, []byte(plaintext))
}

func (this *DHHTool) rsa_tb(modulusStr string, exponent int, data []byte) (string, error) {
	modulus := big.NewInt(0)
	modulus.SetString(modulusStr, 16)

	rsaPublic := rsa.PublicKey{
		N: modulus,
		E: exponent,
	}

	encryptedData, err := rsa.EncryptPKCS1v15(rand.Reader, &rsaPublic, data)

	if err != nil {
		return "", err
	}

	return hex.EncodeToString(encryptedData), nil
}

func (this *DHHTool) loadingTaobaoInitViewData() error {
	this.initViewData = make(map[string]interface{})
	resp, err := http.Get(DHH_LOGIN_URL)

	if err != nil {
		return err
	}

	html, err2 := ioutil.ReadAll(resp.Body)

	if err2 != nil {
		return err2
	}

	reg := regexp.MustCompile("\"loginFormData\":([^}]*})")

	content := reg.FindStringSubmatch(string(html))
	if len(content) < 2 {
		return errors.New("\"loginFormData\" not find in html")

	} else {
		json.Unmarshal([]byte(content[1]), &this.initViewData)
	}

	reg2 := regexp.MustCompile("rsaModulus\":\"([0-9a-z]+)")
	reg3 := regexp.MustCompile("rsaExponent\":\"([0-9a-z]+)")

	content2 := reg2.FindStringSubmatch(string(html))
	content3 := reg3.FindStringSubmatch(string(html))

	if len(content2) < 2 {
		return errors.New("\"rsaModulus\" not find in html")
	}

	if len(content3) < 2 {
		return errors.New("\"rsaExponent\" not find in html")
	}

	tmp, err3 := strconv.ParseInt(content3[1], 16, 64)
	if err3 != nil {
		return err3
	}

	this.rsaModulus = content2[1]
	this.rsaExponent = int(tmp)

	return nil

}
