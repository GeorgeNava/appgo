Welcome to app.go v3.0
======================

**app.go** is a simple web framework for use in Google AppEngine. Just copy the app folder to your working folder and import it from your main program. That's it. A web application ready to run in no time. Also, app.go comes with a powerful datastore manager to simplify your interactions with BigTable, making your code cleaner and safer.


Here is the Guestbook example from AppEngine rewritten using app.go

    package hello

    import(
        "app"
        "db"
    )

    type Greeting struct {
        Author  string
        Content string
        Date    int64
    }

    func init() {
        app.Start()
        app.Get ("/index", index)
        app.Post("/sign" , sign )
    }

    func index(ctx app.Context) {
        DB   := db.New(ctx)
        recs := make([]Greeting, 0, 10)
        qry  := DB.Query("Greeting").Order("-Date").Limit(10)
        DB.Select(qry,&recs)
        ctx.Render("index",recs)
    }

    func sign(ctx app.Context) {
        DB  := db.New(ctx)
        rec := Greeting{
            Author : ctx.User.Nick,
            Content: ctx.GetValue("content"),
            Date   : DB.Now(),
        }
        DB.New(&rec)
        ctx.Redirect("/")
    }


As you can see, with app.go we make it really easy to write web apps in go. We welcome your feedback for any special request or bug fix.

Enjoy!

* [Join us at the wiki for more info](appgo/wiki)
* [How to set up some initial configuration?](appgo/wiki/config)
* [How to use the regexp router?](appgo/wiki/routing)
* [How to use the power of app.Context](appgo/wiki/context)
* [How to use the datastore manager?](appgo/wiki/datastore)
* [My project is getting bigger, how to organize it?](appgo/wiki/organize)


CHANGELOG v3
------------
* separation of app and db packages
* added filters to template parsing (soon to be replaced by new template package)


CHANGELOG v2
------------
* implemented regexp router
* create new instance of DB for every request
* use nanoseconds in db.sequence
* on init: if no templates error/notfound generate default templates.
* cache templates

