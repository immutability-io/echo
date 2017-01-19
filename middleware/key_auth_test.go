package middleware

import (
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestKeyAuth(t *testing.T) {
	e := echo.New()
	req, _ := http.NewRequest(echo.GET, "/", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	config := KeyAuthConfig{
		Validator: func(key string, c echo.Context) bool {
			return key == "valid-key"
		},
	}
	h := KeyAuthWithConfig(config)(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	// Valid key
	auth := DefaultKeyAuthConfig.AuthScheme + " " + "valid-key"
	req.Header.Set(echo.HeaderAuthorization, auth)
	assert.NoError(t, h(c))

	// Invalid key
	auth = DefaultKeyAuthConfig.AuthScheme + " " + "invalid-key"
	req.Header.Set(echo.HeaderAuthorization, auth)
	he := h(c).(*echo.HTTPError)
	assert.Equal(t, http.StatusUnauthorized, he.Code)

	// Missing Authorization header
	req.Header.Del(echo.HeaderAuthorization)
	he = h(c).(*echo.HTTPError)
	assert.Equal(t, http.StatusBadRequest, he.Code)

	// Key from custom header
	config.KeyLookup = "header:API-Key"
	h = KeyAuthWithConfig(config)(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})
	req.Header.Set("API-Key", "valid-key")
	assert.NoError(t, h(c))

	// Key from query string
	config.KeyLookup = "query:key"
	h = KeyAuthWithConfig(config)(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})
	q := req.URL.Query()
	q.Add("key", "valid-key")
	req.URL.RawQuery = q.Encode()
	assert.NoError(t, h(c))

	// Key from cookie
	config.KeyLookup = "cookie:token"
	h = KeyAuthWithConfig(config)(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})
	expiration := time.Now().Add(365 * 24 * time.Hour)
	cookie := http.Cookie{Name: "token", Value: "valid-key", Expires: expiration}
	req.AddCookie(&cookie)
	assert.NoError(t, h(c))

}
