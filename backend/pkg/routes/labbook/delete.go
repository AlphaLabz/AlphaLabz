package labbook

import (
	"alphalabz/pkg/casbin"
	"alphalabz/pkg/pocketbase"
	"net/http"
)

func HandleLabBookDelete(w http.ResponseWriter, r *http.Request, pbClient *pocketbase.PocketBaseClient, ce *casbin.CasbinEnforcer) {
	// ...
}
