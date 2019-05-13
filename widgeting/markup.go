package widgeting

import (
	"io"
	"strings"
)

//PrintCommand definition
type PrintCommand func(...interface{})

//PrintLnCommand definition
type PrintLnCommand func(...interface{})

//OutPrint out print struct
type OutPrint struct {
	printCmd   PrintCommand
	printLnCmd PrintLnCommand
}

//Print method
func (out *OutPrint) Print(a ...interface{}) {
	if out.printCmd != nil {
		var fa []interface{}
		curai := -1
		lastcurai := -1
		for n, d := range a {
			if outfunc, outfuncok := d.(func(*OutPrint)); outfuncok {
				if curai > -1 && lastcurai > -1 {
					out.printCmd(a[curai : lastcurai+1]...)
					curai = -1
					lastcurai = -1
				}
				if fa != nil && len(fa) > 0 {
					out.printCmd(fa...)
					fa = nil
				}
				outfunc(out)
			} else {
				if curai == -1 {
					curai = n
				}
				lastcurai = n
			}
		}
		if curai > -1 && lastcurai > -1 {
			out.printCmd(a[curai : lastcurai+1]...)
			curai = -1
			lastcurai = -1
		}
	}
}

//ReplaceContent replace dynamic content
func (out *OutPrint) ReplaceContent(contentref string, cntnta ...interface{}) {
	out.StartReplaceContent(contentref)
	if len(cntnta) > 0 {
		for _, cntntd := range cntnta {
			if cntntfunc, cntntfuncok := cntntd.(func(*OutPrint)); cntntfuncok {
				cntntfunc(out)
			} else {
				out.Print(cntntd)
			}
		}
	}
	out.EndReplaceContent()
}

//StartReplaceContent indicate start of dynamic content to replace
func (out *OutPrint) StartReplaceContent(contentref string) {
	out.Print("replace-content||" + contentref + "||")
}

//EndReplaceContent indicate end of dynamic content to replace
func (out *OutPrint) EndReplaceContent() {
	out.Print("||replace-content")
}

//StartScriptContent start active script content
func (out *OutPrint) StartScriptContent() {
	out.Print("script||")
}

//EndScriptContent end active script content
func (out *OutPrint) EndScriptContent() {
	out.Print("||script")
}

//ScriptContent active script content
func (out *OutPrint) ScriptContent(contentref string, cntnta ...interface{}) {
	out.StartScriptContent()
	if len(cntnta) > 0 {
		for _, cntntd := range cntnta {
			if cntntfunc, cntntfuncok := cntntd.(func(*OutPrint)); cntntfuncok {
				cntntfunc(out)
			} else {
				out.Print(cntntd)
			}
		}
	}
	out.EndScriptContent()
}

//NewOutPrint invoke new OutPut instance
func NewOutPrint(printCmd PrintCommand, printLnCmd PrintLnCommand) (outPrint *OutPrint) {
	outPrint = &OutPrint{printCmd: printCmd, printLnCmd: printLnCmd}
	return outPrint
}

//Println method
func (out *OutPrint) Println(a ...interface{}) {
	if out.printLnCmd != nil {
		if out.printCmd != nil {
			out.Print(a...)
			out.printLnCmd()
		} else {
			out.printLnCmd(a...)
		}
	}
}

//SingleElem function
func SingleElem(out *OutPrint, tag string, props ...string) {
	out.Print("<", tag)
	ElemProperties(out, props...)
	out.Print("/>")
}

//SingleELEM element
func (out *OutPrint) SingleELEM(tag string, props ...string) {
	SingleElem(out, tag, props...)
}

//StartElem function
func StartElem(out *OutPrint, tag string, props ...string) {
	out.Print("<", tag)
	ElemProperties(out, props...)
	out.Print(">")
}

//StartELEM element
func (out *OutPrint) StartELEM(tag string, props ...string) {
	StartElem(out, tag, props...)
}

//ElemProperties function
func ElemProperties(out *OutPrint, props ...string) {
	if out != nil && (out.printCmd != nil) && len(props) > 0 {
		for _, p := range props {
			if p != "" && strings.Index(p, "=") > -1 {
				out.Print(" ", p[:strings.Index(p, "=")], "=", "\""+p[strings.Index(p, "=")+1:]+"\"")
			}
		}
	}
}

//ElemProperties element properties
func (out *OutPrint) ElemProperties(props ...string) {
	ElemProperties(out, props...)
}

//EndElem function
func EndElem(out *OutPrint, tag string) {
	out.Print("</", tag, ">")
}

//EndELEM element
func (out *OutPrint) EndELEM(tag string) {
	EndElem(out, tag)
}

//MarkupFunction definition
type MarkupFunction = func(*OutPrint)

func stripPropsAndFunctions(a ...interface{}) (props []string, funcs []MarkupFunction, startMarkupElemFunc StartMarkupElementFunction, endMarkupElemFunc EndMarkupElementFunction) {
	if len(a) > 0 {
		for _, d := range a {
			if strtElemFunc, startElemFuncOk := d.(StartMarkupElementFunction); startElemFuncOk && startMarkupElemFunc == nil {
				startMarkupElemFunc = strtElemFunc
			} else if endElemFunc, endElemFuncOk := d.(EndMarkupElementFunction); endElemFuncOk && endMarkupElemFunc == nil {
				endMarkupElemFunc = endElemFunc
			} else if p, pok := d.(string); pok {
				if props == nil {
					props = []string{}
				}
				props = append(props, p)
			} else if f, fok := d.(MarkupFunction); fok {
				if funcs == nil {
					funcs = []MarkupFunction{}
				}
				funcs = append(funcs, f)
			} else if r, rok := d.(io.Reader); rok {
				if funcs == nil {
					funcs = []MarkupFunction{}
				}
				if rs, rsok := r.(io.ReadSeeker); rsok {
					funcs = append(funcs, func(out *OutPrint) {
						rs.Seek(0, 0)
					})
				}
				funcs = append(funcs, func(out *OutPrint) {
					out.Print(r)
				})
			}
		}
	}

	return props, funcs, startMarkupElemFunc, endMarkupElemFunc
}

//StartMarkupElementFunction definition
type StartMarkupElementFunction = func(out *OutPrint, tag string, props ...string)

//EndMarkupElementFunction definition
type EndMarkupElementFunction = func(out *OutPrint, tag string)

func Elem(out *OutPrint, tag string, a ...interface{}) {
	props, funcs, strtElemFunc, endElemFunc := stripPropsAndFunctions(a...)
	if strtElemFunc == nil {
		strtElemFunc = StartElem
	}
	if endElemFunc == nil {
		endElemFunc = EndElem
	}
	strtElemFunc(out, tag, props...)
	if len(funcs) > 0 {
		for _, f := range funcs {
			f(out)
		}
	}

	endElemFunc(out, tag)
}

//ELEM element
func (out *OutPrint) ELEM(tag string, a ...interface{}) {
	Elem(out, tag, a...)
}
