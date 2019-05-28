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
	<script>postForm({"url_ref":"index.html?command=index"});</script>
</body>
</html>
`

func Index(atvpros *lnksworks.ActiveProcessor, path string, a ...string) (err error) {
	var out = atvpros.Out()
	out.ReplaceContent("#mainindex", func(out *widgeting.OutPrint) {
		out.ELEM("ul", "class=nav", func(out *widgeting.OutPrint) {
			out.ELEM("li", "class=nav-item", func(out *widgeting.OutPrint) {
				out.ELEM("a", "class=nav-link active webaction", "href=javascript:void(0)", "url_ref=/index.html?command=inbound&target=mainsection", "INBOUND")
			})
			out.ELEM("li", "class=nav-item", func(out *widgeting.OutPrint) {
				out.ELEM("a", "class=nav-link webaction", "href=javascript:void(0)", "url_ref=/index.html?command=outbound&target=mainsection", "OUTBOUND")
			})
			out.ELEM("li", "class=nav-item", func(out *widgeting.OutPrint) {
				out.ELEM("a", "class=nav-link webaction", "href=javascript:void(0)", "url_ref=/index.html?command=shared&target=mainsection", "SHAREDBOUND")
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

	lnksworks.MapActiveCommand("index.html",
		"inbound", Inbound,
	)
	lnksworks.MapActiveCommand("index.html",
		"outbound", Outbound,
	)
	lnksworks.MapActiveCommand("index.html",
		"shared", Shared,
	)

}

func Inbound(atvpros *lnksworks.ActiveProcessor, path string, a ...string) (err error) {
	IndexSection(atvpros.Out(), "INBOUND", atvpros.Parameters(), atvpros.Parameters().StringParameter("target", ""))
	return
}

func Outbound(atvpros *lnksworks.ActiveProcessor, path string, a ...string) (err error) {
	IndexSection(atvpros.Out(), "OUTBOUND", atvpros.Parameters(), atvpros.Parameters().StringParameter("target", ""))

	return
}

func Shared(atvpros *lnksworks.ActiveProcessor, path string, a ...string) (err error) {
	IndexSection(atvpros.Out(), "SHARED", atvpros.Parameters(), atvpros.Parameters().StringParameter("target", ""))

	return
}

func IndexSection(out *widgeting.OutPrint, title string, params *lnksworks.Parameters, target string, a ...interface{}) {
	if target != "" && !strings.HasPrefix(target, ".") {
		target = "#" + target
	}
	out.ReplaceContent(target, func(out *widgeting.OutPrint) {
		out.ELEM("div", func(out *widgeting.OutPrint) {
			out.Print(strings.ToUpper(title))
		})
	}, a)
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

func ICON(out *widgeting.OutPrint, tag string, class string, title string) {
	out.ELEM(tag, "class="+class, title)
}

func ACTION(out *widgeting.OutPrint, tag string, title string, icon string, navclass string, a ...interface{}) {
	out.ELEM(tag, "class="+navclass, func(out *widgeting.OutPrint) {
		out.ELEM("span", title)
	}, a)
}
