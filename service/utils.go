package service

import (
	"os"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/mengbin92/browser/utils"
)

type pbCache struct {
	cache sync.Map
	time  *time.Ticker
}

func (p *pbCache) checkExpiredTokenTimer() {
	for {
		<-p.time.C
		p.cache.Range(func(key, value interface{}) bool {
			pf := value.(*pbFile)
			if pf.isExpired() {
				p.cache.Delete(key)
				os.Remove(utils.Fullname(pf.Name))
			}
			return true
		})
	}
}

type pbFile struct {
	Name    string
	Expired int64
}

func (p *pbFile) isExpired() bool {
	return p.Expired < time.Now().Unix()
}

func (p *pbFile) renewal() {
	p.Expired = time.Now().Add(300 * time.Second).Unix()
}

func (p *pbFile) Marshal() ([]byte, error) {
	return sonic.Marshal(p)
}

func (p *pbFile) Unmarshal(data []byte) error {
	return sonic.Unmarshal(data, p)
}
