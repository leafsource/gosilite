package consts

const (
	RequestSchemeKeyInCtx = "REQUEST_SCHEME_IN_CTX"
	HostKeyInCtx          = "HOST_KEY_IN_CTX"
	LoginUserDataCtx      = "LOGIN_USER_CTX"

	RequestHeaderIdKey  = "x-request-id"
	SecurityIdentityKey = "security_identity_id"

	SecurityTokenLookupValue = "header: Authorization"

	AuthRefreshTokenPrefix           = "gosilite:auth:refresh:"
	AuthUserRefreshGenerationPrefix  = "gosilite:auth:user_refresh_gen:"
	AuthAccessTokenDenyPrefix        = "gosilite:auth:access_deny:"
	AuthBearerPrefix                 = "Bearer "
)
