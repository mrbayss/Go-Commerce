package router

import (
	"github.com/gofiber/fiber/v3"
	"github.com/mrbayss/Go-Commerce/internal/controller"
)

// SetupProductRoutes registers product endpoints.
// /me MUST be registered before /:id, otherwise Fiber matches "me" as :id.
func SetupProductRoutes(api fiber.Router, pc *controller.ProductController, requireAuth fiber.Handler) {
	products := api.Group("/products")
	products.Post("", requireAuth, pc.Create)
	products.Get("/me", requireAuth, pc.ListMine)
	products.Get("/:id", requireAuth, pc.Get)
	products.Patch("/:id", requireAuth, pc.Update)
	products.Delete("/:id", requireAuth, pc.Delete)

	products.Post("/:id/variants", requireAuth, pc.AddVariant)
	products.Patch("/:id/variants/:variantId", requireAuth, pc.UpdateVariant)
	products.Delete("/:id/variants/:variantId", requireAuth, pc.DeleteVariant)
}
