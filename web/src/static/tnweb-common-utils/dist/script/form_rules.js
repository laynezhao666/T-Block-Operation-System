
/**
 * 用于element的校验rule规则
 * @usage
 * common支持的类型和参数见 ./vdt.js
 * rules: {
 *    keyName1: [formRules.required(), formRules.common('email')]
 *    keyName2: [formRules.required('该项不能为空'), formRules.common('QQ')]
 *    keyName3: [formRules.common('tphone') ]
 *    keyName4: [formRules.common('minlength', 20) ]
 * }
 **/

import { VDT } from './vdt'

let validators = {}

for (let key in VDT) {
  let fn = VDT[key]
  let msg = VDT.messages[key]
  if (typeof fn === 'function' && msg) {
    validators[key] = function ({ rule, value, callback, params }) {
      if (fn(value, params)) {
        callback()
      } else {
        callback(new Error(msg))
      }
    }
  } else if (fn instanceof RegExp && msg) {
    let ffn = text => fn.test(text)
    validators[key] = function ({ rule, value, callback, params }) {
      if (ffn(value)) {
        callback()
      } else {
        callback(new Error(msg))
      }
    }
  }
}

export const REQUIRE_MSG = VDT.messages.required

// 这里只支持blur事件
// 部分组件只支持change事件，自己手写吧
export function required (txt) {
  return { required: true, message: txt || REQUIRE_MSG, trigger: 'blur' }
}

export function changeRequired (txt) {
  return { required: true, message: txt || VDT.messages.changeRequired, trigger: 'change' }
}

export function array (txt) {
  return { required: 'array', message: txt || VDT.messages.array, trigger: 'change' }
}

// 注意
// 为了保证颗粒度，common里的函数
// 有值才会做校验，无值认为是合法
// required条件请单独设置
export function common (type, params) {
  return {
    trigger: 'blur',
    validator (rule, value, callback) {
      // 注意：有值才校验！
      if (value && value.length) {
        validators[type]({ rule, value, callback, params })
      } else {
        callback()
      }
    },
  }
}
