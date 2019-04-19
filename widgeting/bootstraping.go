package widgeting

func (out *OutPrint) NavBar(name string, title string, a ...interface{}) {
	out.Print(`<nav class="navbar sticky-top navbar-light bg-dark">
	<a class="navbar-brand" href="#">`, title, `</a>
	<button class="navbar-toggler" type="button" data-toggle="collapse" data-target="navbar`, name, `controls" aria-controls="navbar`, name, `controls" aria-expanded="false" aria-label="Toggle navigation">
	  <span class="navbar-toggler-icon"></span>
	</button><div class="collapse navbar-collapse" id="navbar`, name, `controls">`, a, `</div></nav>`)
}
