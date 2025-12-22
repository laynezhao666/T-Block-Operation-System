package service

import (
	"context"
	"fmt"
	"agent/entity/config"
	"agent/entity/errcode"
	"agent/logic/setup"
	"agent/utils/file/io"
	"os"
	"os/exec"
	"sync"
	"time"

	pb "trpcprotocol/agent"

	"google.golang.org/protobuf/types/known/emptypb"
	"trpc.group/trpc-go/trpc-go/errs"
	"trpc.group/trpc-go/trpc-go/log"
)

const (
	SyncTimeConf = "conf/time.json"
)

// BoxManager box管理
type BoxManager struct {
	ntpMutex sync.Mutex
}

// SetNtp 设置NTP
func (b BoxManager) SetNtp(ctx context.Context, req *pb.SetNtpReq) (*emptypb.Empty, error) {
	if config.GetRB().IsGatewayMode() {
		return &emptypb.Empty{}, errs.New(errcode.ErrGwMode, "agent-gw mode not support")
	}

	b.ntpMutex.Lock()
	defer b.ntpMutex.Unlock()
	// 配置信息写入json即可，shell脚本负责同步系统时间
	if err := io.JSON.Write(SyncTimeConf, req); err != nil {
		return &emptypb.Empty{}, errs.New(errcode.ErrNorthTBoxNtp, fmt.Sprintf("write time json error: %v", err))
	}

	return &emptypb.Empty{}, nil
}

// SetRealTime 设置系统时间
func (b BoxManager) SetRealTime(ctx context.Context, req *pb.SetRealTimeReq) (*emptypb.Empty, error) {
	if config.GetRB().IsGatewayMode() {
		return &emptypb.Empty{}, errs.New(errcode.ErrGwMode, "agent-gw mode not support")
	}

	cmd := exec.Command("date", "-s", fmt.Sprintf("@%v", req.Time))
	err := cmd.Run()
	if err != nil {
		return &emptypb.Empty{}, errs.New(errcode.ErrNorthTBoxTime, err.Error())
	}
	cmd = exec.Command("hwclock", "-w")
	return &emptypb.Empty{}, cmd.Run()
}

// OSRestart 重启系统
func (b BoxManager) OSRestart(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	if config.GetRB().IsGatewayMode() {
		return &emptypb.Empty{}, errs.New(errcode.ErrGwMode, "agent-gw mode not support")
	}
	log.Warn("OS Reboot")
	cmd := exec.Command("reboot")
	return &emptypb.Empty{}, cmd.Run()
}

// AgentRestart 重启agent
func (b BoxManager) AgentRestart(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	if config.GetRB().IsGatewayMode() {
		return &emptypb.Empty{}, errs.New(errcode.ErrGwMode, "agent-gw mode not support")
	}
	go func() {
		time.Sleep(time.Second * 3)
		log.Warn("Agent Restart")
		setup.UnInit()
		os.Exit(0)
	}()
	return &emptypb.Empty{}, nil
}
