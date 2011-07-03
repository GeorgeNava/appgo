package app

import(
  "os"
  "fmt"
  "log"
  "http"
  "time"
  "strings"
  "template"
  "appengine"
  "appengine/user"
)

var DB = &DSM{}  // use datastore manager from db.go to update context here




//  APP Server  -------------------------------------------

type AppConfig struct{
    Debug     bool
    Root      string
    Media     string
    Templates string
}

type AppServer struct {
  Views     Views
}

type View func(Context)

type Views map[string]View

var server *AppServer

func Run(list Views){
  Log(">> APP running...")
  root,_ := os.Getwd()
  wdir := os.Getenv("APPLICATION_ID")+"/"
  Config.Root = root
  if Config.Media == "*" { // Use app folder
    Config.Media = wdir
  }
  if Config.Templates == "*" { // Use app folder
    Config.Templates = wdir
  }
  server = &AppServer{}
  server.Views = list
  http.HandleFunc("/",router())
  return
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
  ctx.Params      = getParams(r)
  ctx.Values      = getValues(r)
  ctx.Form        = getForm(r)
  ctx.Request     = r
  ctx.Response    = w
  ctx.User        = setUser(r)
  return ctx
}




//  APP Router  ------------------------------------------

func router() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err, ok := recover().(os.Error); ok {
				w.WriteHeader(http.StatusInternalServerError)
				var t = template.MustParseFile(getTemplateName("error"),nil)
				t.Execute(w, err)
			}
		}()
        paths := strings.Split(r.URL.Path,"/",-1)
        path := paths[1]
        if(path==""){ path = "index" }
        fn := server.Views[path]
        if fn==nil{ 
          NotFound(w, "There is no view '"+path+"' in our servers")
        } else { 
          DB.Context = appengine.NewContext(r)
		  ctx := getContext(w,r)
		  fn(ctx)
		}
	}
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

func (ctx Context) Render(file string, data interface{}){
  tmp := template.MustParseFile(getTemplateName(file),nil)
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

func getTemplateName(name string) string {
  if strings.Contains(name,".") { return Config.Templates + name }
  return Config.Templates + name + ".html"
}

func NotFound(w http.ResponseWriter, txt string) {
  w.WriteHeader(http.StatusNotFound)
  tmp := template.MustParseFile(getTemplateName("notfound"),nil)
  tmp.Execute(w, txt)
}

func Log(x interface{}){
  log.Println(x)
}

//  END OF PROGRAM  =======================================

