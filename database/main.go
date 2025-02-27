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

type userRole struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

// type Role struct {
// 	Id          string
// 	Name        string
// 	Discription string
// 	Permission  string
// }

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

		// Return user's role
		se.Router.GET("/api/role/{token}", func(e *core.RequestEvent) error {
			// Hanlde requester permission
			if err := verifyRequester(e.Request.Header.Get("Authorization"), app); err != nil {
				return e.String(http.StatusUnauthorized, err.Error())
			}

			// Grant token from parameter
			token := e.Request.PathValue("token")
			user, err := app.FindAuthRecordByToken(token, core.TokenTypeAuth)
			if err != nil {
				return e.String(http.StatusUnauthorized, "Unauthorized 2")
			} else if user.Collection().Name != "users" {
				return e.String(http.StatusForbidden, "Unauthorized 3")
			}

			userRecord, err := app.FindRecordById("users", user.Id)
			if err != nil {
				return e.String(http.StatusInternalServerError, "Internal Server error")
			}

			return e.JSON(http.StatusOK, map[string]any{"name": userRecord.GetString("name"), "role": userRecord.GetString("role")})
		})

		// se.Router.POST("/api/role/copy/{roleId}", func(e *core.RequestEvent) error {
		// 	if err := verifyRequester(e.Request.Header.Get("Authorization"), app); err != nil {
		// 		return e.String(http.StatusUnauthorized, err.Error())
		// 	}

		// 	roleId := e.Request.PathValue("roleId")
		// 	role, err := app.FindRecordById("roles", roleId)
		// 	if err != nil {
		// 		return e.String(http.StatusNotFound, "Role not found")
		// 	}
		// 	roleName := role.GetString("name")
		// 	roleDiscription := role.GetString("discription")
		// 	rolePermission := role.GetString("permission")

		// 	// Create new role
		// 	collection, err := app.FindCollectionByNameOrId("roles")
		// 	if err != nil {
		// 		return err
		// 	}

		// 	record := core.NewRecord(collection)

		// 	record.Set("name", roleName)
		// 	record.Set("discription", roleDiscription)
		// 	record.Set("permission", rolePermission)

		// 	err = app.Save(record)
		// 	if err != nil {
		// 		return err
		// 	}

		// 	return e.JSON(http.StatusCreated, map[string]any{"newRoleId": "role"})
		// })

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
