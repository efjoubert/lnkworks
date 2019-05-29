package shared

import (
	"strings"

	lnksworks "../../../../lnkworks"
	"github.com/efjoubert/lnkworks/widgeting"
)

const sharedhtml = `<div id="sharedindex"></div>
<div id="sharedsection"></div>
<script>postForm({"url_ref":"shared/index.html?command=index"});</script>`

func init() {
	lnksworks.RegisterEmbededResources("shared/index.html", strings.NewReader(sharedhtml))
	lnksworks.MapActiveCommand("shared/index.html", "index", SharedIndex)
}

func SharedIndex(atvpros *lnksworks.ActiveProcessor, path string, a ...string) (err error) {
	var out = atvpros.Out()
	out.ReplaceContent("#sharedindex", func(out *widgeting.OutPrint) {
		out.ELEM("ul", "class=nav", func(out *widgeting.OutPrint) {
			out.ELEM("li", "class=nav-item", func(out *widgeting.OutPrint) {
				out.ELEM("a", "class=nav-link active webaction", "target=#sharedsection", "url_ref=shared/datasources.html",
					func(out *widgeting.OutPrint) {
						out.ELEM("i", "class=far fa-building")
						out.Print(" DATASOURCE(s)")
					})
			})
		})
	})
	out.ReplaceContent("#sharedsection")
	out.ScriptContent("", func(out *widgeting.OutPrint) {
		out.Print(`$(".webaction").click(function(e){
			e.preventDefault();
			postByElem(this);
		});`)
	})
	return
}
