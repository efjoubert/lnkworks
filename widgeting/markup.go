package widgeting

import (
	"io"
	"strings"
)

type PrintCommand func(...interface{})

type PrintLnCommand func(...interface{})

type OutPrint struct {
	printCmd   PrintCommand
	printLnCmd PrintLnCommand
}

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

func (out *OutPrint) StartReplaceContent(contentref string) {
	out.Print("replace-content||" + contentref + "||")
}

func (out *OutPrint) EndReplaceContent() {
	out.Print("||replace-content")
}

func (out *OutPrint) StartScriptContent() {
	out.Print("script||")
}

func (out *OutPrint) EndScriptContent() {
	out.Print("||script")
}

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

func NewOutPrint(printCmd PrintCommand, printLnCmd PrintLnCommand) (outPrint *OutPrint) {
	outPrint = &OutPrint{printCmd: printCmd, printLnCmd: printLnCmd}
	return outPrint
}

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

func SingleElem(out *OutPrint, tag string, props ...string) {
	out.Print("<", tag)
	ElemProperties(out, props...)
	out.Print("/>")
}

func (out *OutPrint) SingleElem(tag string, props ...string) {
	SingleElem(out, tag, props...)
}

func StartElem(out *OutPrint, tag string, props ...string) {
	out.Print("<", tag)
	ElemProperties(out, props...)
	out.Print(">")
}

func (out *OutPrint) StartElem(tag string, props ...string) {
	StartElem(out, tag, props...)
}

func ElemProperties(out *OutPrint, props ...string) {
	if out != nil && (out.printCmd != nil) && len(props) > 0 {
		for _, p := range props {
			if p != "" && strings.Index(p, "=") > -1 {
				out.Print(" ", p[:strings.Index(p, "=")], "=", "\""+p[strings.Index(p, "=")+1:]+"\"")
			}
		}
	}
}

func (out *OutPrint) ElemProperties(props ...string) {
	ElemProperties(out, props...)
}

func EndElem(out *OutPrint, tag string) {
	out.Print("</", tag, ">")
}

func (out *OutPrint) EndElem(tag string) {
	EndElem(out, tag)
}

type MarkupFunction func(*OutPrint, ...interface{})

func stripPropsAndFunctions(a ...interface{}) (props []string, funcs []MarkupFunction) {
	var unmatched []interface{}
	if len(a) > 0 {
		for _, d := range a {
			if p, pok := d.(string); pok {
				if props == nil {
					props = []string{}
				}
				props = append(props, p)
			} else if f, fok := d.(MarkupFunction); fok {
				if funcs == nil {
					funcs = []MarkupFunction{}
				}
				funcs = append(funcs, f)
			} else if prepcntnt, prepcntntok := d.(PrepContent); prepcntntok {
				if funcs == nil {
					funcs = []MarkupFunction{}
				}
				funcs = append(funcs, func(out *OutPrint, a ...interface{}) {
					if unmatched == nil || len(unmatched) == 0 {
						prepcntnt(out)
					} else {
						prepcntnt(out, unmatched)
						unmatched = nil
					}
				})
			} else if funccntnt, funccntntok := d.(func(*OutPrint, ...interface{})); funccntntok {
				if funcs == nil {
					funcs = []MarkupFunction{}
				}
				funcs = append(funcs, funccntnt)
			} else if r, rok := d.(io.Reader); rok {
				if funcs == nil {
					funcs = []MarkupFunction{}
				}
				if rs, rsok := r.(io.ReadSeeker); rsok {
					funcs = append(funcs, func(out *OutPrint, a ...interface{}) {
						rs.Seek(0, 0)
					})
				}
				funcs = append(funcs, func(out *OutPrint, a ...interface{}) {
					out.Print(r)
				})
			}
		}
	}

	return props, funcs
}

type PrepContent func(*OutPrint, ...interface{})

func Elem(out *OutPrint, tag string, a ...interface{}) {
	props, funcs := stripPropsAndFunctions(a...)
	StartElem(out, tag, props...)
	if len(funcs) > 0 {
		for _, f := range funcs {
			f(out)
		}
	}

	EndElem(out, tag)
}

func (out *OutPrint) Elem(tag string, a ...interface{}) {
	Elem(out, tag, a...)
}
