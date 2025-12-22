package device

import (
	"fmt"
	"agent/utils/flog"
	"agent/utils/osal"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"trpc.group/trpc-go/trpc-go/log"
)

var (
	filterLog *flog.Filter
)

const (
	RecvPrefix = "recv:"
	SendPrefix = "send:"
)

// Infof log
func (d *Device) Infof(format string, args ...interface{}) {
	log.Infof(fmt.Sprintf("[device %v, mark: %v]: ", d.ID(), d.randomMark)+format, args...)
}

// Warn log
func (d *Device) Warn(args ...interface{}) {
	log.Warn(fmt.Sprintf("[device %v, mark: %v]: ", d.ID(), d.randomMark) + fmt.Sprint(args...))
}

// Warnf log
func (d *Device) Warnf(format string, args ...interface{}) {
	log.Warnf(fmt.Sprintf("[device %v, mark: %v]: ", d.ID(), d.randomMark)+format, args...)
}

// Errorf log
func (d *Device) Errorf(format string, args ...interface{}) {
	log.Errorf(fmt.Sprintf("[device %v, mark: %v]: ", d.ID(), d.randomMark)+format, args...)
}

// channel log
const kMaxLogFileSize = 1024 * 1024 * 5 // 5MB

// LogPacket log
func (d *Device) LogPacket(prefix string, data string) {
	_, needLog := osal.Instance().GetWithID(osal.PacketLogKey, d.Info.ChannelID)
	if needLog {
		d.LogRaw([]byte(prefix), []byte(data))
	}
}

// LogRaw log
func (d *Device) LogRaw(prefix []byte, buf []byte) {
	d.logMux.Lock()
	defer d.logMux.Unlock()

	if !d.openLog() {
		return
	}

	d.logTimestamp()
	d.logImpl(prefix)
	d.logImpl(buf)
	d.logNewLine()
}

func (d *Device) openLog() bool {
	if d.logFileSize > kMaxLogFileSize {
		if d.logFile != nil {
			_ = d.logFile.Close()
			d.logFile = nil
		}

		err := os.Remove(d.logFileName)
		if err != nil && !os.IsNotExist(err) {
			log.Errorf("delete log file error: %v\n", err)
			return false
		}
		d.logFileSize = 0
	}

	if d.logFile == nil {
		// 目录创建
		dir := filepath.Dir(d.logFileName)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Errorf("create directory %s error: %v\n", dir, err)
			return false
		}

		f, err := os.OpenFile(d.logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Errorf("open log file %s error: %v\n", d.logFileName, err)
			return false
		}

		info, err := f.Stat()
		if err != nil {
			f.Close()
			log.Errorf("get file info error: %v\n", err)
			return false
		}

		d.logFile = f
		d.logFileSize = info.Size()
	}

	return true
}

func (d *Device) logImpl(data []byte) {
	n, err := d.logFile.Write(data)
	d.logFileSize += int64(n)

	if err != nil || n != len(data) {
		log.Errorf("write log failure - expected:%d actual:%d error:%v",
			len(data), n, err)
	}

	if rand.Intn(100) < 30 { // 30% 概率刷盘
		syncErr := d.logFile.Sync()
		if syncErr != nil {
			log.Warnf("sync log file error: %v", syncErr)
		}
	}
}

func (d *Device) logTimestamp() {
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05Z07:00") + ": "
	d.logImpl([]byte(timestamp))
}

func (d *Device) logNewLine() {
	d.logImpl([]byte{'\n'})
}

// LogClose 关闭文件
func (d *Device) logClose() {
	d.logMux.Lock()
	defer d.logMux.Unlock()

	if d.logFile != nil {
		_ = d.logFile.Close()
		d.logFile = nil
	}
}
