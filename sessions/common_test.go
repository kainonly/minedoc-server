package sessions_test

import (
	"bytes"
	"context"
	"github.com/bytedance/go-tagexpr/v2/binding"
	"github.com/bytedance/go-tagexpr/v2/validator"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/decoder"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/errors"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/redis/go-redis/v9"
	"github.com/weplanx/go/help"
	"github.com/weplanx/go/sessions"
	"github.com/weplanx/go/values"
	"log"
	"net/http"
	"net/url"
	"os"
	"testing"
)

var (
	service *sessions.Service
	rdb     *redis.Client
	engine  *route.Engine
)

type M = map[string]interface{}

func TestMain(m *testing.M) {
	namespace := os.Getenv("NAMESPACE")
	if err := UseRedis(); err != nil {
		log.Fatalln(err)
	}
	service = sessions.New(
		sessions.SetNamespace(namespace),
		sessions.SetRedis(rdb),
		sessions.SetDynamicValues(&values.DEFAULT),
	)

	engine = route.NewEngine(config.NewOptions([]config.Option{}))
	engine.Use(ErrHandler())
	help.RegValidate()
	help.SessionsRoutes(engine.Group(""), &sessions.Controller{Service: service})
	os.Exit(m.Run())
}

func UseRedis() (err error) {
	var opts *redis.Options
	opts, err = redis.ParseURL(os.Getenv("DATABASE_REDIS"))
	if err != nil {
		return
	}
	rdb = redis.NewClient(opts)
	return
}

func ErrHandler() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		c.Next(ctx)
		err := c.Errors.Last()
		if err == nil {
			return
		}

		if err.IsType(errors.ErrorTypePublic) {
			statusCode := http.StatusBadRequest
			result := utils.H{"message": err.Error()}
			if meta, ok := err.Meta.(map[string]interface{}); ok {
				if meta["statusCode"] != nil {
					statusCode = meta["statusCode"].(int)
				}
				if meta["code"] != nil {
					result["code"] = meta["code"]
				}
			}
			c.JSON(statusCode, result)
			return
		}

		switch e := err.Err.(type) {
		case decoder.SyntaxError:
			c.JSON(http.StatusBadRequest, utils.H{
				"message": e.Description(),
			})
			break
		case *binding.Error:
			c.JSON(http.StatusBadRequest, utils.H{
				"message": e.Error(),
			})
			break
		case *validator.Error:
			c.JSON(http.StatusBadRequest, utils.H{
				"message": e.Error(),
			})
			break
		default:
			logger.Error(err)
			c.Status(http.StatusInternalServerError)
		}
	}
}

func R(method string, url string, body interface{}) (resp *protocol.Response, err error) {
	utBody := &ut.Body{}
	utHeaders := []ut.Header{
		{Key: "content-type", Value: "application/json"},
	}
	if body != nil {
		var b []byte
		if b, err = sonic.Marshal(body); err != nil {
			return
		}
		utBody.Body = bytes.NewBuffer(b)
		utBody.Len = len(b)
	}

	w := ut.PerformRequest(engine, method, url,
		utBody,
		utHeaders...,
	)

	resp = w.Result()
	return
}

type Params = [][2]string

func U(path string, params Params) string {
	u := url.URL{Path: path}
	query := u.Query()
	for _, v := range params {
		query.Add(v[0], v[1])
	}
	u.RawQuery = query.Encode()
	return u.RequestURI()
}