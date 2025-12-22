/**
 * 封装postmate，方便调用方绑定事件回调函数
 * https://github.com/dollarshaveclub/postmate
 * @example
 * handlers: {
 *
 * 1.一次性连接
 *    // 不返回值
 *    'eventName': function(childData) {
 *      // 不返回
 *    }
 *    // 直接返回值
 *    'eventName': function(childData) {
 *      return childData
 *    },
 *    // 返回promise
 *    'eventName': function(childData) {
 *      return new Promise((resolve, reject) => {
 *        if (success) {
 *          resolve(backData)
 *        } else {
 *          // 注意，报错需要给code
 *          reject({code: -1, message: '', errorData})
 *        }
 *      })
 *    }
 *    // 自定义返回（注意：只能调用一次！！！）
 *    'eventName': function(childData, callbackId) {
 *      setTimeout(() => {
 *        // 自定义成功回调只需要data
 *        this.$refs.bpmFrame.dispatchSuccess(callbackId, data)
 *        // 注意，完整自定义回调需要给非0code, message可选
 *        this.$refs.bpmFrame.dispatchCallback(callbackId, {code, data})
 *      })
 *    }
 *
 * 2. 长效连接(bpm内部需要绑定connect通信方式)
 *  // 该函数可能被bpm多次调用，可不断接收到新的数据
 *    eventName (childData) {
 *      // 接收信息之后可以多次发送消息
 *      setInterval(() => {
 *        that.$refs.bpmFrame.callback({
 *          event: 'eventName',
 *          result: 'result',
 *        });
 *      }, 1000);
 *    }
 * }
 */
import Postmate from 'postmate';
import commonEvents from './common-events';
import { each, remove, get, isNil } from 'lodash';
import Cookies from 'js-cookie';
// Postmate.debug = true;

export default {
  data() {
    return {
      handshakeLoaded: false,
      handshake: null,
      childId: 0,
      callbackIds: [],
      handlerCache: {},
    };
  },
  beforeDestroy() {
    if (this.handshake) {
      this.handshake.then((child) => {
        child.destroy();
      });
    }
  },
  methods: {
    initHandshake(options) {
      if (this.handshake) {
        this.handshake.then((child) => {
          child.destroy();
        });
      }
      this.handshakeLoaded = false;

      return new Promise((resolve, reject) => {
        this.handshake = new Postmate({
          ...options,
        });

        this.handshake.then((child) => {
          child.call('changestyle');

          window.addEventListener('emitchangeidc', () => {
            child.call('changeidc', { mozuInfo: TNBL.getCurrModule() });
          });

          child.get('_pageId').then((pageId) => {
            this.handshakeLoaded = true;
            this.childId = pageId;
            each(commonEvents, (callback, event) => {
              this.addHandler(event, callback);
            });
            if (Object.keys(this.handlerCache).length) {
              each(this.handlerCache, (callback, event) => {
                this.addHandler(event, callback);
              });
              this.handlerCache = {};
            }
            child.call('_ready', { mozuId: Cookies.get('tnebula_cu_moduleid') });
            resolve(child);
          });
          child.get('height').then((height) => {
            this.setStyle({ height });
          });
        }).catch((e) => {
          reject(e);
        });
      });
    },
    addHandler(event, handler) {
      // if (event === 'test') { debugger; }
      if (!this.handshakeLoaded) {
        this.handlerCache[event] = handler;
        return;
      }
      this.handshake.then((child) => {
        child.on(event, ({ callbackId, data, pageId }) => {
          if (this.childId !== pageId) {
            console.log('childId不匹配', this.childId, pageId);
            return;
          }

          if (event && event[0] !== '_') {
            console.log('接收到iframe的事件', event, data, pageId);
          }

          this.callbackIds.push(callbackId);

          const rt = handler.call(this, data, callbackId);

          if (rt === undefined || !callbackId) return;

          if (rt instanceof Promise) {
            rt.then((res) => {
              this.dispatchCallback(callbackId, { code: 0, data: res });
            }).catch((res) => {
              if (isNil(get(res, 'code'))) {
                // 在登录失败的时候会回调false
                // 也一并处理其他没有code的情况
                this.dispatchCallback(callbackId, {
                  code: -1,
                  message: '发生错误了',
                  data: res,
                });
              } else {
                this.dispatchCallback(callbackId, res);
              }
            });
          } else {
            this.dispatchCallback(callbackId, { code: 0, data: rt });
          }
        });
      });
    },
    // 一次性的成功返回，只需填写data
    dispatchSuccess(callbackId, data) {
      if (callbackId) {
        this.dispatchCallback(callbackId, {
          code: 0,
          data,
        });
      }
    },
    // 一次性的标准返回，需填写完整返回对象
    dispatchCallback(callbackId, result) {
      if (isNil(result) || isNil(get(result, 'code'))) {
        console.warn('dispatchCallback回调必须带code', callbackId, result);
      }
      if (this.callbackIds.indexOf(callbackId) === -1) {
        console.error('回调id错误，id只能使用一次', callbackId, result);
        return;
      }
      this.callback({
        callbackId,
        result,
      });
      remove(this.callbackIds, id => id === callbackId);
    },
    // 任意返回
    callback(obj) {
      if (!obj.event && !obj.callbackId) {
        console.error('回调格式错误，必须含event或者callbackId', obj);
      }
      if (obj.result === undefined) {
        console.warn('回调没有提供接收方处理的result数据', obj);
      }

      this.handshake.then((child) => {
        child.call('callback', obj);
      });
    },
  },
};
