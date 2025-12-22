package http

import (
	"context"
	"etrpc-go/util/httputil"
	"fmt"
	urllib "net/url"
	"sync"
	"time"

	"trpc.group/trpc-go/trpc-go/client"

	"agent/entity/config"
	"agent/entity/model"
	"agent/logic/distribution/distributor"
	"agent/utils/osal"

	"trpc.group/trpc-go/trpc-go/log"
)

const (
	httpScheme         string = "http"
	heartbeatFilterKey string = "client heartbeat"
)

var defaultClientsWhitelist = osal.NewSet[string](
	"p1",
	"p2",
	"p3",
)

type httpClient struct {
	ctx    context.Context
	cancel context.CancelFunc
	config distributor.ClientConfig
	wg     sync.WaitGroup
	//heartbeatTarget  string
	heartbeatTargets []string
}

// GetAllClientsConfig 获取所有客户端配置
func (c *httpDistributor) GetAllClientsConfig() []*distributor.ClientConfig {
	configs := make([]*distributor.ClientConfig, 0)
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	for _, cli := range c.clients {
		configs = append(configs, &cli.config)
	}
	return configs
}

// AddClient 添加客户端
func (c *httpDistributor) AddClient(clientType, name, target, clientId string) error {
	err := validateUrl(target)
	if err != nil {
		return fmt.Errorf("invalid url [%v]: %v. valid url example:[http://127.0.0.1/api/data]", target, err)
	}
	err = c.validateName(name)
	if err != nil {
		return err
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if _, ok := c.clients[name]; ok {
		c.RemoveClient(name)
	}
	ctx, cancel := context.WithCancel(context.Background())
	client := &httpClient{
		ctx:    ctx,
		cancel: cancel,
		config: distributor.ClientConfig{
			Name:              name,
			Target:            target,
			Type:              clientType,
			ClientId:          clientId,
			HeartbeatInterval: 0,
		},
		wg: sync.WaitGroup{},
		//heartbeatTarget: getHeartbeatTarget(target),
		// 添加多个心跳目标
		heartbeatTargets: getHeartbeatTarget(target),
	}
	c.clients[name] = client
	client.start()
	return nil
}

// RemoveClient 移除客户端
func (c *httpDistributor) RemoveClient(name string) {
	if client, ok := c.clients[name]; ok {
		client.stop()
		delete(c.clients, name)
	}
}

func (h *httpClient) start() {
	log.Infof("start http client [%v] report target [%v] heartbeat target[%v]...",
		h.config.Name, h.config.Target, h.heartbeatTargets)
	h.wg.Add(1)
	go h.Heartbeat(h.ctx, distributor.DefaultHeartbeatInterval)
}

func (h *httpClient) stop() {
	log.Infof("http client [%v] stop now...", h.config.Name)
	h.cancel()
	h.wg.Wait()

	log.Infof("http client [%v] stop done", h.config.Name)
}

// Heartbeat 客户端心跳上报
func (h *httpClient) Heartbeat(ctx context.Context, interval time.Duration) {
	defer h.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
			hearbeatMessage := &model.HeartbeatMessage{
				Box: model.BoxInfo{
					Ip:       config.LocalIP,
					ClientId: h.config.ClientId,
				},
				Timestamp: time.Now().Unix(),
			}
			var wg sync.WaitGroup
			for _, target := range h.heartbeatTargets {
				wg.Add(1)
				go func(t string) {
					defer wg.Done()
					err := httputil.PostJson(context.Background(), t, nil, hearbeatMessage, nil,
						client.WithTLS(
							"",     // 证书路径, 留空
							"",     // 私钥路径, 留空
							"none", // CA 证书, 两边都不认证
							"",     // server name, 留空
						))
					if err != nil {
						filterLog.Errorf(heartbeatFilterKey+h.config.Name,
							"http client <%v> send heartbeat to %s error: %v", h.config.Name, t, err)
					}
				}(target)
			}
			wg.Wait()
		}
	}
}

func getHeartbeatTarget(target string) []string {
	url, _ := urllib.Parse(target)
	//return fmt.Sprintf("%s://%s%s/heartbeat", url.Scheme, url.Host, url.Path)
	return []string{
		fmt.Sprintf("%s://%s%s/heartbeat", url.Scheme, url.Host, url.Path),
		fmt.Sprintf("%s://%s/heartbeat", url.Scheme, url.Host),
	}
}

func validateUrl(url string) error {
	if url == "" {
		return fmt.Errorf("url is empty")
	}
	parsed, err := urllib.Parse(url)
	if err != nil {
		return err
	}
	// 接受http/https
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("scheme must be http or https")
	}
	if parsed.Host == "" {
		return fmt.Errorf("url host is empty")
	}
	return nil
}

func (c *httpDistributor) validateName(name string) error {
	if defaultClientsWhitelist.Contains(name) {
		return nil
	}
	if c.whitelist.Contains(name) {
		return nil
	}
	return fmt.Errorf("client name <%v> is not invalid", name)
}
