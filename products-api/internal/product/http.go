package product
import("encoding/json";"errors";"net/http";"strings";"time")
type ErrorResponse struct{Code string `json:"code"`;Message string `json:"message"`;Retryable bool `json:"retryable"`}
type Handler struct{service Service};func NewHandler(s Service)Handler{return Handler{s}}
func(h Handler)Routes()http.Handler{m:=http.NewServeMux();m.HandleFunc("GET /health",func(w http.ResponseWriter,_ *http.Request){writeJSON(w,200,map[string]string{"status":"UP"})});m.HandleFunc("GET /products/{id}",h.get);return m}
func(h Handler)get(w http.ResponseWriter,r *http.Request){switch r.Header.Get("X-Failure-Mode"){case"transient":writeJSON(w,503,ErrorResponse{"DEPENDENCY_UNAVAILABLE","simulated transient failure",true});return;case"permanent":writeJSON(w,500,ErrorResponse{"PERMANENT_FAILURE","simulated permanent failure",false});return;case"timeout":select{case<-time.After(10*time.Second):case<-r.Context().Done():};return};p,e:=h.service.Get(r.Context(),strings.TrimSpace(r.PathValue("id")),r.URL.Query().Get("market"));if e!=nil{if errors.Is(e,ErrNotFound){writeJSON(w,404,ErrorResponse{"PRODUCT_NOT_FOUND","product does not exist",false})}else{writeJSON(w,400,ErrorResponse{"INVALID_REQUEST",e.Error(),false})};return};writeJSON(w,200,p)}
func writeJSON(w http.ResponseWriter,s int,b any){w.Header().Set("Content-Type","application/json");w.WriteHeader(s);_=json.NewEncoder(w).Encode(b)}
