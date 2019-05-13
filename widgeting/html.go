package widgeting

import (
	"strings"
)

func doctype(out *OutPrint) {
	out.Println("<DOCTYPE html>")
}

//DOCTYPE element
func (out *OutPrint) DOCTYPE() {
	doctype(out)
}

func startHTML(out *OutPrint, props ...string) {
	StartElem(out, "html", props...)
}

//StartHTML element
func (out *OutPrint) StartHTML(props ...string) {
	startHTML(out, props...)
}

func html(out *OutPrint, a ...interface{}) {
	Elem(out, "html", a...)
}

//HTML element
func (out *OutPrint) HTML(a ...interface{}) {
	html(out, a...)
}

func startHead(out *OutPrint) {
	StartElem(out, "head")
}

//StartHEAD element
func (out *OutPrint) StartHEAD() {
	startHead(out)
}

func head(out *OutPrint, a ...interface{}) {
	Elem(out, "head", a...)
}

//HEAD element
func (out *OutPrint) HEAD(a ...interface{}) {
	head(out, a...)
}

func scripts(out *OutPrint, src ...string) {
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
				startScript(out, "text/"+stype, "src="+s)
			} else {
				startScript(out, "text/javascript", "src="+s)
			}
			endScript(out)
		}
	}
}

//SCRIPTS multiple script(s)
func (out *OutPrint) SCRIPTS(src ...string) {
	scripts(out, src...)
}

func startScript(out *OutPrint, props ...string) {
	if props == nil {
		props = []string{"type=text/javascript"}
	}
	StartElem(out, "script", props...)
	//EndScript(out)
}

//StartSCRIPT element
func (out *OutPrint) StartSCRIPT(props ...string) {
	startScript(out, props...)
}

func endScript(out *OutPrint) {
	EndElem(out, "script")
}

//EndSCRIPT element
func (out *OutPrint) EndSCRIPT() {
	endScript(out)
}

func script(out *OutPrint, a ...interface{}) {
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

//SCRIPT element
func (out *OutPrint) SCRIPT(a ...interface{}) {
	script(out, a...)
}

func endHead(out *OutPrint) {
	EndElem(out, "head")
}

//EndHEAD element
func (out *OutPrint) EndHEAD() {
	endHead(out)
}

func startBody(out *OutPrint, props ...string) {
	StartElem(out, "body", props...)
}

//StartBODY element
func (out *OutPrint) StartBODY(props ...string) {
	startBody(out, props...)
}

func body(out *OutPrint, a ...interface{}) {
	Elem(out, "body", a...)
}

//BODY element
func (out *OutPrint) BODY(a ...interface{}) {
	body(out, a...)
}

func startTable(out *OutPrint, props ...string) {
	StartElem(out, "table", props...)
}

//StartTABLE element
func (out *OutPrint) StartTABLE(props ...string) {
	startTable(out, props...)
}

func endTable(out *OutPrint) {
	EndElem(out, "table")
}

//EndTABLE element
func (out *OutPrint) EndTABLE() {
	endTable(out)
}

func table(out *OutPrint, a ...interface{}) {
	Elem(out, "table", a...)
}

//TABLE element
func (out *OutPrint) TABLE(a ...interface{}) {
	table(out, a...)
}

func startTr(out *OutPrint, props ...string) {
	StartElem(out, "tr", props...)
}

//StartTR element
func (out *OutPrint) StartTR(props ...string) {
	startTr(out, props...)
}

func endTr(out *OutPrint) {
	EndElem(out, "tr")
}

//EndTR element
func (out *OutPrint) EndTR() {
	endTr(out)
}

func tr(out *OutPrint, a ...interface{}) {
	Elem(out, "tr", a...)
}

//TR element
func (out *OutPrint) TR(a ...interface{}) {
	tr(out, a...)
}

func startTHead(out *OutPrint, props ...string) {
	StartElem(out, "thead", props...)
}

//StartTHEAD element
func (out *OutPrint) StartTHEAD(props ...string) {
	startTHead(out, props...)
}

func endTHead(out *OutPrint) {
	EndElem(out, "thead")
}

//EndTHEAD element
func (out *OutPrint) EndTHEAD() {
	endTHead(out)
}

func tHead(out *OutPrint, a ...interface{}) {
	Elem(out, "thead", a...)
}

//THEAD element
func (out *OutPrint) THEAD(a ...interface{}) {
	tHead(out, a...)
}

func startTFoot(out *OutPrint, props ...string) {
	StartElem(out, "tfoot", props...)
}

//StartTFOOT element
func (out *OutPrint) StartTFOOT(props ...string) {
	startTFoot(out, props...)
}

func endTFoot(out *OutPrint) {
	EndElem(out, "tfoot")
}

//EndTFOOT element
func (out *OutPrint) EndTFOOT() {
	endTFoot(out)
}

func tFoot(out *OutPrint, a ...interface{}) {
	Elem(out, "tfoot", a...)
}

//TFOOT element
func (out *OutPrint) TFOOT(a ...interface{}) {
	tFoot(out, a...)
}

func startTh(out *OutPrint, props ...string) {
	StartElem(out, "th", props...)
}

//StartTH element
func (out *OutPrint) StartTH(props ...string) {
	startTh(out, props...)
}

func endTh(out *OutPrint) {
	EndElem(out, "th")
}

//EndTH element
func (out *OutPrint) EndTH() {
	endTh(out)
}

func th(out *OutPrint, a ...interface{}) {
	Elem(out, "th", a...)
}

//TH element
func (out *OutPrint) TH(a ...interface{}) {
	th(out, a...)
}

func startTd(out *OutPrint, props ...string) {
	StartElem(out, "td", props...)
}

//StartTD element
func (out *OutPrint) StartTD(props ...string) {
	startTd(out, props...)
}

func endTd(out *OutPrint) {
	EndElem(out, "td")
}

//EndTD element
func (out *OutPrint) EndTD() {
	endTd(out)
}

func td(out *OutPrint, a ...interface{}) {
	Elem(out, "td", a...)
}

//TD element
func (out *OutPrint) TD(a ...interface{}) {
	td(out, a...)
}

func startDiv(out *OutPrint, props ...string) {
	StartElem(out, "div", props...)
}

//StartDIV element
func (out *OutPrint) StartDIV(props ...string) {
	startDiv(out, props...)
}

func endDiv(out *OutPrint) {
	EndElem(out, "div")
}

//EndDIV element
func (out *OutPrint) EndDIV() {
	endDiv(out)
}

func div(out *OutPrint, a ...interface{}) {
	Elem(out, "div", a...)
}

//DIV element
func (out *OutPrint) DIV(a ...interface{}) {
	div(out, a...)
}

func endBody(out *OutPrint) {
	EndElem(out, "body")
}

//EndBODY element
func (out *OutPrint) EndBODY() {
	endBody(out)
}

func endHTML(out *OutPrint) {
	EndElem(out, "html")
}

//EndHTML element
func (out *OutPrint) EndHTML() {
	endHTML(out)
}

func startTextArea(out *OutPrint, props ...string) {
	StartElem(out, "textarea", props...)
}

//StartTEXTAREA element
func (out *OutPrint) StartTEXTAREA(props ...string) {
	startTextArea(out, props...)
}

func endTextArea(out *OutPrint) {
	EndElem(out, "textarea")
}

//EndTEXTAREA element
func (out *OutPrint) EndTEXTAREA() {
	endTextArea(out)
}

func textArea(out *OutPrint, a ...interface{}) {
	Elem(out, "textarea", a...)
}

//TEXTAREA element
func (out *OutPrint) TEXTAREA(a ...interface{}) {
	textArea(out, a...)
}

func input(out *OutPrint, props ...string) {
	SingleElem(out, "input", props...)
}

//INPUT input element
func (out *OutPrint) INPUT(props ...string) {
	input(out, props...)
}

func field(out *OutPrint, name string, active bool, ftype string, a ...interface{}) {
	if len(a) == 2 {
		val := a[0]
		if valfunc, valfuncok := a[0].(func(out *OutPrint, name string, active bool, ftype string, value interface{}, a ...interface{})); valfuncok {
			valfunc(out, name, active, ftype, val, a[2:]...)
		}
		a = nil
		val = nil
	}
}

//Field field
func (out *OutPrint) Field(name string, active bool, ftype string, a ...interface{}) {
	field(out, name, active, ftype, a...)
}
