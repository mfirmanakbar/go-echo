// // Root level middleware
// e.Use(middleware.Logger())
// e.Use(middleware.Recover())

// // Group level middleware
// g := e.Group("/admin")
// g.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
//   if username == "joe" && password == "secret" {
//     return true, nil
//   }
//   return false, nil
// }))

// // Route level middleware
// track := func(next echo.HandlerFunc) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		println("request to /users")
// 		return next(c)
// 	}
// }
// e.GET("/users", func(c echo.Context) error {
// 	return c.String(http.StatusOK, "/users")
// }, track)
