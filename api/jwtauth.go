package api

// Orginal work from : github.com/goware/jwtauth
// Corrected a bit according to new jwt-go api and use my own returns

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

var (
	jwtErrorKey key = "jwt.err"
	jwtTokenKey key = "jwt"
	// ErrUnauthorized unauhorized token error
	ErrUnauthorized = errors.New("jwtauth: unauthorized token")
	// ErrExpired expired token error
	ErrExpired = errors.New("jwtauth: expired token")
)

// JwtAuth struct to store JWT auth informations
type JwtAuth struct {
	signKey   []byte
	verifyKey []byte
	signer    jwt.SigningMethod
	parser    *jwt.Parser
}

// New creates a JwtAuth authenticator instance that provides middleware handlers
// and encoding/decoding functions for JWT signing.
func New(alg string, signKey []byte, verifyKey []byte) *JwtAuth {
	return &JwtAuth{
		signKey:   signKey,
		verifyKey: verifyKey,
		signer:    jwt.GetSigningMethod(alg),
	}
}

// NewWithParser is the same as New, except it supports custom parser settings
// introduced in ver. 2.4.0 of jwt-go
func NewWithParser(alg string, parser *jwt.Parser, signKey []byte, verifyKey []byte) *JwtAuth {
	return &JwtAuth{
		signKey:   signKey,
		verifyKey: verifyKey,
		signer:    jwt.GetSigningMethod(alg),
		parser:    parser,
	}
}

// Verifier middleware will verify a JWT passed by a client request.
// The Verifier will look for a JWT token from:
// 1. 'jwt' URI query parameter
// 2. 'Authorization: BEARER T' request header
// 3. Cookie 'jwt' value
//
// The verification processes finishes here and sets the token and
// a error in the request context and calls the next handler.
//
// Make sure to have your own handler following the Validator that
// will check the value of the "jwt" and "jwt.err" in the context
// and respond to the client accordingly. A generic Authenticator
// middleware is provided by this package, that will return a 401
// message for all unverified tokens, see jwtauth.Authenticator.
func (ja *JwtAuth) Verifier(next http.Handler) http.Handler {
	return ja.Verify("")(next)
}

// Verify get token from context
func (ja *JwtAuth) Verify(paramAliases ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			var tokenStr string
			var err error

			// Get token from query params
			tokenStr = r.URL.Query().Get("jwt")

			// Get token from other query param aliases
			if tokenStr == "" && paramAliases != nil && len(paramAliases) > 0 {
				for _, p := range paramAliases {
					tokenStr = r.URL.Query().Get(p)
					if tokenStr != "" {
						break
					}
				}
			}

			// Get token from authorization header
			if tokenStr == "" {
				bearer := r.Header.Get("Authorization")
				if len(bearer) > 7 && strings.ToUpper(bearer[0:6]) == "BEARER" {
					tokenStr = bearer[7:]
				}
			}

			// Get token from cookie
			if tokenStr == "" {
				cookie, err := r.Cookie("jwt")
				if err == nil {
					tokenStr = cookie.Value
				}
			}

			// Token is required, cya
			if tokenStr == "" {
				err = ErrUnauthorized
			}

			// Verify the token
			token, err := ja.Decode(tokenStr)
			if err != nil {
				switch err.Error() {
				case "token is expired":
					err = ErrExpired
				}

				ctx = ja.SetContext(ctx, token, err)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			if token == nil || !token.Valid || token.Method != ja.signer {
				err = ErrUnauthorized
				ctx = ja.SetContext(ctx, token, err)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Check expiry via "exp" claim
			if ja.IsExpired(token) {
				err = ErrExpired
				ctx = ja.SetContext(ctx, token, err)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Valid! pass it down the context to an authenticator middleware
			ctx = ja.SetContext(ctx, token, err)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(hfn)
	}
}

// SetContext set jwt && jwt.err in context
func (ja *JwtAuth) SetContext(ctx context.Context, t *jwt.Token, err error) context.Context {
	ctx = context.WithValue(ctx, jwtTokenKey, t)
	ctx = context.WithValue(ctx, jwtErrorKey, err)
	return ctx
}

// Encode encode claims
func (ja *JwtAuth) Encode(claims Claims) (t *jwt.Token, tokenString string, err error) {
	t = jwt.New(ja.signer)
	t.Claims = claims
	tokenString, err = t.SignedString(ja.signKey)
	t.Raw = tokenString
	return
}

// Decode decode claims
func (ja *JwtAuth) Decode(tokenString string) (t *jwt.Token, err error) {
	if ja.parser != nil {
		return ja.parser.Parse(tokenString, ja.keyFunc)
	}
	return jwt.Parse(tokenString, ja.keyFunc)
}

func (ja *JwtAuth) keyFunc(t *jwt.Token) (interface{}, error) {
	if ja.verifyKey != nil && len(ja.verifyKey) > 0 {
		return ja.verifyKey, nil
	}
	return ja.signKey, nil

}

// IsExpired check if token is expired
func (ja *JwtAuth) IsExpired(t *jwt.Token) bool {
	if expv, ok := t.Claims.(jwt.MapClaims)["exp"]; ok {
		var exp int64
		switch v := expv.(type) {
		case float64:
			exp = int64(v)
		case int64:
			exp = v
		case json.Number:
			exp, _ = v.Int64()
		default:
		}

		if exp < EpochNow() {
			return true
		}
	}

	return false
}

// Authenticator validate that user has a valid user auth token before letting him access.
func Authenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if jwtErr, ok := ctx.Value(jwtErrorKey).(error); ok {
			if jwtErr != nil {
				render.JSON(w, 401, "Token not found. You Are not allowed to proceed without token.")
				return
			}
		}

		jwtToken, ok := ctx.Value(jwtTokenKey).(*jwt.Token)
		if !ok || jwtToken == nil || !jwtToken.Valid {
			render.JSON(w, 401, "token is not valid or does not exist")
			return
		}

		tokenType, ok := jwtToken.Claims.(jwt.MapClaims)["type"]

		if !ok {
			render.JSON(w, 401, "Token is not valid. Type is undifined")
			return
		}

		if tokenType != "userauth" {
			render.JSON(w, 401, "Token is not an user auth one")
			return
		}

		// Token is authenticated, pass it through
		next.ServeHTTP(w, r)
	})
}

// allowUserCreationFromToken check the provided token is an invitation one
func allowUserCreationFromToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if jwtErr, ok := ctx.Value(jwtErrorKey).(error); ok {
			if jwtErr != nil {
				render.JSON(w, 401, "Token not found. You Are not allowed to proceed without token.")
				return
			}
		}

		jwtToken, ok := ctx.Value(jwtTokenKey).(*jwt.Token)
		if !ok || jwtToken == nil || !jwtToken.Valid {
			render.JSON(w, 401, "token is not valid or does not exist")
			return
		}

		tokenType, ok := jwtToken.Claims.(jwt.MapClaims)["type"]

		if !ok {
			render.JSON(w, 401, "Token is not valid. Type is undifined")
			return
		}

		if tokenType != "invitation" {
			render.JSON(w, 401, "Token is not an invitation one")
			return
		}

		// tokenOrganisation, ok := jwtToken.Claims.(jwt.MapClaims)["organisation"].(string)

		// if !ok {
		// render.JSON(w, 401, "Token is not valid. Organisation is undifined")
		// return
		// }
		// apiOrganisation := datastores.Store().Organisation().Get(dbStore.db)

		// if tokenOrganisation != apiOrganisation.OrganisationName {
		// 	render.JSON(w, 401, "Token is not valid. Organisation does not match current organsisation")
		// 	return
		// }
		// Token is authenticated, pass it through
		next.ServeHTTP(w, r)
	})
}

// Claims is a convenience type to manage a JWT claims hash.
type Claims map[string]interface{}

// Valid check if claims map is valid
func (c Claims) Valid() error {
	return jwt.MapClaims(c).Valid()
}

// Set set claim value
func (c Claims) Set(k string, v interface{}) Claims {
	c[k] = v
	return c
}

// Get get claim value
func (c Claims) Get(k string) (interface{}, bool) {
	v, ok := c[k]
	return v, ok
}

// SetIssuedAt Set issued at ("iat") to specified time in the claims
func (c Claims) SetIssuedAt(tm time.Time) Claims {
	c["iat"] = tm.UTC().Unix()
	return c
}

// SetIssuedNow Set issued at ("iat") to present time in the claims
func (c Claims) SetIssuedNow() Claims {
	c["iat"] = EpochNow()
	return c
}

// SetExpiry Set expiry ("exp") in the claims and return itself so it can be chained
func (c Claims) SetExpiry(tm time.Time) Claims {
	c["exp"] = tm.UTC().Unix()
	return c
}

// SetExpiryIn Set expiry ("exp") in the claims to some duration from the present time
// and return itself so it can be chained
func (c Claims) SetExpiryIn(tm time.Duration) Claims {
	c["exp"] = ExpireIn(tm)
	return c
}

// EpochNow Helper function that returns the NumericDate time value used by the spec
func EpochNow() int64 {
	return time.Now().UTC().Unix()
}

// ExpireIn Helper function to return calculated time in the future for "exp" claim.
func ExpireIn(tm time.Duration) int64 {
	return EpochNow() + int64(tm.Seconds())
}
