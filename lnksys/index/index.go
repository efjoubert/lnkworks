package index

import (
	"strings"

	lnksworks "../../../lnkworks"
	widgeting "github.com/efjoubert/lnkworks/widgeting"
	//widgeting "../../../lnkworks/widgeting"
	//"github.com/efjoubert/lnkworks/widgeting"
)

const indexhtml = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>INDEX</title>
	<script type="text/javascript" src="/jquery.js|block-ui.js|webactions.js"></script>
<link rel="stylesheet" type="text/css" href="/bootstrap-all.css|datatables.css|bootnavbar.css"/>
<script type="text/javascript" src="/bootstrap-all.js|bootstrap-datatables.js|bootnavbar.js"></script>
<script type="text/javascript" src="/fontawesome.js"></script>
</head>
<body>    
	<div id="mainindex"></div>
	<div id="mainsection"></div>
	<div class="samplesection"></div>
	<script>postForm({"target":"#mainindex","url_ref":"index.html?command=index"});</script>
</body>
</html>
`

func Index(atvpros *lnksworks.ActiveProcessor, path string, a ...string) (err error) {
	var out = atvpros.Out()
	widgeting.NavBar(out, "id=navbar", func(out *widgeting.OutPrint) {
		widgeting.NavItem(out, "LNK", "class=web-action", "target=#mainsection", "url_ref=/index.html?command=datasources")
		widgeting.NavItem(out, "ADHOC OPTIONS", "class=web-action", "target=#mainsection", "url_ref=/index.html?command=adhocoptions")
		widgeting.Menu(out, "menu1", "MENU 1", func(out *widgeting.OutPrint) {
			widgeting.MenuItem(out, "MENU 1 1", "class=web-action", "url_ref=/index.html?command=samplesection")
			widgeting.MenuItem(out, "MENU 1 2")
			widgeting.Menu(out, "menu11", "MENU 1 3", func(out *widgeting.OutPrint) {
				widgeting.MenuItem(out, "MENU 1 1")
			})
		})
	})
	return
}

func init() {
	lnksworks.RegisterEmbededResources("index.html", strings.NewReader(indexhtml))
	lnksworks.MapActiveCommand("index.html",
		"index", Index,
	)

	lnksworks.MapActiveCommand("index.html",
		"datasources", Datasources,
	)

	lnksworks.MapActiveCommand("index.html",
		"adhocoptions", AdhocOptions,
	)

	lnksworks.MapActiveCommand("index.html",
		"samplesection", SampleSection,
	)
}

func SampleSection(atvpros *lnksworks.ActiveProcessor, path string, a ...string) (err error) {
	var out = atvpros.Out()
	out.ReplaceContent(".samplesection", func(out *widgeting.OutPrint) {
		out.Print("SECTION-TEST")
	})
	return
}

func AdhocOptions(atvpros *lnksworks.ActiveProcessor, path string, a ...string) (err error) {
	var out = atvpros.Out()

	out.Print("ADHOC-OPTIONS")

	return
}

func Datasources(atvpros *lnksworks.ActiveProcessor, path string, a ...string) (err error) {
	var out = atvpros.Out()
	out.ELEM("div", func(out *widgeting.OutPrint) {
		out.Print("DATASOURCE(s)")
	})
	out.DIV(func(out *widgeting.OutPrint) {
		ACTION(out, "button", "DATABASE(s)", "fa fa-database", "", "url_ref=/index?command=dbsources")
		ACTION(out, "button", "FILE(s)", "fa fa-file-import", "", "url_ref=/index?command=filesources")
		ACTION(out, "button", "WEB", "fa fa-basketball-ball", "", "url_ref=/index?command=websources")
		ACTION(out, "button", "SMS-GATEWAY(s)", "fa fa-sms", "", "url_ref=/index?command=smsgatewaysources")
	}, "style=dipslay:inline")

	return
}

func ACTION(out *widgeting.OutPrint, tag string, title string, icon string, navclass string, a ...interface{}) {
	out.ELEM(tag, "class="+navclass, func(out *widgeting.OutPrint) {
		out.ELEM("i", "class="+icon, "role=button")
		out.ELEM("span", title)
	}, a)
}
