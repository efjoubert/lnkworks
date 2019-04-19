package activerulinggui

import (
	"io"
	"strings"

	"../../../lnksworks"
	"../../../lnksworks/widgeting"
)

func ActiveRulingUI(path string, section ...string) {

	activeRulingUI(path, section...)

	lnksworks.MapActiveCommand("test/test.html",
		"testcommand", func(atvpros *lnksworks.ActiveProcessor, path string, a ...string) (err error) {

			atvpros.Out().Elem("span", func(out *widgeting.OutPrint, a ...interface{}) {
				out.Print("content in span")
			})

			return err
		},
	)
}

func ActiveRulingWeb(gui string) (r io.Reader) {
	if rRuling, rRulingOk := mappedRulingWebs[gui]; rRulingOk {
		r = rRuling()
	}
	return r
}

var mappedRulingWebs map[string]func() io.Reader
var mappedRulingWebCommands map[string]map[string]lnksworks.ActiveCommandHandler

func activeRulingUI(path string, gui ...string) {

	if len(gui) > 0 {
		for _, g := range gui {
			if r := ActiveRulingWeb(g); r != nil {
				lnksworks.RegisterEmbededResources(path+"/"+g, r)
				if _, rRulingOk := mappedRulingWebCommands[g]; rRulingOk {
					if guicmds := rulingWebCommand(g); len(guicmds) > 0 {
						lnksworks.MapActiveCommand(path+"/"+g, guicmds...)
					}
				}
			}
		}
	}
}

func rulingWebCommand(gui string) (a []interface{}) {
	if rRuleMappedCmds, rRulingOk := mappedRulingWebCommands[gui]; rRulingOk {
		for ruleCmdName, ruleCmd := range rRuleMappedCmds {
			if a == nil {
				a = []interface{}{}
			}
			a = append(a, ruleCmdName, ruleCmd)
		}
	}
	return a
}

func RegisterWebActiveRuling(readerRuling string, rruling func() io.Reader, rulingCommands ...interface{}) {
	rRulingOk := false
	if _, rRulingOk = mappedRulingWebs[readerRuling]; !rRulingOk {
		mappedRulingWebs[readerRuling] = rruling
		rRulingOk = true
	}
	if rRulingOk {
		if len(rulingCommands) > 0 && len(rulingCommands)%2 == 0 {
			if mappedRulingWebCommands[readerRuling] == nil {
				mappedRulingWebCommands[readerRuling] = map[string]lnksworks.ActiveCommandHandler{}
			}
			rRulingCmdi := 0
			for rRulingCmdi < len(rulingCommands) {
				if rRulingCmdName, rRulingCmdNameOk := rulingCommands[rRulingCmdi].(string); rRulingCmdNameOk {
					if rRulingCmd, rRulingCmdOk := rulingCommands[rRulingCmdi+1].(lnksworks.ActiveCommandHandler); rRulingCmdOk {
						mappedRulingWebCommands[readerRuling][rRulingCmdName] = rRulingCmd
					}
				}
				rRulingCmdi = rRulingCmdi + 2
			}
		}
	}
}

func init() {
	if mappedRulingWebs == nil {
		mappedRulingWebs = map[string]func() io.Reader{}
	}
	if mappedRulingWebCommands == nil {
		mappedRulingWebCommands = map[string]map[string]lnksworks.ActiveCommandHandler{}
	}
	indexgui()
}

//index.html

func indexgui() {
	RegisterWebActiveRuling(
		"index.html",
		IndexHtml,
		"validatelogin", validateLogin)
}

const indexhtml string = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
	<meta http-equiv="X-UA-Compatible" content="ie=edge">
	<link rel="stylesheet" type="text/css" href="../bootstrap-all.css|datatables.css">
	<script src="../jquery.js|block-ui.js|webactions.js"></script>
	<script src="../bootstrap-all.js|bootstrap-datatables.js|fontawesome.js"></script>
    <title>ACTIVE RULING</title>
</head>
<body id="ruling_main" style="font-size:0.8em">
	<div>
		<table>
			<tr><th>LOGIN</th><th><input type="text" name="loginu"/></th></tr>
			<tr><th>PASSWORD</th><th><input type="password" name="loginpw"/></th></tr>
			<tr><td colspan="2"><button id="cmdlogin" url_ref="?command=validatelogin" form_ref="#ruling_main" onclick="postByElem(this)">LOGIN</button></th></tr>
			<tr><td colspan="2" id="loginerror"></th></tr>
		</table>
	</div>
</body>
</html>`

func IndexHtml() io.Reader {
	return strings.NewReader(indexhtml)
}

func validateLogin(atvpros *lnksworks.ActiveProcessor, path string, a ...string) (err error) {
	if atvpros.Parameters().StringParameter("loginu", "") == "" || atvpros.Parameters().StringParameter("loginpw", "") == "" {
		atvpros.Out().ReplaceContent("#loginerror", func(out *widgeting.OutPrint) {
			out.Print("ERROR LOGIN :[", atvpros.Parameters().StringParameter("loginu", ""), ",", atvpros.Parameters().StringParameter("loginpw", ""), "]")
		})
	} else {
		atvpros.Out().ReplaceContent("#ruling_main", func(out *widgeting.OutPrint) {
			out.NavBar("testnav", "TEST NAVIGATION")
		})
	}
	return err
}
