/**
 * common.js
 *
 * 通用公共js
 *
 *
 */

// 全局变量对象,
var TNBL = TNBL || {};
var FO = TNBL;

/**
 * @desc 全局公用js
 * @depends jQuery,artTemplte
 */
(function ($, artTemplate) {

    /**
     * 获取网站绝对地址
     */
	TNBL.getUrl = function (id) {
    	return  $('#'+id).val();
    };

	//禁止ajax缓存
    //默认跨域
    $.ajaxSetup({
        cache: false
    });

    /**
     * 封装 art.template,兼容1.0和2.0版本
     */
    TNBL.template = artTemplate.version >= '2.0' ? artTemplate.compile : artTemplate;

    /**
     * 封装 art.template,兼容1.0和2.0版本(支持eacape,即自动转义html字符)
     */
    TNBL.escapeTemplate = function(args){
        if (artTemplate.version >= '2.0') {
            artTemplate.isEscape =true;
            return artTemplate.compile(args);
        } else {
            return artTemplate(args);
        }
    };

    /**
     * 节流函数
     * @param method 方法
     * @param scope 当前函数执行作用域
     */
    TNBL.throttle_id;
    TNBL.throttle_settimeout;
    TNBL.throttle = function (method, scope) {

        if (typeof TNBL.throttle_id == 'object') {
            //canceled 掉请求
            TNBL.throttle_id.abort();
        }

        clearTimeout(TNBL.throttle_settimeout);
        TNBL.throttle_settimeout= setTimeout(function(){
            method.call(scope);
        }, 300);
    };

    /**
     * 封装 jQuery.ajax
     */
    TNBL.ajax = function (opts) {
        // 打开mask
        opts.useMask && opts.useMask === true && TNBL.mask(null, 'open');

        var options;

        var defaults = {
            dataType: 'text',
            type: 'GET',

            //服务器返回404,500等
            sysError: function (message) {
                opts.useMask && opts.useMask === true && TNBL.mask(null, 'close');
                TNBL.showErr(message || ('请求无返回:' + this.url));
            },
            // 设置request_type,用于处理请求路径错误区分ajax请求和非ajax请求
            beforeSend: function (XMLHttpRequest) {
                XMLHttpRequest.setRequestHeader("request_type", "ajax");
            },
            //正常返回 $this->json->error() ; 子分类请在data中说明
            retError: function (message, data) {
                // 关闭mask
                opts.useMask && opts.useMask === true && TNBL.mask(null, 'close');
            },
            //正常返回 $this->json->ok() ; 子分类请在data中说明
            retOk: function (message, data) {
                // 关闭mask
                opts.useMask && opts.useMask === true && TNBL.mask(null, 'close');
            }
        };

        //拷贝
        options = $.extend({}, defaults, opts);
        if (!options.data && options.params) {
            options.data = options.params;
        }

        // 合并beforesend
        options.beforeSend = function (XMLHttpRequest) {
            defaults.beforeSend(XMLHttpRequest);
            $.isFunction(opts.beforeSend) && opts.beforeSend(XMLHttpRequest);
        };

        //构造error
        if (!options.error) {
            options.error = function (ret) {
                opts.useMask && opts.useMask === true && TNBL.mask(null, 'close');
                //abort操作，用户主动终止
                if (ret.status === 0) {
                    return;
                } else if (ret.responseText) {
                    return options.sysError(ret.responseText);
                }
            };
        }

        //构造success
        if (!options.success) {
            // 返回为 json
            if (options.dataType == 'json') {
                options.success = function (ret) {
                    // 关闭mask
                    opts.useMask && opts.useMask === true && TNBL.mask(null, 'close');

                    var json = ret;// 类型，提示等
                    // Fix Null
                    if (!json) {
                        return options.sysError('return is null');
                    }

                    switch (parseInt(json.code)) {
                        case 0:
                            return options.retOk(json.message, json.data);
                        // case -99:
                            // (window.parent || window).location.href = json.data.r_url;
                            // return;
                        default:
                            return options.retError(json.message, json.data);
                    }
                }
                // 返回为text
            } else if (options.dataType == 'text') {

                options.success = function (ret) {

                    // 关闭mask
                    opts.useMask && opts.useMask === true && TNBL.mask(null, 'close');

                    var json;// 类型，提示等

                    try {
                        json = $.parseJSON(ret);
                    } catch (e) {
                        return options.sysError(ret);
                    }
                    // Fix Null
                    if (!json) {
                        return options.sysError(ret);
                    }

                    switch (parseInt(json.code)) {
                        case 0:
                            return options.retOk(json.message, json.data);
                        // case -99:
                        //     (window.parent || window).location.href = json.data.r_url;
                        //     return;
                        default:
                            return options.retError(json.message, json.data);
                    }
                }
                // 其他类型直接返回
            } else {
                options.success = function (ret) {
                    // 关闭mask
                    opts.useMask && opts.useMask === true && TNBL.mask(null, 'close');
                    // 不校验内容，直接返回
                    return options.retOk(ret);
                }
            }
        }

        return TNBL.throttle_id = $.ajax(options);
    };

    /**
     * 获取url中的get参数
     * @params String name get参数的名字
     */
    TNBL.getVar = function (name) {
        var getVarList = {}, getVarStrList, urlParts, url = window.document.location.href.toString();
        urlParts = url.split("?");
        if (typeof(urlParts[1]) != "string") {
            return undefined;
        }

        getVarStrList = urlParts[1].split("&");

        $.each(getVarStrList, function (index, getVarStr) {
            var kvPair = getVarStr.split("=");
            getVarList[kvPair[0]] = decodeURI(kvPair[1]);
        });

        return getVarList[name];
    }


    /**
     * 获取输入框的值(排除ie下placeholder的内容)
     */
    TNBL.getInputValue = function ($obj) {
        return ( $obj.attr('placeholder') && $obj.val() == $obj.attr('placeholder') ) ? '' : $obj.val();
    }


    /**
     * 获取网站绝对地址(已经包含了'/web/')
     */
    TNBL.siteUrl = function (url, query) {
        var baseURL = $('base').attr('href') + $.trim(url, '/'),
            queryString = '';
        if (typeof query !== 'undefined') {
            queryString = '?' + $.param(query);
        }
        return baseURL + queryString;
    }

    /**
     * 字符串截取
     */
    TNBL.getEllipsisStr = function (preString, length) {
        return preString.length > length ? preString.substr(0, length) + '...' : preString;
    }

    /**
     * 弹出dialog框，默认2个按钮，支持复杂内容
     * @param    content            dialog内容，支持复杂的html
     * @param    title            dialog的标题
     * @param    id                dialog的ID,也作为dom对象的id
     * @param    okValue        dialog"确认"按钮的显示名称
     * @param    okCallback        dialog"取消"按钮点击时的回调函数
     * @param    cancelValue    dialog"确认"按钮的显示名称
     * @param    cancelCallback    dialog"取消"按钮点击时的回调函数
     *
     */
    TNBL.noConfirmDialog = function (content, title, id, width, height) {
        $.artDialog({
            id: id || 'systemDialog',
            title: title || '系统提示',
            width: width || 800,   // 宽度默认800px
            height: height || null,// 高度默认自适应
            content: content,
            lock: true
        }).showModal();
    };

    TNBL.dialog = function (content, title, id, width, height, okValue, okCallback, cancelValue, cancelCallback) {
        $.artDialog({
            id: id || 'systemDialog',
            title: title || '系统提示',
            okValue: okValue || '确认',
            cancelValue: cancelValue || '取消',
            width: width || 800,   // 宽度默认800px
            height: height || null,// 高度默认自适应
            content: content,
            ok: function () {
                if ($.isFunction(okCallback) && okCallback() === false) {
                    return false;
                }
                ;
                this.close().remove();
            },
            cancel: function () {
                if ($.isFunction(cancelCallback) && cancelCallback() === false) {
                    return false;
                }
                ;
                this.close().remove();
            },
            lock: true
        }).showModal();
    }



    /**
     * 弹出信息
     */
    TNBL.showMsg = function (content, title, id, okCallback, width,okValue) {
        $.artDialog({
            id: id || 'systemMessage',
            title: title || '系统提示',
            okValue: okValue||'确定',
            width: width || 320,
            content: '<div class="f14">' + content + '</div>',
            ok: function () {
                $.isFunction(okCallback) && okCallback();
                return true;
            },
            lock: true,
            closeOnEscape:false,
			open:function(event,ui){$(".ui-dialog-titlebar-close").hide();}
        }).showModal();
    }


    /**
     * 弹出信息(操作成功)
     */
    TNBL.showOk = function (content, okCallback) {
        var id = 'systemOk';
        var content = '<div class="info-text fl f14"><span class="inblo wraptext break" style="line-height:24px;">' + ( content || '操作成功！' ) + '</span></div>';
        var title = "系统提示"; //title || '系统消息';
        TNBL.showMsg(content, title, id, okCallback,"", "确定");
    }


    /**
     * 弹出信息(操作失败)
     */
    TNBL.showErr = function (content,  okCallback) {
        var id = 'systemError';
        var content = '<div class="info-text fl f14"><span class="inblo wraptext break" style="line-height:24px;">' + ( content || '操作失败！' ) + '</span></div>';
        var title = "系统提示"; //title || '系统消息';
        TNBL.showMsg(content, title, id, okCallback,320, '确定');
    }


    /**
     * 弹出确认
     */
    TNBL.confirm = function (content, yes, no, title) {
	    var title = title;
        var content = '<div class="info-text fl f14"><span class="inblo wraptext break" style="line-height:20px;">' + ( content || '操作失败！' ) + '</span></div>';
        return $.artDialog({
            id: 'Confirm',
            title: title || "系统提示",
            width: 320,
            lock: true,
            content: content,
            okValue: '确定',
            cancelValue: '取消',
            ok: function (here) {
                return yes && yes.call(this, here);
            },
            cancel: function (here) {
                return no && no.call(this, here);
            }
        }).showModal();
    }

    /**
     * 遮罩
     * @params  {String}  遮罩提示
     * @params  {String}    'open'  打开mask
     *                        'close' 关闭mask
     */
    TNBL.mask = function (tips, type) {
        tips = tips || '请求处理中，请耐心等待...';

        TNBL._maskDialog = TNBL._maskDialog || $.artDialog({
            id: 'waiting',
            title: '系统提示',
            width: 320,
            lock: true,
            content: '<div style="height:72px;"><div class="loadingSpan inblo vm"></div><div class="inblo vm f14" style="width:220px;height:70px;padding-left:10px;overflow:hidden;">' + tips + '</div></div>',
            cancelValue: '关闭',
            cancel: false
        });

        // 设置tips内容
        TNBL._maskDialog.content('<div style="height:72px;"><div class="loadingSpan inblo vm"></div><div class="inblo vm f14" style="width:220px;height:70px;padding-left:10px;overflow:hidden;">' + tips + '</div></div>');
        if (type) {
            if (type === 'open') {
            	TNBL._maskDialog.__zIndex();//解决一个窗口有多个dialog时，最上层遮罩失效的问题
            	TNBL._maskDialog.showModal();
                //return newMaskDialog();
            } else if (type === 'close') {
            	TNBL._maskDialog.close();
            }
        } else if (TNBL._maskDialog && TNBL._maskDialog.open) {
        	TNBL._maskDialog.close();
        } else {
        	TNBL._maskDialog.__zIndex();
        	TNBL._maskDialog.showModal();
        }
    }

    /**
     * 短暂提示
     * @param    {String}    提示内容
     * @param    {Number}    显示时间 (默认1.5秒)
     */
    TNBL.tips = function (content, time, callback) {
        var dialog = $.artDialog({
            id: 'Tips',
            title: false,
            cancel: false,
            content: '<div style="padding: 0 1em;">' + content + '</div>',
            fixed: true,
            lock: false
        });
        dialog.show();
        //重新给dialog body样式赋值
        $("td[i=body]").addClass("tnbl-tips-body");
        setTimeout(function () {
            dialog.remove();
            $.isFunction(callback) && callback();
        }, time || 1500);
    };



    /**
     * 文件下载
     * @param {String}    数据url
     * @param {Array}    请求的参数
     * @param {String}  GET or POST
     */
    TNBL.exportFile = function (url, param, type) {
        $.fileDownload(url, {
            httpMethod: type || "GET",
            data: param || {},
            finishCallback: function () {
                /**/
            },
            successCallback: function () {
                TNBL.tips('下载成功！');
            },
            failCallback: function () {
                TNBL.tips('下载失败');
            }
        });
    }

    /**
     * 文件下载, 存在遮罩
     * @param {String}    数据url
     * @param {Array}    请求的参数
     * @param {String}  GET or POST
     */
    TNBL.exportFileMask = function (url, param, type) {
        TNBL.mask(null,'open');
        $.fileDownload(url, {
            httpMethod: type || "GET",
            data: param || {},
            finishCallback: function () {
                TNBL.mask(null,'close');
                TNBL.tips('下载成功！');
                delCookie('fileDownload');
            },
            successCallback: function () {
                TNBL.mask(null,'close');
                TNBL.tips('下载成功！');
                delCookie('fileDownload');
            },
            failCallback: function () {
                TNBL.mask(null,'close');
                TNBL.tips('导出数据太多，请增加条件');
                delCookie('fileDownload');
            }
        });

    };
    //取cookies函数
    function getCookie(name){
        var arr = document.cookie.match(new RegExp("(^| )"+name+"=([^;]*)(;|$)"));

        if(arr != null) return decodeURIComponent(arr[2]);
        return null;
    }
    function delCookie(name) {
        var exp = new Date();
        exp.setTime(exp.getTime() - 1);
        var cval=getCookie(name);
        if(cval!=null) document.cookie= name + "="+cval+";expires="+exp.toGMTString();
    }

    /**
     *    公共校验库
     */
    TNBL.validate = {
        isEmpty: function (preValue) {
            return preValue === '' || preValue === 0 || preValue === undefined || preValue === [] || preValue === null || ( preValue instanceof Array && preValue.length == 0);
        },
        isEmptyObject: function(preValue){
        	if (preValue === undefined){
        		return true;
        	}
        	for (var key in preValue) {
        		return false;
        	}
        	return true;
        }
    }


    /**
     * 把\r\n转换为<br>
     */
    TNBL.nl2br = function (str, isXhtml) {
        var breakTag = (isXhtml || typeof isXhtml === 'undefined') ? '<br />' : '<br>';
        return (str + '').replace(/([^>\r\n]?)(\r\n|\n\r|\r|\n)/g, '$1' + breakTag + '$2');
    }

    /**
     * 刷新页面
     * @param {bool} isForceReloadRes 是否重新加载资源
     */
    TNBL.refresh = function (isForceReloadRes) {
        if (isForceReloadRes == undefined) {
            isForceReloadRes = false;
        }
        location.reload(isForceReloadRes);
    };

    TNBL.Encrypt = function (input) {
        var keyStr = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=";

        var output = "";
        var chr1, chr2, chr3, enc1, enc2, enc3, enc4;
        var i = 0;
        input = _utf8_encode(input);
        while (i < input.length) {
            chr1 = input.charCodeAt(i++);
            chr2 = input.charCodeAt(i++);
            chr3 = input.charCodeAt(i++);
            enc1 = chr1 >> 2;
            enc2 = ((chr1 & 3) << 4) | (chr2 >> 4);
            enc3 = ((chr2 & 15) << 2) | (chr3 >> 6);
            enc4 = chr3 & 63;
            if (isNaN(chr2)) {
                enc3 = enc4 = 64;
            } else if (isNaN(chr3)) {
                enc4 = 64;
            }
            output = output +
            keyStr.charAt(enc1) + keyStr.charAt(enc2) +
            keyStr.charAt(enc3) + keyStr.charAt(enc4);
        }
        return output;
    };

    //获取当前模组{name: 模组名 , id:模组id}
    TNBL.getCurrModule = function(){
        // if ($("#sel-module-chosen").length == 0){
        //     return false;
        // }
        // return {name:$("#sel-module-chosen option:selected").text(), id: $("#sel-module-chosen").val()};
        var mData = Object.assign({},__GetFrameDataByKey("curMozuData"))
        if(mData){
            return {
                id:mData["id"]  || getCookie('tnebula_cu_moduleid'),
                name:mData["name"]  || getCookie('tnebula_cu_modulename'),
                alias:mData["alias"] || getCookie('tnebula_cu_modulealias'),
                region:mData.region||{},
                campus:mData.campus||{},
                source: mData['source'] || 0, // 1代表老园区，2代表大园区
            }
        }
        return {name:'',id:'',realname:'',alias:'',campus:{},region:{} , source: 0}
    };
    // 获取模组框数据
    TNBL.getMozuData = function(key){
      let mozuData =Object.assign([], __GetFrameDataByKey('mozuData'))
      let keyArr = []
      if(key){
        mozuData.forEach(region=>{
          region.children.forEach(campus=>{
            campus.children.forEach(mozu=>{
              if(mozu.id !== '全部' && mozu[key] ){
                keyArr.push(mozu[key])
              }
            })
          })
        })
        return keyArr
      }
      return mozuData
    }
    TNBL.getCurModuleId = function(){
        return __GetFrameDataByKey('curMozuId')
    }

})(jQuery, template);

//-----------------------------IE8 supprot--------------------------
if (!Array.prototype.indexOf) {
    Array.prototype.indexOf = function (elt /*, from*/) {
        var len = this.length >>> 0;
        var from = Number(arguments[1]) || 0;
        from = (from < 0)
            ? Math.ceil(from)
            : Math.floor(from);
        if (from < 0)
            from += len;
        for (; from < len; from++) {
            if (from in this &&
                this[from] === elt)
                return from;
        }
        return -1;
    };
}
if (!String.prototype.trim) {
    String.prototype.trim = function (c) {
        if (c == undefined) {
            c = '\s';
        }
        var rex = '/(^' + c + '*|' + c + '*$)/g'
        return this.replace(rex, '');
    }
}

/*
 * 对Date的扩展，将 Date 转化为指定格式的String
 * 月(M)、日(d)、小时(h)、分(m)、秒(s)、季度(q) 可以用 1-2 个占位符，
 * 年(y)可以用 1-4 个占位符，毫秒(S)只能用 1 个占位符(是 1-3 位的数字)
 * 例子：
 * (new Date()).format("yyyy-MM-dd hh:mm:ss.S") ==> 2006-07-02 08:09:04.423
 * (new Date()).format("yyyy-M-d h:m:s.S")      ==> 2006-7-2 8:9:4.18
 */
Date.prototype.format = function (fmt) {
    var o = {
        "M+": this.getMonth() + 1,                 //月份
        "d+": this.getDate(),                    //日
        "h+": this.getHours(),                   //小时
        "m+": this.getMinutes(),                 //分
        "s+": this.getSeconds(),                 //秒
        "q+": Math.floor((this.getMonth() + 3) / 3), //季度
        "S": this.getMilliseconds()             //毫秒
    };
    if (/(y+)/.test(fmt))
        fmt = fmt.replace(RegExp.$1, (this.getFullYear() + "").substr(4 - RegExp.$1.length));
    for (var k in o)
        if (new RegExp("(" + k + ")").test(fmt))
            fmt = fmt.replace(RegExp.$1, (RegExp.$1.length == 1) ? (o[k]) : (("00" + o[k]).substr(("" + o[k]).length)));
    return fmt;
};
