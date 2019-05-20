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
	<script type="text/javascript" src="/jquery.js|block-ui.js|webactions.js"></script>
<link rel="stylesheet" type="text/css" href="/bootstrap-all.css|datatables.css|hc-offcanvas-nav.css"/>
<script type="text/javascript" src="/bootstrap-all.js|bootstrap-datatables.js|hc-offcanvas-nav.js"></script>
</head>
<body>    
	<span id="togglemenu" class="toggle"></span>
	<div id="mainindex"></div>
	<script>postForm({"target":"#mainindex","url_ref":"index.html?command=index"});</script>
</body>
</html>
`

func Index(atvpros *lnksworks.ActiveProcessor, path string, a ...string) (err error) {
	var out = atvpros.Out()
	out.ELEM("nav", "id=mainmenu", func(out *widgeting.OutPrint) {
		out.ELEM("ul", func(out *widgeting.OutPrint) {
			var menu = []string{""}
			for _, m := range menu {
				out.ELEM("li", func(out *widgeting.OutPrint) {
					out.ELEM("a", "href=#", func(out *widgeting.OutPrint) {
						out.Print(m)
					})
				})
			}
			/*for_,m:=range menu {
				out.ELEM("li", func(out *widgeting.OutPrint) {
					out.ELEM("a", "href=#", func(out *widgeting.OutPrint) {
						out.Print(m)
					})
				})
			}*/

		})
	})
	out.SCRIPT("type=text/javascript", func(out *widgeting.OutPrint) {
		out.Print(`$("#mainmenu").hcOffcanvasNav(
			{
				maxWidth: false,
				customToggle:$("#togglemenu"),
				navTitle:'All Categories',
				levelTitles:true,
				insertClose:true,
				insertBack:true
			}
		).update(true);`)
	})
	return
}

func init() {
	lnksworks.RegisterEmbededResources("index.html", strings.NewReader(indexhtml))
	lnksworks.MapActiveCommand("index.html",
		"index", Index,
	)
}
