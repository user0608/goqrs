package security

import (
	"context"

	"github.com/ksaucedo002/answer"
	"github.com/labstack/echo/v4"
)

type keyType string

const jwttokenclaimskey = keyType("jwt-values-context-key")

func JwtClaims(c context.Context) (JWTValues, bool) {
	values := c.Value(jwttokenclaimskey)
	if values == nil {
		return JWTValues{}, false
	}
	v, ok := values.(JWTValues)
	return v, ok
}
func UserName(c context.Context) string {
	values, ok := JwtClaims(c)
	if !ok {
		return ""
	}
	return values.Username
}
func Context(ctx context.Context, v JWTValues) context.Context {
	return context.WithValue(ctx, jwttokenclaimskey, v)
}

func JWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")
		if token == "" {
			token = c.QueryParam("token")
		}
		values, err := ValidateToken(token)
		if err != nil {
			return answer.ErrorResponse(c, err)
		}
		ctx := Context(c.Request().Context(), values)
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}
