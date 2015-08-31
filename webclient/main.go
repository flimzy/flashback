// +build js

package main

import (
	"github.com/flimzy/go-pouchdb"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"honnef.co/go/js/console"
	"strings"
	"sync"

	"golang.org/x/net/context"

	"github.com/flimzy/flashback"

	"github.com/flimzy/flashback/clientstate"
	//    "github.com/flimzy/flashback/user"
	"github.com/flimzy/flashback/webclient/pages"
	"github.com/flimzy/flashback/webclient/pages/all"
	"github.com/flimzy/flashback/webclient/pages/login"
	//     "github.com/flimzy/flashback/webclient/pages/index"
)

// Some spiffy shortcuts
var jQuery = jquery.NewJQuery
var jQMobile *js.Object
var document *js.Object = js.Global.Get("document")

func main() {
	console.Log("in main()")

	var db *pouchdb.PouchDB

	var wg sync.WaitGroup

	initPouchDB(&wg, db)
	cordova := initCordova(&wg)
	router := initRoutes(&wg)

	state := clientstate.New()
	api := flashback.New(jQuery("link[rel=flashback]").Get(0).Get("href").String())
	ctx := context.Background()
	ctx = context.WithValue(ctx, "cordova", cordova)
	ctx = context.WithValue(ctx, "AppState", state)
	ctx = context.WithValue(ctx, "db", db)
	ctx = context.WithValue(ctx, "api", api)

	// Wait for the above modules to initialize before we initialize jQuery Mobile
	wg.Wait()
	console.Log("Done with main()")
	initjQueryMobile(ctx, router)
}

func initRoutes(wg *sync.WaitGroup) *pages.Router {
	wg.Add(1)
	p := pages.NewRouter()
	go func() {
		defer wg.Done()
		p.Register("BEFORE", "pagecontainerbeforechange", all_pages.BeforeChange)
		//         p.RegisterBeforeChangeFunc( "index", index_page.BeforeChange )
		p.Register("/login.html", "pagecontainerbeforetransition", login_page.BeforeTransition)
		console.Log("Done setting up routes")
	}()
	return p
}

func initPouchDB(wg *sync.WaitGroup, db *pouchdb.PouchDB) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		db = pouchdb.New("flashback")
		// Then make sure we actually connected successfully
		info, err := db.Info()
		if err != nil {
			console.Log("Found an error: " + err.Error())
		}
		console.Log("PouchDB connected to " + info.Get("db_name").String())
	}()
}

func initCordova(wg *sync.WaitGroup) *js.Object {
	mobile := isMobile()
	if mobile == nil {
		return nil
	}
	wg.Add(1)
	document.Call("addEventListener", "deviceready", func() {
		defer wg.Done()
		console.Log("Cordova device ready")
	}, false)
	return mobile
}

func initjQueryMobile(ctx context.Context, router *pages.Router) {
	jQuery(document).On("mobileinit", func() {
		console.Log("mobileinit")
		MobileInit(ctx, router)
	})
	// This is what actually loads jQuery Mobile. We have to register our 'mobileinit'
	// event handler above first, though.
	js.Global.Call("postInit")
}

func MobileInit(ctx context.Context, router *pages.Router) {
	jQMobile = js.Global.Get("jQuery").Get("mobile")

	// Disable hash features
	jQMobile.Set("hashListeningEnabled", false)
	jQMobile.Set("pushStateEnabled", false)
	jQMobile.Get("changePage").Get("defaults").Set("changeHash", false)

	//    DebugEvents()

	router.Init(ctx)
	jQuery(document).On("pagecontainerbeforechange", func(event *jquery.Event, ui *js.Object) {
		console.Log("last beforechange event handler")
	})
}

func MobileGlobal() *js.Object {
	if m := js.Global.Get("cordova"); m != nil {
		return m
	}
	if m := js.Global.Get("PhoneGap"); m != nil {
		return m
	}
	if m := js.Global.Get("phonegap"); m != nil {
		return m
	}
	return nil
}

func isMobile() *js.Object {
	mobile := MobileGlobal()
	if mobile == nil {
		return nil
	}
	ua := strings.ToLower(js.Global.Get("navigator").Get("userAgent").String())

	if strings.HasPrefix(strings.ToLower(js.Global.Get("location").Get("href").String()), "file:///") &&
		(strings.Contains(ua, "ios") || strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") || strings.Contains(ua, "android")) {
		return mobile
	}
	return nil
}

func ConsoleEvent(name string, event *jquery.Event, data *js.Object) {
	page := data.Get("toPage").String()
	if page == "[object Object]" {
		page = data.Get("toPage").Call("jqmData", "url").String()
	}
	//     ui.Get("toPage").String()
	//    page := jQMobile.Get("pageContainer").Call("pagecontainer", "getActivePage").Get("id").String()
	console.Log("Event: %s, Current page: %s", name, page)
}

func ConsolePageEvent(name string, event *jquery.Event) {
	console.Log("Event: %s", name)
}

func DebugEvents() {
	events := []string{"pagecontainerbeforehide", "pagecontainerbeforechange", "pagecontainerbeforeload", "pagecontainerbeforeshow",
		"pagecontainerbeforetransition", "pagecontainerchange", "pagecontainerchangefailed", "pagecontainercreate", "pagecontainerhide",
		"pagecontainerload", "pagecontainerloadfailed", "pagecontainerremove", "pagecontainershow", "pagecontainertransition"}
	for _, event := range events {
		copy := event // Necessary for each iterration to have an effectively uinque closure
		jQuery(document).On(event, func(e *jquery.Event, d *js.Object) {
			ConsoleEvent(copy, e, d)
		})
	}
	pageEvents := []string{"beforecreate", "create"}
	for _, event := range pageEvents {
		copy := event
		jQuery(document).On(event, func(e *jquery.Event) {
			ConsolePageEvent(copy, e)
		})
	}
}
