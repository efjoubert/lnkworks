package main

import (
	"fmt"
	"strings"
	"time"

	lnksworks "../../lnkworks"
	activeruling "../../lnkworks/activeruling"
	activerulinggui "../../lnkworks/activeruling/gui"
	embed "../../lnkworks/embed"
	widgeting "../../lnkworks/widgeting"
)

func main() {
	i := 0
	for i < 1 {
		schname := fmt.Sprintf("SCHDL%d", i)
		schdl := activeruling.RegisterSchedule(schname, 1000*time.Millisecond, func(schdlname string, tickStamp time.Time) {
			fmt.Println(schdlname, "->", tickStamp)
		})
		schdl.EnableSchedule()

		go func() {
			time.AfterFunc(time.Duration(i+1)*100*time.Millisecond, func() {
				schdl.DisableSchedule()
			})
		}()
		i++
	}

	activerulinggui.ActiveRulingUI("ruling", "index.html")

	lnksworks.RegisterEmbededResources(
		"require.js", embed.RequireJS(),
		"babel.js", embed.BabelJS(),
		"preact.js", embed.PreactJS(),
		"d3.js", embed.D3JS(),
		"three.js", embed.ThreeJS(),
		"vue.js", embed.VueJS(),
		"fontawesome.js", embed.FontAwesomeJS(),
		"bootstrap-all.css", embed.BootstrapAllCSS(),
		"bootstrap-all.js", embed.BootstrapAllJS(),
		"bootstrap-datatables.js", embed.DataTablesJS(true),
		"datatables.js", embed.DataTablesJS(false),
		"datatables.css", embed.DataTablesCSS(),
		"mdb.js", embed.MdbJS(),
		"material-icons.css", embed.MaterialIconsCSS(),
		"material-icons.woff2", embed.MaterialIconsWoff2(),
		"mdb.css", embed.MdbCSS(),
		"jquery.js", embed.JQueryJS(),
		"webactions.js", embed.WebActionsJS(true),
		"block-ui.js", embed.BlockUiJS(),
		"jquery-ui.js", embed.JQueryUiJS(),
		"jquery-ui.css", embed.JQueryUiCSS(),
		"testjquery.html", embed.JQueryUiTestHtml(),
		"ui-icons_444444_256x240.png", embed.JQueryUiImages("ui-icons_444444_256x240.png"),
		"ui-icons_555555_256x240.png", embed.JQueryUiImages("ui-icons_555555_256x240.png"),
		"ui-icons_777777_256x240.png", embed.JQueryUiImages("ui-icons_777777_256x240.png"),
		"ui-icons_777620_256x240.png", embed.JQueryUiImages("ui-icons_777620_256x240.png"),
		"ui-icons_cc0000_256x240.png", embed.JQueryUiImages("ui-icons_cc0000_256x240.png"),
		"ui-icons_ffffff_256x240.png", embed.JQueryUiImages("ui-icons_ffffff_256x240.png"),
	)

	test()

	lnksworks.RegisterEmbededResources(
		"section/index.html", strings.NewReader(`<html><head></head><section:sub/><body></body></html>`),
		"section/sub.html", strings.NewReader(`<span>section->sub</span>`),
	)

	lnksworks.RegisterRoute("/lnks", "")
	svr := lnksworks.NewServer(":1030", false, "", "")
	if err := svr.Listen(); err != nil {
		panic(err)
	}
}

func test() {
	lnksworks.RegisterEmbededResources("test/test.html", strings.NewReader(""))
	lnksworks.MapActiveCommand("test/test.html",
		"testcommand", func(atvpros *lnksworks.ActiveProcessor, path string, a ...string) (err error) {

			atvpros.Out().Elem("span", func(out *widgeting.OutPrint, a ...interface{}) {
				out.Print("content in span")
			})

			return err
		},
	)
}
