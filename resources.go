package lnksworks

import (
	"io"
	"sync"
)

type EmbededResource struct {
	contentRS *IORW
}

var embededRes map[string]*IORW
var embededResLock *sync.RWMutex

func init() {
	embededResLock = &sync.RWMutex{}
	if embededRes == nil {
		embededRes = make(map[string]*IORW)
	}
}

func EmbededResourceByPath(respath string) (resio *IORW) {
	embededResLock.RLock()
	defer embededResLock.RUnlock()

	if emres, embok := embededRes[respath]; embok {
		resio = emres
	}

	return resio
}

func RegisterEmbededResources(resources ...interface{}) {
	if len(resources) > 0 && len(resources)%2 == 0 {
		resi := 0
		embededResLock.RLock()
		defer embededResLock.RUnlock()
		for resi < len(resources) {
			if respath, respathok := resources[resi].(string); respathok {

				if resstream, resstreamok := resources[resi+1].(io.Reader); resstreamok {
					if _, resok := embededRes[respath]; resok {
						embededRes[respath].Close()
						embededRes[respath] = nil
						delete(embededRes, respath)
					}
					iores, _ := NewIORW()
					iores.Print(resstream)
					embededRes[respath] = iores
				}
			}
			resi = resi + 2
		}
	}
}
