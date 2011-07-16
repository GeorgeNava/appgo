package app


import(
  "os"
  "fmt"
  "log"
  "http"
  "path"
  "time"
  "regexp"
  "strings"
  "template"
  "appengine"
  "appengine/user"
)




//  APP Server  -------------------------------------------

type Settings struct{
    Debug     bool
    Root      string
    Media     string
    Templates string
}

type Server struct {
    Views      Views
    Routes     Routes
    Templates  Templates
}

type View  func(Context)
type Views map[string]View

type Route struct{
    Method   string
    Pattern  string
    Handler  View
    regex   *regexp.Regexp
}
type Routes []Route

type Template *template.Template
type Templates map[string]Template

var server *Server

func Start(){
  Log(">> APP running...")
  root,_ := os.Getwd()
  main := os.Getenv("APPLICATION_ID")
  //main = appengine.NewContext(???).AppID()
  Config.Root = root
  if Config.Media == "*" {
    Config.Media = path.Join(root, main)+"/"
  }
  if Config.Templates == "*" {
    Config.Templates = path.Join(root, main)+"/"
  }
  server = &Server{}
  server.initTemplates()
  server.Routes = []Route{}
  http.HandleFunc("/",routeHandler())
}

func (s Server) initTemplates() {
  server.Templates = make(map[string]Template)
  if !fileExist(getTemplateName("error")) {
    tmp,_ := template.Parse(htmlError,nil)
    server.Templates["error"] = tmp
  }
  if !fileExist(getTemplateName("notfound")) {
    tmp,_ := template.Parse(htmlNotFound,nil)
    server.Templates["notfound"] = tmp
  }
}

// Parse pretty urls: /path/value1/value2
func getValues(r *http.Request) []string {
  var vals = []string{}
  path := r.URL.Path
  if(path==""){ path = "/index" }
  if(string(path[0])!="/"){ path = "/"+path }
  parts := strings.Split(path,"/",-1)
  if(len(parts)>2){
    vals = parts[2:]
  }
  return vals
}

// Parse query string: /path?q=val1&id=val2
func getParams(r *http.Request) map[string]string {
  params := make(map[string]string)
  p, _ := http.ParseQuery(r.URL.RawQuery)
  for k,v := range p { 
    params[k] = v[0]
  }
  return params
}

func getForm(r *http.Request) map[string]string {
  form := make(map[string]string)
  r.ParseForm()
  if len(r.Form)>0 {
      for k := range r.Form {
          form[k] = r.FormValue(k)
      }
  }
  return form
}

func getContext(w http.ResponseWriter, r *http.Request) Context {
  ctx := Context{}
  ctx.Method      = r.Method
  ctx.Values      = getValues(r)
  ctx.Params      = getParams(r)
  ctx.Form        = getForm(r)
  ctx.Request     = r
  ctx.Response    = w
  ctx.User        = setUser(r)
  ctx.Context     = appengine.NewContext(r)
  return ctx
}




//  Context  ----------------------------------------------

type Context struct {
  Method     string
  Values     []string
  Params     map[string]string
  Form       map[string]string
  Request   *http.Request
  Response   http.ResponseWriter
  User      *UserType
  Context    appengine.Context
}

func (ctx Context) GetValue(key string) string {
  return ctx.Request.FormValue(key)
}

func (ctx Context) DefValue(key string, def string) string {
  val := ctx.Request.FormValue(key)
  if val=="" { return def }
  return val
}

func (ctx Context) Print(txt string){
  fmt.Println(txt)
}

func (ctx Context) Write(txt string){
  ctx.Response.Write([]byte(txt))
}

func (ctx Context) Show(file string){
  ctx.Render(file, nil)
}

func (ctx Context) Render(name string, data interface{}){
  tmp := getTemplate(name)
  tmp.Execute(ctx.Response, data)
}

func (ctx Context) NotFound(txt string){
  NotFound(ctx.Response,txt)
}

func (ctx Context) Redirect(url string){
  http.Redirect(ctx.Response, ctx.Request, url, http.StatusFound)
}

func (ctx Context) SetHeader(key string, val string) {
  ctx.Response.Header().Set(key, val)
}

func (ctx Context) SetCookie(key string, val string, exp int64) {
  var utc *time.Time
  if exp == 0 {
    utc = time.SecondsToUTC(2147483647)  // year 2038
  } else {
    utc = time.SecondsToUTC(time.UTC().Seconds() + exp)
  }
  cookie := fmt.Sprintf("%s=%s; expires=%s", key, val, webTime(utc))
  ctx.SetHeader("Set-Cookie", cookie)
}




//  APP Router  ------------------------------------------

func Get(pattern string, view View) {
  Handle("GET",pattern,view)
}
func Post(pattern string, view View) {
  Handle("POST",pattern,view)
}
func Put(pattern string, view View) {
  Handle("PUT",pattern,view)
}
func Delete(pattern string, view View) {
  Handle("DELETE",pattern,view)
}

func Handle(method string, pattern string, view View) {
  m  := strings.ToUpper(method)
  if pattern[len(pattern)-1] != 36 { pattern = pattern + "$" }
  rx, err := regexp.Compile(pattern)
  if err!=nil {
    Log(">>> ERROR compiling regexp: ",pattern)
    return
  }
  server.Routes = append(server.Routes,Route{Method:m,Pattern:pattern,Handler:view,regex:rx})
}

func routeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err, ok := recover().(os.Error); ok {
				w.WriteHeader(http.StatusInternalServerError)
				t := getTemplate("error")
				t.Execute(w, err)
			}
		}()
        path := r.URL.Path
        if path=="/" { path = "/index" }
        meth := strings.ToUpper(r.Method)
        req  := meth+" "+path
        var route Route
        var values []string
        found := false
        for i := range server.Routes {
            if server.Routes[i].Method==meth && server.Routes[i].regex.MatchString(path) {
                route  = server.Routes[i]
                params := route.regex.FindAllStringSubmatch(path,-1)
                values = params[0][1:]
                found  = true
                break
            }
        }
        if found { 
		  ctx := getContext(w,r)
		  ctx.Values = values
		  route.Handler(ctx)
        } else { 
          NotFound(w, "Couldn't match any handler for this pattern: "+req)
		}
	}
}




//  APP User  --------------------------------------------
type UserType struct{ 
  Nick     string
  Email    string
  IsAdmin  bool
  Context  appengine.Context
}

func setUser(r *http.Request) *UserType {
  User := UserType{}
  c := appengine.NewContext(r)
  u := user.Current(c)
  if u != nil {
    User.Nick    = u.String()
    User.Email   = u.Email
    User.IsAdmin = user.IsAdmin(c)
    User.Context = c
  } else {
    User.Nick    = ""
    User.Email   = ""
    User.IsAdmin = false
    User.Context = c
  }
  return &User
}

func (u *UserType) GetLoginURL(url string) string {
  login,_ := user.LoginURL(u.Context, url)
  return login
}

func (u *UserType) getLogoutURL(url string ) string {
  logout,_ := user.LogoutURL(u.Context, url)
  return logout
}




//  APP Utils  --------------------------------------------

func webTime(t *time.Time) string {
  gmt := t.Format(time.RFC1123)
  if strings.HasSuffix(gmt, "UTC") {
    gmt = gmt[0:len(gmt)-3] + "GMT"
  }
  return gmt
}


// Cache templates
func getTemplate(name string) *template.Template {
  tmp := server.Templates[name]
  if tmp==nil {
    tmp = template.MustParseFile(getTemplateName(name),filters)
    server.Templates[name] = tmp
  }
  return tmp
}

func getTemplateName(name string) string {
  if strings.Contains(name,".") {
    return Config.Templates +"/"+ name
  }
  return Config.Templates +"/"+ name + ".html"
}

func fileExist(name string) bool { 
  _, err := os.Stat(name) 
  return err==nil
} 

func NotFound(w http.ResponseWriter, txt string) {
  w.WriteHeader(http.StatusNotFound)
  tmp := getTemplate("notfound")
  tmp.Execute(w, txt)
}

func Log(stuff ...interface{}){
  log.Println(stuff...)
}




//  APP templates  ----------------------------------------
/*
    You can use your own templates, just create two files error.html and notfound.html 
    and place them in your Config.Templates folder
*/

var htmlError = `
<html>
<head>
  <title>Error</title>
</head>
<body>
  <h4>Oops! An error occurred:</h4>
  <pre>{@}</pre>
</body>
</html>
`

var htmlNotFound = `
<html>
<head>
  <title>Not found</title>
</head>
<body>
  <h4>The requested resource was not found in our servers</h4>
  <pre>{@}</pre>
</body>
</html>
`

//  END OF PROGRAM  =======================================

