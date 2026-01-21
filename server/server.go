package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"myblog-gogogo/config"

	"github.com/quic-go/quic-go/http3"
)

// Server HTTP 服务器
type Server struct {
	httpServer  *http.Server
	http3Server *http3.Server
	config      *config.Config
}

// New 创建新的服务器实例
func New(handler http.Handler, cfg *config.Config) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         ":" + cfg.Port,
			Handler:      handler,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
			ErrorLog:     log.New(os.Stderr, "HTTP: ", log.LstdFlags),
		},
		config: cfg,
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	if s.config.EnableTLS && s.config.TLSCert != "" && s.config.TLSKey != "" {
		return s.startTLS()
	}
	return s.startHTTP()
}

// startHTTP 启动 HTTP 服务器
func (s *Server) startHTTP() error {
	log.Printf("HTTP server listening on port %s (HTTP1.1 only)", s.config.Port)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("HTTP server error: %v", err)
	}
	return nil
}

// startTLS 启动 HTTPS/HTTP3 服务器
func (s *Server) startTLS() error {
	// 加载TLS证书
	cert, err := tls.LoadX509KeyPair(s.config.TLSCert, s.config.TLSKey)
	if err != nil {
		return fmt.Errorf("failed to load TLS certificate: %v", err)
	}

	// 配置TLS，支持HTTP2和HTTP3
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"h3", "h2", "http/1.1"}, // 优先HTTP3，回退到HTTP2，最后HTTP1.1
		MinVersion:   tls.VersionTLS12,
	}

	// 创建HTTP3服务器
	s.http3Server = &http3.Server{
		Addr:      ":" + s.config.Port,
		Handler:   s.httpServer.Handler,
		TLSConfig: tlsConfig,
	}

	// 启动HTTP3服务器
	http3ErrChan := make(chan error, 1)
	go func() {
		log.Printf("HTTP3 server listening on port %s (QUIC)", s.config.Port)
		if err := s.http3Server.ListenAndServe(); err != nil {
			log.Printf("HTTP3 server error: %v", err)
			http3ErrChan <- err
		}
	}()

	// 创建HTTP2/HTTP1.1服务器（作为回退）
	s.httpServer.TLSConfig = tlsConfig

	// 启动HTTP2/HTTP1.1服务器
	httpErrChan := make(chan error, 1)
	go func() {
		log.Printf("HTTP2/HTTP1.1 server listening on port %s (TCP)", s.config.Port)
		if err := s.httpServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
			httpErrChan <- err
		}
	}()

	// 等待任一服务器启动成功
	select {
	case <-time.After(500 * time.Millisecond):
		log.Printf("All servers started successfully")
	case err := <-http3ErrChan:
		if err != nil {
			log.Printf("HTTP3 failed to start, falling back to HTTP2/HTTP1.1: %v", err)
		}
	case err := <-httpErrChan:
		if err != nil {
			return fmt.Errorf("HTTP2/HTTP1.1 failed to start: %v", err)
		}
	}

	return nil
}

// Shutdown 优雅关闭服务器
func (s *Server) Shutdown() error {
	log.Println("Received shutdown signal")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 关闭HTTP3服务器
	if s.http3Server != nil {
		if err := s.http3Server.Close(); err != nil {
			log.Printf("HTTP3 server Shutdown error: %v", err)
		}
	}

	// 关闭HTTP服务器
	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server Shutdown error: %v", err)
		return err
	}

	log.Println("Server shutdown complete")
	return nil
}

// WaitForShutdown 等待关闭信号
func (s *Server) WaitForShutdown() {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
	<-sigint
}

// GracefulShutdown 优雅关闭服务器
func GracefulShutdown(srv *Server) error {
	srv.WaitForShutdown()
	return srv.Shutdown()
}