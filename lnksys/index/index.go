package index

import (
	"strings"

	"github.com/efjoubert/lnkworks/widgeting"

	lnksworks "../../../lnkworks"
	//widgeting "../../../lnkworks/widgeting"
)

const indexhtml = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>INDEX</title>
	<script src="/jquery.js|block-ui.js|webactions.js"></script>
	<link rel="stylesheet" type="text/css" href="/bootstrap-all.css|datatables.css|hc-offcanvas-nav.css|goldenlayout-dark.css">
	<script src="/bootstrap-all.js"></script>
	<script src="/bootstrap-datatables.js"></script>
	<script src="/hc-offcanvas-nav.js"></script>
	<script src="/goldenlayout.js"></script>
</head>
<body>    
	<div id="mainindex"></div>
	<script>postForm({"target":"#mainindex","url_ref":"index.html?command=index"});</script>
</body>
</html>
`

func Index(atvpros *lnksworks.ActiveProcessor, path string, a ...string) (err error) {
	var out = atvpros.Out()
	out.ELEM("ul", "id=mainmenu", func(out *widgeting.OutPrint) {
		out.ELEM("li", func(out *widgeting.OutPrint) {
			out.ELEM("span", func(out *widgeting.OutPrint) {
				out.Print("TITLE 1")
			})
		})
	})
	out.SCRIPT("type=text/javascript", func(out *widgeting.OutPrint) {
		out.Print(`alert('test');$(#mainmenu).hcOffcanvasNav(
			{maxWidth: 100px,
			levelTitles:true}
		).update();`)
	})
	return
}

func init() {
	lnksworks.RegisterEmbededResources("index.html", strings.NewReader(indexhtml))
	lnksworks.MapActiveCommand("index.html",
		"index", Index,
	)
}
