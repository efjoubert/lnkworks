package widgeting

import (
	"strings"
)

func Doctype(out *OutPrint) {
	out.Println("<DOCTYPE html>")
}

func (out *OutPrint) Doctype() {
	Doctype(out)
}

func StartHtml(out *OutPrint, props ...string) {
	StartElem(out, "html", props...)
}

func (out *OutPrint) StartHtml(props ...string) {
	StartHtml(out, props...)
}

func Html(out *OutPrint, a ...interface{}) {
	Elem(out, "html", a...)
}

func (out *OutPrint) Html(a ...interface{}) {
	Html(out, a...)
}

func StartHead(out *OutPrint) {
	StartElem(out, "head")
}

func (out *OutPrint) StartHead() {
	StartHead(out)
}

func Head(out *OutPrint, a ...interface{}) {
	Elem(out, "head", a...)
}

func Scripts(out *OutPrint, src ...string) {
	if len(src) > 0 {
		si := 0
		for si < len(src) {
			s := src[si]
			if strings.Index(s, ":") > -1 && strings.Index(s, "://") == -1 || strings.Index(s, ":") < strings.Index(s, "://") {
				stype := s[:strings.Index(s, ":")]
				if stype == "js" {
					stype = "javascript"
				}
				s = s[strings.Index(s, ":")+1:]
				StartScript(out, "text/"+stype, "src="+s)
			} else {
				StartScript(out, "text/javascript", "src="+s)
			}
			EndScript(out)
		}
	}
}

func (out *OutPrint) Scripts(src ...string) {
	Scripts(out, src...)
}

func StartScript(out *OutPrint, props ...string) {
	if props == nil {
		props = []string{"type=text/javascript"}
	}
	StartElem(out, "script", props...)
	EndScript(out)
}

func (out *OutPrint) StartScript(props ...string) {
	StartScript(out, props...)
}

func EndScript(out *OutPrint) {
	EndElem(out, "script")
}

func (out *OutPrint) EndScript() {
	EndScript(out)
}

func Script(out *OutPrint, a ...interface{}) {
	if a == nil {
		if a == nil {
			a = []interface{}{}
		}
	}
	ai := 0
	hasType := false
	for ai < len(a) {
		if s, sok := a[ai].(string); sok && strings.Index(s, "=") > -1 && strings.Replace(s[:strings.Index(s, "=")], " ", "", -1) == "type" {
			if s[strings.Index(s, "=")+1:] != "" {
				hasType = true
			}
			break
		}
		ai++
	}
	if !hasType {
		if len(a) == 0 {
			a = append(a, "type=text/javascript")
		} else {
			na := make([]interface{}, len(a))
			na[0] = "type=text/javascript"
			copy(na[1:], a)
			a = nil
			a = na
			na = nil
		}
	}
	Elem(out, "script", a...)
}

func (out *OutPrint) Script(a ...interface{}) {
	Script(out, a...)
}

func EndHead(out *OutPrint) {
	EndElem(out, "head")
}

func (out *OutPrint) EndHead() {
	EndHead(out)
}

func StartBody(out *OutPrint, props ...string) {
	StartElem(out, "body", props...)
}

func (out *OutPrint) StartBody(props ...string) {
	StartBody(out, props...)
}

func Body(out *OutPrint, a ...interface{}) {
	Elem(out, "body", a...)
}

func StartTable(out *OutPrint, props ...string) {
	StartElem(out, "table", props...)
}

func (out *OutPrint) StartTable(props ...string) {
	StartTable(out, props...)
}

func EndTable(out *OutPrint) {
	EndElem(out, "table")
}

func (out *OutPrint) EndTable() {
	EndTable(out)
}

func Table(out *OutPrint, a ...interface{}) {
	Elem(out, "table", a...)
}

func StartTr(out *OutPrint, props ...string) {
	StartElem(out, "tr", props...)
}

func (out *OutPrint) StartTr(props ...string) {
	StartTr(out, props...)
}

func EndTr(out *OutPrint) {
	EndElem(out, "tr")
}

func (out *OutPrint) EndTr() {
	EndTr(out)
}

func Tr(out *OutPrint, a ...interface{}) {
	Elem(out, "tr", a...)
}

func (out *OutPrint) Tr(a ...interface{}) {
	Tr(out, a...)
}

func StartTHead(out *OutPrint, props ...string) {
	StartElem(out, "thead", props...)
}

func (out *OutPrint) StartTHead(props ...string) {
	StartTHead(out, props...)
}

func EndTHead(out *OutPrint) {
	EndElem(out, "thead")
}

func (out *OutPrint) EndTHead() {
	EndTHead(out)
}

func THead(out *OutPrint, a ...interface{}) {
	Elem(out, "thead", a...)
}

func (out *OutPrint) THead(a ...interface{}) {
	THead(out, a...)
}

func StartTFoot(out *OutPrint, props ...string) {
	StartElem(out, "tfoot", props...)
}

func (out *OutPrint) StartTFoot(props ...string) {
	StartTFoot(out, props...)
}

func EndTFoot(out *OutPrint) {
	EndElem(out, "tfoot")
}

func (out *OutPrint) EndTFoot() {
	EndTFoot(out)
}

func TFoot(out *OutPrint, a ...interface{}) {
	Elem(out, "tfoot", a...)
}

func (out *OutPrint) TFoot(a ...interface{}) {
	TFoot(out, a...)
}

func (out *OutPrint) Table(a ...interface{}) {
	Table(out, a...)
}

func StartTh(out *OutPrint, props ...string) {
	StartElem(out, "th", props...)
}

func (out *OutPrint) StartTh(props ...string) {
	StartTh(out, props...)
}

func EndTh(out *OutPrint) {
	EndElem(out, "th")
}

func (out *OutPrint) EndTh() {
	EndTh(out)
}

func Th(out *OutPrint, a ...interface{}) {
	Elem(out, "th", a...)
}

func (out *OutPrint) Th(a ...interface{}) {
	Th(out, a...)
}

func StartTd(out *OutPrint, props ...string) {
	StartElem(out, "td", props...)
}

func (out *OutPrint) StartTd(props ...string) {
	StartTd(out, props...)
}

func EndTd(out *OutPrint) {
	EndElem(out, "td")
}

func (out *OutPrint) EndTd() {
	EndTd(out)
}

func Td(out *OutPrint, a ...interface{}) {
	Elem(out, "td", a...)
}

func (out *OutPrint) Td(a ...interface{}) {
	Td(out, a...)
}

func StartDiv(out *OutPrint, props ...string) {
	StartElem(out, "div", props...)
}

func (out *OutPrint) StartDiv(props ...string) {
	StartDiv(out, props...)
}

func EndDiv(out *OutPrint) {
	EndElem(out, "div")
}

func (out *OutPrint) EndDiv() {
	EndDiv(out)
}

func Div(out *OutPrint, a ...interface{}) {
	Elem(out, "div", a...)
}

func (out *OutPrint) Div(a ...interface{}) {
	Div(out, a...)
}

func EndBody(out *OutPrint) {
	EndElem(out, "body")
}

func (out *OutPrint) EndBody() {
	EndBody(out)
}

func EndHtml(out *OutPrint) {
	EndElem(out, "html")
}

func (out *OutPrint) EndHtml() {
	EndHtml(out)
}

func StartTextArea(out *OutPrint, props ...string) {
	StartElem(out, "textarea", props...)
}

func (out *OutPrint) StartTextArea(props ...string) {
	StartTextArea(out, props...)
}

func EndTextArea(out *OutPrint) {
	EndElem(out, "textarea")
}

func (out *OutPrint) EndTextArea() {
	EndTextArea(out)
}

func TextArea(out *OutPrint, a ...interface{}) {
	Elem(out, "textarea", a...)
}

func (out *OutPrint) TextArea(a ...interface{}) {
	TextArea(out, a...)
}

func Input(out *OutPrint, props ...string) {
	SingleElem(out, "input", props...)
}

func (out *OutPrint) Input(props ...string) {
	Input(out, props...)
}

func Field(out *OutPrint, name string, active bool, ftype string, a ...interface{}) {
	if len(a) == 2 {
		val := a[0]
		if valfunc, valfuncok := a[0].(func(out *OutPrint, name string, active bool, ftype string, value interface{}, a ...interface{})); valfuncok {
			valfunc(out, name, active, ftype, val, a[2:]...)
		}
		a = nil
		val = nil
	}
}

func (out *OutPrint) Field(name string, active bool, ftype string, a ...interface{}) {
	Field(out, name, active, ftype, a...)
}
