package web

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/dimuls/graph/entity"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
)

type Storage interface {
	Graph(graphID int64) (entity.Graph, error)
	Graphs() ([]entity.Graph, error)
	AddGraph(g entity.Graph) (int64, error)
	RemoveGraph(graphID int64) error

	Vertex(vertexID int64) (entity.Vertex, error)
	Vertexes(graphID int64) ([]entity.Vertex, error)
	AddVertex(v entity.Vertex) (int64, error)
	SetVertex(v entity.Vertex) error
	RemoveVertex(vertexID int64) error

	Edge(edgeID int64) (entity.Edge, error)
	Edges(graphID int64) ([]entity.Edge, error)
	AddEdge(e entity.Edge) (int64, error)
	SetEdge(e entity.Edge) error
	RemoveEdge(edgeID int64) error
}

type Server struct {
	bindAddr string
	storage  Storage
	echo     *echo.Echo
	log      *logrus.Entry
	wg       sync.WaitGroup
	stop     chan struct{}

	graphListeners   map[int64]map[*websocket.Conn]chan struct{}
	graphListenersMx sync.RWMutex
}

func NewServer(bindAddr string, s Storage) *Server {
	return &Server{
		bindAddr:       bindAddr,
		storage:        s,
		log:            logrus.WithField("subsystem", "web_server"),
		graphListeners: map[int64]map[*websocket.Conn]chan struct{}{},
	}
}

func (s *Server) Start() {
	s.log.WithField("bind_addr", s.bindAddr).Info("starting")

	e := echo.New()

	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recover())
	e.Use(logrusLogger)

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		var (
			code = http.StatusInternalServerError
			msg  interface{}
		)

		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
			msg = he.Message
		} else if e.Debug {
			msg = err.Error()
		} else {
			msg = http.StatusText(code)
		}
		if _, ok := msg.(string); !ok {
			msg = fmt.Sprintf("%v", msg)
		}

		// Send response
		if !c.Response().Committed {
			if c.Request().Method == http.MethodHead { // Issue #608
				err = c.NoContent(code)
			} else {
				err = c.String(code, msg.(string))
			}
			if err != nil {
				s.log.WithError(err).Error("failed to error response")
			}
		}
	}

	e.GET("/", s.getIndex)
	e.GET("/graphs/:graph_id", s.getGraph)

	api := e.Group("/api")

	api.GET("/graphs", s.getAPIGraphs)
	api.POST("/graphs", s.postAPIGraphs)
	api.GET("/graphs/:graph_id", s.getAPIGraph)
	api.DELETE("/graphs/:graph_id", s.deleteAPIGraph)
	api.GET("/graphs/:graph_id/shortest-path", s.getAPIGraphShortestPath)

	api.POST("/vertexes", s.postAPIVertexes)
	api.PUT("/vertexes", s.putAPIVertexes)
	api.DELETE("/vertexes/:vertex_id", s.deleteAPIVertex)

	api.POST("/edges", s.postAPIEdges)
	api.PUT("/edges", s.putAPIEdges)
	api.DELETE("/edges/:edge_id", s.deleteAPIEdge)

	s.echo = e
	s.stop = make(chan struct{})

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			err := s.echo.Start(s.bindAddr)
			if err != nil {
				if err == http.ErrServerClosed {
					s.log.Info("server is closed")
					return
				}
				s.log.WithError(err).Error("failed to start")
				time.Sleep(3 * time.Second)
			}
		}
	}()
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	close(s.stop)

	err := s.echo.Shutdown(ctx)
	if err != nil {
		s.log.WithError(err).Error("failed to graceful shutdown")
	}

	s.wg.Wait()
}

func logrusLogger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		err := next(c)

		stop := time.Now()

		if err != nil {
			c.Error(err)
		}

		req := c.Request()
		res := c.Response()

		p := req.URL.Path
		if p == "" {
			p = "/"
		}

		bytesIn := req.Header.Get(echo.HeaderContentLength)
		if bytesIn == "" {
			bytesIn = "0"
		}

		entry := logrus.WithFields(map[string]interface{}{
			"subsystem":    "web_server",
			"remote_ip":    c.RealIP(),
			"host":         req.Host,
			"query_params": c.QueryParams(),
			"uri":          req.RequestURI,
			"method":       req.Method,
			"path":         p,
			"referer":      req.Referer(),
			"user_agent":   req.UserAgent(),
			"status":       res.Status,
			"latency":      stop.Sub(start).String(),
			"bytes_in":     bytesIn,
			"bytes_out":    strconv.FormatInt(res.Size, 10),
		})

		const msg = "request handled"

		if res.Status >= 500 {
			if err != nil {
				entry = entry.WithError(err)
			}
			entry.Error(msg)
		} else if res.Status >= 400 {
			entry.Warn(msg)
		} else {
			entry.Info(msg)
		}

		return nil
	}
}
