package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// verifyRequester ensure only superuser can access the endpoints
func verifyRequester(header string, app *pocketbase.PocketBase) error {
	if header == "" {
		return fmt.Errorf("unauthorized, missing header")
	}

	if strings.HasPrefix(header, "Bearer ") {
		requesterToken := strings.TrimSpace(header[7:]) // Trim spaces for safety
		if requesterToken == "" {
			return fmt.Errorf("unauthorized, missing token")
		}

		requester, err := app.FindAuthRecordByToken(requesterToken, core.TokenTypeAuth)
		if err != nil {
			return fmt.Errorf("unauthorized, can't find user")
		}

		if requester.Collection().Name != "_superusers" {
			return fmt.Errorf("Unauthorized")
		}
	}

	return nil
}

func main() {
	app := pocketbase.New()

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// serves static files from the provided public dir (if exists)
		se.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), false))

		se.Router.PATCH("/api/settings/smtp", func(e *core.RequestEvent) error {
			// Hanlde requester permission
			if err := verifyRequester(e.Request.Header.Get("Authorization"), app); err != nil {
				return e.String(http.StatusUnauthorized, err.Error())
			}

			data := struct {
				Service     string `json:"service"`
				Host        string `json:"host"`
				Port        int    `json:"port"`
				Username    string `json:"username"`
				Password    string `json:"password"`
				FromAddress string `json:"from_address"`
				FromName    string `json:"from_name"`
			}{}
			if err := e.BindBody(&data); err != nil {
				return e.BadRequestError("Failed to read request data", err)
			}

			e.App.Settings().SMTP.Host = data.Host
			e.App.Settings().SMTP.Port = data.Port
			e.App.Settings().SMTP.Username = data.Username
			e.App.Settings().SMTP.Password = data.Password
			e.App.Settings().Meta.SenderAddress = data.FromAddress
			e.App.Settings().Meta.SenderName = data.FromName

			return e.JSON(http.StatusOK, map[string]any{"message": "ok"})
		})

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
