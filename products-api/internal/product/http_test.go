package product
import("context";"net/http";"net/http/httptest";"testing")
func TestHandler(t *testing.T){h:=NewHandler(NewService(NewMemoryRepository())).Routes();r:=httptest.NewRequest(http.MethodGet,"/products/PRD-001?market=MX",nil);w:=httptest.NewRecorder();h.ServeHTTP(w,r);if w.Code!=200{t.Fatalf("want 200 got %d",w.Code)};r=httptest.NewRequest(http.MethodGet,"/products/missing?market=MX",nil);w=httptest.NewRecorder();h.ServeHTTP(w,r);if w.Code!=404{t.Fatalf("want 404 got %d",w.Code)}}
func TestValidation(t *testing.T){_,e:=NewService(NewMemoryRepository()).Get(context.Background(),"PRD-001","US");if e==nil{t.Fatal("expected error")}}
