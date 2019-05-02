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

type activeRSActivePassiveContent struct {
	*Seeker
	cders   *codeSeekReader
	cntntrs *contentSeekReader
	//atvRsAtvPsvCntnts   []*activeRSActivePassiveContent
	atvRsAtvPsvCntntMap map[int]*activeRSActivePassiveContent
	atvparse            *activeParser
	atvRsACPi           int
	lastStartRsI        int
	lastEndRsI          int
	cdeIO               *IORW
	atvRsAPCCount       int
	tokenPath           string
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
	atvRsAPC.cders = newCodeSeekReader(atvrs, atvRsAPC, atvRsAPC.cntntrs)
	return
}

func (atvRsAPC *activeRSActivePassiveContent) appendLevels(atvparser *activeParser) (a []int) {
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
			for n, _ := range a {
				aref[n] = a[len(a)-(n+1)]
			}
			a = nil
			a = aref[:]
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
			atvrAPCiCode = fmt.Sprint(atvRsAPC.appendLevels(atvparser))
			atvrAPCiCode = strings.ReplaceAll(strings.Replace(strings.Replace(atvrAPCiCode, "[", "", 1), "]", "", 1), " ", ",")
		}
		if atvrAPCiCode != "" {
			atvrAPCiCode = "," + atvrAPCiCode
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

	var captureCode = func() {
		if ss, sserr := cders.StringSeedPos(cdepos, 0); sserr == nil {
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

	var captureContent = func() {
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
		if err = atvRSAPCCoding(hasCode, atvRsAPC.atvRsAtvPsvCntntMap[atvapcpos], atvparser, w); err == nil {
			if !hasCode && atvRsAPC.topRSAPC().cdeIO != nil && !atvRsAPC.topRSAPC().cdeIO.Empty() {
				hasCode = true
			}
			atvapcpos++
			if atvapcpos < atvapcL {
				atvapcpoint = atvRsAPC.seekis[atvapcpos][:]
			}
		}
	}

	for err == nil && (cdepos < cdeL || cntntpos < cntntL || atvapcpos < atvapcL) {
		if cdepos < cdeL && cntntpos < cntntL && atvapcpos < atvapcL {
			//ALL
			if cdepoint[1] < cntntpoint[0] && cdepoint[1] < atvapcpoint[0] {
				captureCode()
			} else if cntntpoint[1] < cdepoint[0] && cntntpoint[1] < atvapcpoint[0] {
				captureContent()
			} else if atvapcpoint[1] < cdepoint[0] && atvapcpoint[1] < cntntpoint[0] {
				captureAtvRSAPContent()
			}
		} else if cdepos < cdeL && cntntpos < cntntL && atvapcpos == atvapcL {
			//code && content
			if cdepoint[1] < cntntpoint[0] {
				captureCode()
			} else if cntntpoint[1] < cdepoint[0] {
				captureContent()
			}
		} else if cdepos == cdeL && cntntpos < cntntL && atvapcpos < atvapcL {
			//content && atvrsapc
			if cntntpoint[1] < atvapcpoint[0] {
				captureContent()
			} else if atvapcpoint[1] < cntntpoint[0] {
				captureAtvRSAPContent()
			}
		} else if cdepos < cdeL && cntntpos == cntntL && atvapcpos < atvapcL {
			//code && atvrsapc
			if cdepoint[1] < atvapcpoint[0] {
				captureCode()
			} else if atvapcpoint[1] < cdepoint[0] {
				captureAtvRSAPContent()
			}
		} else if cdepos < cdeL {
			//code
			captureCode()
		} else if cntntpos < cntntL {
			//content
			captureContent()
		} else if atvapcpos < atvapcL {
			//atvrsapc
			captureAtvRSAPContent()
		}
	}

	return
}

func (atvRsAPC *activeRSActivePassiveContent) cleanupActiveRSActivePassiveContent() {
	if atvRsAPC.Seeker != nil {
		atvRsAPC.Seeker.ClearSeeker()
		atvRsAPC.Seeker = nil
	}
	if atvRsAPC.atvRsAtvPsvCntntMap != nil {
		for n := range atvRsAPC.atvRsAtvPsvCntntMap {
			atvRsAPC.atvRsAtvPsvCntntMap[n] = nil
			delete(atvRsAPC.atvRsAtvPsvCntntMap, n)
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
	if atvRsAPC.atvparse != nil {
		atvRsAPC.atvparse = nil
	}
	if atvRsAPC.cdeIO != nil {
		atvRsAPC.cdeIO.Close()
		atvRsAPC.cdeIO = nil
	}
}

func (atvRsAPC *activeRSActivePassiveContent) Append(atvRsAtvPsvCntnt *activeRSActivePassiveContent, starti int64, endi int64) {
	if atvRsAPC.atvRsAtvPsvCntntMap == nil {
		atvRsAPC.atvRsAtvPsvCntntMap = map[int]*activeRSActivePassiveContent{}
	}
	/*if atvRsAPC.atvRsAtvPsvCntnts == nil {
		atvRsAPC.atvRsAtvPsvCntnts = []*activeRSActivePassiveContent{}
	}*/

	atvRsAtvPsvCntnt.atvRsACPi = atvRsAPC.atvRsAPCCount

	atvRsAPC.atvRsAtvPsvCntntMap[atvRsAtvPsvCntnt.atvRsACPi] = atvRsAtvPsvCntnt
	//atvRsAPC.atvRsAtvPsvCntnts = append(atvRsAPC.atvRsAtvPsvCntnts, atvRsAtvPsvCntnt)
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
	atvrs        *activeReadSeeker
	atvRsAPC     *activeRSActivePassiveContent
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
	if cntntsr.atvRsAPC != nil {
		cntntsr.atvRsAPC = nil
	}
}

type codeSeekReader struct {
	atvRsAPC *activeRSActivePassiveContent
	atvrs    *activeReadSeeker
	*IOSeekReader
	//cntntsr              *contentSeekReader
	//lastCntntStartI      int64
	//cntntseekristart     map[int64][]int
	//atvrsapcseekristart  map[int64][]int
	//cntntseekriendpos    []int
	//atvrsapcseekriendpos []int
}

/*func (cdesr *codeSeekReader) String() (s string) {
	if len(cdesr.seekis) > 0 {
		atvrAPCiCode := ""

		if cdesr.atvRsAPC.atvparse.atvRsAPCStart != cdesr.atvRsAPC {
			var atvRsAPC = cdesr.atvRsAPC
			var atvparse = cdesr.atvRsAPC.atvparse
			for atvRsAPC != nil {
				atvrAPCiCode += fmt.Sprintf("%d,", atvRsAPC.atvRsACPi)
				atvRsAPC = atvparse.atvRsAPCMap[atvRsAPC]
			}
		}

		if atvrAPCiCode != "" {
			atvrAPCiCode = "," + atvrAPCiCode
		}
		for skrsipos, skrsi := range cdesr.seekis {
			if cntntpos, cntntposok := cdesr.cntntseekristart[skrsi[0]]; cntntposok {
				for _, cntntposi := range cntntpos {
					s = s + "_atvparse.WriteContentByPos(" + fmt.Sprintf("%d", cntntposi) + atvrAPCiCode + ");"
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
				s = s + "_atvparse.WriteContentByPos(" + fmt.Sprintf("%d", cntntseekriendpos) + ");"
			}
		}
	}
	return s
}*/

func (cdesr *codeSeekReader) Append(starti int64, endi int64) {
	cdesr.IOSeekReader.Append(starti, endi)
	/*if !cdesr.cntntsr.Empty() || !cdesr.atvRsAPC.Empty() {
		if cdesr.cntntseekristart == nil {
			cdesr.cntntseekristart = make(map[int64][]int)
		}
		if cdesr.atvrsapcseekristart == nil {
			cdesr.atvrsapcseekristart = make(map[int64][]int)
		}

		if _, istartok := cdesr.cntntseekristart[starti]; !istartok {
			cntids := make([]int, (cdesr.cntntsr.lastEndRsI-cdesr.cntntsr.lastStartRsI)+1)
			atvrsapcids := make([]int, (cdesr.atvRsAPC.lastEndRsI-cdesr.atvRsAPC.lastStartRsI)+1)

			if (cdesr.cntntsr.lastStartRsI > -1 && cdesr.cntntsr.lastStartRsI <= cdesr.cntntsr.lastEndRsI) || (cdesr.atvRsAPC.lastStartRsI > -1 && cdesr.atvRsAPC.lastStartRsI <= cdesr.atvRsAPC.lastEndRsI) {
				if cdesr.cntntsr.lastStartRsI > -1 && cdesr.cntntsr.lastStartRsI <= cdesr.cntntsr.lastEndRsI {
					for cntidsn := range cntids {
						cntids[cntidsn] = cdesr.cntntsr.lastStartRsI
						cdesr.cntntsr.lastStartRsI++
					}
				} else if cdesr.atvRsAPC.lastStartRsI > -1 && cdesr.atvRsAPC.lastStartRsI <= cdesr.atvRsAPC.lastEndRsI {
					for atvapcidsn := range atvrsapcids {
						atvrsapcids[atvapcidsn] = cdesr.atvRsAPC.lastStartRsI
						cdesr.atvRsAPC.lastStartRsI++
					}
				}
			}
			cdesr.cntntseekristart[starti] = cntids[:]
			cdesr.atvrsapcseekristart[starti] = atvrsapcids[:]
			cdesr.cntntsr.lastStartRsI = -1
			cdesr.cntntsr.lastEndRsI = -1
			cdesr.atvRsAPC.lastStartRsI = -1
			cdesr.atvRsAPC.lastEndRsI = -1
			cntids = nil
			atvrsapcids = nil
		}

	}*/
}

func (cdesr *codeSeekReader) clearCodeSeekReader() {
	if cdesr.IOSeekReader != nil {
		cdesr.IOSeekReader.ClearIOSeekReader()
		cdesr.IOSeekReader = nil
	}
	/*if cdesr.cntntsr != nil {
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
	if cdesr.atvrsapcseekriendpos != nil {
		cdesr.atvrsapcseekriendpos = nil
	}*/
	if cdesr.atvRsAPC != nil {
		cdesr.atvRsAPC = nil
	}
	if cdesr.atvrs != nil {
		cdesr.atvrs = nil
	}
}

func newContentSeekReader(atvRsAPC *activeRSActivePassiveContent, atvrs *activeReadSeeker) *contentSeekReader {
	cntntsr := &contentSeekReader{atvrs: atvrs, atvRsAPC: atvRsAPC, IOSeekReader: NewIOSeekReader(atvrs), lastStartRsI: -1, lastEndRsI: -1}
	return cntntsr
}

func newCodeSeekReader(atvrs *activeReadSeeker, atvRsAPC *activeRSActivePassiveContent, cntntsr *contentSeekReader) *codeSeekReader {
	cdesr := &codeSeekReader{atvrs: atvrs, atvRsAPC: atvRsAPC, IOSeekReader: NewIOSeekReader(atvrs) /* lastCntntStartI: -1, cntntsr: cntntsr*/}
	return cdesr
}

var emptyIO = &IORW{cached: false}

func (atvparse *activeParser) WriteContentByPos(pos int, atvRsAPCi ...int) {
	var atvRsAPC = atvparse.atvRsAPCStart
	if len(atvRsAPCi) > 0 {
		for _, atvRACPi := range atvRsAPCi {
			atvRsAPC = atvRsAPC.atvRsAtvPsvCntntMap[atvRACPi]
		}
	}
	if atvRsAPC.cntntrs != nil && len(atvRsAPC.cntntrs.seekis) > 0 {
		atvRsAPC.cntntrs.WriteSeekedPos(atvparse.atv.w, pos, 0)
	}
}

func readatvrs(token *activeParseToken, p []byte) (n int, err error) {
	if token.curPrkdAtvRs != nil && token.curPrkdAtvRs.enabled {
		n, err = token.curPrkdAtvRs.Read(p)
		if err == io.EOF {
			token.curPrkdAtvRs.enabled = false
		}
		n, err = readatvrs(token, p)
	} else if token.atvrs != nil {
		n, err = token.atvrs.Read(p)
	} else {
		err = io.EOF
	}
	return n, err
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
	atvrspath string
}

type parkedActiveReaderSeeker struct {
	*activeReadSeeker
	prkdStartI int64
	prkdEndI   int64
	prdkAtvRsi int
	token      *activeParseToken
	enabled    bool
}

func (prdkAtvRs *parkedActiveReaderSeeker) cleanupParkedActiveReaderSeeker() {
	if prdkAtvRs.activeReadSeeker != nil {
		prdkAtvRs.activeReadSeeker = nil
	}
	if prdkAtvRs.token != nil {
		prdkAtvRs.token = nil
	}
}

func (prdkAtvRs *parkedActiveReaderSeeker) Read(p []byte) (n int, err error) {
	if prdkAtvRs.prkdStartI < prdkAtvRs.prkdEndI {

	}
	return
}

func newParkedActiveReaderSeeker(token *activeParseToken) (prkdAtvRS *parkedActiveReaderSeeker) {
	prkdAtvRS = &parkedActiveReaderSeeker{enabled: false, token: token, activeReadSeeker: token.atvrs, prkdStartI: token.atvRStartIndex, prkdEndI: token.atvREndIndex}
	return
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

func (atvparse *activeParser) parse(rs io.ReadSeeker, root string, path string, retrieveRS func(string, string) (io.ReadSeeker, error)) (parseerr error) {
	if atvparse.retrieveRS == nil || &atvparse.retrieveRS != &retrieveRS {
		atvparse.retrieveRS = retrieveRS
	}
	atvparse.setRS(path, rs)

	parseerr = parseNextToken(nil, atvparse, path, -1, -1)

	if parseerr == nil && atvparse.atvRsAPCStart != nil {
		parseerr = evalStartActiveRSEntryPoint(atvparse, atvparse.atvRsAPCStart, atvparse.atv.w)
	} else {
		fmt.Println(parseerr)
	}

	return
}

func parseNextToken(token *activeParseToken, atvparse *activeParser, rspath string, atvRsAPCStartIndex int64, atvRsAPCEndIndex int64) (parseErr error) {
	atvtoken := nextActiveParseToken(token, atvparse, rspath, atvRsAPCStartIndex, atvRsAPCEndIndex)

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

func nextActiveParseToken(token *activeParseToken, parser *activeParser, rspath string, atvRsAPCStartIndex int64, atvRsAPCEndIndex int64) (nexttoken *activeParseToken) {
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

	nexttoken = &activeParseToken{
		startRIndex:   0,
		endRIndex:     0,
		parse:         parser,
		tknrb:         make([]byte, 1),
		curStartIndex: -1, curEndIndex: -1,
		rsroot:         rsroot,
		rspath:         rspath,
		rspathname:     rspathname,
		rsrootname:     rsrootname,
		rspathext:      rspathext,
		atvRStartIndex: -1, atvREndIndex: -1,
		lbls:               []string{"<@", "@>"},
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

				if token.atvRsAPCStartIndex > -1 && token.atvRsAPCStartIndex < token.atvRsAPCEndIndex {
					var prdkAtvRs = newParkedActiveReaderSeeker(token)
					nexttoken.curPrkdAtvRs = prdkAtvRs
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
	parkedStartIndex   int64
	parkedEndIndex     int64
	parkedLevel        int
	elemName           string
	curPrkdAtvRs       *parkedActiveReaderSeeker
	startRIndex        int64
	lastEndRIndex      int64
	endRIndex          int64
	eofEndRIndex       int64
	atvRStartIndex     int64
	atvREndIndex       int64
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
	atvRSAPC.cders.Append(atvRStartIndex, atvREndIndex)
}

func (token *activeParseToken) appendCntnt(atvRSAPC *activeRSActivePassiveContent, psvRStartIndex int64, psvREndIndex int64) {
	atvRSAPC.cntntrs.Append(psvRStartIndex, psvREndIndex)
}

func (token *activeParseToken) appendAtvRsAPC(atvRsAPC *activeRSActivePassiveContent, atvRsAPCRStartIndex int64, atvRsAPCREndIndex int64) {
	token.atvRsAPC.Append(atvRsAPC, atvRsAPCRStartIndex, atvRsAPCREndIndex)
	token.parse.atvRsAPCMap[atvRsAPC] = token.atvRsAPC
}

func (token *activeParseToken) tokenPathName() string {
	var prevToken = token.parse.atvTkns[token]
	if prevToken == nil {
		return token.rspathname
	} else {
		return prevToken.tokenPathName() + "/" + token.rspathname
	}
}

func (token *activeParseToken) wrapupactiveParseToken() {
	if token.atvRsAPC != nil {
		/*var prevToken = token.parse.atvTkns[token]
		if prevToken != nil {

			if token.atvRsAPCStartIndex > -1 && token.atvRsAPCStartIndex <= token.atvRsAPCEndIndex {
				if token.atvRsAPC.Empty() && token.atvRsAPC.cders.Empty() && token.atvRsAPC.cntntrs.Empty() {
					token.atvRsAPCStartIndex = -1
					token.atvRsAPCEndIndex = -1
				} else {
					prevToken.appendAtvRsAPC(token.atvRsAPC, token.atvRsAPCStartIndex, token.atvRsAPCEndIndex)
					token.atvRsAPCStartIndex = -1
					token.atvRsAPCEndIndex = -1
				}
			}
		}*/
	}
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
		if token.atvRsAPC.Empty() && token.atvRsAPC.cders.Empty() && token.atvRsAPC.cntntrs.Empty() {
			token.atvRsAPC.cleanupActiveRSActivePassiveContent()
		}
		token.atvRsAPC = nil
	}
	if token.atvrs != nil {
		token.atvrs = nil
	}
	if token.curPrkdAtvRs != nil {
		token.curPrkdAtvRs.cleanupParkedActiveReaderSeeker()
		token.curPrkdAtvRs = nil
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
	return token.parsingActive()
}

func (token *activeParseToken) parsingActive() (parsed bool, err error) {
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
				if token.curStartIndex == -1 && token.parkedStartIndex == -1 {
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
			if token.curStartIndex == -1 && token.parkedStartIndex == -1 {
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

			if token.psvCapturedIO != nil && !token.psvCapturedIO.Empty() && token.psvCapturedIO.HasSuffix([]byte(lbls[1][1:])) {
				if token.psvCapturedIO.HasPrefixSuffix([]byte(lbls[0][0:len(lbls)-1]), []byte(lbls[1][1:])) {
					if valid, single, complexStart, complexEnd, elemName, elemPath, elemExt, valErr := validatePassiveCapturedIO(token, lbls); valid {
						if single || complexStart {
							if token.curStartIndex > -1 {
								if token.curEndIndex == -1 {
									token.curEndIndex = token.lastEndRIndex - token.psvCapturedIO.Size()
								}
								if token.tknrb, err = caprureCurrentPassiveStartEnd(token, token.atvrs, token.lastEndRIndex, token.rerr != nil && token.rerr == io.EOF, token.curStartIndex, token.curEndIndex, token.startRIndex); err != nil {
									return nextparse, err
								}
							}
							if token.parkedStartIndex == -1 {
								token.parkedStartIndex = token.lastEndRIndex
							}
						}
						if single || complexEnd {
							if token.parkedStartIndex > -1 {
								if token.parkedEndIndex == -1 {
									if single {
										token.parkedEndIndex = token.lastEndRIndex
									} else if complexEnd {
										token.parkedEndIndex = token.lastEndRIndex - token.psvCapturedIO.Size()
									}
								}
							}
							if elemName != "" && elemPath != "" && elemExt != "" {
								if !strings.HasPrefix(elemPath, "./") {
									if token.rsroot != "" {
										elemPath = token.rsroot + elemPath
									}
								}
								if single && elemName == (".:"+token.rsrootname) {
									token.parkedStartIndex = -1
									token.parkedEndIndex = -1
								} else {
									if strings.HasPrefix(elemPath, "./") {
										elemPath = elemPath[2:]
									}
									elemPath = token.rsroot + elemPath
									if err = token.parse.setRSByPath(elemPath); err == nil {
										atvrsapcsi := token.parkedStartIndex
										atvrsapcei := token.parkedEndIndex
										token.parkedStartIndex = -1
										token.parkedEndIndex = -1
										parseNextToken(token, token.parse, elemPath, atvrsapcsi, atvrsapcei)
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
						token.appendCde(token.atvRsAPC, token.atvRStartIndex, token.atvREndIndex)
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
				if token.tknrb, err = caprureCurrentPassiveStartEnd(token, token.atvrs, token.lastEndRIndex, token.rerr != nil && token.rerr == io.EOF, token.curStartIndex, token.curEndIndex, token.startRIndex); err != nil {
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
			token.tknrb, err = caprureCurrentPassiveStartEnd(token, token.atvrs, token.curEndIndex, true, token.curStartIndex, token.curEndIndex, token.startRIndex)
		}
		nextparse = true
		if token.atvRsAPC.atvparse.atvTkns[token] != nil {
			token = nil
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

func caprureCurrentPassiveStartEnd(token *activeParseToken, curatvrs *activeReadSeeker, lastEndRIndex int64, eof bool, curStartIndex, curEndIndex, startRIndex int64) (lasttknrb []byte, err error) {
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
func (atvpros *ActiveProcessor) Process(rs io.ReadSeeker, root string, path string, retrieveRS RetrieveRSFunc) (err error) {
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
		err = atvpros.atvParser.parse(rs, root, path, retrieveRS)
	}
	return err
}
