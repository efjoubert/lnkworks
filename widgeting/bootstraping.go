package widgeting

import (
	"strings"

	"github.com/efjoubert/lnkworks/widgeting"
)

func NavBar(out *widgeting.OutPrint, a ...interface{}) {
	var navbarid = ""
	var navbara = []interface{}{}
	var nava = []interface{}{}
	var n = 0
	var d widgeting.MarkupFunction
	var dok = false
	for n < len(a) {
		if d, dok = a[n].(widgeting.MarkupFunction); dok {
			navbara = append(nava, d)
		} else {
			nava = append(nava, a[n])
		}
		n++
	}
	out.ELEM("nav", "class=navbar navbar-expand-lg navbar-light bg-light",
		func(out *widgeting.OutPrint, tag string, props ...string) {
			out.StartELEM(tag, props...)
			for _, p := range props {
				if strings.HasPrefix(p, "id=") {
					navbarid = strings.TrimSpace(p[len("id="):])
					break
				}
			}
		},
		func(out *widgeting.OutPrint) {
			if len(a) > 0 {
				out.ELEM("button",
					"class=navbar-toggler",
					"type=button",
					"data-toggle=collapse",
					"data-target=#"+navbarid+"Content",
					"aria-controls="+navbarid+"Content",
					"aria-expanded=false",
					"aria-label=Toggle navigation",
					func(out *widgeting.OutPrint) {
						out.ELEM("span", "class=navbar-toggler-icon")
					})

				out.DIV("class=collapse navbar-collapse", "id="+navbarid+"Content", func(out *widgeting.OutPrint) {
					out.ELEM("ul", "class=navbar-nav mr-auto", navbara)
				})
			}
		}, func(out *widgeting.OutPrint, tag string) {
			out.EndELEM(tag)
			out.SCRIPT(func(out *widgeting.OutPrint) {
				out.Print(`$("#` + navbarid + `").bootnavbar();$("#` + navbarid + ` .web-action").on("click",function(){
					postByElem(this);
				});`)
			})
		}, nava)
}

func NavItem(out *widgeting.OutPrint, title string, a ...interface{}) {
	out.ELEM("li", "class=nav-item", func(out *widgeting.OutPrint) {
		out.ELEM("a", "class=nav-link", "href=javascript:void(0)", title, a)
	})
}

func MenuItem(out *widgeting.OutPrint, title string, a ...interface{}) {
	out.ELEM("li", "class=dropdown-item", func(out *widgeting.OutPrint) {
		out.ELEM("a", "href=javascript:void(0)", title, a)
	})
}

func Menu(out *widgeting.OutPrint, menuid string, title string, a ...interface{}) {
	out.ELEM("li", "class=nav-item dropdown", func(out *widgeting.OutPrint) {
		out.ELEM("a", "class=nav-link dropdown-toggle", "href=#", "role=button", "data-toggle=dropdown", "aria-haspopup=true", "aria-expanded=false", "id="+menuid+"Dropdown", title)
		out.ELEM("ul", "class=dropdown-menu", "aria-labelledby="+menuid+"Dropdown", a)
	})
}
