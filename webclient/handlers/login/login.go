// +build js

package loginhandler

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/flimzy/go-cordova"
	"github.com/flimzy/jqeventrouter"
	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/pkg/errors"

	"github.com/FlashbackSRS/flashback/model"
)

var jQuery = jquery.NewJQuery

// BeforeTransition prepares the logout page before display.
func BeforeTransition(repo *model.Repo, providers map[string]string) jqeventrouter.HandlerFunc {
	return func(_ *jquery.Event, _ *js.Object, _ url.Values) bool {
		log.Debug("login BEFORE")

		cancel := checkLoginStatus(repo)

		container := jQuery(":mobile-pagecontainer")
		for rel, href := range providers {
			li := jQuery("li."+rel, container)
			li.Show()
			a := jQuery("a", li)
			if cordova.IsMobile() {
				a.On("click", func() {
					cancel()
					cordovaLogin(repo)()
				})
			} else {
				a.SetAttr("href", href)
				a.On("click", cancel)
			}
		}
		jQuery(".show-until-load", container).Hide()
		jQuery(".hide-until-load", container).Show()

		return true
	}
}

func checkCtx(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// checkLoginStatus checks for auth in the background
func checkLoginStatus(repo *model.Repo) func() {
	log.Debug("checkLoginStatus\n")
	defer log.Debug("return from checkLoginStatus\n")
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		if cordova.IsMobile() {
			var wg sync.WaitGroup
			wg.Add(1)
			js.Global.Get("facebookConnectPlugin").Call("getLoginStatus", func(response *js.Object) {
				go func() {
					defer wg.Done()
					provider := "facebook"
					if authResponse := response.Get("authResponse"); authResponse != js.Undefined {
						token := authResponse.Get("accessToken").String()
						fmt.Printf("token = %s\n", token)
						if err := repo.Auth(context.TODO(), provider, token); err != nil {
							log.Printf("(cls) Auth error: %s", err)
							return
						}
					}
				}()
			}, func() {
				wg.Done()
			})
			wg.Wait()
		}
		if _, err := repo.CurrentUser(); err != nil {
			log.Debugf("(cls) repo err: %s", err)
			return
		}
		if e := checkCtx(ctx); e != nil {
			log.Debugf("(cls) ctx err: %s", e)
			return
		}

		log.Debugln("(cls) Already authenticated")
		js.Global.Get("jQuery").Get("mobile").Call("changePage", "index.html")
	}()
	return cancel
}

func cordovaLogin(repo *model.Repo) func() bool {
	return func() bool {
		log.Debug("CordovaLogin()")
		js.Global.Get("facebookConnectPlugin").Call("login", []string{}, func(response *js.Object) {
			log.Debug("cl success pre goroutine\n")
			go func() {
				provider := "facebook"
				token := response.Get("authResponse").Get("accessToken").String()
				if err := repo.Auth(context.TODO(), provider, token); err != nil {
					displayError(err.Error())
					return
				}
				fmt.Printf("Auth succeeded!\n")
				js.Global.Get("jQuery").Get("mobile").Call("changePage", "index.html")
			}()
		}, func(err *js.Object) {
			log.Printf("Failure logging in: %s", err.Get("errorMessage").String())
		})
		log.Debug("Leaving CordovaLogin()")
		return true
	}
}

func displayError(msg string) {
	log.Printf("Authentication error: %s\n", msg)
	container := jQuery(":mobile-pagecontainer")
	jQuery("#auth_fail_reason", container).SetText(msg)
	jQuery(".show-until-load", container).Hide()
	jQuery(".hide-until-load", container).Show()
}

func BTCallback(repo *model.Repo, providers map[string]string) jqeventrouter.HandlerFunc {
	return func(event *jquery.Event, ui *js.Object, _ url.Values) bool {
		log.Debug("Auth Callback")
		provider, token, err := extractAuthToken(js.Global.Get("location").String())
		if err != nil {
			displayError(err.Error())
			return true
		}
		go func() {
			if err := repo.Auth(context.TODO(), provider, token); err != nil {
				msg := err.Error()
				if strings.Contains(msg, "Session has expired on") {
					for name, href := range providers {
						if name == provider {
							log.Debugf("Redirecting unauthenticated user to %s\n", href)
							js.Global.Get("location").Call("replace", href)
							event.StopImmediatePropagation()
							return
						}
					}
				}
				displayError(msg)
				return
			}
			fmt.Printf("Auth succeeded!\n")
			ui.Set("toPage", "index.html")
			event.StopImmediatePropagation()
			js.Global.Get("jQuery").Get("mobile").Call("changePage", "index.html")
		}()
		return true
	}
}

func extractAuthToken(uri string) (provider, token string, err error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return "", "", err
	}
	provider = parsed.Query().Get("provider")
	if provider == "" {
		return "", "", errors.New("no provider")
	}
	switch provider {
	case "facebook":
		frag, err := url.ParseQuery(parsed.Fragment)
		if err != nil {
			return "", "", errors.Wrapf(err, "failed to parse URL fragment")
		}
		token = frag.Get("access_token")
	default:
		return "", "", errors.Errorf("Unknown provider '%s'", provider)
	}
	return provider, token, nil
}
