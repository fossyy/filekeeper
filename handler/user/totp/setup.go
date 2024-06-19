package userHandlerTotpSetup

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/fossyy/filekeeper/db"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
	userTotpSetupView "github.com/fossyy/filekeeper/view/user/totp"
	"github.com/skip2/go-qrcode"
	"github.com/xlzd/gotp"
	"image/png"
	"net/http"
	"time"
)

func GET(w http.ResponseWriter, r *http.Request) {
	secret := gotp.RandomSecret(16)
	userSession := r.Context().Value("user").(types.User)
	totp := gotp.NewDefaultTOTP(secret)
	uri := totp.ProvisioningUri(userSession.Email, utils.Getenv("DOMAIN"))
	qr, err := qrcode.New(uri, qrcode.Medium)
	if err != nil {
		fmt.Printf("Failed to generate QR code: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var buffer bytes.Buffer
	err = png.Encode(&buffer, qr.Image(256))
	if err != nil {
		fmt.Printf("Failed to encode QR code to PNG: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	base64Str := base64.StdEncoding.EncodeToString(buffer.Bytes())

	component := userTotpSetupView.Main("Totp setup page", base64Str, secret)
	err = component.Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func POST(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	code := r.Form.Get("totp")
	secret := r.Form.Get("secret")
	totp := gotp.NewDefaultTOTP(secret)
	userSession := r.Context().Value("user").(types.User)
	fmt.Println(userSession)
	if totp.Verify(code, time.Now().Unix()) {
		err := db.DB.InitializeTotp(userSession.Email, secret)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Authentication successful! Access granted.")
		return
	} else {
		uri := totp.ProvisioningUri(userSession.Email, utils.Getenv("DOMAIN"))
		qr, err := qrcode.New(uri, qrcode.Medium)
		if err != nil {
			fmt.Printf("Failed to generate QR code: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var buffer bytes.Buffer
		err = png.Encode(&buffer, qr.Image(256))
		if err != nil {
			fmt.Printf("Failed to encode QR code to PNG: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		base64Str := base64.StdEncoding.EncodeToString(buffer.Bytes())
		component := userTotpSetupView.Main("Totp setup page", base64Str, secret)
		err = component.Render(r.Context(), w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}
}
