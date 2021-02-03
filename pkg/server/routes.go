package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/markbates/pkger"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/c9s/bbgo/pkg/bbgo"
	"github.com/c9s/bbgo/pkg/service"
	"github.com/c9s/bbgo/pkg/types"
)

func Run(ctx context.Context, userConfig *bbgo.Config, environ *bbgo.Environment, trader *bbgo.Trader, setup bool) error {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowMethods:     []string{"GET", "POST"},
		AllowWebSockets:  true,
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/api/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	if setup {
		r.POST("/api/setup/test-db", func(c *gin.Context) {
			payload := struct {
				DSN string `json:"dsn"`
			}{}

			if err := c.BindJSON(&payload); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "missing arguments"})
				return
			}

			dsn := payload.DSN
			if len(dsn) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "missing dsn argument"})
				return
			}

			db, err := bbgo.ConnectMySQL(dsn)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if err := db.Close(); err != nil {
				logrus.WithError(err).Error("db connection close error")
			}

			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		r.POST("/api/setup/configure-db", func(c *gin.Context) {
			payload := struct {
				DSN string `json:"dsn"`
			}{}

			if err := c.BindJSON(&payload); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "missing arguments"})
				return
			}

			dsn := payload.DSN
			if len(dsn) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "missing dsn argument"})
				return
			}

			if err := environ.ConfigureDatabase(ctx, dsn); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		r.POST("/api/setup/strategy/single/:id/session/:session", func(c *gin.Context) {
			sessionName := c.Param("session")
			strategyID := c.Param("id")

			_, ok := environ.Session(sessionName)
			if !ok {
				c.JSON(http.StatusNotFound, "session not found")
				return
			}

			var conf map[string]interface{}

			if err := c.BindJSON(&conf); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "missing arguments"})
				return
			}

			strategy, err := bbgo.NewStrategyFromMap(strategyID, conf)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			mount := bbgo.ExchangeStrategyMount{
				Mounts:   []string{sessionName},
				Strategy: strategy,
			}

			userConfig.ExchangeStrategies = append(userConfig.ExchangeStrategies, mount)

			out, _ := yaml.Marshal(userConfig)
			fmt.Println(string(out))

			c.JSON(http.StatusOK, gin.H{"success": true})
		})

	}

	r.GET("/api/trades", func(c *gin.Context) {
		exchange := c.Query("exchange")
		symbol := c.Query("symbol")
		gidStr := c.DefaultQuery("gid", "0")
		lastGID, err := strconv.ParseInt(gidStr, 10, 64)
		if err != nil {
			logrus.WithError(err).Error("last gid parse error")
			c.Status(http.StatusBadRequest)
			return
		}

		trades, err := environ.TradeService.Query(service.QueryTradesOptions{
			Exchange: types.ExchangeName(exchange),
			Symbol:   symbol,
			LastGID:  lastGID,
			Ordering: "DESC",
		})
		if err != nil {
			c.Status(http.StatusBadRequest)
			logrus.WithError(err).Error("order query error")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"trades": trades,
		})
	})

	r.GET("/api/orders/closed", func(c *gin.Context) {
		exchange := c.Query("exchange")
		symbol := c.Query("symbol")
		gidStr := c.DefaultQuery("gid", "0")

		lastGID, err := strconv.ParseInt(gidStr, 10, 64)
		if err != nil {
			logrus.WithError(err).Error("last gid parse error")
			c.Status(http.StatusBadRequest)
			return
		}

		orders, err := environ.OrderService.Query(service.QueryOrdersOptions{
			Exchange: types.ExchangeName(exchange),
			Symbol:   symbol,
			LastGID:  lastGID,
			Ordering: "DESC",
		})
		if err != nil {
			c.Status(http.StatusBadRequest)
			logrus.WithError(err).Error("order query error")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"orders": orders,
		})
	})

	r.GET("/api/trading-volume", func(c *gin.Context) {
		period := c.DefaultQuery("period", "day")
		segment := c.DefaultQuery("segment", "exchange")
		startTimeStr := c.Query("start-time")

		var startTime time.Time

		if startTimeStr != "" {
			v, err := time.Parse(time.RFC3339, startTimeStr)
			if err != nil {
				c.Status(http.StatusBadRequest)
				logrus.WithError(err).Error("start-time format incorrect")
				return
			}
			startTime = v

		} else {
			switch period {
			case "day":
				startTime = time.Now().AddDate(0, 0, -30)

			case "month":
				startTime = time.Now().AddDate(0, -6, 0)

			case "year":
				startTime = time.Now().AddDate(-2, 0, 0)

			default:
				startTime = time.Now().AddDate(0, 0, -7)

			}
		}

		rows, err := environ.TradeService.QueryTradingVolume(startTime, service.TradingVolumeQueryOptions{
			SegmentBy:     segment,
			GroupByPeriod: period,
		})
		if err != nil {
			logrus.WithError(err).Error("trading volume query error")
			c.Status(http.StatusInternalServerError)
			return
		}

		c.JSON(http.StatusOK, gin.H{"tradingVolumes": rows})
		return
	})

	r.POST("/api/sessions/test", func(c *gin.Context) {
		var sessionConfig bbgo.ExchangeSession
		if err := c.BindJSON(&sessionConfig); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		session, err := bbgo.NewExchangeSessionFromConfig(sessionConfig.ExchangeName, &sessionConfig)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		var anyErr error
		_, openOrdersErr := session.Exchange.QueryOpenOrders(ctx, "BTCUSDT")
		if openOrdersErr != nil {
			anyErr = openOrdersErr
		}

		_, balanceErr := session.Exchange.QueryAccountBalances(ctx)
		if balanceErr != nil {
			anyErr = balanceErr
		}

		c.JSON(http.StatusOK, gin.H{
			"success":    anyErr == nil,
			"error":      anyErr,
			"balance":    balanceErr == nil,
			"openOrders": openOrdersErr == nil,
		})
	})

	r.GET("/api/sessions", func(c *gin.Context) {
		var sessions = []*bbgo.ExchangeSession{}
		for _, session := range environ.Sessions() {
			sessions = append(sessions, session)
		}

		c.JSON(http.StatusOK, gin.H{"sessions": sessions})
	})

	r.POST("/api/sessions", func(c *gin.Context) {
		var sessionConfig bbgo.ExchangeSession
		if err := c.BindJSON(&sessionConfig); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		session, err := bbgo.NewExchangeSessionFromConfig(sessionConfig.ExchangeName, &sessionConfig)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		if userConfig.Sessions == nil {
			userConfig.Sessions = make(map[string]*bbgo.ExchangeSession)
		}
		userConfig.Sessions[sessionConfig.Name] = session

		environ.AddExchangeSession(sessionConfig.Name, session)

		if err := session.Init(ctx, environ); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
		})
	})

	r.GET("/api/assets", func(c *gin.Context) {
		totalAssets := types.AssetMap{}

		for _, session := range environ.Sessions() {
			balances := session.Account.Balances()

			if err := session.UpdatePrices(ctx); err != nil {
				logrus.WithError(err).Error("price update failed")
				c.Status(http.StatusInternalServerError)
				return
			}

			assets := balances.Assets(session.LastPrices())

			for currency, asset := range assets {
				totalAssets[currency] = asset
			}
		}

		c.JSON(http.StatusOK, gin.H{"assets": totalAssets})
	})

	r.GET("/api/sessions/:session", func(c *gin.Context) {
		sessionName := c.Param("session")
		session, ok := environ.Session(sessionName)

		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("session %s not found", sessionName)})
			return
		}

		c.JSON(http.StatusOK, gin.H{"session": session})
	})

	r.GET("/api/sessions/:session/trades", func(c *gin.Context) {
		sessionName := c.Param("session")
		session, ok := environ.Session(sessionName)

		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("session %s not found", sessionName)})
			return
		}

		c.JSON(http.StatusOK, gin.H{"trades": session.Trades})
	})

	r.GET("/api/sessions/:session/open-orders", func(c *gin.Context) {
		sessionName := c.Param("session")
		session, ok := environ.Session(sessionName)

		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("session %s not found", sessionName)})
			return
		}

		marketOrders := make(map[string][]types.Order)
		for symbol, orderStore := range session.OrderStores() {
			marketOrders[symbol] = orderStore.Orders()
		}

		c.JSON(http.StatusOK, gin.H{"orders": marketOrders})
	})

	r.GET("/api/sessions/:session/account", func(c *gin.Context) {
		sessionName := c.Param("session")
		session, ok := environ.Session(sessionName)

		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("session %s not found", sessionName)})
			return
		}

		c.JSON(http.StatusOK, gin.H{"account": session.Account})
	})

	r.GET("/api/sessions/:session/account/balances", func(c *gin.Context) {
		sessionName := c.Param("session")
		session, ok := environ.Session(sessionName)

		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("session %s not found", sessionName)})
			return
		}

		if session.Account == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("the account of session %s is nil", sessionName)})
			return
		}

		c.JSON(http.StatusOK, gin.H{"balances": session.Account.Balances()})
	})

	r.GET("/api/sessions/:session/symbols", func(c *gin.Context) {
		sessionName := c.Param("session")
		session, ok := environ.Session(sessionName)

		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("session %s not found", sessionName)})
			return
		}

		var symbols []string
		for symbol := range session.Markets() {
			symbols = append(symbols, symbol)
		}

		c.JSON(http.StatusOK, gin.H{"symbols": symbols})
	})

	r.GET("/api/sessions/:session/pnl", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	r.GET("/api/sessions/:session/market/:symbol/open-orders", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	r.GET("/api/sessions/:session/market/:symbol/trades", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	r.GET("/api/sessions/:session/market/:symbol/pnl", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	fs := pkger.Dir("/frontend/out")
	r.NoRoute(func(c *gin.Context) {
		http.FileServer(fs).ServeHTTP(c.Writer, c.Request)
	})

	return r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}