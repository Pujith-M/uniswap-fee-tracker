package api

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"uniswap-fee-tracker/api/handlers"
	_ "uniswap-fee-tracker/docs" // This is required for swagger
)

type Server struct {
	router    *gin.Engine
	txHandler *handlers.TransactionHandler
}

func NewServer(txHandler *handlers.TransactionHandler) *Server {
	// Start HTTP server
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	server := &Server{
		router:    r,
		txHandler: txHandler,
	}
	return server
}

func (s *Server) RegisterRoutes() *Server {

	// Add health check endpoint
	s.router.GET("/health", func(c *gin.Context) {
		c.String(200, "Service is healthy")
	})

	// Swagger documentation
	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		v1.GET("/transactions/:txHash", s.txHandler.GetTransactionFee)
	}
	return s
}

func (s *Server) Start(addr string) error {
	return s.router.Run(addr)
}
