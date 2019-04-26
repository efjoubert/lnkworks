package lnksworks

import (
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dop251/goja"
	"github.com/efjoubert/lnkworks/widgeting"
)

//RetrieveRSFunc definition if function that retrieve external active resource e.g file
type RetrieveRSFunc = func(root string, path string) (rsfound io.ReadSeeker, rsfounderr error)

type activeParser struct {
	atv              *ActiveProcessor
	retrieveRS       RetrieveRSFunc
	atvParseFunc     func(*activeParseToken) bool
	psvParseFunc     func(*activeParseToken) bool
	atvrsmap         map[string]*activeReadSeeker
	atvrscntntcdemap map[int]*activeReadSeeker
	atvtrsstart      *activeReadSeeker
	atvTkns          map[*activeParseToken]*activeParseToken
}

func (atvparse *activeParser) appendActiveRSPoint(atvrs *activeReadSeeker, isStart bool) {
	if isStart {
		atvparse.atvtrsstart = atvrs
	}
}

func (atvparse *activeParser) atvrs(path string) (atvrs *activeReadSeeker) {
	if atvparse.atvrsmap != nil {
		atvrs, _ = atvparse.atvrsmap[path]
	}
	return atvrs
}

func (atvparse *activeParser) setRSByPath(path string) (err error) {
	if atvparse.atvrsmap == nil {
		if rstomap, rstomaperr := atvparse.retrieveRS("", path); rstomap != nil && rstomaperr == nil {
			atvparse.setRS(path, rstomap)
		} else if rstomaperr != nil {
			err = rstomaperr
		}
	} else {
		if _, rsok := atvparse.atvrsmap[path]; !rsok {
			if rstomap, rstomaperr := atvparse.retrieveRS("", path); rstomap != nil && rstomaperr == nil {
				atvparse.setRS(path, rstomap)
			} else if rstomaperr != nil {
				err = rstomaperr
			}
		}
	}
	return
}

func (atvparse *activeParser) setRS(path string, rstomap io.ReadSeeker) {
	if rstomap != nil {
		if atvparse.atvrsmap == nil {
			atvparse.atvrsmap = map[string]*activeReadSeeker{}
		}
		if atvparse.atvrscntntcdemap == nil {
			atvparse.atvrscntntcdemap = map[int]*activeReadSeeker{}
		}
		if _, atvrsok := atvparse.atvrsmap[path]; !atvrsok {
			atvrsio, _ := NewIORW(rstomap)
			atvrs := &activeReadSeeker{atvrsio: atvrsio, atvparse: atvparse}
			cntntrs := newContentSeekReader(atvrs)
			atvrs.cntntRS = cntntrs
			cders := newCodeSeekReader(atvrs, cntntrs)
			atvrs.cdeRS = cders
			atvparse.atvrsmap[path] = atvrs
			atvparse.atvrscntntcdemap[len(atvparse.atvrsmap)-1] = atvrs
		}
	}
}

type activeRSItem struct {
	cders   *contentSeekReader
	cntntrs *contentSeekReader
}

type contentSeekReader struct {
	atvrs        *activeReadSeeker
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
	if cntntsr.atvrs != nil {
		cntntsr.atvrs = nil
	}
}

type codeSeekReader struct {
	atvrs *activeReadSeeker
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
				for cntidsn := range cntids {
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
		for cntntseekristartKey := range cdesr.cntntseekristart {
			cdesr.cntntseekristart[cntntseekristartKey] = nil
			delete(cdesr.cntntseekristart, cntntseekristartKey)
		}
		cdesr.cntntseekristart = nil
	}
	if cdesr.cntntseekriendpos != nil {
		cdesr.cntntseekriendpos = nil
	}
}

func newContentSeekReader(atvrs *activeReadSeeker) *contentSeekReader {
	cntntsr := &contentSeekReader{atvrs: atvrs, IOSeekReader: NewIOSeekReader(atvrs), lastStartRsI: -1, lastEndRsI: -1}
	return cntntsr
}

func newCodeSeekReader(atvrs *activeReadSeeker, cntntsr *contentSeekReader) *codeSeekReader {
	cdesr := &codeSeekReader{atvrs: atvrs, IOSeekReader: NewIOSeekReader(atvrs), lastCntntStartI: -1, cntntsr: cntntsr}
	return cdesr
}

var emptyIO = &IORW{cached: false}

func (atvparse *activeParser) WriteContentByPos(rsi int, pos int) {
	if atvparse.atvrsmap != nil && atvparse.atvrscntntcdemap != nil {
		if atvrs, atvrsok := atvparse.atvrscntntcdemap[rsi]; atvrsok && atvrs.cntntRS != nil && len(atvrs.cntntRS.seekis) > 0 {
			atvrs.cntntRS.WriteSeekedPos(atvparse.atv.w, pos, 0)
		}
	}
}

func (atvparse *activeParser) readCurrentUnparsedTokenIO(token *activeParseToken, p []byte) (n int, err error) {
	for {
		if n, err = token.atvrs.Read(p); n == 0 && err == nil {
			err = io.EOF
			return
		}
		break
	}
	if err == io.EOF {
		if token.endRIndex > 0 {
			token.eofEndRIndex = token.endRIndex - 1
		} else {
			token.eofEndRIndex = token.endRIndex
		}
		token.startRIndex = 0
		token.endRIndex = 0
	} else {
		token.lastEndRIndex = token.endRIndex
		token.startRIndex = token.endRIndex
		token.endRIndex += int64(n)
	}
	return n, err
}

func (atvparse *activeParser) cleanupactiveParser() {
	if atvparse.atvrscntntcdemap != nil {
		for atvcntk := range atvparse.atvrscntntcdemap {
			atvparse.atvrscntntcdemap[atvcntk].cleanupactiveReadSeeker()
			atvparse.atvrscntntcdemap[atvcntk] = nil
			delete(atvparse.atvrscntntcdemap, atvcntk)
		}
		atvparse.atvrscntntcdemap = nil
	}
	if atvparse.atvrsmap != nil {
		for rspath := range atvparse.atvrsmap {
			atvparse.atvrsmap[rspath] = nil
			delete(atvparse.atvrsmap, rspath)
		}
		atvparse.atvrsmap = nil
	}
}

func (atvparse *activeParser) readActive(token *activeParseToken) (tknrn int, tknrnerr error) {
	tknrn, tknrnerr = atvparse.readCurrentUnparsedTokenIO(token, token.tknrb)
	return tknrn, tknrnerr
}

func (atvparse *activeParser) readPassive(token *activeParseToken) (tknrn int, tknrnerr error) {
	if token.psvCapturedIO != nil {
		tknrn, tknrnerr = token.psvCapturedIO.Read(token.tknrb)
	} else {
		tknrn, tknrnerr = emptyIO.Read(token.tknrb)
	}
	return tknrn, tknrnerr
}

type activeReadSeeker struct {
	atvpros   *ActiveProcessor
	atvparse  *activeParser
	atvrsio   *IORW
	cntntRS   *contentSeekReader
	cdeRS     *codeSeekReader
	atvrsi    int
	atvrspath string
}

func (atvrs *activeReadSeeker) writeAllContent(w io.Writer) {
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

func (atvrs *activeReadSeeker) cleanupactiveReadSeeker() {
	if atvrs.atvparse != nil {
		atvrs.atvparse = nil
	}
	if atvrs.atvpros != nil {
		atvrs.atvpros = nil
	}
	if atvrs.atvrsio != nil {
		atvrs.atvrsio.Close()
		atvrs.atvrsio = nil
	}
	if atvrs.cdeRS != nil {
		atvrs.cdeRS.clearCodeSeekReader()
		atvrs.cdeRS = nil
	}
	if atvrs.cntntRS != nil {
		atvrs.cntntRS.clearContentSeekReader()
		atvrs.cntntRS = nil
	}
}

func (atvrs *activeReadSeeker) Seek(offset int64, whence int) (n int64, err error) {
	return atvrs.atvrsio.Seek(offset, whence)
}

func (atvrs *activeReadSeeker) Read(p []byte) (n int, err error) {
	return atvrs.atvrsio.Read(p)
}

func (atvparse *activeParser) code() (s string) {

	return s
}

func (atvparse *activeParser) parse(rs io.ReadSeeker, root string, path string, retrieveRS func(string, string) (io.ReadSeeker, error)) (parseerr error) {
	if atvparse.retrieveRS == nil || &atvparse.retrieveRS != &retrieveRS {
		atvparse.retrieveRS = retrieveRS
	}
	atvparse.setRS(path, rs)

	parseerr = parseNextToken(nil, atvparse, path)

	if parseerr == nil && atvparse.atvtrsstart != nil {
		atvparse.evalStartActiveEntryPoint(atvparse.atvtrsstart)
	} else {
		fmt.Println(parseerr)
	}
	return
}

func parseNextToken(token *activeParseToken, atvparse *activeParser, rspath string) (parseErr error) {
	atvtoken := nextactiveParseToken(token, atvparse, rspath)

	var tokenparsed bool
	var tokenerr error
	var prevtoken *activeParseToken
	for {
		if tokenparsed, tokenerr = atvtoken.parsing(); tokenparsed || tokenerr != nil {
			if tokenparsed && tokenerr == nil {
				atvtoken.wrapupactiveParseToken()
			}
			prevtoken, tokenerr = atvtoken.cleanupactiveParseToken()
			if tokenerr != nil {
				parseErr = tokenerr
				for prevtoken != nil {
					prevtoken, _ = prevtoken.cleanupactiveParseToken()
				}
				break
			}
			atvtoken = prevtoken
			if atvtoken == token {
				break
			}
		}
	}
	return
}

func (atvparse *activeParser) evalStartActiveEntryPoint(atvrsstart *activeReadSeeker) {
	if !atvrsstart.cdeRS.Empty() {
		s := ""
		s += atvrsstart.cdeRS.String()
		atvparse.atv.evalCode(func() string {
			return s
		}, map[string]interface{}{"_out": atvparse.atv.Out(), "_atvparse": atvparse, "_parameters": atvparse.atv.params, "@db@execute": func(alias string, query string, args ...interface{}) *DBExecuted {
			return DatabaseManager().Execute(alias, query, args...)
		}, "@db@query": func(alias string, query string, args ...interface{}) *DBQuery {
			return DatabaseManager().Query(alias, query, args...)
		}})
	} else {
		atvrsstart.writeAllContent(atvparse.atv.w)
	}
}

func nextactiveParseToken(token *activeParseToken, parser *activeParser, rspath string) (nexttoken *activeParseToken) {
	rspathext := filepath.Ext(rspath)
	rspathname := rspath
	rsroot := rspath
	if strings.LastIndex(rspath, "/") > -1 {
		rsroot = rspath[0:strings.LastIndex(rspath, "/")]
	} else {
		rsroot = ""
	}
	if rspathext == "" {
		if token != nil {
			rspathext = token.rspathext
		}
		rspathname = strings.ReplaceAll(rspathname, "/", ":")
	}

	if parser.atvTkns == nil {
		parser.atvTkns = map[*activeParseToken]*activeParseToken{}
	}

	nexttoken = &activeParseToken{startRIndex: 0, endRIndex: 0, parse: parser, tknrb: make([]byte, 1), curStartIndex: -1, curEndIndex: -1, rsroot: rsroot, rspath: rspath, atvRStartIndex: -1, atvREndIndex: -1, tokenMde: tokenActive}
	nexttoken.atvlbls = []string{"<@", "@>"}
	nexttoken.psvlbls = []string{nexttoken.atvlbls[0][0 : len(nexttoken.atvlbls)-1], nexttoken.atvlbls[1][1:]}
	nexttoken.atvlblsi = []int{0, 0}
	nexttoken.parkedStartIndex = -1
	nexttoken.parkedEndIndex = -1
	nexttoken.parkedLevel = 0
	parser.atvTkns[nexttoken] = token
	//nexttoken.psvlblsi = []int{0, 0}
	nexttoken.atvrs = parser.atvrs(rspath)
	return nexttoken
}

type tokenMode int

const (
	tokenActive  tokenMode = 0
	tokenPassive tokenMode = 1
)

type activeParseToken struct {
	parse            *activeParser
	atvlbls          []string
	atvlblsi         []int
	atvprevb         byte
	hasAtv           bool
	psvlbls          []string
	psvCapturedIO    *IORW
	psvUnvalidatedIO *IORW
	tknrb            []byte
	nr               int
	rerr             error
	atvrs            *activeReadSeeker //   io.ReadSeeker
	curStartIndex    int64
	curEndIndex      int64
	rspath           string
	rsroot           string
	rspathname       string
	rspathext        string
	// ELEM VALID SETTINGS
	parkedStartIndex int64
	parkedEndIndex   int64
	parkedLevel      int
	elemName         string
	//
	startRIndex    int64
	lastEndRIndex  int64
	endRIndex      int64
	eofEndRIndex   int64
	atvRStartIndex int64
	atvREndIndex   int64
	tokenMde       tokenMode
}

func (token *activeParseToken) prevToken() *activeParseToken {
	if token.parse != nil && token.parse.atvTkns != nil {
		return token.parse.atvTkns[token]
	}
	return nil
}

func (token *activeParseToken) appendCde(atvrs *activeReadSeeker, atvRStartIndex int64, atvREndIndex int64) {
	atvrs.cdeRS.Append(atvRStartIndex, atvREndIndex)
}

func (token *activeParseToken) appendCntnt(atvrs *activeReadSeeker, psvRStartIndex int64, psvREndIndex int64) {
	atvrs.cntntRS.Append(psvRStartIndex, psvREndIndex)
}

func (token *activeParseToken) wrapupactiveParseToken() {
	if token.atvrs != nil {
		if !token.atvrs.cdeRS.Empty() {
			if !token.atvrs.cdeRS.Empty() {
				if token.atvrs.cdeRS.cntntseekriendpos == nil || len(token.atvrs.cdeRS.cntntseekriendpos) == 0 {
					if token.atvrs.cntntRS.lastStartRsI > -1 && token.atvrs.cntntRS.lastStartRsI <= token.atvrs.cntntRS.lastEndRsI {
						cntntrsseeki := make([]int, (token.atvrs.cntntRS.lastEndRsI-token.atvrs.cntntRS.lastStartRsI)+1)
						for cntntrsseekin := range cntntrsseeki {
							cntntrsseeki[cntntrsseekin] = token.atvrs.cntntRS.lastStartRsI
							token.atvrs.cntntRS.lastStartRsI++
						}
						token.atvrs.cntntRS.lastStartRsI = -1
						token.atvrs.cntntRS.lastEndRsI = -1
						token.atvrs.cdeRS.cntntseekriendpos = append(token.atvrs.cdeRS.cntntseekriendpos, cntntrsseeki[:]...)
						cntntrsseeki = nil
					}
				}
				token.parse.appendActiveRSPoint(token.atvrs, token.parse.atvTkns == nil || token.parse.atvTkns[token] == nil)
			}
		} else if !token.atvrs.cntntRS.Empty() {
			token.parse.appendActiveRSPoint(token.atvrs, token.parse.atvTkns == nil || token.parse.atvTkns[token] == nil)
		}
	}
}

func (token *activeParseToken) cleanupactiveParseToken() (prevtoken *activeParseToken, err error) {
	if token.atvlbls != nil {
		token.atvlbls = nil
	}
	if token.atvlblsi != nil {
		token.atvlblsi = nil
	}
	if token.parse != nil {
		if token.parse.atvTkns != nil {
			if _, tokenOk := token.parse.atvTkns[token]; tokenOk {
				prevtoken = token.parse.atvTkns[token]
				delete(token.parse.atvTkns, token)
			}
		}
		token.parse = nil
	}

	if token.psvlbls != nil {
		token.psvlbls = nil
	}
	if token.rerr != nil {
		token.rerr = nil
	}
	if token.psvCapturedIO != nil {
		token.psvCapturedIO.Close()
		token.psvCapturedIO = nil
	}
	if token.psvUnvalidatedIO != nil {
		token.psvUnvalidatedIO.Close()
		token.psvUnvalidatedIO = nil
	}
	if token.tknrb != nil {
		token.tknrb = nil
	}

	return prevtoken, err
}

func (token *activeParseToken) passiveCapturedIO() *IORW {
	if token.psvCapturedIO == nil {
		token.psvCapturedIO, _ = NewIORW()
	}
	return token.psvCapturedIO
}

func (token *activeParseToken) passiveUnvalidatedIO() *IORW {
	if token.psvUnvalidatedIO == nil {
		token.psvUnvalidatedIO, _ = NewIORW()
	}
	return token.psvUnvalidatedIO
}

func (token *activeParseToken) parsing() (parsed bool, err error) {
	if token.tokenMde == tokenActive {
		return token.parsingActive()
	}
	err = fmt.Errorf("INVALID PARSING POINT READ")
	return parsed, err
}

func (token *activeParseToken) parsingActive() (parsed bool, err error) {
	token.nr, token.rerr = token.parse.readActive(token)
	return parseActiveToken(token, token.atvlbls, token.atvlblsi)
}

func parseActiveToken(token *activeParseToken, lbls []string, lblsi []int) (nextparse bool, err error) {
	if token.nr > 0 {
		if lblsi[1] == 0 && lblsi[0] < len(lbls[0]) {
			if lblsi[0] > 1 && lbls[0][lblsi[0]-1] == token.atvprevb && lbls[0][lblsi[0]] != token.tknrb[0] {
				if token.psvCapturedIO != nil && !token.psvCapturedIO.Empty() {
					token.psvCapturedIO.Close()
				}
				token.passiveCapturedIO().Print(lbls[0][:lblsi[0]])
				if token.curStartIndex == -1 {
					token.curStartIndex = token.startRIndex - int64(lblsi[0])
				}
				lblsi[0] = 0
				token.atvprevb = 0
			}
			if lbls[0][lblsi[0]] == token.tknrb[0] {
				lblsi[0]++
				if len(lbls[0]) == lblsi[0] {
					token.atvprevb = 0
					return nextparse, err
				}
				token.atvprevb = token.tknrb[0]
				return nextparse, err
			}
			if token.curStartIndex == -1 {
				if lblsi[0] > 0 {
					token.curStartIndex = token.startRIndex - int64(lblsi[0])
				} else {
					token.curStartIndex = token.startRIndex
				}
			}
			if lblsi[0] > 0 {
				if token.psvCapturedIO != nil && !token.psvCapturedIO.Empty() {
					token.psvCapturedIO.Close()
				}
				token.passiveCapturedIO().Print(lbls[0][:lblsi[0]])
				lblsi[0] = 0
			}
			token.atvprevb = token.tknrb[0]
			token.passiveCapturedIO().Print(token.tknrb)

			if token.psvCapturedIO != nil && !token.psvCapturedIO.Empty() && token.psvCapturedIO.HasSuffix([]byte(token.psvlbls[1])) {
				if token.psvCapturedIO.HasPrefixSuffix([]byte(token.psvlbls[0]), []byte(token.psvlbls[1])) {
					if valid, single, complexStart, complexEnd, elemName, elemPath, elemExt, valErr := validatePassiveCapturedIO(token); valid {
						if single || complexStart {
							if token.curStartIndex > -1 {
								if token.curEndIndex == -1 {
									token.curEndIndex = token.lastEndRIndex - token.psvCapturedIO.Size()
								}
								if token.tknrb, err = parseCurrentPassiveTokenStartEnd(token, token.atvrs, token.lastEndRIndex, token.rerr != nil && token.rerr == io.EOF, token.curStartIndex, token.curEndIndex, token.startRIndex); err != nil {
									return nextparse, err
								}
							}
							if single || complexEnd {
								if elemName != "" && elemPath != "" && elemExt != "" {
									if !strings.HasPrefix(elemPath, "./") {
										if token.rsroot != "" {
											elemPath = token.rsroot + elemPath
										}
									}
									if err = token.parse.setRSByPath(elemPath); err == nil {
										parseNextToken(token, token.parse, elemPath)
									}
								}
							}
						}
					} else if valErr != nil {
						err = valErr
						return nextparse, err
					}
				}
				token.psvCapturedIO.Close()
			}

			return nextparse, err

		} else if lblsi[0] == len(lbls[0]) && lblsi[1] < len(lbls[1]) {
			if lbls[1][lblsi[1]] == token.tknrb[0] {
				lblsi[1]++
				if lblsi[1] == len(lbls[1]) {
					if token.atvRStartIndex > -1 && token.atvREndIndex > -1 {
						token.appendCde(token.atvrs, token.atvRStartIndex, token.atvREndIndex)
					}

					if token.atvRStartIndex > -1 {
						token.atvRStartIndex = -1
					}
					if token.atvREndIndex > -1 {
						token.atvREndIndex = -1
					}
					if token.hasAtv {
						token.hasAtv = false
					}
					lblsi[0] = 0
					lblsi[1] = 0
					return nextparse, err
				}
				return nextparse, err
			}
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

func validatePassiveCapturedIO(token *activeParseToken) (valid bool, single bool, comlexStart bool, complexEnd bool, elemName string, elemPath string, elemExt string, err error) {
	if actualSize := (token.psvCapturedIO.Size() - int64(len(token.psvlbls[0])+len(token.psvlbls[1]))); actualSize >= 1 {
		if valid = (actualSize == 1 && token.psvCapturedIO.String() == "/"); !valid {
			token.psvCapturedIO.Seek(int64(len(token.psvlbls[0])), 0)

			actualSizei := int64(0)

			foundFSlash := false

			elemName = ""

			for actualSizei < actualSize {
				if r, _, _ := token.psvCapturedIO.ReadRune(); r > 0 {
					actualSizei += int64(len(string(r)))
					if strings.TrimSpace(string(r)) != "" {
						if r == 47 {
							if !foundFSlash {
								foundFSlash = true
								if token.psvUnvalidatedIO != nil && !token.psvUnvalidatedIO.Empty() {
									single = true
								} else {
									complexEnd = true
								}
							} else {
								err = fmt.Errorf("Invalid element - " + token.psvCapturedIO.String())
							}
						} else {
							token.passiveUnvalidatedIO().WriteRune(r)
							if actualSizei < actualSize {
								continue
							}
						}
					}
					if strings.TrimSpace(string(r)) == "" || actualSizei == actualSize {
						if token.psvUnvalidatedIO.MatchExp(regexptagstart) {
							comlexStart = !(single || complexEnd)
							elemName = token.psvUnvalidatedIO.String()
							if elemExt = filepath.Ext(elemName); elemExt != "" {
								elemName = elemName[0 : len(elemName)-len(elemExt)]
							} else if elemExt == "" {
								elemExt = token.rspathext
							}
							elemPath = strings.ReplaceAll(elemName, ":", "/") + elemExt
							valid = true
						} else {
							break
						}
					}
				} else {
					break
				}
			}

			if token.psvUnvalidatedIO != nil && !token.psvUnvalidatedIO.Empty() {
				token.psvUnvalidatedIO.Close()
			}
		}
	}
	if valid {
		if err != nil {
			valid = false
		} else {
			if single || complexEnd {
				if complexEnd {
					if token.parkedLevel > 0 {
						if token.elemName == elemName {
							token.parkedLevel--
						}

						if valid = token.parkedLevel == 0; valid {
							token.elemName = ""
						}
					} else {
						valid = false
					}
				} else if single {
					valid = token.parkedLevel == 0
				}
			} else if comlexStart {
				if token.elemName == "" {
					if token.parkedLevel == 0 {
						token.elemName = elemName
					}
				}
				if token.elemName == elemName {
					valid = token.parkedLevel == 0
					token.parkedLevel++
				} else {
					valid = false
				}
			}
		}
	}
	return
}

func parseCurrentPassiveTokenStartEnd(token *activeParseToken, curatvrs *activeReadSeeker, lastEndRIndex int64, eof bool, curStartIndex, curEndIndex, startRIndex int64) (lasttknrb []byte, err error) {
	token.curStartIndex = -1
	token.curEndIndex = -1
	lasttknrb = token.tknrb[:]
	if curStartIndex > -1 && curEndIndex > -1 && curStartIndex <= curEndIndex {
		token.appendCntnt(token.atvrs, curStartIndex, curEndIndex)
	}

	return lasttknrb, err
}

//ActiveProcessor - ActiveProcessor
type ActiveProcessor struct {
	vm               *goja.Runtime
	atvParser        *activeParser
	w                io.Writer
	out              *IORW
	outprint         *widgeting.OutPrint
	canCleanupParams bool
	params           *Parameters
}

//Parameters current active process Parameters container
func (atvpros *ActiveProcessor) Parameters() *Parameters {
	return atvpros.params
}

//Out current acive process out put handle
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

func (atvpros *ActiveProcessor) evalCode(cdefunc func() string, refelems ...map[string]interface{}) (err error) {
	if atvpros.vm == nil {
		atvpros.vm = goja.New()
	}
	if len(refelems) > 0 {
		for elemname, elem := range refelems[0] {
			atvpros.vm.Set(elemname, elem)
		}
	}
	s := cdefunc()

	vmeval := &vmeval{vm: atvpros.vm, code: s, done: make(chan bool, 1)}
	vmelalqueue <- vmeval

	if <-vmeval.done {
		close(vmeval.done)
		if vmeval.err != nil {
			fmt.Println(vmeval.err.Error())
		}
		vmeval = nil
	}

	if len(refelems) > 0 {
		for elemname := range refelems[0] {
			atvpros.vm.Set(elemname, nil)
		}
	}
	refelems = nil

	return err
}

//NewActiveProcessor new ActiveProcessor
func NewActiveProcessor(w io.Writer) *ActiveProcessor {
	var atv = &ActiveProcessor{w: w, canCleanupParams: true}
	atv.out, _ = NewIORW(atv.w)
	atv.atvParser = &activeParser{atv: atv, atvTkns: map[*activeParseToken]*activeParseToken{}}
	return atv
}

func (atvpros *ActiveProcessor) cleanupActiveProcessor() {
	if atvpros.atvParser != nil {
		atvpros.atvParser.cleanupactiveParser()
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

//MapActiveCommand map a resource path to a list if ActiveCommandHandler(s) by command name
//e.g.
// lnksworks.MapActiveCommand("test/test.html",
//		"testcommand", func(atvpros *lnksworks.ActiveProcessor, path string, a ...string) (err error) {
//		atvpros.Out().Elem("span", func(out *widgeting.OutPrint, a ...interface{}) {
//			out.Print("content in span")
//		})
//		return err
//		},
//	)
//
func MapActiveCommand(path string, a ...interface{}) {
	if len(a) > 0 && len(a)%2 == 0 {
		ai := 0

		if atvcmddefs, atvcmddefsok := activeModuledCommands[path]; !atvcmddefsok {
			atvcmddefs = map[string]ActiveCommandHandler{}
			activeModuledCommands[path] = atvcmddefs
			for ai < len(a) {
				if atvcmdname, atvcmdnameok := a[ai].(string); atvcmdnameok {
					if atvcmdhndlr, atvcmdhndlrok := a[ai+1].(ActiveCommandHandler); atvcmdhndlrok {
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

const tagstartregexp string = `^((.:(([a-z]|[A-Z])\w*)+)|(([a-z]|[A-Z])+(:(([a-z]|[A-Z])\w*)+)+))+(:(([a-z]|[A-Z])\w*)+)*(-(([a-z]|[A-Z])\w*)+)?(.([a-z]|[A-Z])+)?$`

var regexptagstart *regexp.Regexp

const propregexp string = `^-?-?(([a-z]+[0-9]*)[a-z]*)+(-([a-z]+[0-9]*)[a-z]*)?$`

var regexprop *regexp.Regexp

const propvalnumberexp string = `^[-+]?\d+([.]\d+)?$`

var regexpropvalnumberexp *regexp.Regexp

func init() {

	if regexptagstart == nil {
		regexptagstart = regexp.MustCompile(tagstartregexp)
	}
	if regexprop == nil {
		regexprop = regexp.MustCompile(propregexp)
	}

	if regexpropvalnumberexp == nil {
		regexpropvalnumberexp = regexp.MustCompile(propvalnumberexp)
	}

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

//ActiveCommandHandler definition of func that impliments the function that needs to be applied
// by the command mapped in MapActiveCommand
type ActiveCommandHandler = func(atvpros *ActiveProcessor, path string, a ...string) error

func execCommand(atvpros *ActiveProcessor, path string, atvCmdHndlr ActiveCommandHandler, a ...string) (err error) {
	err = atvCmdHndlr(atvpros, path, a...)
	return err
}

//Process main method that apply the active process
//rs - io.ReadSeeker of active content
//root - root path of rs
//path - path that active content can be found
//retrieveRS - func reference to a implementation base on RetrieveRSFunc definition
func (atvpros *ActiveProcessor) Process(rs io.ReadSeeker, root string, path string, retrieveRS RetrieveRSFunc) {
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

		var executeCommand = execCommand

		for cmdtoexecn := range commands {
			if err := executeCommand(atvpros, path, cmdhnlrs[cmdtoexecn], cmdhnlrparams[cmdtoexecn]...); err != nil {
				fmt.Println(err)
				break
			}
		}
	} else {
		atvpros.atvParser.parse(rs, root, path, retrieveRS)
	}
}
