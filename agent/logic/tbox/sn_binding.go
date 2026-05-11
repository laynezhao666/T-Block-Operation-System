package tbox

import (
	"agent/entity/config"
	"agent/repo/cm"
	"context"
	"fmt"
	"time"

	"git.woa.com/idc/etrpc-go/etrpc-go/util/httputil"
	collectorPb "git.woa.com/idc/trpcprotocol/tbos/idc-tbos-collector"
	"trpc.group/trpc-go/trpc-go/log"
)

func (t *TboxManager) snBindingLoop(ctx context.Context) {
	_, errSupportSn := cm.NewSnReader(config.GetRB().Project.Source)
	if errSupportSn != nil {
		log.Infof("snBindingLoop exit, source not support sn reader: %v", config.GetRB().Project.Source)
		// 只有支持SnReader的Source才需要上报
		return
	}
	for {
		if config.GetRB().Tbox.SnReportEnabled {
			t.reportSnBinding(ctx)
		}
		select {
		case <-ctx.Done():
			log.Info("stop heartbeatLoop.")
			return
		case <-time.After(t.snInterval):
			break
		}
	}
}

type RequestSnBinding struct {
	Sn string `json:"sn"`
	Ip string `json:"ip"`
}

func (t *TboxManager) reportSnBinding(ctx context.Context) {
	ip := t.GetIP()
	sn := t.GetSN()
	if ip == "" || sn == "" {
		log.Warnf("ip or sn invalid,ip:%s,sn:%s", ip, sn)
		return
	}
	cli := collectorPb.NewConfigBusClientProxy()
	_, err := cli.ReportSn(ctx, &collectorPb.ReqReportSn{
		Sn: sn,
		Ip: ip,
	})
	if err == nil {
		log.Debugf("send sn binding ok, sn:%s, ip:%s", sn, ip)
		return
	} else {
		log.Errorf("send sn binding failed, sn:%s, ip:%s, err: %v", sn, ip, err)
	}
	if len(t.elvdbTarget) == 0 || len(t.elvdbSnUrl) == 0 {
		log.Debugf("elvdbTarget or elvdbSnUrl invalid, elvdbTarget:%s, elvdbSnUrl:%s",
			t.elvdbTarget, t.elvdbSnUrl)
		return
	}
	req := &RequestSnBinding{
		Sn: sn,
		Ip: ip,
	}
	targetUrls := []string{fmt.Sprintf("http://%s%s", t.elvdbTarget, t.elvdbSnUrl)}
	var lastErr error
	for _, targetUrl := range targetUrls {
		errSend := httputil.PostJson(context.Background(), targetUrl,
			nil, req, nil)
		log.Debugf("send sn binding req:%+v,targetUrl:%s", req, targetUrl)
		if errSend == nil {
			log.Debugf("send sn binding ok, sn:%s, ip:%s, target:%s", req.Sn, req.Ip, targetUrl)
			return
		}
		log.Warnf("send sn failed to %s, err: %v", targetUrl, errSend)
		lastErr = errSend
	}
	if lastErr != nil {
		log.Errorf("send sn binding failed to all targets, last error: %v", lastErr)
	}
}
