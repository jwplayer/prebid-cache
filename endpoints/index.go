package endpoints

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

//Handle Default route for the prebid-cache
func Index(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	//Default routing instead of showing 404 not found error
	w.WriteHeader(http.StatusNoContent)
}
