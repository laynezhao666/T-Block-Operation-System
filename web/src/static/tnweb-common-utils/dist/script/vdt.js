export const VDT = {
  messages: {
    required: '必填',
    changeRequired: '必选',
    array: '至少选一项',
    word: '该字段仅支持英文、中文',
    text: '该字段支持英文、数字、中文',
    upperText: '该字段支持大写英文、数字、中文、-',
    codeName: '该字段支持英文、数字、-',
    aschar: '该字段仅支持英文、数字',
    remote: '请修正此字段',
    email: '请输入有效的电子邮件地址',
    url: '请输入有效的网址',
    date: '请输入有效的日期',
    dateISO: '请输入有效的日期 (YYYY-MM-DD)',
    number: '请输入有效的数字',
    digits: '只能输入数字',
    creditcard: '请输入有效的信用卡号码',
    equalTo: '你的输入不相同',
    extension: '请输入有效的后缀',
    minlength: '输入字数过短',
    maxlength: '输入字数过长',
    rangelength: '长度不符合要求',
    mphone: '请输入正确的手机号格式',
    tphone: '请输入正确的电话格式',
    anyPhone: '请输入正确的电话号码',
    fax: '请输入正确的传真号码',
    QQ: '请输入正确的QQ格式',
    ID: '请输入正确的身份证格式',
    postal: '请输入正确的邮编格式',
    upper: '只能输入大写字母',
    wechat: '请输入正确的微信号',
    multiIp: '请输入正确的IP',
    'cen-s-': '内容为中文，英文字母，数字或符号的组合，不允许全部为数字，不允许全部为符号',
    cens: '内容为中文，英文字母，数字或符号的组合',
    e: '内容为英文字母组合',
    E: '内容为大写英文字母组合',
    'e+n+': '内容全为英文字母，或者全为数字',
    'en-': '内容为英文字母，数字的组合，不允许全部为数字',
    'cen-': '内容为中文，英文字母，数字的组合，不允许全部为数字',
    'en-s-': '内容为英文字母，数字或符号的组合，不允许全部为数字，不允许全部为符号',
    'ens-': '内容为英文字母，数字或符号的组合，不允许全部为符号',
    'cens-': '内容为中文，英文字母，数字或符号的组合，不允许全部为符号',
    deviceno: '内容只能为1个英文字母或者2个数字',
    nospace: '内容不能包含空格'
  },
  required: function (value) {
    // eslint-disable-next-line eqeqeq
    return value != undefined ? (value.toString().length > 0) : false
  },

  array: function (value) {
    return value.length > 0
  },
  changeRequired: function (value) {
    return value.toString().length > 0
  },
  word (value) {
    return /^[A-Za-z\u4e00-\u9fa5]+$/.test(value)
  },
  text (value) {
    return /^[A-Za-z0-9\u4e00-\u9fa5]+$/.test(value)
  },
  upperText (value) {
    return /^[A-Z0-9-\u4e00-\u9fa5]+$/.test(value)
  },
  codeName (value) {
    return /^[A-Za-z0-9-]+$/.test(value)
  },
  aschar (value) {
    return /^[A-Za-z0-9]+$/.test(value)
  },
  email: function (value) {
    return /^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$/.test(value)
  },
  url: function (value) {
    return /^(?:(?:(?:https?|ftp):)?\/\/)(?:\S+(?::\S*)?@)?(?:(?!(?:10|127)(?:\.\d{1,3}){3})(?!(?:169\.254|192\.168)(?:\.\d{1,3}){2})(?!172\.(?:1[6-9]|2\d|3[0-1])(?:\.\d{1,3}){2})(?:[1-9]\d?|1\d\d|2[01]\d|22[0-3])(?:\.(?:1?\d{1,2}|2[0-4]\d|25[0-5])){2}(?:\.(?:[1-9]\d?|1\d\d|2[0-4]\d|25[0-4]))|(?:(?:[a-z\u00a1-\uffff0-9]-*)*[a-z\u00a1-\uffff0-9]+)(?:\.(?:[a-z\u00a1-\uffff0-9]-*)*[a-z\u00a1-\uffff0-9]+)*(?:\.(?:[a-z\u00a1-\uffff]{2,})).?)(?::\d{2,5})?(?:[/?#]\S*)?$/i.test(value)
  },
  date: function (value) {
    return !/Invalid|NaN/.test(new Date(value).toString())
  },
  dateISO: function (value) {
    return /^\d{4}[/-](0?[1-9]|1[012])[/-](0?[1-9]|[12][0-9]|3[01])$/.test(value)
  },
  number: function (value) {
    return /^(?:-?\d+|-?\d{1,3}(?:,\d{3})+)?(?:\.\d+)?$/.test(value)
  },
  digits: function (value) {
    return /^\d+$/.test(value)
  },
  isarr: function (o) {
    // eslint-disable-next-line eqeqeq
    return Object.prototype.toString.call(o) == '[object Array]'
  },
  minlength: function (value, param) {
    return value.length >= param
  },
  maxlength: function (value, param) {
    return value.length <= param
  },
  rangelength: function (value, param) {
    // getLength方法不知道最初是什么了，就先直接用length吧
    // var length = value instanceof Array ? value.length : this.getLength(value)
    const length = value.length
    return (length >= param[0] && length <= param[1])
  },
  min: function (value, param) {
    return value >= param
  },
  max: function (value, param) {
    return value <= param
  },
  range: function (value, param) {
    return (value >= param[0] && value <= param[1])
  },
  equalTo: function (value, param) {
    return value === param
  },
  mphone: function (value) {
    return /^1[3-9](\d{9})$/.test(value)
  },
  // 123都可以通过，可能有问题，建议换下面的telephone
  tphone: function (value) {
    return /^[+]{0,1}(\d){1,3}[ ]?([-]?((\d)|[ ]){1,12})+$/.test(value)
  },
  telephone: function (value) {
    return /^(([0+]\d{2,3}-)?(0\d{2,3})-)?(\d{7,8})(-(\d{3,}))?$/.test(value)
  },
  anyPhone: function (value) {
    return VDT.mphone(value) || VDT.telephone(value)
  },
  fax: function (value) {
    return this.tphone.test(value)
  },
  postal: function (value) {
    return /^[a-zA-Z0-9 ]{3,12}$/g.test(value)
  },
  vdata: function (value, config) { // 返回正确错误对象 提示 与结果
    for (var fun in config) {
      // eslint-disable-next-line eqeqeq
      if (typeof this[fun] === 'function' && (!(config[fun].param == undefined ? this[fun](value) : this[fun](value, config[fun].param)))) {
        if (typeof config[fun] === 'object') {
          return {
            msg: config[fun].msg ? config[fun].msg : this.messages[fun],
            result: false
          }
        } else {
          return {
            msg: typeof config[fun] === 'string' ? config[fun] : this.messages[fun],
            result: false
          }
        }
      } else if (typeof config[fun] === 'function') {
        var tmpr = config[fun](value)
        // eslint-disable-next-line eqeqeq
        if (tmpr != '' && tmpr != undefined && tmpr != false) {
          return {
            msg: tmpr,
            result: false
          }
        }
      }
    }
    return {
      msg: '',
      result: true
    }
  },
  QQ: function (qq) {
    return /^[1-9]\d{4,9}$/.test(qq)
  },
  ID: function (id) {
    return /(^\d{15}$)|(^\d{18}$)|(^\d{17}(\d|X|x)$)/.test(id)
  },
  upper (text) {
    return /^[A-Z]*$/.test(text)
  },
  wechat (text) {
    return /^[a-zA-Z]{1}[-_a-zA-Z0-9]{5,19}$/.test(text)
  },
  'cen-s-': /^(?!\d+$)(?![~!@#$%^&*()_\-+=<>?:"{}|,./;'\\[\]·~！@#￥%……&*（）——\-+={}|《》？：“”【】、；‘’，。、]+$)[A-Za-z\d\u4e00-\u9fa5~!@#$%^&*()_\-+=<>?:"{}|,./;'\\[\]·~！@#￥%……&*（）——\-+={}|《》？：“”【】、；‘’，。、]+$/,
  cens: /^[A-Za-z\d\u4e00-\u9fa5~!@#$%^&*()_\-+=<>?:"{}|,./;'\\[\]·~！@#￥%……&*（）——\-+={}|《》？：“”【】、；‘’，。、]+$/,
  e: /^[A-Za-z]+$/,
  E: /^[A-Z]+$/,
  'en-': /^(?!\d+$)[A-Za-z\d]+$/,
  'cen-': /^(?!\d+$)[A-Za-z\d\u4e00-\u9fa5]+$/,
  'en-s-': /^(?!\d+$)(?![~!@#$%^&*()_\-+=<>?:"{}|,./;'\\[\]·~！@#￥%……&*（）——\-+={}|《》？：“”【】、；‘’，。、]+$)[A-Za-z\d~!@#$%^&*()_\-+=<>?:"{}|,./;'\\[\]·~！@#￥%……&*（）——\-+={}|《》？：“”【】、；‘’，。、]+$/,
  'ens-': /^(?![~!@#$%^&*()_\-+=<>?:"{}|,./;'\\[\]·~！@#￥%……&*（）——\-+={}|《》？：“”【】、；‘’，。、]+$)[A-Za-z\d~!@#$%^&*()_\-+=<>?:"{}|,./;'\\[\]·~！@#￥%……&*（）——\-+={}|《》？：“”【】、；‘’，。、]+$/,
  'cens-': /^(?![~!@#$%^&*()_\-+=<>?:"{}|,./;'\\[\]·~！@#￥%……&*（）——\-+={}|《》？：“”【】、；‘’，。、]+$)[A-Za-z\d\u4e00-\u9fa5~!@#$%^&*()_\-+=<>?:"{}|,./;'\\[\]·~！@#￥%……&*（）——\-+={}|《》？：“”【】、；‘’，。、]+$/,
  'e+n+': /^((?=\d+$)|(?=[A-Za-z]+$))/,
  multiIp: /^((([1-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.)(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){2}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5]);)*(([1-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.)(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){2}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$/,
  deviceno: /^((?:[a-zA-Z])|(?:\d{2}))$/,
  nospace (text) {
    return text.indexOf(' ') === -1
  }
}
