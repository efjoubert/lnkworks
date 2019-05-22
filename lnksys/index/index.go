package index

import (
	"strings"

	lnksworks "../../../lnkworks"
	//widgeting "../../../lnkworks/widgeting"
	"github.com/efjoubert/lnkworks/widgeting"
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
	<script>postForm({"target":"#mainindex","url_ref":"index.html?command=index"});</script>
</body>
</html>
`

func NavBar(out *widgeting.OutPrint, navbarid string, a ...interface{}) {
	var navbara = a
	out.ELEM("nav", "class=navbar navbar-expand-lg navbar-light bg-light", "id="+navbarid, func(out *widgeting.OutPrint, tag string, props ...string) {
		out.StartELEM(tag, props...)
		out.ELEM("button",
			"class=navbar-toggler",
			"type=button",
			"data-toggle=#"+navbarid+"content",
			"aria-controls=#"+navbarid+"content",
			"aria-expanded=false", "aria-label=", func(out *widgeting.OutPrint) {
				out.ELEM("span", "class=navbar-toggle-icon")
			})
	}, func(out *widgeting.OutPrint) {
		out.ELEM("div", "class=collapse navbar-collapse", func(out *widgeting.OutPrint) {
			out.ELEM("ul", "class=navbar-nav mr-auto", navbara)
		})
	}, func(out *widgeting.OutPrint, tag string) {

		out.EndELEM(tag)
		out.SCRIPT(func(out *widgeting.OutPrint) {
			out.Print(`$("#` + navbarid + `").bootnavbar();$("#` + navbarid + ` .webaction").on("click",function(){
				alert($(this).html());
			});`)
		})
	})
}

func NavWebAction(out *widgeting.OutPrint, a ...interface{}) {

}

func Index(atvpros *lnksworks.ActiveProcessor, path string, a ...string) (err error) {
	var out = atvpros.Out()

	NavBar(out, "thenavbar")

	/*out.Print(`<nav class="navbar navbar-expand-lg navbar-light bg-light" id="main_navbar">
		<a class="navbar-brand" href="#">Navbar</a>
		<button class="navbar-toggler" type="button" data-toggle="collapse" data-target="#navbarSupportedContent"
			aria-controls="navbarSupportedContent" aria-expanded="false" aria-label="Toggle navigation">
			<span class="navbar-toggler-icon"></span>
		</button>

		<div class="collapse navbar-collapse" id="navbarSupportedContent">
			<ul class="navbar-nav mr-auto">
				<li class="nav-item active">
					<a class="nav-link" href="#">Home <span class="sr-only">(current)</span></a>
				</li>
				<li class="nav-item">
					<a class="nav-link" href="#">Link</a>
				</li>
				<li class="nav-item dropdown">
					<a class="nav-link dropdown-toggle" href="#" id="navbarDropdown" role="button" data-toggle="dropdown"
						aria-haspopup="true" aria-expanded="false">
						Dropdown
					</a>
					<ul class="dropdown-menu" aria-labelledby="navbarDropdown">
						<li><a class="dropdown-item" href="#">Action</a></li>
						<li><a class="dropdown-item" href="#">Another action</a></li>
						<div class="dropdown-divider"></div>
						<li></li><a class="dropdown-item" href="#">Something else here</a></li>
						<li class="nav-item dropdown">
								<a class="dropdown-item dropdown-toggle" href="#" id="navbarDropdown1" role="button" data-toggle="dropdown"
									aria-haspopup="true" aria-expanded="false">
									Dropdown
								</a>
								<ul class="dropdown-menu" aria-labelledby="navbarDropdown1">
									<li><a class="dropdown-item" href="#">Action</a></li>
									<li><a class="dropdown-item" href="#">Another action</a></li>
									<div class="dropdown-divider"></div>
									<li></li><a class="dropdown-item" href="#">Something else here</a></li>
									<li class="nav-item dropdown">
											<a class="dropdown-item dropdown-toggle" href="#" id="navbarDropdown2" role="button" data-toggle="dropdown"
												aria-haspopup="true" aria-expanded="false">
												Dropdown
											</a>
											<ul class="dropdown-menu" aria-labelledby="navbarDropdown2">
												<li><a class="dropdown-item" href="#">Action</a></li>
												<li><a class="dropdown-item" href="#">Another action</a></li>
												<div class="dropdown-divider"></div>
												<li></li><a class="dropdown-item" href="#">Something else here</a></li>
											</ul>
										</li>
								</ul>
							</li>
					</ul>
				</li>
				<li class="nav-item">
					<a class="nav-link disabled" href="#">Disabled</a>
				</li>
			</ul>
			<form class="form-inline my-2 my-lg-0">
				<input class="form-control mr-sm-2" type="search" placeholder="Search" aria-label="Search">
				<button class="btn btn-outline-success my-2 my-sm-0" type="submit">Search</button>
			</form>
		</div>
	</nav><script type="text/javascript">$('#main_navbar').bootnavbar();</script>`)*/
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
		out.ELEM("i", "class="+icon)
		out.Print(title)
	}, a)
}
