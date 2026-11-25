package router

import (
	"github.com/gofiber/fiber/v3"
	"github.com/mrbayss/Go-Commerce/internal/controller"
)

// SetupStoreRoutes registers store endpoints.
// Static segments (/me) must be registered before /:id routes.
func SetupStoreRoutes(api fiber.Router, sc *controller.StoreController, requireAuth, requireAdmin fiber.Handler) {
	stores := api.Group("/stores")
	stores.Post("", requireAuth, sc.Create)
	stores.Get("/me", requireAuth, sc.GetMine)
	stores.Patch("/me", requireAuth, sc.UpdateMine)
	stores.Put("/me/logo", requireAuth, sc.UpdateLogo)
	stores.Delete("/me/logo", requireAuth, sc.DeleteLogo)
	stores.Put("/me/banner", requireAuth, sc.UpdateBanner)
	stores.Delete("/me/banner", requireAuth, sc.DeleteBanner)

	admin := api.Group("/admin/stores", requireAuth, requireAdmin)
	admin.Get("", sc.List)
	admin.Patch("/:id/approve", sc.Approve)
	admin.Patch("/:id/reject", sc.Reject)
}
