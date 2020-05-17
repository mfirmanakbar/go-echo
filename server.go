package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	jwtGo "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	myMiddleware "github.com/labstack/echo/v4/middleware"
)

var (
	CookieSessionLogin    = "SessionLogin"
	LoginSuccessCookieVal = "LoginSuccess"
	UserIdExample         = "20022012"
	SecretKeyExample      = "mLmHu8f1IxFo4dWurBG3jEf1Ex0wDZvvwND6eFmcaX"
	SigningMethodExample  = "HS512"
)

func main() {
	e := echo.New()

	// Custom Middleware
	e.Use(CustomMiddleWareForServerHeader)

	adminGroup := e.Group("/admin")
	cookieGroup := e.Group("/cookie")
	jwtGroup := e.Group("/jwt")

	// Default Logging Middleware when call API /admin/**
	// adminGroup.Use(myMiddleware.Logger())

	// Custom Logging Middleware when call API /admin/**
	adminGroup.Use(myMiddleware.LoggerWithConfig(myMiddleware.LoggerConfig{
		Format: `[${time_rfc3339} ${status} ${method} ${host}${path} ${latency_human}]` + "\n",
	}))

	// BASIC Auth Middleware
	adminGroup.Use(myMiddleware.BasicAuth(validateUser))

	// JWT Middleware
	jwtGroup.Use(myMiddleware.JWTWithConfig(myMiddleware.JWTConfig{
		SigningMethod: SigningMethodExample,
		SigningKey:    []byte(SecretKeyExample),
	}))

	// COOKIE Middleware
	cookieGroup.Use(checkCookie)

	// ROUTING `cookieGroup`
	cookieGroup.GET("/main", mainCookie)

	// ROUTING `adminGroup`
	adminGroup.GET("/main", mainAdmin)

	// ROUTING `jwtGroup`
	jwtGroup.GET("/main", mainJwt)

	// ROUTING `E`
	e.GET("/login", login)
	e.GET("/", welcome)
	e.GET("/users/:id", getUser)
	e.POST("/users/form", saveUserByForm)
	e.POST("/users/json", saveUserByJson)
	e.PUT("/users/:id", updateUser)
	e.DELETE("/users/:id", deleteUser)
	e.POST("/upload", fileUpload)

	// SERVER HOST
	e.Logger.Fatal(e.Start(":1323"))
}

func mainJwt(c echo.Context) error {
	return c.String(http.StatusOK, "SUCCESS: you are on the top secret jwt page!")
}

func login(c echo.Context) error {
	username := c.QueryParam("username")
	password := c.QueryParam("password")

	if username == "firman" && password == "secret" {
		cookie := &http.Cookie{} // this is same like --> cookie := new(http.Cookie)
		cookie.Name = CookieSessionLogin
		cookie.Value = LoginSuccessCookieVal
		cookie.Expires = time.Now().Add(48 * time.Hour)

		c.SetCookie(cookie)

		// TODO: create jwt token
		token, err := createJwtToken()
		if err != nil {
			log.Println("Error when creating JWT token", err)
			return c.String(http.StatusInternalServerError, "ERROR: something went wrong while creating JWT token!")
		}

		return c.JSON(http.StatusOK, map[string]string{
			"message": "You ware logged in!",
			"token":   token,
		})
	}

	return c.String(http.StatusUnauthorized, "WARNING: Make sure your account is coorect!")
}

type JwtClaims struct {
	Name string `json:"name"`
	jwtGo.StandardClaims
}

func createJwtToken() (string, error) {
	claims := JwtClaims{
		"Firman",
		jwtGo.StandardClaims{
			Id:        UserIdExample,
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		},
	}

	// we hash the jwt claims
	rawToken := jwtGo.NewWithClaims(jwtGo.SigningMethodHS512, claims)

	token, err := rawToken.SignedString([]byte(SecretKeyExample))
	if err != nil {
		return "", err
	}

	return token, nil
}

func checkCookie(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie(CookieSessionLogin)

		if err != nil {
			if strings.Contains(err.Error(), "named cookie not present") {
				return c.String(http.StatusUnauthorized, "WARNING: You don't have any cookie")
			}
			log.Println(err)
			return err
		}

		if cookie.Value == LoginSuccessCookieVal {
			return next(c)
		}

		return c.String(http.StatusUnauthorized, "WARNING: You don't have the right cookie")
	}
}

func mainCookie(c echo.Context) error {
	return c.String(http.StatusOK, "SUCCESS: you are on the secret cookie page!")
}

func CustomMiddleWareForServerHeader(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderServer, "ServerFirman/1.0")
		c.Response().Header().Set("My-Custom-Header", "ThisHaveNoMeaning")
		return next(c)
	}
}

func mainAdmin(c echo.Context) error {
	return c.String(http.StatusOK, "SUCCESS: hello you are in the admin page")
}

func validateUser(username, password string, c echo.Context) (bool, error) {
	if username == "firman" && password == "secret" {
		return true, nil
	}
	return false, nil
}

func welcome(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func getUser(c echo.Context) error {
	id := c.Param("id")
	team := c.QueryParam("team")
	member := c.QueryParam("member")
	// return c.String(http.StatusOK, "id: "+id+", team: "+team+", member: "+member)
	return c.JSON(http.StatusOK, map[string]string{
		"id":     id,
		"team":   team,
		"member": member,
	})
}

func saveUserByForm(c echo.Context) error {
	name := c.FormValue("name")
	email := c.FormValue("email")
	return c.String(http.StatusOK, "name: "+name+", email: "+email)
}

func updateUser(c echo.Context) error {
	id := c.Param("id")
	return c.String(http.StatusOK, "ID: "+id+", successfully updated.")
}

func deleteUser(c echo.Context) error {
	id := c.Param("id")
	return c.String(http.StatusNoContent, id)
}

type User struct {
	Name  string `json:"name" xml:"name" form:"name" query:"name"`
	Email string `json:"email" xml:"email" form:"email" query:"email"`
}

func saveUserByJson(c echo.Context) error {
	u := new(User)
	if err := c.Bind(u); err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, u)
}

func fileUpload(c echo.Context) error {
	avatar, err := c.FormFile("avatar")
	if err != nil {
		return err
	}

	// Source
	src, err := avatar.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Destination
	dst, err := os.Create("upload-files/" + avatar.Filename)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	return c.HTML(http.StatusOK, "Your file `"+avatar.Filename+"` successfully uploaded")
}
