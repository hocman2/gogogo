package cookies

import (
	"github.com/hocman2/gogogo/pkg/env"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
);

var secret []byte;
func InitializeCookieSigner() error {
	if env.Get(env.CookieSecret) == "" {
		return errors.New(fmt.Sprintf("Missing %s environment variable, required to sign cookies.", string(env.CookieSecret)));
	}
	secret = []byte(env.Get(env.CookieSecret));
	return nil;
}

func Sign(cookie http.Cookie) http.Cookie {
	value := cookie.Value;
	hasher := hmac.New(sha256.New, secret);
	hasher.Write([]byte(value));
	mac := hasher.Sum(nil);
	cookie.Value = value + "|" + base64.URLEncoding.EncodeToString(mac);
	return cookie;
}

func Verify(cookie *http.Cookie) (*http.Cookie, error) {
	idx := strings.LastIndex(cookie.Value, "|");
	if idx < 0 {
		return nil, errors.New("Couldn't find value/signature separator: |");
	}

	value := cookie.Value[:idx];
	signatureb64 := cookie.Value[idx+1:];

	given, err := base64.URLEncoding.DecodeString(signatureb64);
	if err != nil {
		return nil, err;
	}

	hasher := hmac.New(sha256.New, secret);
	hasher.Write([]byte(value));
	expected := hasher.Sum(nil);
	if hmac.Equal(expected, given) {
		verifiedCookie := *cookie;
		verifiedCookie.Value = value;
		return &verifiedCookie, nil;
	} else {
		return nil, errors.New("Verification failed");
	}
}
