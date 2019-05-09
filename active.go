package lnksworks

import (
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/dop251/goja"
	"github.com/efjoubert/lnkworks/widgeting"
)

//Exception interface
type Exception interface{}

type tcfblock struct {
	Try     func()
	Catch   func(Exception)
	Finally func()
}

func (tcf tcfblock) Do() {
	if tcf.Finally != nil {

		defer tcf.Finally()
	}
	if tcf.Catch != nil {
		defer func() {
			if r := recover(); r != nil {
				tcf.Catch(r)
			}
		}()
	}
	tcf.Try()
}

//RetrieveRSFunc definition if function that retrieve external active resource e.g file
type RetrieveRSFunc = func(root string, path string) (rsfound io.ReadSeeker, rsfounderr error)

type activeParser struct {
	atv              *ActiveProcessor
	retrieveRS       RetrieveRSFunc
	atvParseFunc     func(*activeParseToken) bool
	psvParseFunc     func(*activeParseToken) bool
	atvrsmap         map[string]*activeReadSeeker
	atvrscntntcdemap map[int]*activeReadSeeker
	atvRsAPCStart    *activeRSActivePassiveContent
	atvTkns          map[*activeParseToken]*activeParseToken
	atvRsAPCMap      map[*activeRSActivePassiveContent]*activeRSActivePassiveContent
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
			atvparse.atvrsmap[path] = atvrs
			atvparse.atvrscntntcdemap[len(atvparse.atvrsmap)-1] = atvrs
		}
	}
}

type parkedRsAPCPoint struct {
	atvRsAPC *activeRSActivePassiveContent
	atvrs    *activeReadSeeker
	*Seeker
	prkdcders       map[int]*codeSeekReader
	prkdcntntrs     map[int]*contentSeekReader
	prkdAtvrsATCmap map[int]bool
	lastAppendi     int
	prkdStartI      int64
	prkdEndI        int64
	prkdCurI        int64
	enabled         bool
	startRIndex     int64
	lastEndRIndex   int64
	endRIndex       int64
	eofEndRIndex    int64
}

func (prkdPoint *parkedRsAPCPoint) empty() bool {
	return len(prkdPoint.seekis) == 0 && len(prkdPoint.prkdAtvrsATCmap) == 0 && (prkdPoint.prkdcders == nil || len(prkdPoint.prkdcders) == 0 && len(prkdPoint.prkdcntntrs) == 0)
}

func (prkdPoint *parkedRsAPCPoint) enable() (err error) {
	if !prkdPoint.enabled {
		prkdPoint.enabled = true
		prkdPoint.prkdStartI, err = prkdPoint.atvrs.Seek(prkdPoint.prkdStartI, 0)
		prkdPoint.prkdCurI = prkdPoint.prkdStartI
	}
	return
}

func (prkdPoint *parkedRsAPCPoint) Read(p []byte) (n int, err error) {
	if prkdPoint.prkdCurI < (prkdPoint.prkdEndI + 1) {
		if pl := len(p); ((prkdPoint.prkdEndI + 1) - prkdPoint.prkdCurI) >= int64(pl) {
			n, err = prkdPoint.atvrs.Read(p)
		} else if prdkl := (prkdPoint.prkdEndI + 1) - prkdPoint.prkdCurI; prdkl > 0 {
			n, err = prkdPoint.atvrs.Read(p[:int(prdkl)])
		} else {
			err = io.EOF
		}
		if err == nil {
			prkdPoint.prkdCurI += int64(n)
		}
	} else {
		err = io.EOF
	}
	return
}

func (prkdPoint *parkedRsAPCPoint) cleanupParksRsAPCPoint() {
	if prkdPoint.atvrs != nil {
		prkdPoint.atvrs = nil
	}

	if prkdPoint.prkdcders != nil {
		for _, cders := range prkdPoint.prkdcders {
			cders.clearCodeSeekReader()
		}
		prkdPoint.prkdcders = nil
	}

	if prkdPoint.prkdcntntrs != nil {
		for _, cntntrs := range prkdPoint.prkdcntntrs {
			cntntrs.clearContentSeekReader()
		}
		prkdPoint.prkdcntntrs = nil
	}

	if prkdPoint.Seeker != nil {

		prkdPoint.Seeker.ClearSeeker()
		prkdPoint.Seeker = nil
	}

	if prkdPoint.prkdAtvrsATCmap != nil {
		for n := range prkdPoint.prkdAtvrsATCmap {
			delete(prkdPoint.prkdAtvrsATCmap, n)
		}
		prkdPoint.prkdAtvrsATCmap = nil
	}
}

func newParkedRsAPCPoint(atvrs *activeReadSeeker, prkdStartI int64, prkdEndI int64) (prkdPoint *parkedRsAPCPoint) {
	prkdPoint = &parkedRsAPCPoint{
		Seeker:          &Seeker{},
		atvrs:           atvrs,
		lastAppendi:     -1,
		prkdcders:       map[int]*codeSeekReader{},
		prkdcntntrs:     map[int]*contentSeekReader{},
		prkdStartI:      prkdStartI,
		prkdEndI:        prkdEndI,
		prkdCurI:        prkdStartI,
		prkdAtvrsATCmap: map[int]bool{},
		enabled:         false}
	return prkdPoint
}

func (prkdPoint *parkedRsAPCPoint) Append(starti int64, endi int64) {
	if !prkdPoint.enabled {
		prkdPoint.lastAppendi = len(prkdPoint.seekis)
		prkdPoint.Seeker.Append(starti, endi)
		if prkdPoint.prkdCurI != prkdPoint.prkdStartI {
			prkdPoint.prkdCurI = prkdPoint.prkdStartI
		}
		if prkdPoint.startRIndex != prkdPoint.prkdStartI {
			prkdPoint.endRIndex = prkdPoint.prkdStartI
			prkdPoint.startRIndex = prkdPoint.prkdStartI
			prkdPoint.endRIndex = prkdPoint.prkdStartI
		}
	}
}

func (prkdPoint *parkedRsAPCPoint) AppendCntnt(starti int64, endi int64) {
	if cntntrs, cntntrsok := prkdPoint.prkdcntntrs[prkdPoint.lastAppendi]; cntntrsok {
		cntntrs.Append(starti, endi)
	} else {
		prkdPoint.prkdcntntrs[prkdPoint.lastAppendi] = newContentSeekReader(prkdPoint.atvRsAPC, prkdPoint.atvrs)
		prkdPoint.prkdcntntrs[prkdPoint.lastAppendi].Append(starti, endi)
	}
}

func (prkdPoint *parkedRsAPCPoint) AppendCde(starti int64, endi int64) {
	if cders, cdersok := prkdPoint.prkdcders[prkdPoint.lastAppendi]; cdersok {
		cders.Append(starti, endi)
	} else {
		prkdPoint.prkdcders[prkdPoint.lastAppendi] = newCodeSeekReader(prkdPoint.atvrs, prkdPoint.atvRsAPC)
		prkdPoint.prkdcders[prkdPoint.lastAppendi].Append(starti, endi)
	}
}

type activeRSActivePassiveContent struct {
	*Seeker
	cders   *codeSeekReader
	cntntrs *contentSeekReader

	prkdPoint     *parkedRsAPCPoint
	atvRSAPCMap   map[int]*activeRSActivePassiveContent
	atvparse      *activeParser
	atvRsACPi     int
	lastStartRsI  int
	lastEndRsI    int
	cdeIO         *IORW
	atvRsAPCCount int
	tokenPath     string
}

func (atvRsAPC *activeRSActivePassiveContent) code() *IORW {
	if atvRsAPC.cdeIO == nil {
		atvRsAPC.cdeIO, _ = NewIORW()
	}
	return atvRsAPC.cdeIO
}

func (atvRsAPC *activeRSActivePassiveContent) topRSAPC() *activeRSActivePassiveContent {
	return atvRsAPC.atvparse.atvRsAPCStart
}

func (atvRsAPC *activeRSActivePassiveContent) Empty() bool {
	return atvRsAPC.Seeker == nil || len(atvRsAPC.Seeker.seekis) == 0
}

func newActiveRSActivePassiveContent(atvrs *activeReadSeeker, atvparse *activeParser) (atvRsAPC *activeRSActivePassiveContent) {
	atvRsAPC = &activeRSActivePassiveContent{Seeker: &Seeker{}, atvRsAPCCount: 0, atvparse: atvparse, atvRsACPi: -1}
	atvRsAPC.cntntrs = newContentSeekReader(atvRsAPC, atvrs)
	atvRsAPC.cders = newCodeSeekReader(atvrs, atvRsAPC)
	return
}

func (atvRsAPC *activeRSActivePassiveContent) appendLevels(atvparser *activeParser, ignoreLvls int) (a []int) {
	var atvRsApcRef = atvRsAPC
	if atvRsApcRef != atvparser.atvRsAPCStart {
		var atvRsApcRef = atvRsAPC
		for atvRsApcRef != nil {
			a = append(a, atvRsApcRef.atvRsACPi)
			if atvRsApcRef = atvparser.atvRsAPCMap[atvRsApcRef]; atvRsApcRef == nil || atvRsApcRef.atvRsACPi == -1 {
				break
			}
		}

		if len(a) > 1 {
			var aref = make([]int, len(a))
			for n := range a {
				aref[n] = a[len(a)-(n+1)]
			}
			a = nil
			a = aref[:len(aref)-ignoreLvls]
			aref = nil
		}
		atvRsApcRef = nil
	}
	return
}

func atvRSAPCCoding(hasCode bool, atvRsAPC *activeRSActivePassiveContent, atvparser *activeParser, w io.Writer) (err error) {
	var atvrAPCiCode = ""
	var cntntrs = atvRsAPC.cntntrs
	var cntntpos = 0
	var cntntL = len(cntntrs.seekis)
	var cntntpoint []int64
	if cntntpos < cntntL {
		cntntpoint = cntntrs.seekis[cntntpos][:]
	}

	if cntntpos < cntntL {
		if atvRsAPC.atvRsACPi != -1 {
			atvrAPCiCode = fmt.Sprint(atvRsAPC.appendLevels(atvparser, 0))
			atvrAPCiCode = strings.ReplaceAll(strings.Replace(strings.Replace(atvrAPCiCode, "[", "", 1), "]", "", 1), " ", ",")
		}
		if atvrAPCiCode != "" {
			atvrAPCiCode = "," + atvrAPCiCode
		}
	}

	var prkdPoint = atvRsAPC.prkdPoint
	var prkdPointPos = 0
	var prkdPointL = 0
	var prkdpnt []int64
	var atvrPrkdAPCiCode = ""
	if prkdPoint != nil {
		if atvRsAPC.atvRsACPi != -1 {
			atvrPrkdAPCiCode = fmt.Sprint(prkdPoint.atvRsAPC.appendLevels(atvparser, 0))
			atvrPrkdAPCiCode = strings.ReplaceAll(strings.Replace(strings.Replace(atvrPrkdAPCiCode, "[", "", 1), "]", "", 1), " ", ",")
		}
		if atvrPrkdAPCiCode != "" {
			atvrPrkdAPCiCode = "," + atvrPrkdAPCiCode
		}
		prkdPointL = len(prkdPoint.seekis)
		if prkdPointPos < prkdPointL {
			prkdpnt = prkdPoint.seekis[prkdPointPos][:]
		}
	}
	var cders = atvRsAPC.cders
	var cdepos = 0
	var cdeL = len(cders.seekis)
	var cdepoint []int64
	if cdepos < cdeL {
		cdepoint = cders.seekis[cdepos][:]
	}

	var atvapcpos = 0
	var atvapcL = len(atvRsAPC.seekis)
	var atvapcpoint []int64
	if atvapcpos < atvapcL {
		atvapcpoint = atvRsAPC.seekis[atvapcpos][:]
	}

	var isPrkdAtv = func() bool {
		return atvapcpos < atvapcL && prkdPoint != nil && prkdPoint.prkdAtvrsATCmap[atvapcpos]
	}

	var captureCode = func(prkd ...bool) {
		if ss, sserr := cders.StringSeekPos(cdepos, 0); sserr == nil {
			_, err = atvRsAPC.topRSAPC().code().Print(ss)
			cdepos++
			if cdepos < cdeL {
				cdepoint = cders.seekis[cdepos][:]
			}
			if !hasCode {
				hasCode = true
			}
		}
	}

	var captureContent = func(prkd ...bool) {
		if hasCode {
			if _, err = atvRsAPC.topRSAPC().code().Print("_atvparse.WriteContentByPos(" + fmt.Sprintf("%d", cntntpos) + atvrAPCiCode + ");"); err == nil {
				cntntpos++
				if cntntpos < cntntL {
					cntntpoint = cntntrs.seekis[cntntpos]
				}
			}
		} else if err = cntntrs.WriteSeekedPos(w, cntntpos, 0); err == nil {
			cntntpos++
			if cntntpos < cntntL {
				cntntpoint = cntntrs.seekis[cntntpos]
			}
		}
	}

	var captureAtvRSAPContent = func() {
		if err = atvRSAPCCoding(hasCode, atvRsAPC.atvRSAPCMap[atvapcpos], atvparser, w); err == nil {
			if !hasCode && atvRsAPC.topRSAPC().cdeIO != nil && !atvRsAPC.topRSAPC().cdeIO.Empty() {
				hasCode = true
			}
			atvapcpos++
			if atvapcpos < atvapcL {
				atvapcpoint = atvRsAPC.seekis[atvapcpos][:]
			}
		}
	}

	var captureParkedPoint = func(prkdpi int) {

		var prkdcders = prkdPoint.prkdcders[prkdpi]
		var prkdcdepos = 0
		var prkdcdeL = 0
		var prkdcdepoint []int64
		if prkdcders != nil {
			if prkdcdeL = len(prkdcders.seekis); prkdcdepos < prkdcdeL {
				prkdcdepoint = prkdcders.seekis[prkdcdepos][:]
			}
		}

		var prkdcntntrs = prkdPoint.prkdcntntrs[prkdpi]
		var prkdcntntpos = 0
		var prkdcntntL = 0
		var prkdcntntpoint []int64
		if prkdcntntrs != nil {
			if prkdcntntL = len(prkdcntntrs.seekis); prkdcntntpos < prkdcntntL {
				prkdcntntpoint = prkdcntntrs.seekis[prkdcntntpos][:]
			}
		}

		for err == nil && (prkdcdepos < prkdcdeL || prkdcntntpos < prkdcntntL || isPrkdAtv()) {
			if prkdcdepos < prkdcdeL && (prkdcntntpos == prkdcntntL || prkdcdepoint[1] < prkdcntntpoint[0]) && (atvapcpos == atvapcL || isPrkdAtv() && prkdcdepoint[1] < atvapcpoint[0] || !isPrkdAtv()) {
				if ss, sserr := prkdcders.StringSeekPos(prkdcdepos, 0); sserr == nil {
					_, err = atvRsAPC.topRSAPC().code().Print(ss)
					prkdcdepos++
					if prkdcdepos < prkdcdeL {
						prkdcdepoint = prkdcders.seekis[prkdcdepos][:]
					}
					if !hasCode {
						hasCode = true
					}
				}
			} else if prkdcntntpos < prkdcntntL && (prkdcdepos == prkdcdeL || prkdcntntpoint[1] < prkdcdepoint[0]) && (atvapcpos == atvapcL || isPrkdAtv() && prkdcntntpoint[1] < atvapcpoint[0] || !isPrkdAtv()) {
				if hasCode {
					if _, err = atvRsAPC.topRSAPC().code().Print("_atvparse.WriteParkedContentByPos(" + fmt.Sprintf("%d,%d", prkdPointPos, prkdcntntpos) + atvrAPCiCode + ");"); err == nil {
						prkdcntntpos++
						if prkdcntntpos < prkdcntntL {
							prkdcntntpoint = prkdcntntrs.seekis[prkdcntntpos]
						}
					}
				} else if err = prkdcntntrs.WriteSeekedPos(w, prkdcntntpos, 0); err == nil {
					prkdcntntpos++
					if prkdcntntpos < prkdcntntL {
						prkdcntntpoint = prkdcntntrs.seekis[prkdcntntpos]
					}
				}
			} else if isPrkdAtv() && (prkdcdepos == prkdcdeL || atvapcpoint[1] < prkdcdepoint[0]) && (prkdcntntpos == prkdcntntL || atvapcpoint[1] < prkdcntntpoint[0]) {
				captureAtvRSAPContent()
			}
		}
	}

	for err == nil && (cdepos < cdeL || cntntpos < cntntL || atvapcpos < atvapcL || prkdPointPos < prkdPointL) {
		if prkdPointPos < prkdPointL && (cdepos == cdeL || prkdpnt[1] < cdepoint[0]) && (cntntpos == cntntL || prkdpnt[1] < cntntpoint[0]) && (atvapcpos == atvapcL || prkdpnt[1] < atvapcpoint[0]) {
			captureParkedPoint(prkdPointPos)
			prkdPointPos++
			if prkdPointPos < prkdPointL {
				prkdpnt = prkdPoint.seekis[prkdPointPos][:]
			}
		} else if cdepos < cdeL && (cntntpos == cntntL || cdepoint[1] < cntntpoint[0]) && (atvapcpos == atvapcL || cdepoint[1] < atvapcpoint[0]) && (prkdPointPos == prkdPointL || cdepoint[1] < prkdpnt[0]) {
			captureCode()
		} else if cntntpos < cntntL && (cdepos == cdeL || cntntpoint[1] < cdepoint[0]) && (atvapcpos == atvapcL || cntntpoint[1] < atvapcpoint[0]) && (prkdPointPos == prkdPointL || cntntpoint[1] < prkdpnt[0]) {
			captureContent()
		} else if atvapcpos < atvapcL && (cdepos == cdeL || atvapcpoint[1] < cdepoint[0]) && (cntntpos == cntntL || atvapcpoint[1] < cntntpoint[0]) && (prkdPointPos == prkdPointL || atvapcpoint[1] < prkdpnt[0]) {
			captureAtvRSAPContent()
		} else {
			err = fmt.Errorf("Should not get here")
		}
	}

	return
}

func (atvRsAPC *activeRSActivePassiveContent) cleanupActiveRSActivePassiveContent() {
	if atvRsAPC.Seeker != nil {
		atvRsAPC.Seeker.ClearSeeker()
		atvRsAPC.Seeker = nil
	}
	if atvRsAPC.atvRSAPCMap != nil {
		for n := range atvRsAPC.atvRSAPCMap {
			atvRsAPC.atvRSAPCMap[n] = nil
			delete(atvRsAPC.atvRSAPCMap, n)
		}
	}
	if atvRsAPC.cders != nil {
		atvRsAPC.cders.clearCodeSeekReader()
		atvRsAPC.cders = nil
	}
	if atvRsAPC.cntntrs != nil {
		atvRsAPC.cntntrs.clearContentSeekReader()
		atvRsAPC.cntntrs = nil
	}

	if atvRsAPC.prkdPoint != nil {
		atvRsAPC.prkdPoint.cleanupParksRsAPCPoint()
		atvRsAPC.prkdPoint = nil
	}

	if atvRsAPC.atvparse != nil {
		atvRsAPC.atvparse = nil
	}
	if atvRsAPC.cdeIO != nil {
		atvRsAPC.cdeIO.Close()
		atvRsAPC.cdeIO = nil
	}
}

func (atvRsAPC *activeRSActivePassiveContent) Append(atvRsAtvPsvCntnt *activeRSActivePassiveContent, starti int64, endi int64) {
	if atvRsAPC.atvRSAPCMap == nil {
		atvRsAPC.atvRSAPCMap = map[int]*activeRSActivePassiveContent{}
	}

	atvRsAtvPsvCntnt.atvRsACPi = atvRsAPC.atvRsAPCCount

	if atvRsAPC.prkdPoint != nil && atvRsAPC.prkdPoint.enabled {
		atvRsAPC.prkdPoint.prkdAtvrsATCmap[atvRsAtvPsvCntnt.atvRsACPi] = true
	}

	atvRsAPC.atvRSAPCMap[atvRsAtvPsvCntnt.atvRsACPi] = atvRsAtvPsvCntnt

	atvRsAPC.Seeker.Append(starti, endi)
	if atvRsAPC.lastStartRsI == -1 {
		atvRsAPC.lastStartRsI = len(atvRsAPC.seekis) - 1
	}
	if atvRsAPC.lastStartRsI > -1 {
		atvRsAPC.lastEndRsI = len(atvRsAPC.seekis) - 1
	}
	atvRsAPC.atvRsAPCCount++
}

type contentSeekReader struct {
	atvrs    *activeReadSeeker
	atvRsAPC *activeRSActivePassiveContent
	*IOSeekReader
}

func (cntntsr *contentSeekReader) empty() bool {
	return cntntsr.IOSeekReader.Empty()
}

func (cntntsr *contentSeekReader) Append(starti int64, endi int64) {
	cntntsr.IOSeekReader.Append(starti, endi)
}

func (cntntsr *contentSeekReader) clearContentSeekReader() {
	if cntntsr.IOSeekReader != nil {
		cntntsr.IOSeekReader.ClearIOSeekReader()
		cntntsr.IOSeekReader = nil
	}
	if cntntsr.atvRsAPC != nil {
		cntntsr.atvRsAPC = nil
	}
}

type codeSeekReader struct {
	atvRsAPC *activeRSActivePassiveContent
	atvrs    *activeReadSeeker
	*IOSeekReader
}

func (cdesr *codeSeekReader) empty() bool {
	return cdesr.IOSeekReader.Empty()
}

func (cdesr *codeSeekReader) Append(starti int64, endi int64) {
	cdesr.IOSeekReader.Append(starti, endi)
}

func (cdesr *codeSeekReader) clearCodeSeekReader() {
	if cdesr.IOSeekReader != nil {
		cdesr.IOSeekReader.ClearIOSeekReader()
		cdesr.IOSeekReader = nil
	}
	if cdesr.atvRsAPC != nil {
		cdesr.atvRsAPC = nil
	}
	if cdesr.atvrs != nil {
		cdesr.atvrs = nil
	}
}

func newContentSeekReader(atvRsAPC *activeRSActivePassiveContent, atvrs *activeReadSeeker) *contentSeekReader {
	cntntsr := &contentSeekReader{atvrs: atvrs, atvRsAPC: atvRsAPC, IOSeekReader: NewIOSeekReader(atvrs)}
	return cntntsr
}

func newCodeSeekReader(atvrs *activeReadSeeker, atvRsAPC *activeRSActivePassiveContent) *codeSeekReader {
	cdesr := &codeSeekReader{atvrs: atvrs, atvRsAPC: atvRsAPC, IOSeekReader: NewIOSeekReader(atvrs)}
	return cdesr
}

var emptyIO = &IORW{cached: false}

func (atvparse *activeParser) WriteContentByPos(pos int, atvRsAPCi ...int) {
	var atvRsAPC = atvparse.atvRsAPCStart
	if len(atvRsAPCi) > 0 {
		for _, atvRACPi := range atvRsAPCi {
			atvRsAPC = atvRsAPC.atvRSAPCMap[atvRACPi]
		}
	}
	if atvRsAPC.cntntrs != nil && len(atvRsAPC.cntntrs.seekis) > 0 {
		atvRsAPC.cntntrs.WriteSeekedPos(atvparse.atv.w, pos, 0)
	}
}

func (atvparse *activeParser) WriteParkedContentByPos(prkdpi int, pos int, atvRsAPCi ...int) {
	var atvRsAPC = atvparse.atvRsAPCStart
	if len(atvRsAPCi) > 0 {
		for _, atvRACPi := range atvRsAPCi {
			atvRsAPC = atvRsAPC.atvRSAPCMap[atvRACPi]
		}
	}
	if atvRsAPC.prkdPoint != nil && atvRsAPC.prkdPoint.prkdcntntrs[prkdpi] != nil && len(atvRsAPC.prkdPoint.prkdcntntrs[prkdpi].seekis) > 0 {
		atvRsAPC.prkdPoint.prkdcntntrs[prkdpi].WriteSeekedPos(atvparse.atv.w, pos, 0)
	}
}

func readatvrs(token *activeParseToken, p []byte) (n int, err error) {
	if token.atvRsAPC.prkdPoint != nil && token.atvRsAPC.prkdPoint.enabled {
		n, err = token.atvRsAPC.prkdPoint.Read(p)
		if err == io.EOF {
			if token.atvRsAPC.prkdPoint.endRIndex > token.atvRsAPC.prkdPoint.prkdStartI {
				token.atvRsAPC.prkdPoint.eofEndRIndex = token.atvRsAPC.prkdPoint.endRIndex - 1
			} else {
				token.atvRsAPC.prkdPoint.eofEndRIndex = token.atvRsAPC.prkdPoint.endRIndex
			}
			token.atvRsAPC.prkdPoint.startRIndex = token.atvRsAPC.prkdPoint.prkdStartI
			token.atvRsAPC.prkdPoint.endRIndex = token.atvRsAPC.prkdPoint.prkdStartI
			token.atvRsAPC.prkdPoint.prkdCurI = token.atvRsAPC.prkdPoint.prkdStartI
		} else {
			token.atvRsAPC.prkdPoint.lastEndRIndex = token.atvRsAPC.prkdPoint.endRIndex
			token.atvRsAPC.prkdPoint.startRIndex = token.atvRsAPC.prkdPoint.endRIndex
			token.atvRsAPC.prkdPoint.endRIndex += int64(n)
		}
	} else {
		if token.atvrs != nil {
			n, err = token.atvrs.Read(p)
		} else {
			err = io.EOF
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
	for {
		if tknrn, tknrnerr = readatvrs(token, token.tknrb); tknrn == 0 && tknrnerr == nil {
			tknrnerr = io.EOF
			return
		}
		return tknrn, tknrnerr
	}
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
	atvrspath string
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
}

func (atvrs *activeReadSeeker) Seek(offset int64, whence int) (n int64, err error) {
	return atvrs.atvrsio.Seek(offset, whence)
}

func (atvrs *activeReadSeeker) Read(p []byte) (n int, err error) {
	return atvrs.atvrsio.Read(p)
}

func (atvparse *activeParser) code() (s string) {
	if atvparse.atvRsAPCStart != nil && atvparse.atvRsAPCStart.cdeIO != nil && !atvparse.atvRsAPCStart.cdeIO.Empty() {
		return atvparse.atvRsAPCStart.cdeIO.String()
	}
	return
}

func (atvparse *activeParser) parse(rs io.ReadSeeker, root string, path string, retrieveRS func(string, string) (io.ReadSeeker, error), altlbls ...string) (parseerr error) {
	if atvparse.retrieveRS == nil || &atvparse.retrieveRS != &retrieveRS {
		atvparse.retrieveRS = retrieveRS
	}
	atvparse.setRS(path, rs)

	parseerr = parseNextToken(nil, atvparse, path, -1, -1, altlbls...)

	if parseerr == nil && atvparse.atvRsAPCStart != nil {
		parseerr = evalStartActiveRSEntryPoint(atvparse, atvparse.atvRsAPCStart, atvparse.atv.w)
	} else {
		fmt.Println(parseerr)
	}
	return
}

func parseNextToken(token *activeParseToken, atvparse *activeParser, rspath string, atvRsAPCStartIndex int64, atvRsAPCEndIndex int64, altlbls ...string) (parseErr error) {
	atvtoken := nextActiveParseToken(token, atvparse, rspath, atvRsAPCStartIndex, atvRsAPCEndIndex, altlbls...)

	var tokenparsed bool
	var tokenerr error
	var prevtoken *activeParseToken
	for {
		if tokenparsed, tokenerr = atvtoken.parsing(); tokenparsed || tokenerr != nil {
			if tokenparsed && tokenerr == nil {
				tokenerr = atvtoken.wrapupActiveParseToken()
			}
			if tokenerr == nil {
				prevtoken, tokenerr = atvtoken.cleanupactiveParseToken()
			}
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

func evalStartActiveRSEntryPoint(atvparse *activeParser, atvRSAPCStart *activeRSActivePassiveContent, w io.Writer) (err error) {
	if err = atvRSAPCCoding(false, atvRSAPCStart, atvparse, w); err == nil {
		var s = atvparse.code()

		if s != "" {
			err = atvparse.atv.evalCode(func() string {
				return s
			}, map[string]interface{}{"_out": atvparse.atv.Out(), "_atvparse": atvparse, "_parameters": atvparse.atv.params, "@db@execute": func(alias string, query string, args ...interface{}) *DBExecuted {
				return DatabaseManager().Execute(alias, query, args...)
			}, "@db@query": func(alias string, query string, args ...interface{}) *DBQuery {
				return DatabaseManager().Query(alias, query, args...)
			}})
		}
	}

	return
}

func nextActiveParseToken(token *activeParseToken, parser *activeParser, rspath string, atvRsAPCStartIndex int64, atvRsAPCEndIndex int64, altlbls ...string) (nexttoken *activeParseToken) {
	rspathext := filepath.Ext(rspath)
	rspathname := rspath
	rsroot := rspath
	if strings.LastIndex(rspath, "/") > -1 {
		rsroot = rspath[0 : strings.LastIndex(rspath, "/")+1]
	} else {
		rsroot = ""
	}

	if rspathext == "" {
		if token != nil {
			rspathext = token.rspathext
		}
	}

	rspathname = strings.ReplaceAll(rspathname, "/", ":")

	if strings.HasSuffix(rspathname, rspathext) {
		rspathname = rspathname[:len(rspathname)-len(rspathext)]
	}
	rsrootname := rspathname[len(rsroot):len(rspathname)]

	if parser.atvTkns == nil {
		parser.atvTkns = map[*activeParseToken]*activeParseToken{}
	}
	if parser.atvRsAPCMap == nil {
		parser.atvRsAPCMap = map[*activeRSActivePassiveContent]*activeRSActivePassiveContent{}
	}

	if altlbls == nil || len(altlbls) != 2 {
		altlbls = []string{"<@", "@>"}
	}

	nexttoken = &activeParseToken{
		startRIndex:   0,
		endRIndex:     0,
		parse:         parser,
		tknrb:         make([]byte, 1),
		curStartIndex: -1, curEndIndex: -1,
		rsroot:          rsroot,
		rspath:          rspath,
		rspathname:      rspathname,
		rsrootname:      rsrootname,
		rspathext:       rspathext,
		procRStartIndex: -1, procREndIndex: -1,
		lbls:               altlbls[:2],
		lblsi:              []int{0, 0},
		parkedStartIndex:   -1,
		parkedEndIndex:     -1,
		atvRsAPCStartIndex: atvRsAPCStartIndex,
		atvRsAPCEndIndex:   atvRsAPCEndIndex,
		parkedLevel:        0}
	parser.atvTkns[nexttoken] = token

	if nexttoken.atvrs = parser.atvrs(rspath); nexttoken.atvrs != nil {
		nexttoken.atvrs.Seek(0, 0)
	}

	var atvRsAPC = newActiveRSActivePassiveContent(nexttoken.atvrs, parser)

	if token == nil {
		parser.atvRsAPCMap[atvRsAPC] = nil
	} else {
		if token != nil {
			if atvRsAPCStartIndex >= 0 && atvRsAPCStartIndex <= atvRsAPCEndIndex {
				token.appendAtvRsAPC(atvRsAPC, atvRsAPCStartIndex, atvRsAPCEndIndex)
				if atvRsAPCStartIndex > -1 && atvRsAPCStartIndex < atvRsAPCEndIndex {
					if atvRsAPC.prkdPoint == nil {
						if token.atvRsAPC.prkdPoint != nil && token.atvRsAPC.prkdPoint.enabled {
							atvRsAPC.prkdPoint = newParkedRsAPCPoint(token.atvRsAPC.prkdPoint.atvrs, atvRsAPCStartIndex+1, atvRsAPCEndIndex)
						} else {
							atvRsAPC.prkdPoint = newParkedRsAPCPoint(token.atvrs, atvRsAPCStartIndex+1, atvRsAPCEndIndex)
						}
					}
				}
			} else {
				atvRsAPC.cleanupActiveRSActivePassiveContent()
				atvRsAPC = nil
			}
		}
	}

	if atvRsAPC != nil {
		atvRsAPC.tokenPath = nexttoken.tokenPathName()
		nexttoken.atvRsAPC = atvRsAPC
	}
	if parser.atvRsAPCStart == nil && atvRsAPC != nil {
		parser.atvRsAPCStart = atvRsAPC
	}

	return nexttoken
}

type activeParseToken struct {
	parse            *activeParser
	lbls             []string
	lblsi            []int
	atvprevb         byte
	hasAtv           bool
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
	rsrootname       string
	rspathext        string

	// ELEM VALID SETTINGS
	parkedStartIndex int64
	parkedEndIndex   int64
	parkedLevel      int
	elemName         string
	//UNPARKED
	startRIndex   int64
	lastEndRIndex int64
	endRIndex     int64
	eofEndRIndex  int64
	//PARSE
	procRStartIndex    int64
	procREndIndex      int64
	atvRsAPC           *activeRSActivePassiveContent
	atvRsAPCStartIndex int64
	atvRsAPCEndIndex   int64
}

func (token *activeParseToken) prevToken() *activeParseToken {
	if token.parse != nil && token.parse.atvTkns != nil {
		return token.parse.atvTkns[token]
	}
	return nil
}

func (token *activeParseToken) appendCde(atvRSAPC *activeRSActivePassiveContent, atvRStartIndex int64, atvREndIndex int64) {
	if token.atvRsAPC.prkdPoint != nil && token.atvRsAPC.prkdPoint.enabled {
		atvRSAPC.prkdPoint.AppendCde(atvRStartIndex, atvREndIndex)
	} else {
		atvRSAPC.cders.Append(atvRStartIndex, atvREndIndex)
	}
}

func (token *activeParseToken) appendCntnt(atvRSAPC *activeRSActivePassiveContent, psvRStartIndex int64, psvREndIndex int64) {
	if atvRSAPC.prkdPoint != nil && atvRSAPC.prkdPoint.enabled {
		atvRSAPC.prkdPoint.AppendCntnt(psvRStartIndex, psvREndIndex)
	} else {
		atvRSAPC.cntntrs.Append(psvRStartIndex, psvREndIndex)
	}
}

func (token *activeParseToken) appendAtvRsAPC(atvRsAPC *activeRSActivePassiveContent, atvRsAPCRStartIndex int64, atvRsAPCREndIndex int64) {
	token.atvRsAPC.Append(atvRsAPC, atvRsAPCRStartIndex, atvRsAPCREndIndex)
	token.parse.atvRsAPCMap[atvRsAPC] = token.atvRsAPC
}

func (token *activeParseToken) tokenPathName() string {
	var prevToken = token.parse.atvTkns[token]
	if prevToken == nil {
		return token.rspathname
	}
	return prevToken.tokenPathName() + "/" + token.rspathname
}

func (token *activeParseToken) wrapupActiveParseToken() (err error) {
	if token.atvRsAPC != nil {
		if token.atvRsAPC.prkdPoint != nil {
			var prevToken = token.parse.atvTkns[token]
			if prevToken != nil {
				if prevToken.atvRsAPC.prkdPoint != nil && prevToken.atvRsAPC.prkdPoint.enabled {
					prevToken.atvRsAPC.prkdPoint.endRIndex, err = prevToken.atvRsAPC.prkdPoint.atvrs.Seek(prevToken.atvRsAPC.prkdPoint.endRIndex, 0)
				} else {
					prevToken.endRIndex, err = prevToken.atvrs.Seek(prevToken.endRIndex, 0)
				}
			}
		}
	}
	return
}

func (token *activeParseToken) cleanupactiveParseToken() (prevtoken *activeParseToken, err error) {
	if token.lbls != nil {
		token.lbls = nil
	}
	if token.lblsi != nil {
		token.lblsi = nil
	}
	if token.parse != nil {
		if token.parse.atvTkns != nil {
			if token.parkedLevel > 0 {
				err = fmt.Errorf("[parse-error]" + token.tokenPathName() + "-" + "token note closed")
			}
			if _, tokenOk := token.parse.atvTkns[token]; tokenOk {
				prevtoken = token.parse.atvTkns[token]
				delete(token.parse.atvTkns, token)
			}
		}
		token.parse = nil
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
	if token.atvRsAPC != nil {
		if token.atvRsAPC.Empty() && token.atvRsAPC.cders.Empty() && token.atvRsAPC.cntntrs.Empty() && (token.atvRsAPC.prkdPoint == nil || token.atvRsAPC.prkdPoint.empty()) {
			token.atvRsAPC.cleanupActiveRSActivePassiveContent()
		}
		token.atvRsAPC = nil
	}
	if token.atvrs != nil {
		token.atvrs = nil
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
	token.nr, token.rerr = token.parse.readActive(token)
	return parseActiveToken(token, token.lbls, token.lblsi)
}

func parseActiveToken(token *activeParseToken, lbls []string, lblsi []int) (nextparse bool, err error) {
	if token.nr > 0 {
		if lblsi[1] == 0 && lblsi[0] < len(lbls[0]) {
			if lblsi[0] > 1 && lbls[0][lblsi[0]-1] == token.atvprevb && lbls[0][lblsi[0]] != token.tknrb[0] {
				if token.psvCapturedIO != nil && !token.psvCapturedIO.Empty() {
					token.psvCapturedIO.Close()
				}
				token.passiveCapturedIO().Print(lbls[0][:lblsi[0]])
				if token.atvRsAPC.prkdPoint != nil && token.atvRsAPC.prkdPoint.enabled {
					if token.curStartIndex == -1 && token.parkedStartIndex == -1 {
						token.curStartIndex = token.atvRsAPC.prkdPoint.prkdCurI - 1 - int64(lblsi[0])
					}
				} else {
					if token.curStartIndex == -1 && token.parkedStartIndex == -1 {
						token.curStartIndex = token.startRIndex - int64(lblsi[0])
					}
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
			if token.curStartIndex == -1 && token.parkedStartIndex == -1 {
				if lblsi[0] > 0 {
					if token.atvRsAPC.prkdPoint != nil && token.atvRsAPC.prkdPoint.enabled {
						token.curStartIndex = token.atvRsAPC.prkdPoint.prkdCurI - 1 - int64(lblsi[0])
					} else {
						token.curStartIndex = token.startRIndex - int64(lblsi[0])
					}
				} else {
					if token.atvRsAPC.prkdPoint != nil && token.atvRsAPC.prkdPoint.enabled {
						token.curStartIndex = token.atvRsAPC.prkdPoint.prkdCurI - 1
					} else {
						token.curStartIndex = token.startRIndex
					}
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

			if token.psvCapturedIO != nil && !token.psvCapturedIO.Empty() && token.psvCapturedIO.HasSuffix([]byte(lbls[1][1:])) {
				if token.psvCapturedIO.HasPrefixSuffix([]byte(lbls[0][0:len(lbls)-1]), []byte(lbls[1][1:])) {
					if valid, single, complexStart, complexEnd, elemName, elemPath, elemExt, valErr := validatePassiveCapturedIO(token, lbls); valid {
						if single || complexStart {
							if token.curStartIndex > -1 {
								if token.curEndIndex == -1 {
									if token.atvRsAPC.prkdPoint != nil && token.atvRsAPC.prkdPoint.enabled {
										token.curEndIndex = token.atvRsAPC.prkdPoint.lastEndRIndex - token.psvCapturedIO.Size()
									} else {
										token.curEndIndex = token.lastEndRIndex - token.psvCapturedIO.Size()
									}
								}
								if token.tknrb, err = captureCurrentPassiveStartEnd(token, token.rerr != nil && token.rerr == io.EOF, token.curStartIndex, token.curEndIndex, token.startRIndex); err != nil {
									return nextparse, err
								}
							}
							if token.parkedStartIndex == -1 {
								if token.atvRsAPC.prkdPoint != nil && token.atvRsAPC.prkdPoint.enabled {
									token.parkedStartIndex = token.atvRsAPC.prkdPoint.lastEndRIndex
								} else {
									token.parkedStartIndex = token.lastEndRIndex
								}
							}
						}
						if single || complexEnd {
							if token.parkedStartIndex > -1 {
								if token.parkedEndIndex == -1 {
									if single {
										if token.atvRsAPC.prkdPoint != nil && token.atvRsAPC.prkdPoint.enabled {
											token.parkedEndIndex = token.atvRsAPC.prkdPoint.lastEndRIndex
										} else {
											token.parkedEndIndex = token.lastEndRIndex
										}
									} else if complexEnd {
										if token.atvRsAPC.prkdPoint != nil && token.atvRsAPC.prkdPoint.enabled {
											token.parkedEndIndex = token.atvRsAPC.prkdPoint.lastEndRIndex - token.psvCapturedIO.Size()
										} else {
											token.parkedEndIndex = token.lastEndRIndex - token.psvCapturedIO.Size()
										}
									}
								}
							}
							if elemName != "" && elemPath != "" && elemExt != "" {
								if !(strings.HasPrefix(elemPath, "./") || strings.HasPrefix(elemPath, "/")) {
									if token.rsroot != "" {
										elemPath = token.rsroot + elemPath
									}
								}
								if single && (elemName == (".:"+token.rsrootname) || elemName == (":"+token.rsrootname)) && token.rspathext == elemExt {
									if token.atvRsAPC.prkdPoint != nil {
										var atvrsapcei = token.endRIndex
										/*if token.atvRsAPC.prkdPoint != nil && token.atvRsAPC.prkdPoint.enabled {
											atvrsapcei = token.atvRsAPC.prkdPoint.endRIndex
										}*/
										var atvrsapcsi = atvrsapcei - token.psvCapturedIO.Size() + 1

										token.atvRsAPC.prkdPoint.Append(atvrsapcsi, atvrsapcei-1)
										err = token.atvRsAPC.prkdPoint.enable()
									}
									token.parkedStartIndex = -1
									token.parkedEndIndex = -1
								} else {
									/*if strings.HasPrefix(elemPath, "./") {
										elemPath = elemPath[2:]
									} else if strings.HasPrefix(elemPath, "/") {
										elemPath = elemPath[1:]
									}
									elemPath = token.rsroot + elemPath
									*/
									if err = token.parse.setRSByPath(elemPath); err == nil {
										var atvrsapcsi = token.parkedStartIndex
										var atvrsapcei = token.parkedEndIndex
										token.parkedStartIndex = -1
										token.parkedEndIndex = -1
										err = parseNextToken(token, token.parse, elemPath, atvrsapcsi, atvrsapcei, token.lbls...)
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
					if token.parkedLevel == 0 && token.procRStartIndex > -1 && token.procREndIndex > -1 {
						token.appendCde(token.atvRsAPC, token.procRStartIndex, token.procREndIndex)
					}

					if token.procRStartIndex > -1 {
						token.procRStartIndex = -1
					}
					if token.procREndIndex > -1 {
						token.procREndIndex = -1
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
					if token.atvRsAPC.prkdPoint != nil && token.atvRsAPC.prkdPoint.enabled {
						token.curEndIndex = token.atvRsAPC.prkdPoint.lastEndRIndex - 1 - int64(len(lbls[1]))
					} else {
						token.curEndIndex = token.lastEndRIndex - 1 - int64(len(lbls[1]))
					}
				}
				if token.tknrb, err = captureCurrentPassiveStartEnd(token, token.rerr != nil && token.rerr == io.EOF, token.curStartIndex, token.curEndIndex, token.startRIndex); err != nil {
					return nextparse, err
				}
			}
			if token.parkedLevel == 0 && !token.hasAtv && strings.TrimSpace(string(token.tknrb)) != "" {
				token.hasAtv = true
			}

			if lblsi[1] > 0 {
				lblsi[1] = 0
			}

			if token.parkedLevel == 0 && token.hasAtv {
				if token.procRStartIndex == -1 {
					if token.atvRsAPC.prkdPoint != nil && token.atvRsAPC.prkdPoint.enabled {
						token.procRStartIndex = token.atvRsAPC.prkdPoint.startRIndex
					} else {
						token.procRStartIndex = token.startRIndex
					}
				}
				if token.hasAtv && token.procRStartIndex > -1 && strings.TrimSpace(string(token.tknrb)) != "" {
					if token.atvRsAPC.prkdPoint != nil && token.atvRsAPC.prkdPoint.enabled {
						token.procREndIndex = token.atvRsAPC.prkdPoint.startRIndex
					} else {
						token.procREndIndex = token.startRIndex
					}
				}
			}
			return nextparse, err
		}
	} else if token.rerr == io.EOF {
		if token.curStartIndex > -1 && token.curEndIndex == -1 {
			if token.atvRsAPC.prkdPoint != nil && token.atvRsAPC.prkdPoint.enabled {
				token.curEndIndex = token.atvRsAPC.prkdPoint.eofEndRIndex
			} else {
				token.curEndIndex = token.eofEndRIndex
			}
		}
		if token.curStartIndex > -1 && token.curEndIndex > -1 {
			token.tknrb, err = captureCurrentPassiveStartEnd(token, true, token.curStartIndex, token.curEndIndex, token.startRIndex)
		}

		if token.atvRsAPC.prkdPoint != nil && token.atvRsAPC.prkdPoint.enabled {
			token.atvRsAPC.prkdPoint.enabled = false
			if token.psvCapturedIO != nil && !token.psvCapturedIO.Empty() {
				token.psvCapturedIO.Close()
			}
		} else {
			nextparse = true
		}
	}
	return nextparse, err
}

func validatePassiveCapturedIO(token *activeParseToken, lbls []string) (valid bool, single bool, comlexStart bool, complexEnd bool, elemName string, elemPath string, elemExt string, err error) {
	if actualSize := (token.psvCapturedIO.Size() - int64(len(lbls[0:len(lbls[0])-1])+len(lbls[1][1:]))); actualSize >= 1 {
		if valid = (actualSize == 1 && token.psvCapturedIO.String() == "/"); !valid {
			token.psvCapturedIO.Seek(int64(len(lbls[0:len(lbls[0])-1])), 0)

			actualSizei := int64(0)

			foundFSlash := false

			elemName = ""

			var validElemName = false

			for actualSizei < actualSize && !validElemName {
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
								break
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
							if strings.HasPrefix(elemName, ".:") {
								if elemExt = filepath.Ext(elemName[2:]); elemExt != "" {
									elemName = elemName[0 : len(elemName)-len(elemExt)]
								} else if elemExt == "" {
									elemExt = token.rspathext
								}
							} else {
								if elemExt = filepath.Ext(elemName); elemExt != "" {
									elemName = elemName[0 : len(elemName)-len(elemExt)]
								} else if elemExt == "" {
									elemExt = token.rspathext
								}
							}
							if strings.HasPrefix(elemName, ".:") {
								elemPath = strings.ReplaceAll(elemName[2:], ":", "/") + elemExt
							} else if strings.HasPrefix(elemName, ":") {
								elemPath = strings.ReplaceAll(elemName[1:], ":", "/") + elemExt
							} else {
								elemPath = strings.ReplaceAll(elemName, ":", "/") + elemExt
							}

							if err = token.parse.setRSByPath(token.rsroot + elemPath); err == nil {
								if token.parse.atvrs(token.rsroot+elemPath) != nil {
									validElemName = true
								}
							}
							valid = validElemName
							break
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

func captureCurrentPassiveStartEnd(token *activeParseToken, eof bool, curStartIndex, curEndIndex, startRIndex int64) (lasttknrb []byte, err error) {
	token.curStartIndex = -1
	token.curEndIndex = -1
	lasttknrb = token.tknrb[:]
	if curStartIndex > -1 && curEndIndex > -1 && curStartIndex <= curEndIndex {
		token.appendCntnt(token.atvRsAPC, curStartIndex, curEndIndex)
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
			err = vmeval.err
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

const tagstartregexp string = `^((:(([a-z]|[A-Z])\w*)+)|(.:(([a-z]|[A-Z])\w*)+)|(([a-z]|[A-Z])+(:(([a-z]|[A-Z])\w*)+)+))+(:(([a-z]|[A-Z])\w*)+)*(-(([a-z]|[A-Z])\w*)+)?(.([a-z]|[A-Z])+)?$`

var regexptagstart *regexp.Regexp

//const fulltagstartregexp string = `^(([<]|[</])+((:(([a-z]|[A-Z])\w*)+)|(.:(([a-z]|[A-Z])\w*)+)|(([a-z]|[A-Z])+(:(([a-z]|[A-Z])\w*)+)+))+(:(([a-z]|[A-Z])\w*)+)+)*(-(([a-z]|[A-Z])\w*)+)?(.([a-z]|[A-Z])+)?([>]|[/>])+$`

//var regexpfulltagstart *regexp.Regexp

const propregexp string = `^-?-?(([a-z]+[0-9]*)[a-z]*)+(-([a-z]+[0-9]*)[a-z]*)?$`

var regexprop *regexp.Regexp

const propvalnumberexp string = `^[-+]?\d+([.]\d+)?$`

var regexpropvalnumberexp *regexp.Regexp

func init() {

	if regexptagstart == nil {
		regexptagstart = regexp.MustCompile(tagstartregexp)
	}

	//if regexpfulltagstart == nil {
	//	regexpfulltagstart = regexp.MustCompile(fulltagstartregexp)
	//}

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
func (atvpros *ActiveProcessor) Process(rs io.ReadSeeker, root string, path string, retrieveRS RetrieveRSFunc, altlbls ...string) (err error) {
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
			if err = executeCommand(atvpros, path, cmdhnlrs[cmdtoexecn], cmdhnlrparams[cmdtoexecn]...); err != nil {
				break
			}
		}
	} else {
		err = atvpros.atvParser.parse(rs, root, path, retrieveRS, altlbls...)
	}
	if err == nil {
		runtime.GC()
	}
	return err
}
