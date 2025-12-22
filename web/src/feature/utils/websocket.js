
function processData(data, processFunc, host) {
  if (data.cmd === 'pong' || data.cmd === 'ping') {
  } else if (data.code === 0) {
    processFunc.call(host, data);
  } else if (data.code === -999) {
    host.$message.error('登录超时，请重新登录');
  } else {
    host.$message.error('服务返回异常，请稍候重试或联系管理员');
  }
}

// websocket管理
// 创建连接，接收数据，保持状态（状态同步）
// todo,发送数据包装/接收数据reqid使用

// 使用：
// 1，创建与接收数据：提供data.webSocketConfigs配置信息
// 1）接收数据处理函数（必选），参数为服务器端推送的数据
// 2）状态变化处理函数（可选），参数为true或false，表示socket是否在线
// 3）连接后处理函数（可选），多用于初始化，向后台发送命令
// 2，发送数据：
// 1）调用const ws = this.webSocketInstances[url].ins来获取相应的ws对象
// 2）判断ws.readyState === WebSocket.OPEN，
// 3）ws.send(JSON.stringify(data))发送
// 3，手动重连 this.retrySockets(url, true)
export default {
  data() {
    return {
      maxRetryTimes: 30000, // 最大重连次数
      heartBeatTime: 10 * 1000, // 后台心跳间隔
      // webSocketConfigs: {
      //   $url: {// websocket url
      //     dataProcess: () => {}, // 服务器推送的数据处理，参数为data（json格式）
      //     statusProcess: () => {}, // websocket状态处理，参数为true或false，可用于显示在线或离线
      //     onConnected: () => {},//连接后执行的方法，重新连接也会执行
      //   },
      // },
      detecteTime: null,
      connectStatus: false,
    };
  },
  created() {
    if (this.webSocketConfigs && !window.tnwebServices?.tboxObsService?.enable) {
      console.log('socket initing.');
      const webSocketInstances = {};
      Object.keys(this.webSocketConfigs).forEach((url) => {
        const ins = this.createSocket(url);
        webSocketInstances[url] = {
          lastActive: undefined,
          ins,
          retriedTimes: 0,
        };
      });
      this.webSocketInstances = webSocketInstances;

      this.intervalTimeId = window.setInterval(() => {
        Object.keys(webSocketInstances).forEach((url) => {
          const wi = webSocketInstances[url];
          const dtNow = new Date();
          if ((dtNow - wi.lastActive) > this.heartBeatTime && wi.ins.readyState === WebSocket.OPEN) {
            const data = {
              cmd: 'ping',
              reqid: 0,
              timsstamp: new Date() / 1000,
            };
            console.log(`websocket send heartbeat:${url}`);
            wi.ins.send(JSON.stringify(data));
          }
        });
      }, 3000);
    }
  },
  methods: {
    createSocket(url) {
      const conf = this.webSocketConfigs[url];
      const ins = new WebSocket(url);
      ins.addEventListener('open', () => {
        console.log(`websocket connected:${url}`);
        const wi = this.webSocketInstances[url];
        wi.retriedTimes = 0;
        wi.lastActive = new Date();
        if (conf.onConnected) {
          conf.onConnected.call(this, url);
        }
        // if (conf.statusProcess) {
        //   conf.statusProcess.call(this, true);
        // }
      });
      ins.addEventListener('message', (event) => {
        if (ins.readyState === WebSocket.OPEN && conf.statusProcess && !this.connectStatus) {
          this.connectStatus = true;
          conf.statusProcess.call(this, this.connectStatus);
        }
        const { data } = event;
        this.webSocketInstances[url].lastActive = new Date();
        if (window.____DEBUG) {
          console.log(`websocket recieved:${event.data},${url}`);
        }
        let json;
        try {
          json = JSON.parse(data);
        } catch (ex) {
          console.log(`websocket recieved data format error:${ex},${url},${event.data}`);
          return;
        }

        if (window.location.origin.includes('ha')) {
          clearTimeout(this.detecteTime);
          this.detecteTime = setTimeout(() => {
            ins.close();
            this.createSocket(url);
          }, (json.heartbeat || 3) * 1000);
        }

        processData(json, conf.dataProcess, this);
      });
      ins.addEventListener('close', (event) => {
        console.log(`websocket closed:${event.message},${url}`);
        if (conf.statusProcess) {
          this.connectStatus = false;
          conf.statusProcess.call(this, this.connectStatus);
        }
        // 重试
        this.retrySockets(url);
      });
      ins.addEventListener('error', (event) => {
        console.log(`websocket error:${event.message},${url}`);
        // 连接重试
        // this.retrySockets(url);
      });
      return ins;
    },
    retrySockets(url, isForce) {
      const wi = this.webSocketInstances[url];
      if (wi.ins.readyState !== WebSocket.CLOSED) {
        console.log(`retrySockets failed:ins is not in closed.readyState:${wi.ins.readyState},${url}`);
      } else if (!isForce && wi.retriedTimes >= this.maxRetryTimes) {
        console.log(`retrySockets failed,retriedTimes:${wi.retriedTimes},${url}`);
        const conf = this.webSocketConfigs[url];
        if (conf.statusProcess) {
          this.connectStatus = false;
          conf.statusProcess.call(this, this.connectStatus);
        }
      } else {
        console.log(`retrySockets,isForce:${isForce},retriedTimes:${wi.retriedTimes},${url}`);
        if (isForce) {
          wi.retriedTimes = 0;
        } else {
          wi.retriedTimes = wi.retriedTimes + 1;
        }
        setTimeout(() => {
          const ins = this.createSocket(url);
          wi.ins = ins;
        }, 1000);
      }
    },
  },
  mounted() {
  },
  beforeDestroy() {
    clearInterval(this.intervalTimeId);
    const { webSocketInstances } = this;
    Object.keys(webSocketInstances).forEach((url) => {
      const wi = webSocketInstances[url];
      wi.ins.close();
    });
  },
};
