/* eslint-disable no-underscore-dangle */
/**
 * Socket: websocket通信 - 前端组件
 * @author 
 * 【注意】需要求后端使用以下通信规则方可使用本组件
 * 1. 通信规则：
 * cmd：区分操作命令
 * reqId：对请求进行唯一标识，用于回复
 * 请求/推送：{cmd, reqId, data, timestamp}
 * 回复：{reqId, cmd, code, data, timestamp}
 * 心跳：后端自行维护，不需要前端发送
 * 超时时间：30s
 * 连接重试时间：1s/次 * 100次
 * 注意：暂未提供针对后端单个请求的回复功能
 * @usage
 * 1. 初始化并绑定对cmd的处理函数
 * const socket = new Socket({
 *  url: 'ws://localhost:8080/ws',
 *  listeners: {
 *    _status(statusOk) {
 *      // 连接成功后的回调
 *    },
 *    _error(error) {
 *      // 发生错误的回调
 *    },
 *    cmdName({data, timestamp}) {
 *      // do something
 *    }
 *    // ...more cmd callbacks
 *  }
 * });
 * 2. 发送请求并接收回复
 * socket.send(cmdName, data).then(({data, timestamp}) => {
 *  // do something
 * }).catch(({code, data, message, timestamp}) => {
 *  // do something when error occurred
 * });
 * 3. 断开连接
 * socket.close();
 */
import { each } from 'lodash';
import { redirectToLogin } from 'common/script/passport_login';
class Socket {
  #RETRY_INTERVAL = 1000
  #TIME_OUT = 1000 * 30
  // #MAX_CALLBACK_LENGTH = 1000
  #MAX_RETRY_TIMES = 100
  #NOT_LOGIN_CODE = 4999
  #url = ''
  #socket
  #state = false
  #retryTimes = 0
  #timestampOffset = 0
  // 按cmd的监听
  #listeners = {}
  // 回调参数缓存
  #callbackMap = {}
  // 回调reqId缓存
  #reqIdList = new Set()
  // 连接建立前的请求缓存
  #messageQueue = [];

  #manuallyClose = false

  #interval = 0

  constructor({
    url,
    listeners,
  }) {
    this.#url = this.#buildUrl(url);

    this.#listeners = listeners;

    if (!this.#listeners._status) {
      console.warn('没有绑定_status监听');
      this.#listeners._status = () => {};
    }

    if (!this.#listeners._error) {
      console.warn('没有绑定error监听');
      this.#listeners._error = () => {};
    }

    this.#createConnection();
  }
  #buildUrl(url) {
    const {
      protocol,
      hostname,
    } = window.location;
    const p = protocol === 'https:' ? 'wss:' : 'ws:';
    const u = new URL(url, `${p}//${hostname}`);
    u.protocol = p;
    return u.href;
  }
  #createConnection() {
    if (this.#manuallyClose) return;

    if (this.#socket) {
      if (this.#state) return;
      this.#retryTimes += 1;
      if (this.#retryTimes >= this.#MAX_RETRY_TIMES) {
        this.close();
        const error = new Error(`connection try times more than ${this.#MAX_RETRY_TIMES}, fail`);
        this.#listeners._error(error);
        console.error(error);
        return;
      }
      console.warn(`Retry connection for ${this.#retryTimes} times`, this.#manuallyClose);
    }

    const socket = new WebSocket(this.#url);

    socket.addEventListener('open', () => {
      this.#state = true;
      this.#retryTimes = 0;
      this.#listeners._status(true);
      if (this.#messageQueue.length) {
        each(this.#messageQueue, (msg) => {
          this.#send(msg);
        });
        this.#messageQueue = [];
      }
    });

    socket.addEventListener('message', (event) => {
      this.#handle(event.data);
    });

    socket.addEventListener('error', (e) => {
      console.log('websocket error', e);
      this.#listeners._error(new Error('Websocket网络连接发生错误'));
    });

    socket.addEventListener('close', (e) => {
      console.error('socket.close', e);
      const { code, reason } = e;
      if (code === this.#NOT_LOGIN_CODE) {
        this.#listeners._error(new Error(`连接已断开：${reason || '未知原因'}`));
        redirectToLogin();
        return;
      }
      this.#state = false;
      this.#listeners._status(false);
      if (this.#manuallyClose) return;
      setTimeout(() => {
        this.#createConnection();
      }, this.#RETRY_INTERVAL);
    });

    this.#socket = socket;

    this.#addInterval();
  }
  #generateId() {
    return `${Date.now()}${Math.floor(Math.random() * 1000)}`;
  }
  // 前端：ID的时间
  #getTimestampById(id) {
    return (`${id}`).slice(0, 13) * 1;
  }
  // 后端时间戳
  #getTimestamp() {
    return Date.now() + this.#timestampOffset;
  }
  #addCallback({ msg, resolve, reject }) {
    const { reqId, cmd } = msg;
    this.#callbackMap[reqId] = {
      cmd,
      resolve,
      reject,
    };
    this.#reqIdList.add(reqId);
    // this.#checkCallbackCount();
  }
  #addInterval() {
    if (this.#interval) {
      clearInterval(this.#interval);
    }
    this.#interval = setInterval(() => {
      const now = Date.now();
      each(Array.from(this.#reqIdList), (reqId) => {
        const ts = this.#getTimestampById(reqId);
        if (now - ts >= this.#TIME_OUT) {
          console.log('socket,clear timeout req', reqId);
          this.#callbackMap[reqId].reject({
            code: -1,
            message: 'timeout',
          });
          this.#clearCallback(reqId);
        }
      });
    }, 1000);
  }
  // // 去掉超时的请求
  // #checkCallbackCount() {
  //   if (Object.keys(this.#callbackMap).length > this.#MAX_CALLBACK_LENGTH) {
  //     const now = Date.now();
  //     this.#callbackMap = filter(this.#callbackMap, ({ reject }, reqId) => {
  //       const ts = this.#getTimestampById(reqId);
  //       return now - ts < this.#TIME_OUT;
  //     });
  //   }
  // }
  #buildMessage(cmd, data) {
    const reqId = this.#generateId() * 1;
    return {
      reqId,
      cmd,
      timestamp: this.#getTimestamp(),
      data,
    };
  }
  #clearCallback(reqId) {
    delete this.#callbackMap[reqId];
    this.#reqIdList.delete(reqId);
  }
  #parseMessage(str) {
    try {
      const msg = JSON.parse(str);
      return msg;
    } catch (e) {
      console.error('Parse message failed', str);
      return {};
    }
  }
  #handle(msg) {
    const {
      code,
      cmd,
      message,
      timestamp,
      reqId,
      data,
    } = this.#parseMessage(msg);

    if (timestamp) {
      this.#timestampOffset = timestamp - Date.now();
    }

    // 优先按reqId处理
    if (reqId) {
      const callbackParams = this.#callbackMap[reqId];

      if (callbackParams) {
        if (callbackParams.cmd !== cmd) {
          console.warn(`reqId of cmd ${callbackParams.cmd} is not equal to response cmd ${cmd}`);
        };

        const { resolve, reject } = callbackParams;

        if (code === 0) {
          resolve({
            data,
            timestamp,
          });
        } else {
          reject({
            code,
            message,
            data,
          });
        }
        this.#clearCallback(reqId);
        return;
      }
    }

    if (cmd) {
      const callback = this.#listeners[cmd];

      if (!callback) return;

      // eslint-disable-next-line standard/no-callback-literal
      callback(code === 0 ? {
        timestamp,
        data,
      } : {
        code,
        timestamp,
        message,
        data,
      });
    }
  }
  #send(msg) {
    this.#socket.send(JSON.stringify(msg));
  }
  send(event, data) {
    return new Promise((resolve, reject) => {
      const msg = this.#buildMessage(event, data);
      this.#addCallback({ msg, resolve, reject });
      if (this.#state) {
        this.#send(msg);
      } else {
        this.#messageQueue.push(msg);
      }
    });
  }
  close() {
    this.#manuallyClose = true;
    this.#socket.close();
  }
}

export default Socket;
