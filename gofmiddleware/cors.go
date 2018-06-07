/*******************************************************************************
 * Copyright (c) 2018  charles
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NON INFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 * -------------------------------------------------------------------------
 * created at 2018-06-04 17:59:04
 ******************************************************************************/

package gofmiddleware

import (
	"net/http"
	"strconv"
	"strings"

	"gitee.com/goframe/gof/gof-errors"
	"github.com/gin-gonic/gin"
	"github.com/labstack/echo"
	"github.com/spf13/viper"
)

type (
	// CORSConfig defines the config for CORS middleware.
	CORSConfig struct {
		// AllowOrigin defines a list of origins that may access the resource.
		// Optional. Default value []string{"*"}.
		AllowOrigins []string `mapstructure:"allow_origins" yaml:"allow_origins"`

		// AllowMethods defines a list methods allowed when accessing the resource.
		// This is used in response to a preflight request.
		// Optional. Default value DefaultCORSConfig.AllowMethods.
		AllowMethods []string `mapstructure:"allow_methods" yaml:"allow_methods"`

		// AllowHeaders defines a list of request headers that can be used when
		// making the actual request. This in response to a preflight request.
		// Optional. Default value []string{}.
		AllowHeaders []string `mapstructure:"allow_headers" yaml:"allow_headers"`

		// AllowCredentials indicates whether or not the response to the request
		// can be exposed when the credentials flag is true. When used as part of
		// a response to a preflight request, this indicates whether or not the
		// actual request can be made using credentials.
		// Optional. Default value false.
		AllowCredentials bool `mapstructure:"allow_credentials" yaml:"allow_credentials"`

		// ExposeHeaders defines a whitelist headers that clients are allowed to
		// access.
		// Optional. Default value []string{}.
		ExposeHeaders []string `mapstructure:"expose_headers" yaml:"expose_headers"`

		// MaxAge indicates how long (in seconds) the results of a preflight request
		// can be cached.
		// Optional. Default value 0.
		MaxAge int `mapstructure:"max_age" yaml:"max_age"`
	}
)

//ReadIn ...
func (p *CORSConfig) ReadIn() error {
	key := "cors"
	if viper.IsSet(key) {
		return viper.UnmarshalKey(key, p)
	}
	viper.Set(key, DefaultCORSConfig)
	return goferrors.ErrInitConfig
}

var (
	// DefaultCORSConfig is the default CORS middleware config.
	DefaultCORSConfig = CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE, echo.HeaderAuthorization},
		AllowHeaders:     []string{},
		AllowCredentials: true,
		MaxAge:           172800,
	}
)

func CORS(configs ...CORSConfig) gin.HandlerFunc {
	var config CORSConfig
	if len(configs) == 0 {
		config = DefaultCORSConfig
	} else {
		config = configs[0]
	}
	if len(config.AllowOrigins) == 0 {
		config.AllowOrigins = DefaultCORSConfig.AllowOrigins
	}
	if len(config.AllowMethods) == 0 {
		config.AllowMethods = DefaultCORSConfig.AllowMethods
	}
	allowMethods := strings.Join(config.AllowMethods, ",")
	allowHeaders := strings.Join(config.AllowHeaders, ",")
	exposeHeaders := strings.Join(config.ExposeHeaders, ",")
	maxAge := strconv.Itoa(config.MaxAge)
	return func(c *gin.Context) {
		req := c.Request
		res := c.Writer
		origin := req.Header.Get(echo.HeaderOrigin)
		allowOrigin := ""
		// Check allowed origins
		for _, o := range config.AllowOrigins {
			if o == "*" || o == origin {
				allowOrigin = origin
				break
			}
		}
		// Simple request
		if req.Method != echo.OPTIONS {
			res.Header().Add(echo.HeaderVary, echo.HeaderOrigin)
			res.Header().Set(echo.HeaderAccessControlAllowOrigin, allowOrigin)
			if config.AllowCredentials {
				res.Header().Set(echo.HeaderAccessControlAllowCredentials, "true")
			}
			if exposeHeaders != "" {
				res.Header().Set(echo.HeaderAccessControlExposeHeaders, exposeHeaders)
			}
			c.Next()
			return
		}
		// Preflight request
		res.Header().Add(echo.HeaderVary, echo.HeaderOrigin)
		res.Header().Add(echo.HeaderVary, echo.HeaderAccessControlRequestMethod)
		res.Header().Add(echo.HeaderVary, echo.HeaderAccessControlRequestHeaders)
		res.Header().Set(echo.HeaderAccessControlAllowOrigin, allowOrigin)
		res.Header().Set(echo.HeaderAccessControlAllowMethods, allowMethods)
		if config.AllowCredentials {
			res.Header().Set(echo.HeaderAccessControlAllowCredentials, "true")
		}
		if allowHeaders != "" {
			res.Header().Set(echo.HeaderAccessControlAllowHeaders, allowHeaders)
		} else {
			h := req.Header.Get(echo.HeaderAccessControlRequestHeaders)
			if h != "" {
				res.Header().Set(echo.HeaderAccessControlAllowHeaders, h)
			}
		}
		if config.MaxAge > 0 {
			res.Header().Set(echo.HeaderAccessControlMaxAge, maxAge)
		}
		c.Writer.WriteHeader(http.StatusNoContent)
	}
}
