package api

import (
	"github.com/alexedwards/scs/v2"
	"github.com/erikgmatos/gobid/internal/services"
	"github.com/go-chi/chi/v5"
)

type Api struct {
	Router          *chi.Mux
	UserServices    services.UserService
	ProductServices services.ProductService
	Sessions        *scs.SessionManager
}
