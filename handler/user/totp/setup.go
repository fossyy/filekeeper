package userHandlerTotpSetup

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/fossyy/filekeeper/cache"
	userTotpSetupView "github.com/fossyy/filekeeper/view/user/totp"
	"image/png"
	"net/http"
	"time"

	"github.com/fossyy/filekeeper/db"
	"github.com/fossyy/filekeeper/types"
	"github.com/skip2/go-qrcode"
	"github.com/xlzd/gotp"
)

func generateQRCode(uri string) (string, error) {
	qr, err := qrcode.New(uri, qrcode.Medium)
	if err != nil {
		return "", fmt.Errorf("failed to generate QR code: %w", err)
	}

	var buffer bytes.Buffer
	if err := png.Encode(&buffer, qr.Image(256)); err != nil {
		return "", fmt.Errorf("failed to encode QR code to PNG: %w", err)
	}

	return base64.StdEncoding.EncodeToString(buffer.Bytes()), nil
}

func GET(w http.ResponseWriter, r *http.Request) {
	secret := gotp.RandomSecret(16)
	userSession := r.Context().Value("user").(types.User)
	totp := gotp.NewDefaultTOTP(secret)
	uri := totp.ProvisioningUri(userSession.Email, "filekeeper")
	base64Str, err := generateQRCode(uri)
	if err != nil {
		fmt.Printf("%v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	component := userTotpSetupView.Main("Filekeeper - 2FA Setup Page", base64Str, secret, userSession)
	if err := component.Render(r.Context(), w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func POST(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	code := r.Form.Get("totp")
	secret := r.Form.Get("secret")
	totp := gotp.NewDefaultTOTP(secret)
	userSession := r.Context().Value("user").(types.User)
	if totp.Verify(code, time.Now().Unix()) {
		if err := db.DB.InitializeTotp(userSession.Email, secret); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		cache.DeleteUser(userSession.Email)
		fmt.Fprint(w, "Authentication successful! Access granted.")
		return
	} else {
		uri := totp.ProvisioningUri(userSession.Email, "filekeeper")

		base64Str, err := generateQRCode(uri)
		if err != nil {
			fmt.Printf("%v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		component := userTotpSetupView.Main("Filekeeper - 2FA Setup Page", base64Str, secret, userSession)
		if err := component.Render(r.Context(), w); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
