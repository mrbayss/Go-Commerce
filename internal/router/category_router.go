package router

import (
	"github.com/gofiber/fiber/v3"
	"github.com/mrbayss/Go-Commerce/internal/controller"
)

// SetupCategoryRoutes registers category endpoints.
// Static segments must precede /:id routes (StrictRouting is on).
func SetupCategoryRoutes(api fiber.Router, cc *controller.CategoryController, requireAuth, requireAdmin fiber.Handler) {
	cats := api.Group("/categories")
	cats.Get("", requireAuth, cc.List) // active categories only
	cats.Post("", requireAuth, requireAdmin, cc.Create)
	cats.Get("/:id", requireAuth, cc.Get)
	cats.Patch("/:id", requireAuth, requireAdmin, cc.Update)
	cats.Delete("/:id", requireAuth, requireAdmin, cc.Delete)

	adminCats := api.Group("/admin/categories")
	adminCats.Get("", requireAuth, requireAdmin, cc.ListAll) // includes inactive
}
