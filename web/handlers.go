package web

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/dimuls/graph/dijkstra"
	"github.com/dimuls/graph/entity"
	"github.com/labstack/echo"
	"golang.org/x/net/websocket"
)

// language=html
const indexPage = `<!DOCTYPE html>
<html>
	<head>
		<title>graph</title>
		<link rel="stylesheet"
			href="https://cdnjs.cloudflare.com/ajax/libs/normalize/8.0.1/normalize.min.css"
			integrity="sha256-l85OmPOjvil/SOvVt3HnSSjzF1TUMyT9eV0c2BzEGzU="
			crossorigin="anonymous" />
		<style>
			body {
				padding: 0 1em 1em 1em;
			}
			.graph {
				margin-bottom: 0.5em;
			}
			.graph:last-child {
				margin-bottom: 0;
			}
			.new-graph {
				margin-top: 1em;
			}
		</style>
	</head>
	<body>
		<h1>graph</h1>
		<div data-bind="visible: false">
			<i>Loading...</i>
		</div>
		<div style="display: none" data-bind="visible: true">
			<div data-bind="foreach: graphs">
				<div class="graph">
					<span data-bind="text: id"></span>.
					<a data-bind="text: name, attr: { href: '/graphs/'+id }"></a>
					<button data-bind="click: $root.removeGraph">remove</button>
				</div>
			</div>
			<div class="new-graph">
				<input placeholder="new graph name" data-bind="textInput: newGraphName"/>
				<button data-bind="click: addGraph">add</button>
			</div>
		</div>
		<script
			src="https://code.jquery.com/jquery-3.4.1.min.js"
			integrity="sha256-CSXorXvZcTkaix6Yvo6HppcZGetbYMGWSFlBw8HfCJo="
			crossorigin="anonymous"></script>
		<script
			src="https://cdnjs.cloudflare.com/ajax/libs/knockout/3.5.0/knockout-min.js"
			integrity="sha256-Tjl7WVgF1hgGMgUKZZfzmxOrtoSf8qltZ9wMujjGNQk="
			crossorigin="anonymous"></script>
		<script>
		$(function() {
		    $.get('/api/graphs', function(graphs) {
		        app = {
		        	graphs: ko.observableArray(graphs),
		        	removeGraph: function(graph) {
		        	    $.ajax({
		        	    	url: '/api/graphs/'+graph.id,
		        	    	type: 'DELETE',
		        	    	success: function() {
								app.graphs.remove(graph);		        	    	    
		        	    	}
		        	    })
		        	},
		        	newGraphName: ko.observable(''),
		        	addGraph: function() {
		        	    $.ajax({
		        	    	url: '/api/graphs',
		        	    	type: 'POST',
		        	    	contentType: 'application/json',
		        	    	data: JSON.stringify({
		        	    		name: app.newGraphName(),
		        	    	}),
		        	    	success: function(id) {
		        	    	    app.graphs.push({
		        	    	    	id: id,
		        	    	    	name: app.newGraphName(),
		        	    	    });
		        	    	    app.newGraphName('')
		        	     	}
		        	    });
		        	}	        	
		        };
		        ko.applyBindings(app)
		    });
		});	
		</script>
	</body>
</html>`

func (s *Server) getIndex(c echo.Context) error {
	return c.HTML(http.StatusOK, indexPage)
}

// language=html
const graphPage = `<!DOCTYPE>
<html>
	<head>
		<title>graph</title>
		<style>
			html, body {
				margin: 0;
				padding: 0;
				width: 100%;
				height: 100%;
			}
			.vis-close {
				display: none !important;
			}
		</style>
	</head>
	<body>
		<div id="graph"></div>
		<script
			src="https://code.jquery.com/jquery-3.4.1.min.js"
			integrity="sha256-CSXorXvZcTkaix6Yvo6HppcZGetbYMGWSFlBw8HfCJo="
			crossorigin="anonymous"></script>
		<script src="https://unpkg.com/vis-network/standalone/umd/vis-network.min.js"></script>
		<script>
		$(function() {
		    var graphID = parseInt(window.location.pathname.split('/')[2]);
		    var graph;
		    var data;
		    
		    function initGraph(graph) {
		        var nodes = new vis.DataSet(graph.vertexes.map(function(v) {
		            return {
		                id: v.id,
		                x: v.x,
		                y: v.y,
		                physics: false
		            };
		        }));
		        
		        var edges = new vis.DataSet(graph.edges.map(function(e) {
		            return {
		                id: e.id,
		                from: e.from,
		                to: e.to,
		                label: e.weight.toString(),
		                arrows: 'to',
		            };
		        }));
		        
		        var container = document.getElementById('graph');
		        
		        data = {
		            nodes: nodes,
		            edges: edges,
		        };
		        
		        var options = {
		            autoResize: true,
					height: '100%',
					width: '100%',
					interaction: {
		              	multiselect: true,
		              	selectConnectedEdges: false,  
					},
					manipulation: {
		                enabled: true,
		                initiallyActive: true,
		                addNode: function(node, callback) {
							$.ajax({
								url: '/api/vertexes',
								type: 'POST',
								contentType: 'application/json',
								data: JSON.stringify({
									graph_id: graphID,
									x: node.x,
									y: node.y							
								})
							});
		                	callback(null);
		                },
						addEdge: function(edge, callback) {
		                    var weightStr = prompt('enter edge weight');
		                    var weight = parseFloat(weightStr);
		                    if (weight !== 0 && !weight) {
		                        alert('invalid weight: '+weightStr);
		                    } else {
								$.ajax({
									url: '/api/edges',
									type: 'POST',
									contentType: 'application/json',
									data: JSON.stringify({
										graph_id: graphID,
										from: edge.from,
										to: edge.to,
										weight: weight							
									})
								});
		                	}
		                	callback(null);
		                },
						deleteNode: function(params, callback) {
		                    params.nodes.forEach(function(nodeID) {
		                      	$.ajax({
									url: '/api/vertexes/'+nodeID,
									type: 'DELETE'
								});
		                    });
		                	callback(null);
		                },
						deleteEdge: function(params, callback) {
		                    params.edges.forEach(function(edgeID) {
		                    	$.ajax({
									url: '/api/edges/'+edgeID,
									type: 'DELETE'
								});
		                    });
		                	callback(null);
		                },
						editEdge: false,
				  	}
		        };
		        
		        graph = new vis.Network(container, data, options);
		        
		        graph.on('dragEnd', function(params) {
		            if (params.nodes.length === 1) {
						var node = nodes.get(params.nodes[0]);
		                var positions = graph.getPositions([node.id]);
		                $.ajax({
							url: '/api/vertexes',
							type: 'PUT',
							contentType: 'application/json',
							data: JSON.stringify({
								id: node.id,
								graph_id: graphID,
								x: positions[node.id].x,
								y: positions[node.id].y							
							})
						});		                
		            }
		        });
		        
		        graph.on('doubleClick', function(params) {
		            if (params.nodes.length === 0 && params.edges.length === 1) {
		                var edge = edges.get(params.edges[0]);
		            	var weightStr = prompt('enter edge weight');
						var weight = parseFloat(weightStr);
						if (weight !== 0 && !weight) {
							alert('invalid weight: '+weightStr);
							return
						}
		            	$.ajax({
							url: '/api/edges',
							type: 'PUT',
							contentType: 'application/json',
							data: JSON.stringify({
								id: edge.id,
								graph_id: graphID,
								weight: weight							
							})
						});
					}
		        });
		        
		        var from;
		        
		        graph.on('selectNode', function(params) {
		            if (params.nodes.length === 1) {
		                from = params.nodes[0];
		                return
		            }
		            
		            var to;
		            params.nodes.forEach(function(id) {
		              	if (id !== from) {
		                    to = id;
		                }
		            });
		            
		            graph.unselectAll();
		            
		            $.ajax({
		            	url: "/api/graphs/"+graphID+"/shortest-path",
		            	type: "GET",
		            	data: { from: from, to: to },
		            	success: function(edges) {
		            		graph.selectEdges(edges);
		            	}
		            });
		        });
		    }
		    
		    function connect() {
				var ws = new WebSocket('ws://localhost:8080/api/graphs/'+graphID);
				
				ws.onmessage = function(e) {
					var msg = JSON.parse(e.data);
					switch (msg.type) {
					 	case 'set-graph':
					 	    initGraph(msg.data);
					    	break;
					    case 'graph-removed':
					        window.location.href = "/";
					        break;
						case 'new-vertex':
						    data.nodes.add([{
								id: msg.data.id,
								x: msg.data.x,
								y: msg.data.y,
								physics: false
							}]);
						    break;
						case 'vertex-update':
						    data.nodes.update({
						    	id: msg.data.id,
						    	x: msg.data.x,
								y: msg.data.y
						    });
						    break;
						case 'vertex-removed':
						    data.nodes.remove(msg.data.id);
						    break;
						case 'new-edge':
						    data.edges.add([{
								id: msg.data.id,
								from: msg.data.from,
								to: msg.data.to,
								label: msg.data.weight.toString(),
								arrows: 'to',
							}]);
						    break;
						case 'edge-update':
						    data.edges.update({
						    	id: msg.data.id,
						    	label: msg.data.weight.toString()
						    });
						    break;
						case 'edge-removed':
						    data.edges.remove(msg.data.id);
						    break;
						default:
					    	console.warn('unknown message: ', msg);
					    	break;
					}
				};
				
				ws.onclose = function(e) {
				    if (graph) {
				    	graph.destroy();
				    }
					console.log('Socket is closed. Reconnect will be attempted in 1 second.', e.reason);
					setTimeout(connect, 1000);
				};
				
				ws.onerror = function(err) {
					console.error('Socket encountered error: ', err.message, 'Closing socket');
					ws.close();
				};
			}
				
			connect();
		});
		</script>
	</body>
</html>`

func (s *Server) getGraph(c echo.Context) error {
	return c.HTML(http.StatusOK, graphPage)
}

func (s *Server) getAPIGraphs(c echo.Context) error {
	gs, err := s.storage.Graphs()
	if err != nil {
		return fmt.Errorf("get graphs from storage: %w", err)
	}

	if gs == nil {
		gs = []entity.Graph{}
	}

	return c.JSON(http.StatusOK, gs)
}

func (s *Server) postAPIGraphs(c echo.Context) error {
	var g entity.Graph

	err := c.Bind(&g)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest,
			"bind graph: "+err.Error())
	}

	if g.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest,
			"empty name")
	}

	id, err := s.storage.AddGraph(g)
	if err != nil {
		if err == entity.ErrDuplicatedGraphName {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}
		return fmt.Errorf("add graph to storage: %w", err)
	}

	return c.JSON(http.StatusCreated, id)
}

func (s *Server) getAPIGraph(c echo.Context) error {
	graphID, err := strconv.ParseInt(c.Param("graph_id"),
		10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest,
			"invalid graph_id")
	}

	g, err := s.storage.Graph(graphID)
	if err != nil {
		if err == entity.ErrGraphNotFound {
			return echo.NewHTTPError(http.StatusNotFound, err)
		}
		return fmt.Errorf("get graph from storage: %w", err)
	}

	websocket.Handler(func(ws *websocket.Conn) {
		s.wg.Add(1)
		defer s.wg.Done()

		closeWS := make(chan struct{})

		s.graphListenersMx.Lock()
		if _, exists := s.graphListeners[graphID]; !exists {
			s.graphListeners[graphID] = map[*websocket.Conn]chan struct{}{}
		}
		s.graphListeners[graphID][ws] = closeWS
		s.graphListenersMx.Unlock()

		defer func() {
			s.graphListenersMx.Lock()
			delete(s.graphListeners[graphID], ws)
			if len(s.graphListeners[graphID]) == 0 {
				delete(s.graphListeners, graphID)
			}
			s.graphListenersMx.Unlock()
			err = ws.Close()
			if err != nil {
				s.log.WithError(err).Error(
					"failed to close websocket connection")
			}
		}()

		vs, err := s.storage.Vertexes(graphID)
		if err != nil {
			s.log.WithError(err).Error(
				"failed to get vertexes from storage")
			return
		}

		if vs == nil {
			vs = []entity.Vertex{}
		}

		es, err := s.storage.Edges(graphID)
		if err != nil {
			s.log.WithError(err).Error(
				"failed to get edges from storage")
			return
		}

		if es == nil {
			es = []entity.Edge{}
		}

		err = websocket.JSON.Send(ws, echo.Map{
			"type": "set-graph",
			"data": echo.Map{
				"graph":    g,
				"vertexes": vs,
				"edges":    es,
			},
		})
		if err != nil {
			s.log.WithError(err).Error(
				"failed to send graph to websocket")
			return
		}

		select {
		case <-s.stop:
		case <-closeWS:
		}
	}).ServeHTTP(c.Response(), c.Request())

	return nil
}

func (s *Server) deleteAPIGraph(c echo.Context) error {
	graphID, err := strconv.ParseInt(c.Param("graph_id"),
		10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest,
			"invalid graph_id")
	}

	err = s.storage.RemoveGraph(graphID)
	if err != nil {
		return fmt.Errorf("remove graph from storage: %w", err)
	}

	s.graphListenersMx.RLock()
	if listeners, exists := s.graphListeners[graphID]; exists {
		for ws, closeWS := range listeners {
			websocket.JSON.Send(ws, echo.Map{
				"type": "graph-removed",
			})
			close(closeWS)
		}
		delete(s.graphListeners, graphID)
	}
	s.graphListenersMx.RUnlock()

	return c.NoContent(http.StatusNoContent)
}

func (s *Server) postAPIVertexes(c echo.Context) error {
	var v entity.Vertex

	err := c.Bind(&v)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest,
			"bind vertex: "+err.Error())
	}

	id, err := s.storage.AddVertex(v)
	if err != nil {
		return fmt.Errorf("add vertex to storage: %w", err)
	}

	s.graphListenersMx.RLock()
	if listeners, exists := s.graphListeners[v.GraphID]; exists {
		v.ID = id
		for ws, closeWS := range listeners {
			err = websocket.JSON.Send(ws, echo.Map{
				"type": "new-vertex",
				"data": v,
			})
			if err != nil {
				s.log.WithError(err).
					WithField("method", "postAPIVertexes").
					Error("failed to send JSON to websocket")
				close(closeWS)
			}
		}
	}
	s.graphListenersMx.RUnlock()

	return c.NoContent(http.StatusCreated)
}

func (s *Server) putAPIVertexes(c echo.Context) error {
	var v entity.Vertex

	err := c.Bind(&v)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest,
			"bind vertex: "+err.Error())
	}

	err = s.storage.SetVertex(v)
	if err != nil {
		if err == entity.ErrVertexNotFound {
			return echo.NewHTTPError(http.StatusNotFound, err)
		}
		return fmt.Errorf("set vertex in storage: %w", err)
	}

	s.graphListenersMx.RLock()
	if listeners, exists := s.graphListeners[v.GraphID]; exists {
		for ws, closeWS := range listeners {
			err = websocket.JSON.Send(ws, echo.Map{
				"type": "vertex-update",
				"data": v,
			})
			if err != nil {
				s.log.WithError(err).
					WithField("method", "putAPIVertexes").
					Error("failed to send JSON to websocket")
				close(closeWS)
			}
		}
	}
	s.graphListenersMx.RUnlock()

	return c.NoContent(http.StatusNoContent)
}

func (s *Server) deleteAPIVertex(c echo.Context) error {
	vertexID, err := strconv.ParseInt(c.Param("vertex_id"),
		10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest,
			"invalid vertex_id")
	}

	v, err := s.storage.Vertex(vertexID)
	if err != nil {
		if err == entity.ErrVertexNotFound {
			return echo.NewHTTPError(http.StatusNotFound, err)
		}
		return fmt.Errorf("get vertex from storage: %w", err)
	}

	err = s.storage.RemoveVertex(vertexID)
	if err != nil {
		return fmt.Errorf("remove vertex from storage: %w", err)
	}

	s.graphListenersMx.RLock()
	if listeners, exists := s.graphListeners[v.GraphID]; exists {
		for ws, closeWS := range listeners {
			err = websocket.JSON.Send(ws, echo.Map{
				"type": "vertex-removed",
				"data": v,
			})
			if err != nil {
				s.log.WithError(err).
					WithField("method", "deleteAPIVertex").
					Error("failed to send JSON to websocket")
				close(closeWS)
			}
		}
	}
	s.graphListenersMx.RUnlock()

	return c.NoContent(http.StatusNoContent)
}

func (s *Server) postAPIEdges(c echo.Context) error {
	var e entity.Edge

	err := c.Bind(&e)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest,
			"bind edge: "+err.Error())
	}

	id, err := s.storage.AddEdge(e)
	if err != nil {
		return fmt.Errorf("add edge to storage: %w", err)
	}

	s.graphListenersMx.RLock()
	if listeners, exists := s.graphListeners[e.GraphID]; exists {
		e.ID = id
		for ws, closeWS := range listeners {
			err = websocket.JSON.Send(ws, echo.Map{
				"type": "new-edge",
				"data": e,
			})
			if err != nil {
				s.log.WithError(err).
					WithField("method", "postAPIEdges").
					Error("failed to send JSON to websocket")
				close(closeWS)
			}
		}
	}
	s.graphListenersMx.RUnlock()

	return c.NoContent(http.StatusCreated)
}

func (s *Server) putAPIEdges(c echo.Context) error {
	var e entity.Edge

	err := c.Bind(&e)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest,
			"bind edge: "+err.Error())
	}

	err = s.storage.SetEdge(e)
	if err != nil {
		if err == entity.ErrEdgeNotFound {
			return echo.NewHTTPError(http.StatusNotFound, err)
		}
		return fmt.Errorf("set edge in storage: %w", err)
	}

	s.graphListenersMx.RLock()
	if listeners, exists := s.graphListeners[e.GraphID]; exists {
		for ws, closeWS := range listeners {
			err = websocket.JSON.Send(ws, echo.Map{
				"type": "edge-update",
				"data": e,
			})
			if err != nil {
				s.log.WithError(err).
					WithField("method", "putAPIEdges").
					Error("failed to send JSON to websocket")
				close(closeWS)
			}
		}
	}
	s.graphListenersMx.RUnlock()

	return c.NoContent(http.StatusNoContent)
}

func (s *Server) deleteAPIEdge(c echo.Context) error {
	edgeID, err := strconv.ParseInt(c.Param("edge_id"),
		10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest,
			"invalid edge_id")
	}

	e, err := s.storage.Edge(edgeID)
	if err != nil {
		if err == entity.ErrEdgeNotFound {
			return echo.NewHTTPError(http.StatusNotFound, err)
		}
		return fmt.Errorf("get edge from storage: %v", err)
	}

	err = s.storage.RemoveEdge(edgeID)
	if err != nil {
		return fmt.Errorf("remove edge from storage: %w", err)
	}

	s.graphListenersMx.RLock()
	if listeners, exists := s.graphListeners[e.GraphID]; exists {
		for ws, closeWS := range listeners {
			err = websocket.JSON.Send(ws, echo.Map{
				"type": "edge-removed",
				"data": e,
			})
			if err != nil {
				s.log.WithError(err).
					WithField("method", "deleteAPIEdge").
					Error("failed to send JSON to websocket")
				close(closeWS)
			}
		}
	}
	s.graphListenersMx.RUnlock()

	return c.NoContent(http.StatusNoContent)
}

func (s *Server) getAPIGraphShortestPath(c echo.Context) error {
	graphID, err := strconv.ParseInt(c.Param("graph_id"),
		10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest,
			"invalid graph_id")
	}

	from, err := strconv.ParseInt(c.QueryParam("from"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest,
			"failed to parse from: "+err.Error())
	}

	to, err := strconv.ParseInt(c.QueryParam("to"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest,
			"failed to parse to: "+err.Error())
	}

	vs, err := s.storage.Vertexes(graphID)
	if err != nil {
		return fmt.Errorf("get vertexes from storage: %w", err)
	}

	es, err := s.storage.Edges(graphID)
	if err != nil {
		return fmt.Errorf("get edges from storage: %w", err)
	}

	path, err := dijkstra.ShortestPath(vs, es, from, to)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	return c.JSON(http.StatusOK, path)
}
