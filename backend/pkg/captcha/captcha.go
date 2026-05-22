package captcha

import (
	"github.com/mojocn/base64Captcha"
)

var store = base64Captcha.DefaultMemStore

// Generate returns captcha id and base64 PNG (data URI prefix included in library output).
func Generate() (id, b64 string, err error) {
	driver := base64Captcha.DefaultDriverDigit
	c := base64Captcha.NewCaptcha(driver, store)
	id, b64, _, err = c.Generate()
	return id, b64, err
}

// Verify checks captcha answer; clears after success when clear is true.
func Verify(id, answer string, clear bool) bool {
	return store.Verify(id, answer, clear)
}
