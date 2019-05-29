package index

import (
	"strings"

	lnksworks "../../../lnkworks"
	_ "./shared"
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
<style>.webaction{
	cursor:pointer
}</style>
</head>
<body>    
	<div id="mainindex"></div>
	<div id="mainsection"></div>
	<script>postForm({"url_ref":"index.html?command=index"});</script>
</body>
</html>
`

func Index(atvpros *lnksworks.ActiveProcessor, path string, a ...string) (err error) {
	var out = atvpros.Out()
	out.ReplaceContent("#mainindex", func(out *widgeting.OutPrint) {
		out.ELEM("ul", "class=nav", func(out *widgeting.OutPrint) {
			out.ELEM("li", "class=nav-item", func(out *widgeting.OutPrint) {
				out.ELEM("a", "class=nav-link active webaction", "url_ref=/index.html?command=inbound&target=mainsection", func(out *widgeting.OutPrint) {
					out.ELEM("i", "class=far fa-arrow-alt-circle-left")
					out.Print(" ", "INBOUND")
				})
			})
			out.ELEM("li", "class=nav-item", func(out *widgeting.OutPrint) {
				out.ELEM("a", "class=nav-link webaction", "url_ref=/index.html?command=outbound&target=mainsection", func(out *widgeting.OutPrint) {
					out.ELEM("i", "class=far fa-arrow-alt-circle-right")
					out.Print(" ", "OUTBOUND")
				})
			})
			out.ELEM("li", "class=nav-item", func(out *widgeting.OutPrint) {
				out.ELEM("a", "class=nav-link webaction", "target=#mainsection", "url_ref=/shared/index.html", func(out *widgeting.OutPrint) {
					out.ELEM("i", "class=far fa-share-square")
					out.Print(" ", "SHARED")
				})
			})
		})
	})
	out.ReplaceContent("#mainsection")
	out.ScriptContent("", func(out *widgeting.OutPrint) {
		out.Print(`$(".webaction").click(function(e){
			e.preventDefault();
			postByElem(this);
		});`)
	})
	return
}

func init() {
	lnksworks.RegisterEmbededResources("index.html", strings.NewReader(indexhtml))
	lnksworks.MapActiveCommand("index.html",
		"index", Index,
	)
}
