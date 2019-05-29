package shared

import (
	"strings"

	lnksworks "../../../../lnkworks"
	"github.com/efjoubert/lnkworks/widgeting"
)

const datasourceshtml = `<div class="d-flex"><div id="datasourcesindex"></div>
	<div id="datasourcessection" class="flex-grow-1 border"></div>
</div>
<script>postForm({"url_ref":"shared/datasources.html?command=index"});</script>`

func init() {
	lnksworks.RegisterEmbededResources("shared/datasources.html", strings.NewReader(datasourceshtml))
	lnksworks.MapActiveCommand("shared/datasources.html", "index", DatasourcesIndex)
}

func DatasourcesIndex(atvpros *lnksworks.ActiveProcessor, path string, a ...string) (err error) {
	var out = atvpros.Out()
	out.ReplaceContent("#datasourcesindex", func(out *widgeting.OutPrint) {
		out.ELEM("ul", "class=nav nav-bar flex-column", func(out *widgeting.OutPrint) {
			out.ELEM("li", "class=nav-item", func(out *widgeting.OutPrint) {
				out.ELEM("a", "class=nav-link webaction firsttab", "target=#datasourcessection", "url_ref=shared/datasources/databases.html",
					func(out *widgeting.OutPrint) {
						out.ELEM("i", "class=fas fa-database")
						out.Print(" ", "DATABASE(s)")
					})
			})
			out.ELEM("li", "class=nav-item", func(out *widgeting.OutPrint) {
				out.ELEM("a", "class=nav-link webaction", "target=#datasourcessection", "url_ref=shared/datasources/files.html",
					func(out *widgeting.OutPrint) {
						out.ELEM("i", "class=far fa-file-code")
						out.Print(" ", "FILE(s)")
					})
			})
			out.ELEM("li", "class=nav-item", func(out *widgeting.OutPrint) {
				out.ELEM("a", "class=nav-link webaction", "target=#datasourcessection", "url_ref=shared/datasources/urls.html",
					func(out *widgeting.OutPrint) {
						out.ELEM("i", "class=fas fa-globe")
						out.Print(" ", "WEB")
					})
			})
			out.ELEM("li", "class=nav-item", func(out *widgeting.OutPrint) {
				out.ELEM("a", "class=nav-link webaction", "target=#datasourcessection", "url_ref=shared/datasources/messaging.html",
					func(out *widgeting.OutPrint) {
						out.ELEM("i", "class=far fa-envelope")
						out.Print(" ", "MESSAGING")
					})
			})
		})
	})
	out.ReplaceContent("#datasourcessection")
	out.ScriptContent("", func(out *widgeting.OutPrint) {
		out.Print(`$(".webaction").click(function(e){
			e.preventDefault();
			$(this).tab('show');
			postByElem(this);
		});
		$(".webaction.firsttab").click();
		`)
	})
	return
}
