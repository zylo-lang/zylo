package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// HTTPResponse representa una respuesta HTTP
type HTTPResponse struct {
	StatusCode int                 `json:"status_code"`
	Headers    map[string]string   `json:"headers"`
	Body       string              `json:"body"`
	Error      string              `json:"error,omitempty"`
}

// HTTPRequest representa una petición HTTP
type HTTPRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
	Timeout int               `json:"timeout"` // en segundos
}

// HTTPGet realiza una petición GET
func HTTPGet(url string, headers map[string]string, timeout int) *HTTPResponse {
	return httpRequest("GET", url, "", headers, timeout)
}

// HTTPPost realiza una petición POST
func HTTPPost(url, data string, headers map[string]string, timeout int) *HTTPResponse {
	return httpRequest("POST", url, data, headers, timeout)
}

// HTTPPostJSON realiza una petición POST con JSON
func HTTPPostJSON(url string, data interface{}, headers map[string]string, timeout int) *HTTPResponse {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return &HTTPResponse{Error: fmt.Sprintf("Error marshaling JSON: %v", err)}
	}

	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Content-Type"] = "application/json"

	return httpRequest("POST", url, string(jsonData), headers, timeout)
}

// HTTPGetAsync realiza una petición GET asíncrona
func HTTPGetAsync(url string, headers map[string]string, timeout int) chan *HTTPResponse {
	resultChan := make(chan *HTTPResponse, 1)
	go func() {
		resultChan <- HTTPGet(url, headers, timeout)
	}()
	return resultChan
}

// HTTPPostJSONAsync realiza una petición POST JSON asíncrona
func HTTPPostJSONAsync(url string, data interface{}, headers map[string]string, timeout int) chan *HTTPResponse {
	resultChan := make(chan *HTTPResponse, 1)
	go func() {
		resultChan <- HTTPPostJSON(url, data, headers, timeout)
	}()
	return resultChan
}

// httpRequest es la función interna para hacer peticiones HTTP
func httpRequest(method, url, body string, headers map[string]string, timeout int) *HTTPResponse {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	var reqBody *strings.Reader
	if body != "" {
		reqBody = strings.NewReader(body)
	} else {
		reqBody = strings.NewReader("")
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return &HTTPResponse{Error: fmt.Sprintf("Error creating request: %v", err)}
	}

	// Agregar headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return &HTTPResponse{Error: fmt.Sprintf("Error making request: %v", err)}
	}
	defer resp.Body.Close()

	// Leer el body
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &HTTPResponse{Error: fmt.Sprintf("Error reading response body: %v", err)}
	}

	// Convertir headers a map[string]string
	respHeaders := make(map[string]string)
	for key, values := range resp.Header {
		respHeaders[key] = strings.Join(values, ", ")
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    respHeaders,
		Body:       string(bodyBytes),
	}
}

// HTTPServer representa un servidor HTTP simple
type HTTPServer struct {
	port    int
	handler func(*HTTPRequest) *HTTPResponse
	server  *http.Server
}

// NewHTTPServer crea un nuevo servidor HTTP
func NewHTTPServer(port int, handler func(*HTTPRequest) *HTTPResponse) *HTTPServer {
	return &HTTPServer{
		port:    port,
		handler: handler,
	}
}

// Start inicia el servidor HTTP
func (s *HTTPServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRequest)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	fmt.Printf("Servidor HTTP escuchando en puerto %d\n", s.port)
	return s.server.ListenAndServe()
}

// Stop detiene el servidor HTTP
func (s *HTTPServer) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// handleRequest maneja las peticiones HTTP entrantes
func (s *HTTPServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Leer el body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	// Convertir headers
	headers := make(map[string]string)
	for key, values := range r.Header {
		headers[key] = strings.Join(values, ", ")
	}

	// Crear HTTPRequest
	req := &HTTPRequest{
		Method:  r.Method,
		URL:     r.URL.String(),
		Headers: headers,
		Body:    string(body),
	}

	// Llamar al handler de Zylo
	resp := s.handler(req)

	// Enviar respuesta
	for key, value := range resp.Headers {
		w.Header().Set(key, value)
	}
	w.WriteHeader(resp.StatusCode)
	w.Write([]byte(resp.Body))
}

// JSONToZyloObject convierte JSON a un objeto Zylo (simulado)
func JSONToZyloObject(jsonStr string) interface{} {
	var data interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}
	return data
}

// ZyloObjectToJSON convierte un objeto Zylo a JSON (simulado)
func ZyloObjectToJSON(obj interface{}) (string, error) {
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}