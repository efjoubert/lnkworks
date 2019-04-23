package lnksworks

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/dop251/goja"
	"github.com/efjoubert/lnkworks/widgeting"
)

type activeReadSeekerPoint struct {
	atvrsi      int
	atvrs       *ActiveReadSeeker
	atvrsstarti int64
	atvrsendi   int64
	atvrspointi int
}

type ActiveParser struct {
	atv          *ActiveProcessor
	retrieveRS   func(string, string) (io.ReadSeeker, error)
	atvParseFunc func(*ActiveParseToken) bool
	psvParseFunc func(*ActiveParseToken) bool
	tknrb        []byte
	//unparsedIOs      []*IORW
	atvrsmap         map[string]*ActiveReadSeeker
	atvrscntntcdemap map[int]*ActiveReadSeeker
	atvrspntsentries []*activeReadSeekerPoint
	atvtrsstartpoint *activeReadSeekerPoint
}

func (atvparse *ActiveParser) appendActiveRSPoint(atvrs *ActiveReadSeeker, isStart bool) {
	if atvparse.atvrspntsentries == nil {
		atvparse.atvrspntsentries = []*activeReadSeekerPoint{}
	}
	atvrspoint := &activeReadSeekerPoint{atvrs: atvrs, atvrsi: atvrs.atvrsi, atvrspointi: len(atvparse.atvrspntsentries)}
	atvparse.atvrspntsentries = append(atvparse.atvrspntsentries, atvrspoint)
	if isStart {
		atvparse.atvtrsstartpoint = atvrspoint
	}
	atvrspoint = nil
}

func (atvparse *ActiveParser) atvrs(path string) (atvrs *ActiveReadSeeker) {
	if atvparse.atvrsmap != nil {
		atvrs, _ = atvparse.atvrsmap[path]
	}
	return atvrs
}

func (atvparse *ActiveParser) setRS(path string, rstomap io.ReadSeeker) {
	if rstomap != nil {
		if atvparse.atvrsmap == nil {
			atvparse.atvrsmap = map[string]*ActiveReadSeeker{}
		}
		if atvparse.atvrscntntcdemap == nil {
			atvparse.atvrscntntcdemap = map[int]*ActiveReadSeeker{}
		}
		if _, atvrsok := atvparse.atvrsmap[path]; !atvrsok {
			atvrsio, _ := NewIORW(rstomap)
			atvrs := &ActiveReadSeeker{atvrsio: atvrsio, atvparse: atvparse}
			cntntrs := newContentSeekReader(atvrs)
			atvrs.cntntRS = cntntrs
			cders := newCodeSeekReader(atvrs, cntntrs)
			atvrs.cdeRS = cders
			atvparse.atvrsmap[path] = atvrs
			atvparse.atvrscntntcdemap[len(atvparse.atvrsmap)-1] = atvrs
		}
	}
}

type contentSeekReader struct {
	atvrs        *ActiveReadSeeker
	lastStartRsI int
	lastEndRsI   int
	*IOSeekReader
}

func (cntntsr *contentSeekReader) Append(starti int64, endi int64) {
	cntntsr.IOSeekReader.Append(starti, endi)
	if cntntsr.lastStartRsI == -1 {
		cntntsr.lastStartRsI = len(cntntsr.seekis) - 1
	}
	if cntntsr.lastStartRsI > -1 {
		cntntsr.lastEndRsI = len(cntntsr.seekis) - 1
	}
}

func (cntntsr *contentSeekReader) clearContentSeekReader() {
	if cntntsr.IOSeekReader != nil {
		cntntsr.IOSeekReader.ClearIOSeekReader()
		cntntsr.IOSeekReader = nil
	}
}

type codeSeekReader struct {
	atvrs *ActiveReadSeeker
	*IOSeekReader
	cntntsr           *contentSeekReader
	lastCntntStartI   int64
	cntntseekristart  map[int64][]int
	cntntseekriendpos []int
}

func (cdesr *codeSeekReader) String() (s string) {
	if len(cdesr.seekis) > 0 {
		for skrsipos, skrsi := range cdesr.seekis {
			if cntntpos, cntntposok := cdesr.cntntseekristart[skrsi[0]]; cntntposok {
				for _, cntntposi := range cntntpos {
					s = s + "_atvparse.WriteContentByPos(" + fmt.Sprintf("%d,%d", cdesr.atvrs.atvrsi, cntntposi) + ");"
				}
			}
			if ss, sserr := cdesr.StringSeedPos(skrsipos, 0); sserr == nil {
				s = s + ss
			} else {
				panic(sserr)
			}
		}
		if len(cdesr.cntntseekriendpos) > 0 {
			for _, cntntseekriendpos := range cdesr.cntntseekriendpos {
				s = s + "_atvparse.WriteContentByPos(" + fmt.Sprintf("%d,%d", cdesr.atvrs.atvrsi, cntntseekriendpos) + ");"
			}
		}
	}
	return s
}

func (cdesr *codeSeekReader) Append(starti int64, endi int64) {
	if !cdesr.cntntsr.Empty() {
		if cdesr.cntntseekristart == nil {
			cdesr.cntntseekristart = make(map[int64][]int)
		}
		if _, istartok := cdesr.cntntseekristart[starti]; !istartok {
			if cdesr.cntntsr.lastStartRsI > -1 && cdesr.cntntsr.lastStartRsI <= cdesr.cntntsr.lastEndRsI {
				cntids := make([]int, (cdesr.cntntsr.lastEndRsI-cdesr.cntntsr.lastStartRsI)+1)
				for cntidsn, _ := range cntids {
					cntids[cntidsn] = cdesr.cntntsr.lastStartRsI
					cdesr.cntntsr.lastStartRsI++
				}
				cdesr.cntntseekristart[starti] = cntids[:]
				cdesr.cntntsr.lastStartRsI = -1
				cdesr.cntntsr.lastEndRsI = -1
				cntids = nil
			}
		}
	}
	cdesr.IOSeekReader.Append(starti, endi)
}

func (cdesr *codeSeekReader) clearCodeSeekReader() {
	if cdesr.IOSeekReader != nil {
		cdesr.IOSeekReader.ClearIOSeekReader()
		cdesr.IOSeekReader = nil
	}
	if cdesr.cntntsr != nil {
		cdesr.cntntsr = nil
	}
	if cdesr.cntntseekristart != nil {
		for cntntseekristartKey, _ := range cdesr.cntntseekristart {
			cdesr.cntntseekristart[cntntseekristartKey] = nil
			delete(cdesr.cntntseekristart, cntntseekristartKey)
		}
		cdesr.cntntseekristart = nil
	}
	if cdesr.cntntseekriendpos != nil {
		cdesr.cntntseekriendpos = nil
	}
}

func newContentSeekReader(atvrs *ActiveReadSeeker) *contentSeekReader {
	cntntsr := &contentSeekReader{atvrs: atvrs, IOSeekReader: NewIOSeekReader(atvrs), lastStartRsI: -1, lastEndRsI: -1}
	return cntntsr
}

func newCodeSeekReader(atvrs *ActiveReadSeeker, cntntsr *contentSeekReader) *codeSeekReader {
	cdesr := &codeSeekReader{atvrs: atvrs, IOSeekReader: NewIOSeekReader(atvrs), lastCntntStartI: -1, cntntsr: cntntsr}
	return cdesr
}

var emptyIO *IORW = &IORW{cached: false}

func (atvparse *ActiveParser) WriteContentByPos(rsi int, pos int) {
	if atvparse.atvrsmap != nil && atvparse.atvrscntntcdemap != nil {
		if atvrs, atvrsok := atvparse.atvrscntntcdemap[rsi]; atvrsok && atvrs.cntntRS != nil && len(atvrs.cntntRS.seekis) > 0 {
			atvrs.cntntRS.WriteSeekedPos(atvparse.atv.w, pos, 0)
		}
	}
}

func (atvparse *ActiveParser) readCurrentUnparsedTokenIO(token *ActiveParseToken, p []byte) (n int, err error) {
	n, err = token.atvrs.Read(p)
	if err == io.EOF {
		if token.endRIndex > 0 {
			token.eofEndRIndex = token.endRIndex - 1
		} else {
			token.eofEndRIndex = token.endRIndex
		}
		token.startRIndex = 0
		token.endRIndex = 0
		/*if len(atvparse.unparsedIOs) == 1 {
			atvparse.unparsedIOs[0].Seek(0, 0)
		} else {
			atvparse.unparsedIOs[0].Close()
		}
		atvparse.unparsedIOs[0] = nil
		if len(atvparse.unparsedIOs) > 1 {
			atvparse.unparsedIOs = atvparse.unparsedIOs[1:]
		} else {
			atvparse.unparsedIOs = nil
		}*/
	} else {
		token.lastEndRIndex = token.endRIndex
		token.startRIndex = token.endRIndex
		token.endRIndex += int64(n)
	}
	return n, err
}

func (atvparse *ActiveParser) cleanupActiveParser() {
	if atvparse.atvrscntntcdemap != nil {
		for atvcntk, _ := range atvparse.atvrscntntcdemap {
			atvparse.atvrscntntcdemap[atvcntk].cleanupActiveReadSeeker()
			atvparse.atvrscntntcdemap[atvcntk] = nil
			delete(atvparse.atvrscntntcdemap, atvcntk)
		}
		atvparse.atvrscntntcdemap = nil
	}
	if atvparse.atvrsmap != nil {
		for rspath, _ := range atvparse.atvrsmap {
			atvparse.atvrsmap[rspath] = nil
			delete(atvparse.atvrsmap, rspath)
		}
		atvparse.atvrsmap = nil
	}
}

func (atvparse *ActiveParser) readActive(token *ActiveParseToken) (tknrn int, tknrnerr error) {
	tknrn, tknrnerr = atvparse.readCurrentUnparsedTokenIO(token, token.tknrb)
	return tknrn, tknrnerr
}

func (atvparse *ActiveParser) readPassive(token *ActiveParseToken) (tknrn int, tknrnerr error) {
	if token.psvCapturedIO != nil {
		tknrn, tknrnerr = token.psvCapturedIO.Read(token.tknrb)
	} else {
		tknrn, tknrnerr = emptyIO.Read(token.tknrb)
	}
	return tknrn, tknrnerr
}

type ActiveReadSeeker struct {
	atvpros   *ActiveProcessor
	atvparse  *ActiveParser
	atvrsio   *IORW
	cntntRS   *contentSeekReader
	cdeRS     *codeSeekReader
	atvrsi    int
	atvrspath string
}

func (atvrs *ActiveReadSeeker) writeAllContent(w io.Writer) {
	if !atvrs.cntntRS.Empty() {
		if len(atvrs.cntntRS.seekis) > 0 {
			if atvrs.atvparse.atv.out == nil {
				for spos := range atvrs.cntntRS.seekis {
					atvrs.cntntRS.WriteSeekedPos(w, spos, 0)
				}
			} else {
				for spos := range atvrs.cntntRS.seekis {
					atvrs.cntntRS.WriteSeekedPos(atvrs.atvparse.atv.out, spos, 0)
				}
			}
		}
	}
}

func (atvRS *ActiveReadSeeker) cleanupActiveReadSeeker() {
	if atvRS.atvparse != nil {
		atvRS.atvparse = nil
	}
	if atvRS.atvpros != nil {
		atvRS.atvpros = nil
	}
	if atvRS.atvrsio != nil {
		atvRS.atvrsio.Close()
		atvRS.atvrsio = nil
	}
	if atvRS.cdeRS != nil {
		atvRS.cdeRS.clearCodeSeekReader()
		atvRS.cdeRS = nil
	}
	if atvRS.cntntRS != nil {
		atvRS.cntntRS.clearContentSeekReader()
		atvRS.cntntRS = nil
	}
}

func (atvrs *ActiveReadSeeker) Seek(offset int64, whence int) (n int64, err error) {
	return atvrs.atvrsio.Seek(offset, whence)
}

func (atvrs *ActiveReadSeeker) Read(p []byte) (n int, err error) {
	return atvrs.atvrsio.Read(p)
}

func (atvparse *ActiveParser) code() (s string) {

	return s
}

func (atvparse *ActiveParser) parse(rs io.ReadSeeker, root string, path string, retrieveRS func(string, string) (io.ReadSeeker, error), atvparsfunc func(*ActiveParseToken, []string, []int) (bool, error), psvparsefunc func(*ActiveParseToken, []string, []int) (bool, error)) {
	fmt.Println("start parsing:" + path)
	if atvparse.retrieveRS == nil || &atvparse.retrieveRS != &retrieveRS {
		atvparse.retrieveRS = retrieveRS
	}
	atvparse.setRS(path, rs)
	atvtoken := nextActiveParseToken(nil, atvparse, path, atvparsfunc, psvparsefunc, true, isJsExtension(filepath.Ext(path)))
	//queueActiveToken(atvtoken)
	//if atvparse.unparsedIOs == nil {
	//	atvparse.unparsedIOs = []*IORW{}
	//}
	var tokenparsed bool
	var tokenerr error
	var prevtoken *ActiveParseToken
	for {
		if tokenparsed, tokenerr = atvtoken.parsing(); tokenparsed || tokenerr != nil {
			if tokenparsed && tokenerr == nil {
				atvtoken.wrapupActiveParseToken()
			}
			prevtoken, tokenerr = atvtoken.cleanupActiveParseToken()
			if tokenerr != nil {
				for prevtoken != nil {
					prevtoken, _ = prevtoken.cleanupActiveParseToken()
				}
			}
			atvtoken = prevtoken
			if atvtoken == nil {
				break
			}
		}
	}

	if atvparse.atvrspntsentries != nil && len(atvparse.atvrspntsentries) > 0 {
		atvparse.evalStartActiveEntryPoint(atvparse.atvtrsstartpoint)
	}
}

func (atvparse *ActiveParser) evalStartActiveEntryPoint(atvrsstartpoint *activeReadSeekerPoint) {
	if !atvrsstartpoint.atvrs.cdeRS.Empty() {
		s := ""
		s = atvrsstartpoint.atvrs.cdeRS.String()
		atvparse.atv.evalCode(func() string {
			return s
		}, map[string]interface{}{"_out": atvparse.atv.Out(), "_atvparse": atvparse, "_parameters": atvparse.atv.params, "@db@execute": func(alias string, query string, args ...interface{}) *DBExecuted {
			return DatabaseManager().Execute(alias, query, args...)
		}, "@db@query": func(alias string, query string, args ...interface{}) *DBQuery {
			return DatabaseManager().Query(alias, query, args...)
		}})
	} else {
		atvrsstartpoint.atvrs.writeAllContent(atvparse.atv.w)
	}
}

func nextActiveParseToken(token *ActiveParseToken, parser *ActiveParser, rspath string, atvparsefunc func(*ActiveParseToken, []string, []int) (bool, error), psvparsefunc func(*ActiveParseToken, []string, []int) (bool, error), isactive bool, isjs bool) (nexttoken *ActiveParseToken) {
	nexttoken = &ActiveParseToken{startRIndex: 0, endRIndex: 0, prevtoken: token, parse: parser, atvparsefunc: atvparsefunc, psvparsefunc: psvparsefunc, isactive: isactive, tknrb: make([]byte, 1), curStartIndex: -1, curEndIndex: -1, rspath: rspath, atvRStartIndex: -1, atvREndIndex: -1, psvRStartIndex: -1, psvREndIndex: -1, tokenMde: tokenActive}
	nexttoken.atvlbls = []string{"<@", "@>"}
	nexttoken.psvlbls = []string{nexttoken.atvlbls[0][0 : len(nexttoken.atvlbls)-1], nexttoken.atvlbls[1][1:]}
	nexttoken.atvlblsi = []int{0, 0}
	nexttoken.psvlblsi = []int{0, 0}
	nexttoken.atvrs = parser.atvrs(rspath)
	return nexttoken
}

type tokenMode int

const (
	tokenActive  tokenMode = 0
	tokenPassive tokenMode = 1
)

type ActiveParseToken struct {
	parse    *ActiveParser
	atvlbls  []string
	atvlblsi []int
	atvprevb byte
	hasAtv   bool
	//atvCapturedIO *IORW
	psvlbls       []string
	psvlblsi      []int
	psvprevb      byte
	psvCapturedIO *IORW
	tknrb         []byte
	nr            int
	rerr          error
	atvrs         *ActiveReadSeeker //   io.ReadSeeker
	atvparsefunc  func(*ActiveParseToken, []string, []int) (bool, error)
	psvparsefunc  func(*ActiveParseToken, []string, []int) (bool, error)
	curStartIndex int64
	curEndIndex   int64
	prevtoken     *ActiveParseToken
	isactive      bool
	rspath        string
	//
	startRIndex    int64
	lastEndRIndex  int64
	endRIndex      int64
	eofEndRIndex   int64
	atvRStartIndex int64
	atvREndIndex   int64
	psvRStartIndex int64
	psvREndIndex   int64
	//cdeSR          *codeSeekReader
	//cntntSR        *contentSeekReader
	tokenMde tokenMode
}

func (token *ActiveParseToken) appendCde(atvrs *ActiveReadSeeker, atvRStartIndex int64, atvREndIndex int64) {
	atvrs.cdeRS.Append(atvRStartIndex, atvREndIndex)
}

func (token *ActiveParseToken) appendCntnt(atvrs *ActiveReadSeeker, psvRStartIndex int64, psvREndIndex int64) {
	atvrs.cntntRS.Append(psvRStartIndex, psvREndIndex)
}

func (token *ActiveParseToken) wrapupActiveParseToken() {
	if token.atvrs != nil {
		if !token.atvrs.cdeRS.Empty() {
			if !token.atvrs.cdeRS.Empty() {
				if token.atvrs.cdeRS.cntntseekriendpos == nil || len(token.atvrs.cdeRS.cntntseekriendpos) == 0 {
					if token.atvrs.cntntRS.lastStartRsI > -1 && token.atvrs.cntntRS.lastStartRsI <= token.atvrs.cntntRS.lastEndRsI {
						cntntrsseeki := make([]int, (token.atvrs.cntntRS.lastEndRsI-token.atvrs.cntntRS.lastStartRsI)+1)
						for cntntrsseekin, _ := range cntntrsseeki {
							cntntrsseeki[cntntrsseekin] = token.atvrs.cntntRS.lastStartRsI
							token.atvrs.cntntRS.lastStartRsI++
						}
						token.atvrs.cntntRS.lastStartRsI = -1
						token.atvrs.cntntRS.lastEndRsI = -1
						token.atvrs.cdeRS.cntntseekriendpos = append(token.atvrs.cdeRS.cntntseekriendpos, cntntrsseeki[:]...)
						cntntrsseeki = nil
					}
				}
				token.parse.appendActiveRSPoint(token.atvrs, token.prevtoken == nil)
			}
		} else if !token.atvrs.cntntRS.Empty() {
			token.parse.appendActiveRSPoint(token.atvrs, token.prevtoken == nil)
		}
	}
}

func (token *ActiveParseToken) cleanupActiveParseToken() (prevtoken *ActiveParseToken, err error) {
	if token.atvlbls != nil {
		token.atvlbls = nil
	}
	if token.atvlblsi != nil {
		token.atvlblsi = nil
	}
	if token.atvparsefunc != nil {
		token.atvparsefunc = nil
	}
	if token.parse != nil {
		token.parse = nil
	}
	if token.prevtoken != nil {
		prevtoken = token.prevtoken
		token.prevtoken = nil
	}
	if token.psvlbls != nil {
		token.psvlbls = nil
	}
	if token.psvlblsi != nil {
		token.psvlblsi = nil
	}
	if token.psvparsefunc != nil {
		token.psvparsefunc = nil
	}
	if token.rerr != nil {
		token.rerr = nil
	}
	if token.psvCapturedIO != nil {
		token.psvCapturedIO.Close()
		token.psvCapturedIO = nil
	}
	if token.tknrb != nil {
		token.tknrb = nil
	}
	return prevtoken, err
}

func (token *ActiveParseToken) passiveCapturedIO() *IORW {
	if token.psvCapturedIO == nil {
		token.psvCapturedIO, _ = NewIORW()
	}
	return token.psvCapturedIO
}

func (token *ActiveParseToken) parsing() (parsed bool, err error) {
	if token.tokenMde == tokenActive {
		return token.parsingActive()
	} else if token.tokenMde == tokenPassive {
		return token.parsingPassive()
	} else {
		err = fmt.Errorf("INVALID PARSING POINT READ")
	}
	return parsed, err
}

func (token *ActiveParseToken) parsingActive() (parsed bool, err error) {
	token.nr, token.rerr = token.parse.readActive(token)
	return token.atvparsefunc(token, token.atvlbls, token.atvlblsi)
}

func (token *ActiveParseToken) parsingPassive() (parsed bool, err error) {
	token.nr, token.rerr = token.parse.readPassive(token)
	return token.psvparsefunc(token, token.psvlbls, token.psvlblsi)
}

//ParseActiveToken - Default ParseActiveToken method
func ParseActiveToken(token *ActiveParseToken, lbls []string, lblsi []int) (nextparse bool, err error) {
	if token.nr > 0 {
		if lblsi[1] == 0 && lblsi[0] < len(lbls[0]) {
			if lblsi[0] > 1 && lbls[0][lblsi[0]-1] == token.atvprevb && lbls[0][lblsi[0]] != token.tknrb[0] {
				token.passiveCapturedIO().Print(lbls[0][:lblsi[0]])
				if token.curStartIndex == -1 {
					token.curEndIndex = token.curStartIndex - int64(lblsi[0])
				}
				lblsi[0] = 0
				token.atvprevb = 0
			}
			if lbls[0][lblsi[0]] == token.tknrb[0] {
				lblsi[0]++
				if len(lbls[0]) == lblsi[0] {
					token.atvprevb = 0
					return nextparse, err
				} else {
					token.atvprevb = token.tknrb[0]
					return nextparse, err
				}
			} else {
				if token.curStartIndex == -1 {
					if lblsi[0] > 0 {
						token.curStartIndex = token.startRIndex - int64(lblsi[0])
					} else {
						token.curStartIndex = token.startRIndex
					}
				}
				if lblsi[0] > 0 {
					token.passiveCapturedIO().Print(lbls[0][:lblsi[0]])
					lblsi[0] = 0
				}
				token.atvprevb = token.tknrb[0]
				token.passiveCapturedIO().Print(token.tknrb)
				return nextparse, err
			}
		} else if lblsi[0] == len(lbls[0]) && lblsi[1] < len(lbls[1]) {
			if lbls[1][lblsi[1]] == token.tknrb[0] {
				lblsi[1]++
				if lblsi[1] == len(lbls[1]) {
					if token.atvRStartIndex > -1 && token.atvREndIndex > -1 {
						token.appendCde(token.atvrs, token.atvRStartIndex, token.atvREndIndex)
						//IOSeekReaderOutput(token.parse.cdeSR).Append(token.atvRStartIndex, token.atvREndIndex)
					}

					if token.atvRStartIndex > -1 {
						token.atvRStartIndex = -1
					}
					if token.atvREndIndex > -1 {
						token.atvREndIndex = -1
					}
					if token.psvRStartIndex > -1 {
						token.psvRStartIndex = -1
					}
					if token.psvREndIndex > -1 {
						token.psvREndIndex = -1
					}
					if token.hasAtv {
						token.hasAtv = false
					}
					lblsi[0] = 0
					lblsi[1] = 0
					return nextparse, err
				} else {
					return nextparse, err
				}
			} else {
				if token.curStartIndex > -1 {
					if token.curEndIndex == -1 {
						token.curEndIndex = token.lastEndRIndex - 1 - int64(len(lbls[1]))
					}
					if token.tknrb, err = parseCurrentPassiveTokenStartEnd(token, token.atvrs, token.lastEndRIndex, token.rerr != nil && token.rerr == io.EOF, token.curStartIndex, token.curEndIndex, token.startRIndex); err != nil {
						return nextparse, err
					}
				}
				if !token.hasAtv && strings.TrimSpace(string(token.tknrb)) != "" {
					token.hasAtv = true
				}

				if lblsi[1] > 0 {
					lblsi[1] = 0
				}

				if token.hasAtv {
					if token.atvRStartIndex == -1 {
						token.atvRStartIndex = token.startRIndex
					}

					if token.hasAtv && token.atvRStartIndex > -1 && strings.TrimSpace(string(token.tknrb)) != "" {
						token.atvREndIndex = token.startRIndex
					}
				}
				return nextparse, err
			}
		}
	} else if token.rerr == io.EOF {
		if token.curStartIndex > -1 && token.curEndIndex == -1 {
			token.curEndIndex = token.eofEndRIndex
		}
		if token.curStartIndex > -1 && token.curEndIndex > -1 {
			token.tknrb, err = parseCurrentPassiveTokenStartEnd(token, token.atvrs, token.curEndIndex, true, token.curStartIndex, token.curEndIndex, token.startRIndex)
		}
		nextparse = true
	}
	return nextparse, err
}

func parseCurrentPassiveTokenStartEnd(token *ActiveParseToken, curatvrs *ActiveReadSeeker, lastEndRIndex int64, eof bool, curStartIndex, curEndIndex, startRIndex int64) (lasttknrb []byte, err error) {
	token.curStartIndex = curStartIndex
	token.curEndIndex = curEndIndex

	if curStartIndex > -1 && curStartIndex <= curEndIndex {
		if token.psvCapturedIO != nil && !token.psvCapturedIO.Empty() {
			token.tokenMde = tokenPassive
			var tokenparsed bool
			var tokenerr error
			var prevtoken *ActiveParseToken
			for {
				if tokenparsed, tokenerr = token.parsing(); tokenparsed || tokenerr != nil {
					if tokenparsed && tokenerr == nil {
						if token.psvCapturedIO != nil && !token.psvCapturedIO.Empty() {
							token.psvCapturedIO.Close()
						}
						break
					}
					prevtoken, tokenerr = token.cleanupActiveParseToken()
					if tokenerr != nil {
						for prevtoken != nil {
							prevtoken, _ = prevtoken.cleanupActiveParseToken()
						}
					}
					token = prevtoken
					if token == nil {
						break
					}
				}
			}
		}
	}

	token.tokenMde = tokenActive
	token.psvRStartIndex = token.curStartIndex
	token.psvREndIndex = token.curEndIndex

	token.curStartIndex = -1
	token.curEndIndex = -1

	if token.psvRStartIndex > -1 && token.psvREndIndex > -1 {
		token.appendCntnt(token.atvrs, token.psvRStartIndex, token.psvREndIndex)
		//IOSeekReaderOutput(token.parse.cntntSR).Append(token.psvRStartIndex, token.psvREndIndex)
		token.psvRStartIndex = -1
		token.psvREndIndex = -1
	}

	return lasttknrb, err
}

//ParsePassiveToken - Default ParsePassiveToken method
func ParsePassiveToken(token *ActiveParseToken, lbls []string, lblsi []int) (parsed bool, err error) {
	if token.nr > 0 {
		fmt.Print(string(token.tknrb))
		if lblsi[1] == 0 && lblsi[0] < len(lbls[0]) {
			if lblsi[0] > 1 && lbls[0][lblsi[0]-1] == token.atvprevb && lbls[0][lblsi[0]] != token.tknrb[0] {
				lblsi[0] = 0
				token.psvprevb = 0
			}
			if lbls[0][lblsi[0]] == token.tknrb[0] {
				lblsi[0]++
				if len(lbls[0]) == lblsi[0] {
					token.psvprevb = 0
					return parsed, err
				} else {
					token.psvprevb = token.tknrb[0]
					return parsed, err
				}
			} else {
				if lblsi[0] > 0 {

					lblsi[0] = 0
				}
				token.psvprevb = token.tknrb[0]
				return parsed, err
			}
		} else if lblsi[0] == len(lbls[0]) && lblsi[1] < len(lbls[1]) {
			if lbls[1][lblsi[1]] == token.tknrb[0] {
				lblsi[1]++
				if lblsi[1] == len(lbls[1]) {
					if validPassiveParse(token) {

					} else {

					}
					lblsi[1] = 0
					lblsi[1] = 0
					return parsed, err
				} else {
					return parsed, err
				}
			} else {
				if lblsi[1] > 0 {

					lblsi[1] = 0
				}

				return parsed, err
			}
		}
	} else if token.rerr == io.EOF {
		if token.psvCapturedIO != nil && !token.psvCapturedIO.Empty() {
			token.psvCapturedIO.Close()
		}
	}
	return true, err
}

func validPassiveParse(token *ActiveParseToken) (valid bool) {

	return valid
}

//ActiveProcessor - ActiveProcessor
type ActiveProcessor struct {
	vm               *goja.Runtime
	atvParser        *ActiveParser
	w                io.Writer
	out              *IORW
	outprint         *widgeting.OutPrint
	canCleanupParams bool
	params           *Parameters
}

func (atvpros *ActiveProcessor) Parameters() *Parameters {
	return atvpros.params
}

func (atvpros *ActiveProcessor) Out() *widgeting.OutPrint {
	if atvpros.outprint == nil {
		atvpros.outprint = widgeting.NewOutPrint(func(a ...interface{}) {
			if atvpros.out != nil {
				atvpros.out.Print(a...)
			}
		}, func(a ...interface{}) {
			if atvpros.out != nil {
				atvpros.out.Println(a...)
			}
		})
	}
	return atvpros.outprint
}

func isActiveExtension(extfound string) bool {
	if strings.HasPrefix(extfound, ".") {
		extfound = extfound[1:]
	}
	return strings.Index(","+"html,htm,xml,svg,js,json,css"+",", ","+extfound+",") > -1
}

type vmeval struct {
	vm   *goja.Runtime
	done chan bool
	code string
	err  error
}

func (atvpro *ActiveProcessor) evalCode(cdefunc func() string, refelems ...map[string]interface{}) (err error) {
	if atvpro.vm == nil {
		atvpro.vm = goja.New()
	}
	if len(refelems) > 0 {
		for elemname, elem := range refelems[0] {
			atvpro.vm.Set(elemname, elem)
		}
	}
	s := cdefunc()
	//fmt.Println(s)

	vmeval := &vmeval{vm: atvpro.vm, code: s, done: make(chan bool, 1)}
	vmelalqueue <- vmeval

	if <-vmeval.done {
		close(vmeval.done)
		if vmeval.err != nil {
			fmt.Println(vmeval.err.Error())
		}
		vmeval = nil
	}
	/*if _, err = atvpro.vm.RunString(s); err != nil {
		fmt.Println(err.Error())
	}*/

	if len(refelems) > 0 {
		for elemname, _ := range refelems[0] {
			atvpro.vm.Set(elemname, nil)
		}
	}
	refelems = nil

	return err
}

func isJsExtension(extfound string) bool {
	if strings.HasPrefix(extfound, ".") {
		extfound = extfound[1:]
	}
	return strings.Index(","+",js,json,css"+",", ","+extfound+",") > -1
}

//NewActiveProcessor new ActiveProcessor
func NewActiveProcessor(w io.Writer) *ActiveProcessor {
	var atv *ActiveProcessor = &ActiveProcessor{w: w, canCleanupParams: true}
	atv.out, _ = NewIORW(atv.w)
	atv.atvParser = &ActiveParser{atv: atv}
	return atv
}

func (atvpros *ActiveProcessor) cleanupActiveProcessor() {
	if atvpros.atvParser != nil {
		atvpros.atvParser.cleanupActiveParser()
		atvpros.atvParser = nil
	}
	if atvpros.out != nil {
		atvpros.out.Close()
		atvpros.out = nil
	}

	if atvpros.outprint != nil {
		atvpros.outprint = nil
	}
	if atvpros.params != nil {
		if atvpros.canCleanupParams {
			atvpros.params.CleanupParameters()
		}
		atvpros.params = nil
	}
}

var activeModuledCommands map[string]map[string]ActiveCommandHandler

type activeCommandDefinition struct {
	atvCmdName  string
	atvCmdHndlr ActiveCommandHandler
}

func MapActiveCommand(path string, a ...interface{}) {
	if len(a) > 0 && len(a)%2 == 0 {
		ai := 0

		if atvcmddefs, atvcmddefsok := activeModuledCommands[path]; !atvcmddefsok {
			atvcmddefs = map[string]ActiveCommandHandler{}
			activeModuledCommands[path] = atvcmddefs
			for ai < len(a) {
				if atvcmdname, atvcmdnameok := a[ai].(string); atvcmdnameok {
					if atvcmdhndlr, atvcmdhndlrok := a[ai+1].(func(*ActiveProcessor, string, ...string) error); atvcmdhndlrok {
						atvcmddefs[atvcmdname] = atvcmdhndlr
					}
				}
				ai = ai + 2
			}
		} else {
			for ai < len(a) {
				if atvcmdname, atvcmdnameok := a[ai].(string); atvcmdnameok {
					if atvcmdhndlr, atvcmdhndlrok := a[ai+1].(func(*ActiveProcessor, string, ...string) error); atvcmdhndlrok {
						atvcmddefs[atvcmdname] = atvcmdhndlr
					}
				}
				ai = ai + 2
			}
		}

	}
}

var vmelalqueue chan *vmeval

func init() {
	if activeModuledCommands == nil {
		activeModuledCommands = map[string]map[string]ActiveCommandHandler{}
	}

	MapActiveCommand("index.html",
		"testcommand", func(atvpros *ActiveProcessor, path string, a ...string) (err error) {

			return err
		},
	)
	if vmelalqueue == nil {
		vmelalqueue = make(chan *vmeval)
		go func() {
			for {
				select {
				case vmeval := <-vmelalqueue:
					go func() {
						_, vmeval.err = vmeval.vm.RunString(vmeval.code)
						vmeval.done <- true
					}()
				}
			}
		}()
	}
}

type ActiveCommandHandler = func(*ActiveProcessor, string, ...string) error

func execCommand(atvpros *ActiveProcessor, path string, atvCmdHndlr ActiveCommandHandler, a ...string) (err error) {
	err = atvCmdHndlr(atvpros, path, a...)
	return err
}

func (atvpros *ActiveProcessor) Process(rs io.ReadSeeker, root string, path string, retrieveRS func(string, string) (io.ReadSeeker, error)) {
	if atvcmddefs, atvcmdefsok := activeModuledCommands[path]; atvcmdefsok && atvpros.params.ContainsParameter("COMMAND") {
		commands := atvpros.params.Parameter("COMMAND")
		cmdn := 0

		cmdhnlrs := []ActiveCommandHandler{}
		cmdhnlrparams := [][]string{}
		for cmdn < len(commands) {
			if atvcmddef, atvcmddefok := atvcmddefs[strings.ToLower(commands[cmdn])]; atvcmddefok {
				commands[cmdn] = strings.ToLower(commands[cmdn])
				cmdhnlrs = append(cmdhnlrs, atvcmddef)
				cmdhnlrparams = append(cmdhnlrparams, atvpros.params.Parameter(commands[cmdn]))
				cmdn++
			} else {
				if len(commands) > 1 {
					if cmdn == 0 {
						commands = commands[cmdn+1:]
					} else {
						commands = append(commands[:cmdn], commands[cmdn+1:]...)
					}
				}
			}
		}

		var executeCommand func(*ActiveProcessor, string, ActiveCommandHandler, ...string) error = execCommand

		for cmdtoexecn, _ := range commands {
			if err := executeCommand(atvpros, path, cmdhnlrs[cmdtoexecn], cmdhnlrparams[cmdtoexecn]...); err != nil {
				fmt.Println(err)
				break
			}
		}
	} else {
		atvpros.atvParser.parse(rs, root, path, retrieveRS, ParseActiveToken, ParsePassiveToken)
	}
}

func (atvpros *ActiveProcessor) Seek(offset int64, whence int) (seekedi int64, err error) {
	return seekedi, err
}

func (atvpros *ActiveProcessor) Read(p []byte) (n int, err error) {
	return n, err
}
