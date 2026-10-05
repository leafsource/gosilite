package permission

import (
	"context"
	"errors"
	"fmt"
	gjwt "github.com/golang-jwt/jwt/v4"
	userModel "gosilite/api/model/framework/user"
	"gosilite/application/base/ctxutil"
	userService "gosilite/application/framework/service/user"
	"gosilite/conf"
	"gosilite/pkg/errorx"
	R "gosilite/pkg/utils/response"
	"gosilite/types/consts"
	"gosilite/types/errno"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/jwt"
)

type login struct {
	Username string `form:"username,required" json:"username,required"`
	Password string `form:"password,required" json:"password,required"`
}

// loginUserIDKey caches the authenticated user id on the request context
// so the JWT LoginResponse callback can issue a refresh token for the same login.
const loginUserIDKey = "security_login_user_id"

var Security *jwt.HertzJWTMiddleware

func NewSecurity() []app.HandlerFunc {
	return []app.HandlerFunc{
		func(ctx context.Context, c *app.RequestContext) {
			token := BearerToken(c)
			if token == "" {
				return
			}
			denied, err := userService.SystemUserSVC.IsAccessTokenDenied(ctx, token)
			if err != nil {
				// fail open on cache errors; JWT validation still applies
				return
			}
			if denied {
				R.Unauthorized(c, "access token has been revoked")
			}
		},
		Security.MiddlewareFunc(),
	}
}

// BearerToken returns the raw access token from the Authorization header.
func BearerToken(c *app.RequestContext) string {
	auth := string(c.GetHeader("Authorization"))
	if !strings.HasPrefix(auth, consts.AuthBearerPrefix) {
		return ""
	}
	return strings.TrimSpace(auth[len(consts.AuthBearerPrefix):])
}

// MintAccessToken signs a Bearer access token for the given user with the
// same claims shape produced by PayloadFunc at login time.
func MintAccessToken(userId int64) (string, time.Time, error) {
	if Security == nil {
		return "", time.Time{}, errors.New("security middleware is not registered")
	}
	now := time.Now()
	expire := now.Add(Security.Timeout)
	claims := gjwt.MapClaims{
		consts.SecurityIdentityKey: strconv.FormatInt(userId, 10),
		"exp":                      expire.Unix(),
		"orig_iat":                 now.Unix(),
	}
	method := gjwt.GetSigningMethod(Security.SigningAlgorithm)
	if method == nil {
		return "", time.Time{}, fmt.Errorf("unsupported signing algorithm %q", Security.SigningAlgorithm)
	}
	tokenString, err := gjwt.NewWithClaims(method, claims).SignedString(Security.Key)
	if err != nil {
		return "", time.Time{}, err
	}
	return tokenString, expire, nil
}

// DenyCurrentAccessToken puts the access token carried by this request on the
// deny list until its natural expiry, so logout invalidates it immediately.
func DenyCurrentAccessToken(ctx context.Context, c *app.RequestContext) error {
	token := BearerToken(c)
	if token == "" || Security == nil {
		return nil
	}
	parsed, err := Security.ParseTokenString(token)
	if err != nil || parsed == nil || !parsed.Valid {
		// invalid or already expired: nothing left to revoke
		return nil
	}
	claims, ok := parsed.Claims.(gjwt.MapClaims)
	if !ok {
		return nil
	}
	var exp int64
	if v, ok := claims["exp"].(float64); ok {
		exp = int64(v)
	}
	return userService.SystemUserSVC.DenyAccessToken(ctx, token, exp)
}

func Register(conf conf.JWT) {

	var err error
	// the jwt middleware
	Security, err = jwt.New(&jwt.HertzJWTMiddleware{
		Realm:       conf.Issuer,
		Key:         []byte(conf.Secret),
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour,
		TokenLookup: consts.SecurityTokenLookupValue,
		IdentityKey: consts.SecurityIdentityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(string); ok {
				return jwt.MapClaims{
					consts.SecurityIdentityKey: v,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(ctx context.Context, c *app.RequestContext) interface{} {
			claims := jwt.ExtractClaims(ctx, c)
			return claims[consts.SecurityIdentityKey]
		},
		Authenticator: func(ctx context.Context, c *app.RequestContext) (interface{}, error) {
			var loginVals login
			if err := c.BindAndValidate(&loginVals); err != nil {
				return "", jwt.ErrMissingLoginValues
			}
			username := loginVals.Username
			password := loginVals.Password

			userId, err := userService.SystemUserSVC.Login(ctx, username, password)
			if err != nil {
				return nil, jwt.ErrFailedAuthentication
			}
			c.Set(loginUserIDKey, *userId)
			return fmt.Sprintf("%d", *userId), nil
		},
		Authorizator: func(data interface{}, ctx context.Context, c *app.RequestContext) bool {

			if v, ok := data.(string); ok {
				userId, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					panic(err)
				}
				loginUser, err := userService.SystemUserSVC.Authorizator(ctx, userId)
				if err != nil {
					return false
				}
				ctxutil.SetLoginUser(ctx, loginUser)
				c.Next(ctx)
				return true
			}

			return false
		},
		Unauthorized: func(ctx context.Context, c *app.RequestContext, code int, message string) {
			fmt.Printf("Unauthorized: %d %v\n", code, message)

			if code == http.StatusUnauthorized {
				R.Unauthorized(c, message)
				return
			}
			R.Error(ctx, c, getErrorMessage(message))
		},
		LoginResponse: func(ctx context.Context, c *app.RequestContext, code int, token string, expire time.Time) {
			expiresIn := int64(time.Until(expire).Seconds())
			if expiresIn < 0 {
				expiresIn = 0
			}

			rawId, ok := c.Get(loginUserIDKey)
			if !ok {
				R.Error(ctx, c, errorx.New(errno.ErrUserAuthenticationFailed, errorx.KV("reason", "login identity is missing")))
				return
			}
			userId, ok := rawId.(int64)
			if !ok {
				R.Error(ctx, c, errorx.New(errno.ErrUserAuthenticationFailed, errorx.KV("reason", "login identity is invalid")))
				return
			}
			tokens, err := userService.SystemUserSVC.IssueSessionTokens(ctx, userId)
			if err != nil {
				R.Error(ctx, c, err)
				return
			}

			R.SuccessWithData(ctx, c, &userModel.SessionTokenResponse{
				AccessToken:  token,
				TokenType:    "Bearer",
				ExpiresIn:    expiresIn,
				RefreshToken: tokens.RefreshToken,
			})
		},
	})
	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}

	// When you use jwt.New(), the function is already automatically called for checking,
	// which means you don't need to call it again.
	errInit := Security.MiddlewareInit()

	if errInit != nil {
		log.Fatal("authMiddleware.MiddlewareInit() Error:" + errInit.Error())
	}
}

func getErrorMessage(message string) error {
	switch message {
	case jwt.ErrMissingLoginValues.Error():
		return jwt.ErrMissingLoginValues
	case jwt.ErrFailedAuthentication.Error():
		return errorx.New(errno.ErrUserInfoInvalidateCode)
	case jwt.ErrExpiredToken.Error():
		return jwt.ErrExpiredToken
	case jwt.ErrEmptyAuthHeader.Error():
		return jwt.ErrEmptyAuthHeader
	case jwt.ErrForbidden.Error():
		return jwt.ErrForbidden
	default:
		return errors.New(message)
	}
}
