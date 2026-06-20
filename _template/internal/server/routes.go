package server

// routes.go defines all route groups and wires them to handlers.
//
// Route mounting is called from App.mountRoutes() in app.go.
// Each module's routes are registered in its own init-style function
// or by directly calling handler constructors.
//
// Layout:
//
//	/auth/*        → auth handlers (Phase 6)
//	/settings/*    → settings handlers (Phase 8)
//	/              → redirect to /dashboard
//	/dashboard     → dashboard handler (Phase 7)
//	/health        → health check
//
// TODO: wire auth handlers
//   r.Get("/login", authHandler.LoginPage)
//   r.Post("/login", authHandler.LoginPost)
//   r.Get("/register", authHandler.RegisterPage)
//   ...
//
// TODO: wire settings handlers
//   r.Get("/", settingsHandler.Index)
//   r.Post("/profile", settingsHandler.ProfilePost)
//   ...
