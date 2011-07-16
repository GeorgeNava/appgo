package app


import(
  "io"
  "fmt"
  "math"
  "time"
  "reflect"
  "strings"
  "template"
)


// Add more filters here if needed
var filters = template.FormatterMap{ 
    "upper"   : upperFilter,
    "lower"   : lowerFilter,
    "title"   : titleFilter,
    "break"   : breakFilter,
    "unbreak" : unbreakFilter,
    "plural"  : pluralFilter,
    "ellipsis": ellipsisFilter,
    "date"    : dateFilter,
    "time"    : timeFilter,
    "now"     : nowFilter,
    "today"   : todayFilter,
    "year"    : yearFilter,
    "ago"     : agoFilter,
    "decimal" : decimalFilter,
    "money"   : moneyFilter,
    "pointer" : pointerFilter, 
    "html"    : template.HTMLFormatter,
}


// {Name|upper}
func upperFilter(w io.Writer, format string, data ...interface{}) {
  for _, item := range data {
    fmt.Fprintf(w, strings.ToUpper(item.(string)))
  }
}


// {Name|lower}  
func lowerFilter(w io.Writer, format string, data ...interface{}) {
  for _, item := range data {
    fmt.Fprintf(w, strings.ToLower(item.(string)))
  }
}


// {Name|title}
func titleFilter(w io.Writer, format string, data ...interface{}) {
  for _, item := range data {
    fmt.Fprintf(w, strings.Title(item.(string)))
  }
}


// {Content|break}
func breakFilter(w io.Writer, format string, data ...interface{}) {
  for _, item := range data {
    fmt.Fprintf(w, strings.Replace(item.(string),"\n","<br>",-1))
  }
}


// {Content|unbreak}
func unbreakFilter(w io.Writer, format string, data ...interface{}) {
  for _, item := range data {
    fmt.Fprintf(w, strings.Replace(item.(string),"<br>","\n",-1))
  }
}


// {Count|plural}
func pluralFilter(w io.Writer, format string, data ...interface{}) {
  for _, item := range data {
    if item.(int)!=1 { fmt.Fprintf(w, "s") }
  }
}


// {Description|ellipsis}
func ellipsisFilter(w io.Writer, format string, data ...interface{}) {
  n:=40  // change to your needs
  for _, item := range data {
    s := item.(string)
    if len(s)>n { fmt.Fprintf(w, s[:n]+"&hellip;")
    } else { fmt.Fprintf(w, s) }
  }
}


// {Date|date}  2001/02/03
func dateFilter(w io.Writer, format string, data ...interface{}) {
  for _, item := range data {
    date := time.SecondsToUTC(item.(int64))
    fmt.Fprintf(w, date.Format("2006/01/02"))
  }
}


// {Date|time}  2001/02/03 12:59:59
func timeFilter(w io.Writer, format string, data ...interface{}) {
  for _, item := range data {
    date := time.SecondsToUTC(item.(int64))
    fmt.Fprintf(w, date.Format("2006/01/02 15:04:05"))
  }
}


// {@|now}    2001/02/03 12:59:59
func nowFilter(w io.Writer, format string, data ...interface{}) {
  date := time.UTC()
  fmt.Fprintf(w, date.Format("2006/01/02 15:04:05"))
}


// {@|today}  2001/02/03
func todayFilter(w io.Writer, format string, data ...interface{}) {
  date := time.UTC()
  fmt.Fprintf(w, date.Format("2006/01/02"))
}


// {@|year}  2011
func yearFilter(w io.Writer, format string, data ...interface{}) {
  date := time.UTC()
  fmt.Fprintf(w, date.Format("2006"))
}


// Posted {Date|ago} 5 mins ago
func agoFilter(w io.Writer, format string, data ...interface{}){
  for _, item := range data {
    ago := timeago(item.(int64))
    fmt.Fprintf(w, ago)
  }
}


// {Pi|decimal}  3.14
func decimalFilter(w io.Writer, format string, data ...interface{}){
  for _, item := range data {
    dec := math.Floor(item.(float64) * 100) / 100
    str := strings.Trim(fmt.Sprintf("%9.2f",dec)," ")
    fmt.Fprint(w, str)
  }
}


// {Price|money}  $123.45
func moneyFilter(w io.Writer, format string, data ...interface{}){
  for _, item := range data {
    dec := math.Floor(item.(float64) * 100) / 100
    str := "$"+strings.Trim(fmt.Sprintf("%9.2f",dec)," ")
    fmt.Fprint(w, str)
  }
}


// Dereference pointers
func pointerFilter(w io.Writer, format string, data ...interface{}) { 
    for i := range data {
        data[i] = reflect.Indirect(reflect.ValueOf(data[i])).Interface()
    }
    fmt.Fprint(w, data...)
}


func timeago(t int64) string {
  const ( 
    mm = 60 
    hh = 60 * mm 
    dd = 24 * hh 
    ww =  7 * dd 
  )
  now := time.Seconds()
  dif := now-t
  if dif>dd {
    if dif>ww { return time.SecondsToUTC(t).Format("2006/01/02")
    } else {
      if dif<dd*2 { return "Yesterday"
      } else { return fmt.Sprintf("%d days ago",(int(dif/dd))) }
    }
  }
  s := ""
  if dif>=mm {
    m := int(dif/mm)
    if m>59 {
      h := dif/hh
      if h>1 { s="s" }
      return fmt.Sprintf("%d hour%s ago",h,s)
    }
    if m>1 { s="s" }
    return fmt.Sprintf("%d minute%s ago",m,s)
  }
  if dif<5 { return "just now" }
  return fmt.Sprintf("%d seconds ago",dif)
}

// END

