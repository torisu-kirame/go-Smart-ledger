package captcha

import (
	"strings"

	"github.com/mojocn/base64Captcha"
)

var store = base64Captcha.DefaultMemStore

// DevPassToken bypasses the image captcha when both captchaId and captchaCode
// equal this value. Intended for local AI / automation (see root login.json).
const DevPassToken = "sl-ai-captcha-pass"

// Generate returns captcha id and base64 PNG (data URI prefix included in library output).
func Generate() (id, b64 string, err error) {
	driver := base64Captcha.DefaultDriverDigit
	c := base64Captcha.NewCaptcha(driver, store)
	id, b64, _, err = c.Generate()
	return id, b64, err
}

// Verify checks captcha answer; clears after success when clear is true.
func Verify(id, answer string, clear bool) bool {
	id = strings.TrimSpace(id)
	answer = strings.TrimSpace(answer)
	if id != "" && id == DevPassToken && answer == DevPassToken {
		return true
	}
	return store.Verify(id, answer, clear)
}
