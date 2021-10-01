package route

import (
	"github.com/gin-gonic/gin"
	"github.com/huandu/xstrings"
	"net/http"
	"reflect"
)

type Option struct {
	Path        string
	Middlewares []MiddlewareOption
}

type OptionFunc func(*Option)

func SetPath(path string) OptionFunc {
	return func(option *Option) {
		option.Path = path
	}
}

type MiddlewareOption struct {
	Handler gin.HandlerFunc
	Effects []string
}

func SetMiddleware(middleware gin.HandlerFunc, effect ...string) OptionFunc {
	return func(option *Option) {
		option.Middlewares = append(option.Middlewares, MiddlewareOption{
			Handler: middleware,
			Effects: effect,
		})
	}
}

// Auto 生成路由
func Auto(r *gin.RouterGroup, i interface{}, options ...OptionFunc) *gin.RouterGroup {
	typ := reflect.TypeOf(i)
	val := reflect.ValueOf(i)
	opt := new(Option)
	for _, option := range options {
		option(opt)
	}
	s := r.Group(opt.Path)
	{
		scopes := make(map[string][]gin.HandlerFunc)
		for _, x := range opt.Middlewares {
			if len(x.Effects) == 0 {
				s.Use(x.Handler)
			} else {
				for _, v := range x.Effects {
					scopes[v] = append(scopes[v], x.Handler)
				}
			}
		}
		for x := 0; x < typ.NumMethod(); x++ {
			name := typ.Method(x).Name
			method := val.MethodByName(name).Interface()
			scopes[name] = append(scopes[name], Returns(method))
			s.POST(xstrings.ToSnakeCase(name), scopes[name]...)
		}
	}
	return s
}

// Returns 返回统一结果
func Returns(handlerFn interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if fn, ok := handlerFn.(func(c *gin.Context) interface{}); ok {
			switch result := fn(c).(type) {
			case string:
				c.JSON(http.StatusOK, gin.H{
					"code":    0,
					"message": result,
				})
				break
			case error:
				code, exists := c.Get("code")
				if !exists {
					code = 1
				}
				c.JSON(http.StatusOK, gin.H{
					"code":    code,
					"message": result.Error(),
				})
				break
			default:
				if result != nil {
					c.JSON(http.StatusOK, gin.H{
						"code":    0,
						"data":    result,
						"message": "ok",
					})
				} else {
					c.Status(http.StatusNotFound)
				}
			}
		}
	}
}
