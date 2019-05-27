package main

import (
	"fmt"

	lnksworks "../../lnkworks"
	embed "../../lnkworks/embed"
	_ "./index"
	_ "github.com/jackc/pgx/stdlib"
)

func main() {

	if err := lnksworks.DatabaseManager().RegisterConnection("modulation", "pgx", "user=lnksworks password=lnksworks56579757 host=localhost port=5432 database=lnksworks sslmode=disable"); err != nil {
		fmt.Println(err)
	}

	lnksworks.RegisterEmbededResources(
		//"require.js", embed.RequireJS(),
		//"babel.js", embed.BabelJS(),
		//"typescript.js", embed.TypeScriptJS(),
		//"preact.js", embed.PreactJS(),
		"pdf.js", embed.PdfJS(),
		//"d3.js", embed.D3JS(),
		//"three.js", embed.ThreeJS(),
		//"vue.js", embed.VueJS(),
		"fontawesome.js", embed.FontAwesomeJS(),
		"fontawesome.css", embed.FontAwesomeCSS(),
		"webfonts/fa-brands-400.eot", embed.FontAwesomeFont("400", "brands", "eot"),
		"webfonts/fa-brands-400.ttf", embed.FontAwesomeFont("400", "brands", "ttf"),
		"webfonts/fa-brands-400.woff", embed.FontAwesomeFont("400", "brands", "woff"),
		"webfonts/fa-brands-400.woff2", embed.FontAwesomeFont("400", "brands", "woff2"),
		"webfonts/fa-regular-400.eot", embed.FontAwesomeFont("400", "regular", "eot"),
		"webfonts/fa-regular-400.ttf", embed.FontAwesomeFont("400", "regular", "ttf"),
		"webfonts/fa-regular-400.woff", embed.FontAwesomeFont("400", "regular", "woff"),
		"webfonts/fa-regular-400.woff2", embed.FontAwesomeFont("400", "regular", "woff2"),
		"webfonts/fa-solid-900.eot", embed.FontAwesomeFont("900", "solid", "eot"),
		"webfonts/fa-soldi-900.ttf", embed.FontAwesomeFont("900", "solid", "ttf"),
		"webfonts/fa-solid-900.woff", embed.FontAwesomeFont("900", "solid", "woff"),
		"webfonts/fa-solid-900.woff2", embed.FontAwesomeFont("900", "solid", "woff2"),
		"bootstrap-all.css", embed.BootstrapAllCSS(),
		"bootstrap-all.js", embed.BootstrapAllJS(),
		"bootstrap-datatables.js", embed.DataTablesJS(true),
		"datatables.css", embed.DataTablesCSS(false),
		"datatables.js", embed.DataTablesJS(false),
		"MaterialIcons-Regular.eot", embed.MaterialIconsRegularEOT(),
		"MaterialIcons-Regular.woff", embed.MaterialIconsRegularWOFF(),
		"MaterialIcons-Regular.woff2", embed.MaterialIconsRegularWOFF2(),
		"MaterialIcons-Regular.ttf", embed.MaterialIconsRegularTTF(),
		"material-design-icons.css", embed.MaterialDesignIconsCSS(),

		"jquery.js", embed.JQueryJS(),
		"webactions.js", embed.WebActionsJS(false),
		"block-ui.js", embed.BlockUiJS(),
		"bootnavbar.css", embed.BootNavbarCSS(),
		"bootnavbar.js", embed.BootNavbarJS(),
		"jquery-terminal.js", embed.JQueryTerminalJS(),
		"jquery-terminal.css", embed.JQueryTerminalCSS(),
		"jquery-ui.js", embed.JQueryUiJS(),
		"jquery-ui.css", embed.JQueryUiCSS(),
		"goldenlayout-dark.css", embed.GoldenLayoutBaseCSS("dark"),
		"goldenlayout-light.css", embed.GoldenLayoutBaseCSS("light"),
		"goldenlayout.js", embed.GoldenLayoutJS(),
		"testjquery.html", embed.JQueryUiTestHtml(),
		"ui-icons_444444_256x240.png", embed.JQueryUiImages("ui-icons_444444_256x240.png"),
		"ui-icons_555555_256x240.png", embed.JQueryUiImages("ui-icons_555555_256x240.png"),
		"ui-icons_777777_256x240.png", embed.JQueryUiImages("ui-icons_777777_256x240.png"),
		"ui-icons_777620_256x240.png", embed.JQueryUiImages("ui-icons_777620_256x240.png"),
		"ui-icons_cc0000_256x240.png", embed.JQueryUiImages("ui-icons_cc0000_256x240.png"),
		"ui-icons_ffffff_256x240.png", embed.JQueryUiImages("ui-icons_ffffff_256x240.png"),
	)

	lnksworks.RegisterRoute("/", "")
	svr := lnksworks.NewServer(":1030", false, "", "")
	if err := svr.Listen(); err != nil {
		panic(err)
	}
}
